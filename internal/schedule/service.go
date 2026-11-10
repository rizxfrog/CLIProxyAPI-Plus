package schedule

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// historyLimit is how many run entries are kept per task, newest first.
	historyLimit = 10
	// maxConcurrentRuns bounds in-flight requests across all tasks.
	maxConcurrentRuns = 4
	// wakeCap keeps the loop responsive when the clock is adjusted, so a long
	// delay is never computed once and then trusted for hours.
	wakeCap = 30 * time.Second
)

// Sentinel errors the management handlers map onto HTTP status codes.
var (
	ErrNotConfigured = errors.New("schedule: automation.schedule.dir is not configured")
	ErrGroupNotFound = errors.New("schedule: group not found")
	ErrGroupExists   = errors.New("schedule: group already exists")
	ErrTaskNotFound  = errors.New("schedule: task not found")
	ErrTaskRunning   = errors.New("schedule: task is already running")
)

// SkippedTask reports one entry of a batch add that was not accepted, and why.
// A partial accept is deliberate: a page pasting several curl commands must be
// able to say which one was unusable instead of failing the whole paste.
type SkippedTask struct {
	Index  int    `json:"index"`
	Reason string `json:"reason"`
}

// taskState is the mutable scheduling state of one task, keyed by "group/task".
type taskState struct {
	groupID string
	taskID  string
	// fingerprint captures everything that decides when the task next fires, so
	// an edited trigger recomputes its next run while the task keeps its history.
	fingerprint string
	scheduled   bool
	source      string
	triggerType string
	interval    time.Duration
	cron        *Cron
	nextRun     time.Time
	running     bool

	lastRun        *time.Time
	lastStatus     string
	lastStatusCode int
	lastDurationMS int64
	lastError      string
	runCount       int64
	failureCount   int64
	runs           []RunEntry
}

func (st *taskState) isFixedDelay() bool {
	return st.triggerType == TriggerFixedDelay
}

// taskRunner performs one task's request. It exists as a seam so scheduling tests
// can observe runs without real network I/O or wall-clock waits.
type taskRunner interface {
	run(ctx context.Context, task Task) RunEntry
}

// Service owns the schedule store and the timers that drive it.
type Service struct {
	store  *Store
	runner taskRunner

	// mutateMu serializes read-modify-write cycles so two concurrent edits cannot
	// lose one another's changes.
	mutateMu sync.Mutex

	mu       sync.Mutex
	groups   []Group
	problems []LoadProblem
	state    map[string]*taskState
	running  bool
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	wake     chan struct{}
	sem      chan struct{}
	nowFn    func() time.Time
	// onRunComplete is invoked after a run is folded into the state. Tests use it
	// to synchronize on completion instead of sleeping.
	onRunComplete func(groupID, taskID string, entry RunEntry)
}

// NewService builds a service over a store. The directory may be applied later
// with ApplyDir, which is what a configuration reload does.
func NewService(store *Store) *Service {
	if store == nil {
		store = NewStore("")
	}
	return &Service{
		store:  store,
		runner: newExecutor(),
		state:  make(map[string]*taskState),
		wake:   make(chan struct{}, 1),
		sem:    make(chan struct{}, maxConcurrentRuns),
		nowFn:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) now() time.Time {
	if s.nowFn != nil {
		return s.nowFn()
	}
	return time.Now().UTC()
}

// setClock replaces the time source. Tests use it to advance time without
// sleeping, which the repository forbids in scheduling tests.
func (s *Service) setClock(nowFn func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nowFn = nowFn
}

// Start loads the directory and begins the scheduling loop. An empty directory
// leaves the service idle but still queryable.
func (s *Service) Start(ctx context.Context, dir string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	s.mu.Lock()
	previous := s.cancel
	s.cancel = nil
	s.running = true
	s.store = NewStore(dir)
	now := s.now()
	s.mu.Unlock()
	if previous != nil {
		previous()
		s.wg.Wait()
	}

	s.reload()

	loopCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.cancel = cancel
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.loop(loopCtx)
	}()
	s.applyScheduleAnchors(now)
	log.Infof("schedule: scheduler started (dir=%q)", strings.TrimSpace(dir))
	return nil
}

// applyScheduleAnchors arms interval triggers from the activation time, so a
// restart never fires a burst of catch-up runs. Cron triggers keep their
// wall-clock alignment and are left alone.
func (s *Service) applyScheduleAnchors(anchor time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, st := range s.state {
		if !st.scheduled {
			continue
		}
		if st.cron != nil {
			continue
		}
		if st.interval > 0 && !st.running {
			st.nextRun = anchor.Add(st.interval)
		}
	}
}

// ApplyDir points the service at a new directory and reloads. A no-op when the
// directory is unchanged.
func (s *Service) ApplyDir(dir string) {
	trimmed := strings.TrimSpace(dir)
	s.mu.Lock()
	changed := s.store.Dir() != trimmed
	if changed {
		s.store = NewStore(trimmed)
	}
	s.mu.Unlock()
	if !changed {
		return
	}
	log.Infof("schedule: directory changed to %q", trimmed)
	s.reload()
}

// Stop ends the scheduling loop. In-flight requests are allowed to finish.
func (s *Service) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.running = false
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	s.wg.Wait()
}

// Notify asks the loop to reload and recompute its timers.
func (s *Service) Notify() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// Dir is the configured directory.
func (s *Service) Dir() string { return s.store.Dir() }

// Configured reports whether a directory is set.
func (s *Service) Configured() bool { return s.store.Configured() }

// loop wakes either when the earliest task is due or when a mutation signals.
func (s *Service) loop(ctx context.Context) {
	for {
		delay, pending := s.nextWakeDelay()
		if !pending {
			select {
			case <-ctx.Done():
				return
			case <-s.wake:
				s.reload()
				continue
			}
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-s.wake:
			timer.Stop()
			s.reload()
		case <-timer.C:
			s.fireDue(ctx)
		}
	}
}

// nextWakeDelay reports how long until the earliest scheduled run. pending is
// false only when no task is armed, which lets the loop block instead of
// spinning.
//
// A task that is already running is still armed: its next occurrence is either
// in the future (an interval schedule advances before it launches) or will be
// re-armed when the run finishes, and the due pass decides whether the
// occurrence runs or is recorded as a skip. Excluding running tasks here would
// leave the loop with nothing to wait for and stall it until the next mutation.
func (s *Service) nextWakeDelay() (time.Duration, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	var earliest time.Time
	for _, st := range s.state {
		if !st.scheduled || st.nextRun.IsZero() {
			continue
		}
		if earliest.IsZero() || st.nextRun.Before(earliest) {
			earliest = st.nextRun
		}
	}
	if earliest.IsZero() {
		return 0, false
	}
	delay := earliest.Sub(now)
	if delay < 0 {
		delay = 0
	}
	if delay > wakeCap {
		delay = wakeCap
	}
	return delay, true
}

// reload re-reads every group file and reconciles the runtime state.
func (s *Service) reload() {
	groups, problems, errList := s.store.List()
	s.mu.Lock()
	defer s.mu.Unlock()
	if errList != nil {
		// Keep the previous view rather than pretending the schedules vanished.
		log.Errorf("schedule: reload failed, keeping the previous view: %v", errList)
		return
	}
	s.groups = groups
	s.problems = problems
	s.reconcileLocked()
	for _, problem := range problems {
		log.Warnf("schedule: ignoring %s: %s", problem.File, problem.Message)
	}
}

// reconcileLocked rebuilds the task state from the loaded groups, preserving the
// history and counters of tasks that still exist.
func (s *Service) reconcileLocked() {
	now := s.now()
	live := make(map[string]struct{}, len(s.state))
	for _, group := range s.groups {
		for _, task := range group.Tasks {
			key := stateKey(group.ID, task.ID)
			live[key] = struct{}{}
			trigger, source := EffectiveTrigger(&group, &task)
			enabled := group.Enabled && task.Enabled && trigger != nil
			if trigger != nil {
				if _, errTrigger := trigger.Interval(); errTrigger != nil && trigger.Type != TriggerCron {
					enabled = false
				}
			}
			fingerprint := scheduleFingerprint(trigger, source, enabled)

			st, exists := s.state[key]
			if !exists {
				st = &taskState{groupID: group.ID, taskID: task.ID, lastStatus: StatusNever}
				s.state[key] = st
			}
			if exists && st.fingerprint == fingerprint {
				// Nothing about when it fires changed; keep nextRun as it stands.
				continue
			}
			st.fingerprint = fingerprint
			st.source = source
			st.scheduled = enabled
			st.triggerType = ""
			st.interval = 0
			st.cron = nil
			st.nextRun = time.Time{}
			if !enabled {
				continue
			}
			st.triggerType = trigger.Type
			switch trigger.Type {
			case TriggerFixedRate, TriggerFixedDelay:
				interval, errInterval := trigger.Interval()
				if errInterval != nil {
					st.scheduled = false
					log.Errorf("schedule: %s: %v", key, errInterval)
					continue
				}
				st.interval = interval
				// Neither interval type fires before one full interval has
				// elapsed since the schedule was activated.
				st.nextRun = now.Add(interval)
			case TriggerCron:
				parsed, errCron := ParseCron(trigger.Value)
				if errCron != nil {
					st.scheduled = false
					log.Errorf("schedule: %s: %v", key, errCron)
					continue
				}
				st.cron = parsed
				st.nextRun = parsed.Next(now)
			}
		}
	}
	for key := range s.state {
		if _, ok := live[key]; !ok {
			delete(s.state, key)
		}
	}
}

func stateKey(groupID, taskID string) string { return groupID + "/" + taskID }

func scheduleFingerprint(trigger *Trigger, source string, enabled bool) string {
	if trigger == nil {
		return fmt.Sprintf("%s|none|%t", source, enabled)
	}
	return fmt.Sprintf("%s|%s|%s|%t", source, trigger.Type, trigger.Value, enabled)
}

// fireDue advances every due task and launches the ones that are idle.
//
// The next run is advanced before launching, so a task that is still running
// records a skip instead of being found due again on the next tick.
func (s *Service) fireDue(ctx context.Context) {
	s.mu.Lock()
	now := s.now()
	type launch struct {
		groupID string
		task    Task
	}
	launches := make([]launch, 0, 4)
	for _, key := range sortedStateKeys(s.state) {
		st := s.state[key]
		if !st.scheduled || st.nextRun.IsZero() || st.nextRun.After(now) {
			continue
		}
		// The next run is advanced first, so an occurrence that overlaps a run in
		// flight is skipped rather than queued behind it.
		s.advanceLocked(st, now)
		task, found := s.taskLocked(st.groupID, st.taskID)
		if !found {
			st.scheduled = false
			continue
		}
		if st.running {
			s.recordLocked(st, RunEntry{
				StartedAt:  now,
				FinishedAt: now,
				Status:     StatusSkipped,
				Error:      "previous run still in flight",
			})
			continue
		}
		st.running = true
		launches = append(launches, launch{groupID: st.groupID, task: task})
	}
	s.mu.Unlock()

	for _, item := range launches {
		go s.execute(ctx, item.groupID, item.task)
	}
}

// advanceLocked moves a due task to its next fire time.
func (s *Service) advanceLocked(st *taskState, now time.Time) {
	switch {
	case st.cron != nil:
		st.nextRun = st.cron.Next(now)
	case st.isFixedDelay():
		// A delay schedule is re-armed when the run finishes, not now, so the
		// interval is measured from completion.
		st.nextRun = time.Time{}
	case st.interval > 0:
		// Fixed rate keeps its cadence regardless of run duration and skips
		// missed occurrences instead of replaying them.
		next := st.nextRun
		if !next.After(now) {
			steps := now.Sub(next)/st.interval + 1
			next = next.Add(steps * st.interval)
		}
		st.nextRun = next
	default:
		st.nextRun = time.Time{}
	}
}

// execute performs one run and folds the outcome into the state.
func (s *Service) execute(ctx context.Context, groupID string, task Task) {
	s.sem <- struct{}{}
	defer func() { <-s.sem }()
	entry := s.runner.run(ctx, task)
	s.completeRun(groupID, task.ID, entry)
}

// completeRun records a finished run and re-arms a delay trigger.
func (s *Service) completeRun(groupID, taskID string, entry RunEntry) {
	s.mu.Lock()
	st, ok := s.state[stateKey(groupID, taskID)]
	if ok {
		st.running = false
		if st.isFixedDelay() && st.interval > 0 {
			st.nextRun = entry.FinishedAt.Add(st.interval)
		}
		s.recordLocked(st, entry)
	}
	hook := s.onRunComplete
	s.mu.Unlock()
	if hook != nil {
		hook(groupID, taskID, entry)
	}
	// A finished run arms a delay schedule and clears the in-flight state, so the
	// loop must recompute what to wait for.
	s.Notify()
}

func (s *Service) recordLocked(st *taskState, entry RunEntry) {
	finished := entry.FinishedAt
	st.lastRun = &finished
	st.lastStatus = entry.Status
	st.lastStatusCode = entry.StatusCode
	st.lastDurationMS = entry.DurationMS
	st.lastError = entry.Error
	st.runCount++
	if entry.Status == StatusFailed {
		st.failureCount++
	}
	st.runs = append([]RunEntry{entry}, st.runs...)
	if len(st.runs) > historyLimit {
		st.runs = st.runs[:historyLimit]
	}
	if entry.Status == StatusFailed {
		log.Warnf("schedule: %s/%s failed: %s", st.groupID, st.taskID, entry.Error)
	}
}

// taskLocked finds a stored task by id. The caller must hold the lock.
func (s *Service) taskLocked(groupID, taskID string) (Task, bool) {
	for index := range s.groups {
		if s.groups[index].ID != groupID {
			continue
		}
		if task := s.groups[index].FindTask(taskID); task != nil {
			return *task, true
		}
	}
	return Task{}, false
}

// groupLocked finds a stored group by id. The caller must hold the lock.
func (s *Service) groupLocked(groupID string) *Group {
	for index := range s.groups {
		if s.groups[index].ID == groupID {
			return &s.groups[index]
		}
	}
	return nil
}

func sortedStateKeys(state map[string]*taskState) []string {
	keys := make([]string, 0, len(state))
	for key := range state {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// NewTaskID mints a task id that satisfies ValidateID.
func NewTaskID() string {
	var buffer [6]byte
	if _, errRead := rand.Read(buffer[:]); errRead != nil {
		return "t" + strings.ToLower(hex.EncodeToString([]byte(time.Now().UTC().Format("150405.000"))))
	}
	return "t_" + hex.EncodeToString(buffer[:])
}

// ---------------------------------------------------------------------------
// Reads
// ---------------------------------------------------------------------------

// Group returns one stored group.
func (s *Service) Group(id string) (Group, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	group := s.groupLocked(strings.TrimSpace(id))
	if group == nil {
		return Group{}, ErrGroupNotFound
	}
	return *group, nil
}

// Snapshot is the whole scheduler view the management API serves.
func (s *Service) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	snapshot := Snapshot{
		Configured:       s.store.Configured(),
		Dir:              s.store.Dir(),
		SchedulerRunning: s.running,
		ServerTime:       now,
		GroupCount:       len(s.groups),
		Problems:         append([]LoadProblem(nil), s.problems...),
		Groups:           make([]GroupSnapshot, 0, len(s.groups)),
	}
	if snapshot.Problems == nil {
		snapshot.Problems = []LoadProblem{}
	}
	for _, group := range s.groups {
		item := GroupSnapshot{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			Enabled:     group.Enabled,
			Schedule:    group.Schedule,
			Tasks:       make([]TaskSnapshot, 0, len(group.Tasks)),
			Runtime: GroupRuntime{
				Scheduled:  group.Schedule != nil && group.Enabled,
				LastStatus: StatusNever,
				TaskCount:  len(group.Tasks),
			},
		}
		if group.Schedule != nil {
			item.Runtime.TriggerSource = SourceGroup
		} else {
			item.Runtime.TriggerSource = SourceNone
		}
		groupLast := time.Time{}
		groupStatus := StatusNever
		for _, task := range group.Tasks {
			st := s.state[stateKey(group.ID, task.ID)]
			item.Tasks = append(item.Tasks, TaskSnapshot{Task: task, Runtime: runtimeOf(st)})
			snapshot.TaskCount++
			if group.Enabled && task.Enabled {
				snapshot.EnabledTaskCount++
				item.Runtime.EnabledTasks++
			}
			if st == nil {
				continue
			}
			if st.scheduled {
				snapshot.ScheduledTaskCount++
			}
			if st.running {
				item.Runtime.RunningTasks++
			}
			item.Runtime.RunCount += st.runCount
			item.Runtime.FailureCount += st.failureCount
			if st.nextRun != (time.Time{}) && (item.Runtime.NextRun == nil || st.nextRun.Before(*item.Runtime.NextRun)) {
				next := st.nextRun
				item.Runtime.NextRun = &next
			}
			if st.lastRun != nil && st.lastRun.After(groupLast) {
				groupLast = *st.lastRun
				groupStatus = st.lastStatus
			}
		}
		if !groupLast.IsZero() {
			item.Runtime.LastRun = &groupLast
			item.Runtime.LastStatus = groupStatus
		}
		snapshot.Groups = append(snapshot.Groups, item)
	}
	return snapshot
}

// GroupSnapshotOf returns one group's view, so a mutation response and a list
// response can never disagree about what is stored.
func (s *Service) GroupSnapshotOf(id string) (GroupSnapshot, error) {
	trimmed := strings.TrimSpace(id)
	for _, item := range s.Snapshot().Groups {
		if item.ID == trimmed {
			return item, nil
		}
	}
	return GroupSnapshot{}, ErrGroupNotFound
}

func runtimeOf(st *taskState) TaskRuntime {
	if st == nil {
		return TaskRuntime{TriggerSource: SourceNone, LastStatus: StatusNever, Runs: []RunEntry{}}
	}
	runtime := TaskRuntime{
		Scheduled:      st.scheduled,
		TriggerSource:  st.source,
		LastStatus:     st.lastStatus,
		LastStatusCode: st.lastStatusCode,
		LastDurationMS: st.lastDurationMS,
		LastError:      st.lastError,
		RunCount:       st.runCount,
		FailureCount:   st.failureCount,
		Running:        st.running,
		Runs:           append([]RunEntry{}, st.runs...),
	}
	if runtime.TriggerSource == "" {
		runtime.TriggerSource = SourceNone
	}
	if runtime.LastStatus == "" {
		runtime.LastStatus = StatusNever
	}
	if runtime.Runs == nil {
		runtime.Runs = []RunEntry{}
	}
	if st.lastRun != nil {
		last := *st.lastRun
		runtime.LastRun = &last
	}
	if st.nextRun != (time.Time{}) {
		next := st.nextRun
		runtime.NextRun = &next
	}
	return runtime
}

// ---------------------------------------------------------------------------
// Mutations
// ---------------------------------------------------------------------------

// CreateGroup stores a new group, refusing to replace an existing one.
func (s *Service) CreateGroup(group Group) (Group, error) {
	s.mutateMu.Lock()
	defer s.mutateMu.Unlock()
	if !s.Configured() {
		return Group{}, ErrNotConfigured
	}
	normalized, errNormalize := NormalizeGroup(group)
	if errNormalize != nil {
		return Group{}, errNormalize
	}
	if _, found, errGet := s.store.Get(normalized.ID); errGet != nil {
		return Group{}, errGet
	} else if found {
		return Group{}, ErrGroupExists
	}
	saved, errPut := s.store.Put(normalized)
	if errPut != nil {
		return Group{}, errPut
	}
	s.reload()
	s.Notify()
	return saved, nil
}

// PutGroup replaces an existing group. The path id wins over the document id.
func (s *Service) PutGroup(id string, group Group) (Group, error) {
	s.mutateMu.Lock()
	defer s.mutateMu.Unlock()
	if !s.Configured() {
		return Group{}, ErrNotConfigured
	}
	group.ID = strings.TrimSpace(id)
	if _, found, errGet := s.store.Get(group.ID); errGet != nil {
		return Group{}, errGet
	} else if !found {
		return Group{}, ErrGroupNotFound
	}
	saved, errPut := s.store.Put(group)
	if errPut != nil {
		return Group{}, errPut
	}
	s.reload()
	s.Notify()
	return saved, nil
}

// DeleteGroup removes a group file and its runtime state.
func (s *Service) DeleteGroup(id string) error {
	s.mutateMu.Lock()
	defer s.mutateMu.Unlock()
	if !s.Configured() {
		return ErrNotConfigured
	}
	deleted, errDelete := s.store.Delete(strings.TrimSpace(id))
	if errDelete != nil {
		return errDelete
	}
	if !deleted {
		return ErrGroupNotFound
	}
	s.reload()
	s.Notify()
	return nil
}

// AddTasks appends tasks to a group, generating an id for any task that has none.
// Entries that cannot be used are reported instead of failing the whole batch.
func (s *Service) AddTasks(groupID string, tasks []Task) ([]Task, []SkippedTask, Group, error) {
	s.mutateMu.Lock()
	defer s.mutateMu.Unlock()
	if !s.Configured() {
		return nil, nil, Group{}, ErrNotConfigured
	}
	group, errGroup := s.Group(groupID)
	if errGroup != nil {
		return nil, nil, Group{}, errGroup
	}
	existing := make(map[string]struct{}, len(group.Tasks))
	for _, task := range group.Tasks {
		existing[task.ID] = struct{}{}
	}
	added := make([]Task, 0, len(tasks))
	skipped := make([]SkippedTask, 0)
	for index, candidate := range tasks {
		if strings.TrimSpace(candidate.ID) == "" {
			candidate.ID = NewTaskID()
		}
		normalized, errNormalize := NormalizeTask(candidate)
		if errNormalize != nil {
			skipped = append(skipped, SkippedTask{Index: index, Reason: errNormalize.Error()})
			continue
		}
		if _, duplicate := existing[normalized.ID]; duplicate {
			skipped = append(skipped, SkippedTask{Index: index, Reason: fmt.Sprintf("duplicate id %q", normalized.ID)})
			continue
		}
		existing[normalized.ID] = struct{}{}
		group.Tasks = append(group.Tasks, normalized)
		added = append(added, normalized)
	}
	if len(added) == 0 {
		return nil, skipped, group, nil
	}
	saved, errPut := s.store.Put(group)
	if errPut != nil {
		return nil, nil, Group{}, errPut
	}
	s.reload()
	s.Notify()
	return added, skipped, saved, nil
}

// PutTask replaces one task in place, keeping its position in the group.
func (s *Service) PutTask(groupID, taskID string, task Task) (Task, error) {
	s.mutateMu.Lock()
	defer s.mutateMu.Unlock()
	if !s.Configured() {
		return Task{}, ErrNotConfigured
	}
	group, errGroup := s.Group(groupID)
	if errGroup != nil {
		return Task{}, errGroup
	}
	task.ID = strings.TrimSpace(taskID)
	normalized, errNormalize := NormalizeTask(task)
	if errNormalize != nil {
		return Task{}, errNormalize
	}
	replaced := false
	for index := range group.Tasks {
		if group.Tasks[index].ID == normalized.ID {
			group.Tasks[index] = normalized
			replaced = true
			break
		}
	}
	if !replaced {
		return Task{}, ErrTaskNotFound
	}
	if _, errPut := s.store.Put(group); errPut != nil {
		return Task{}, errPut
	}
	s.reload()
	s.Notify()
	return normalized, nil
}

// DeleteTask removes one task from a group.
func (s *Service) DeleteTask(groupID, taskID string) error {
	s.mutateMu.Lock()
	defer s.mutateMu.Unlock()
	if !s.Configured() {
		return ErrNotConfigured
	}
	group, errGroup := s.Group(groupID)
	if errGroup != nil {
		return errGroup
	}
	trimmed := strings.TrimSpace(taskID)
	kept := make([]Task, 0, len(group.Tasks))
	for _, task := range group.Tasks {
		if task.ID != trimmed {
			kept = append(kept, task)
		}
	}
	if len(kept) == len(group.Tasks) {
		return ErrTaskNotFound
	}
	group.Tasks = kept
	if _, errPut := s.store.Put(group); errPut != nil {
		return errPut
	}
	s.reload()
	s.Notify()
	return nil
}

// RunTask performs one task's request now, independently of its trigger, and
// records the outcome in its history.
func (s *Service) RunTask(ctx context.Context, groupID, taskID string) (RunEntry, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	s.mu.Lock()
	task, found := s.taskLocked(groupID, strings.TrimSpace(taskID))
	if !found {
		s.mu.Unlock()
		return RunEntry{}, ErrTaskNotFound
	}
	st, ok := s.state[stateKey(groupID, strings.TrimSpace(taskID))]
	if ok && st.running {
		s.mu.Unlock()
		return RunEntry{}, ErrTaskRunning
	}
	if ok {
		st.running = true
	}
	s.mu.Unlock()

	s.sem <- struct{}{}
	defer func() { <-s.sem }()
	entry := s.runner.run(ctx, task)
	s.completeRun(groupID, task.ID, entry)
	return entry, nil
}
