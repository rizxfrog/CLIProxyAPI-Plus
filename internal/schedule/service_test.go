package schedule

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeClock makes scheduling assertions exact. The repository forbids wall-clock
// sleeps in scheduling tests, so time is advanced explicitly.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(delta time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(delta)
}

// recordingRunner stands in for the HTTP executor so a test can observe exactly
// which tasks ran and control the outcome.
type recordingRunner struct {
	mu      sync.Mutex
	calls   []Task
	outcome func(task Task, startedAt time.Time) RunEntry
}

func (r *recordingRunner) run(_ context.Context, task Task) RunEntry {
	r.mu.Lock()
	r.calls = append(r.calls, task)
	outcome := r.outcome
	r.mu.Unlock()
	started := time.Now().UTC()
	if outcome != nil {
		return outcome(task, started)
	}
	return RunEntry{StartedAt: started, FinishedAt: started, Status: StatusSuccess, StatusCode: 200}
}

func (r *recordingRunner) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.calls)
}

type harness struct {
	t       *testing.T
	clock   *fakeClock
	store   *Store
	runner  *recordingRunner
	service *Service
	done    chan RunEntry
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{
		t:      t,
		clock:  newFakeClock(),
		store:  NewStore(t.TempDir()),
		runner: &recordingRunner{},
		done:   make(chan RunEntry, 32),
	}
	h.service = NewService(h.store)
	h.service.setClock(h.clock.Now)
	h.service.runner = h.runner
	h.service.onRunComplete = func(_, _ string, entry RunEntry) {
		select {
		case h.done <- entry:
		default:
		}
	}
	return h
}

// put stores a group and reloads, mirroring what a management write does.
func (h *harness) put(group Group) Group {
	h.t.Helper()
	saved, errPut := h.store.Put(group)
	if errPut != nil {
		h.t.Fatalf("store.Put: %v", errPut)
	}
	h.service.reload()
	return saved
}

// waitRun blocks until one scheduled run completes.
func (h *harness) waitRun() RunEntry {
	h.t.Helper()
	select {
	case entry := <-h.done:
		return entry
	case <-time.After(5 * time.Second):
		h.t.Fatal("no run completed")
		return RunEntry{}
	}
}

// taskRuntime reads one task's runtime state.
func (h *harness) taskRuntime(groupID, taskID string) TaskRuntime {
	h.t.Helper()
	snapshot, errSnapshot := h.service.GroupSnapshotOf(groupID)
	if errSnapshot != nil {
		h.t.Fatalf("GroupSnapshotOf: %v", errSnapshot)
	}
	for _, task := range snapshot.Tasks {
		if task.ID == taskID {
			return task.Runtime
		}
	}
	h.t.Fatalf("task %s/%s is absent", groupID, taskID)
	return TaskRuntime{}
}

func rateGroup(id string, interval string, taskIDs ...string) Group {
	group := Group{
		ID:      id,
		Name:    id,
		Enabled: true,
		Schedule: &Trigger{
			Type:  TriggerFixedRate,
			Value: interval,
		},
	}
	for _, taskID := range taskIDs {
		group.Tasks = append(group.Tasks, Task{
			ID:      taskID,
			Name:    taskID,
			Enabled: true,
			Request: Request{Method: "GET", URL: "https://example.com/" + taskID},
		})
	}
	return group
}

func TestServiceWithoutDirectoryServesAnUnconfiguredSnapshot(t *testing.T) {
	service := NewService(NewStore(""))
	snapshot := service.Snapshot()
	if snapshot.Configured {
		t.Fatal("Configured = true without a directory")
	}
	if snapshot.GroupCount != 0 || len(snapshot.Groups) != 0 {
		t.Fatalf("unexpected groups: %+v", snapshot.Groups)
	}
	if _, errCreate := service.CreateGroup(sampleGroup("x")); !errors.Is(errCreate, ErrNotConfigured) {
		t.Fatalf("CreateGroup error = %v, want ErrNotConfigured", errCreate)
	}
}

// A group trigger drives every task that has none of its own.
func TestGroupTriggerArmsItsTasks(t *testing.T) {
	h := newHarness(t)
	h.put(rateGroup("g", "1h", "a", "b"))

	for _, taskID := range []string{"a", "b"} {
		runtime := h.taskRuntime("g", taskID)
		if !runtime.Scheduled {
			t.Fatalf("task %s is not scheduled", taskID)
		}
		if runtime.TriggerSource != SourceGroup {
			t.Fatalf("task %s source = %q, want %q", taskID, runtime.TriggerSource, SourceGroup)
		}
		if runtime.NextRun == nil {
			t.Fatalf("task %s has no next run", taskID)
		}
		want := h.clock.Now().Add(time.Hour)
		if !runtime.NextRun.Equal(want) {
			t.Fatalf("task %s next run = %s, want %s", taskID, runtime.NextRun, want)
		}
	}
	// A group runtime reports the earliest task and the enabled count.
	group, errGroup := h.service.GroupSnapshotOf("g")
	if errGroup != nil {
		t.Fatalf("GroupSnapshotOf: %v", errGroup)
	}
	if group.Runtime.EnabledTasks != 2 || group.Runtime.TaskCount != 2 {
		t.Fatalf("group runtime = %+v", group.Runtime)
	}
	if !group.Runtime.Scheduled || group.Runtime.TriggerSource != SourceGroup {
		t.Fatalf("group scheduled = %t, source = %q", group.Runtime.Scheduled, group.Runtime.TriggerSource)
	}
}

func TestTaskTriggerOverridesGroupTrigger(t *testing.T) {
	h := newHarness(t)
	group := rateGroup("g", "1h", "inherit")
	group.Tasks = append(group.Tasks, Task{
		ID:      "own",
		Name:    "own",
		Enabled: true,
		Schedule: &Trigger{
			Type:  TriggerCron,
			Value: "30 4 * * *",
		},
		Request: Request{Method: "GET", URL: "https://example.com/own"},
	})
	h.put(group)

	if source := h.taskRuntime("g", "inherit").TriggerSource; source != SourceGroup {
		t.Fatalf("inherit source = %q", source)
	}
	if source := h.taskRuntime("g", "own").TriggerSource; source != SourceTask {
		t.Fatalf("own source = %q", source)
	}
}

// A task with neither its own trigger nor a group trigger must stay manual-only
// rather than being given a default cadence.
func TestTaskWithoutAnyTriggerIsNeverScheduled(t *testing.T) {
	h := newHarness(t)
	group := rateGroup("g", "1h", "manual")
	group.Schedule = nil
	h.put(group)

	runtime := h.taskRuntime("g", "manual")
	if runtime.Scheduled || runtime.NextRun != nil {
		t.Fatalf("runtime = %+v, want unscheduled", runtime)
	}
	if runtime.TriggerSource != SourceNone {
		t.Fatalf("source = %q, want %q", runtime.TriggerSource, SourceNone)
	}
	entry, errRun := h.service.RunTask(context.Background(), "g", "manual")
	if errRun != nil {
		t.Fatalf("RunTask: %v", errRun)
	}
	if entry.Status != StatusSuccess {
		t.Fatalf("manual run status = %q", entry.Status)
	}
	if h.taskRuntime("g", "manual").RunCount != 1 {
		t.Fatal("manual run was not recorded")
	}
}

func TestDisabledGroupAndTaskAreNotScheduled(t *testing.T) {
	h := newHarness(t)
	group := rateGroup("g", "1h", "a", "b")
	group.Tasks[1].Enabled = false
	group.Enabled = false
	h.put(group)

	for _, taskID := range []string{"a", "b"} {
		if runtime := h.taskRuntime("g", taskID); runtime.Scheduled {
			t.Fatalf("task %s is scheduled inside a disabled group", taskID)
		}
	}

	group.Enabled = true
	h.put(group)
	if runtime := h.taskRuntime("g", "a"); !runtime.Scheduled {
		t.Fatal("task a did not become scheduled once the group was enabled")
	}
	if runtime := h.taskRuntime("g", "b"); runtime.Scheduled {
		t.Fatal("disabled task b became scheduled")
	}
}

// Fixed rate keeps its cadence and must not replay the occurrences missed while
// the process was down.
func TestFixedRateSkipsMissedOccurrencesWithoutReplay(t *testing.T) {
	h := newHarness(t)
	h.put(rateGroup("g", "1h", "a"))

	h.clock.Advance(10 * time.Hour)
	h.service.fireDue(context.Background())
	h.waitRun()

	if count := h.runner.count(); count != 1 {
		t.Fatalf("ran %d times, want exactly 1 after a 10 interval gap", count)
	}
	runtime := h.taskRuntime("g", "a")
	if runtime.RunCount != 1 {
		t.Fatalf("RunCount = %d", runtime.RunCount)
	}
	if runtime.NextRun == nil || !runtime.NextRun.After(h.clock.Now()) {
		t.Fatalf("NextRun = %v, want a future time", runtime.NextRun)
	}
	// The next run stays on the original cadence, not now+interval.
	want := h.clock.Now().Add(time.Hour)
	if !runtime.NextRun.Equal(want) {
		t.Fatalf("NextRun = %s, want %s", runtime.NextRun, want)
	}
}

// Fixed delay is measured from completion, not from the previous start.
func TestFixedDelayRearmsFromCompletion(t *testing.T) {
	h := newHarness(t)
	finished := h.clock.Now().Add(5 * time.Minute)
	h.runner.outcome = func(_ Task, started time.Time) RunEntry {
		return RunEntry{StartedAt: started, FinishedAt: finished, Status: StatusSuccess}
	}
	group := Group{
		ID:       "g",
		Name:     "g",
		Enabled:  true,
		Schedule: &Trigger{Type: TriggerFixedDelay, Value: "30m"},
		Tasks: []Task{{
			ID: "a", Name: "a", Enabled: true,
			Request: Request{Method: "GET", URL: "https://example.com/a"},
		}},
	}
	h.put(group)
	if got := h.taskRuntime("g", "a").NextRun; got == nil || !got.Equal(h.clock.Now().Add(30*time.Minute)) {
		t.Fatalf("first NextRun = %v, want one interval after activation", got)
	}

	h.clock.Advance(30 * time.Minute)
	h.service.fireDue(context.Background())
	h.waitRun()

	got := h.taskRuntime("g", "a").NextRun
	want := finished.Add(30 * time.Minute)
	if got == nil || !got.Equal(want) {
		t.Fatalf("NextRun after completion = %v, want %s", got, want)
	}
}

// A run that is still in flight makes the next occurrence a recorded skip, never
// a silent drop.
func TestOverlappingOccurrenceIsRecordedAsSkipped(t *testing.T) {
	h := newHarness(t)
	h.put(rateGroup("g", "1h", "a"))

	h.service.mu.Lock()
	h.service.state["g/a"].running = true
	h.service.mu.Unlock()

	h.clock.Advance(time.Hour)
	h.service.fireDue(context.Background())

	if count := h.runner.count(); count != 0 {
		t.Fatalf("launched %d runs while one was in flight", count)
	}
	runtime := h.taskRuntime("g", "a")
	if runtime.RunCount != 1 || runtime.LastStatus != StatusSkipped {
		t.Fatalf("runtime = %+v, want one skipped entry", runtime)
	}
	if runtime.Runs[0].Error == "" {
		t.Fatal("skipped entry does not say why")
	}
}

func TestCronFiresAtTheWallClockTime(t *testing.T) {
	h := newHarness(t)
	group := Group{
		ID:       "g",
		Name:     "g",
		Enabled:  true,
		Schedule: &Trigger{Type: TriggerCron, Value: "0 9 * * *"},
		Tasks: []Task{{
			ID: "a", Name: "a", Enabled: true,
			Request: Request{Method: "GET", URL: "https://example.com/a"},
		}},
	}
	h.put(group)

	// The clock starts exactly at 09:00, so the first fire is tomorrow.
	tomorrow := time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)
	if got := h.taskRuntime("g", "a").NextRun; got == nil || !got.Equal(tomorrow) {
		t.Fatalf("NextRun = %v, want %s", got, tomorrow)
	}
	h.clock.Advance(24 * time.Hour)
	h.service.fireDue(context.Background())
	h.waitRun()

	if count := h.runner.count(); count != 1 {
		t.Fatalf("ran %d times, want 1", count)
	}
	next := time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC)
	if got := h.taskRuntime("g", "a").NextRun; got == nil || !got.Equal(next) {
		t.Fatalf("NextRun = %v, want %s", got, next)
	}
}

func TestRunHistoryIsBounded(t *testing.T) {
	h := newHarness(t)
	h.put(rateGroup("g", "1h", "a"))
	for range historyLimit + 2 {
		if _, errRun := h.service.RunTask(context.Background(), "g", "a"); errRun != nil {
			t.Fatalf("RunTask: %v", errRun)
		}
	}
	runtime := h.taskRuntime("g", "a")
	if runtime.RunCount != int64(historyLimit+2) {
		t.Fatalf("RunCount = %d", runtime.RunCount)
	}
	if len(runtime.Runs) != historyLimit {
		t.Fatalf("kept %d entries, want %d", len(runtime.Runs), historyLimit)
	}
}

func TestRunTaskRejectsAConcurrentRun(t *testing.T) {
	h := newHarness(t)
	h.put(rateGroup("g", "1h", "a"))
	h.service.mu.Lock()
	h.service.state["g/a"].running = true
	h.service.mu.Unlock()
	if _, errRun := h.service.RunTask(context.Background(), "g", "a"); !errors.Is(errRun, ErrTaskRunning) {
		t.Fatalf("RunTask error = %v, want ErrTaskRunning", errRun)
	}
}

// A batch add reports per-entry problems instead of rejecting the whole paste.
func TestAddTasksReportsSkippedEntries(t *testing.T) {
	h := newHarness(t)
	h.put(rateGroup("g", "1h", "existing"))

	added, skipped, group, errAdd := h.service.AddTasks("g", []Task{
		{ID: "existing", Name: "duplicate", Enabled: true, Request: Request{URL: "https://example.com/dup"}},
		{ID: "fresh", Name: "fresh", Enabled: true, Request: Request{URL: "https://example.com/fresh"}},
		{ID: "", Name: "generated", Enabled: true, Request: Request{URL: "https://example.com/gen"}},
		{ID: "broken", Name: "broken", Enabled: true, Request: Request{URL: "ftp://example.com"}},
	})
	if errAdd != nil {
		t.Fatalf("AddTasks: %v", errAdd)
	}
	if len(added) != 2 {
		t.Fatalf("added %d tasks, want 2", len(added))
	}
	if len(skipped) != 2 || skipped[0].Index != 0 || skipped[1].Index != 3 {
		t.Fatalf("skipped = %+v, want indexes 0 and 3", skipped)
	}
	for _, problem := range skipped {
		if problem.Reason == "" {
			t.Fatalf("skipped entry %d has no reason", problem.Index)
		}
	}
	if added[1].ID == "" {
		t.Fatal("a task with no id was stored without one")
	}
	if len(group.Tasks) != 3 {
		t.Fatalf("group holds %d tasks, want 3", len(group.Tasks))
	}
	// The generated task must be able to run, so it needs a valid URL.
	for _, task := range group.Tasks {
		if _, errGroup := NormalizeTask(task); errGroup != nil {
			t.Fatalf("stored task %q is invalid: %v", task.ID, errGroup)
		}
	}
}

func TestAddTasksRejectsAnEntirelyUnusableBatch(t *testing.T) {
	h := newHarness(t)
	h.put(rateGroup("g", "1h", "existing"))
	added, skipped, _, errAdd := h.service.AddTasks("g", []Task{
		{ID: "bad", Name: "bad", Enabled: true, Request: Request{URL: "not-a-url"}},
	})
	if errAdd != nil {
		t.Fatalf("AddTasks: %v", errAdd)
	}
	if len(added) != 0 || len(skipped) != 1 {
		t.Fatalf("added = %d, skipped = %+v", len(added), skipped)
	}
	if group, errGroup := h.service.Group("g"); errGroup != nil || len(group.Tasks) != 1 {
		t.Fatalf("the rejected batch changed the group: %v, %v", group.Tasks, errGroup)
	}
}

// Editing a trigger re-arms the task but keeps its history, which is about the
// task rather than about the schedule.
func TestTriggerEditRearmsAndKeepsHistory(t *testing.T) {
	h := newHarness(t)
	h.put(rateGroup("g", "1h", "a"))
	if _, errRun := h.service.RunTask(context.Background(), "g", "a"); errRun != nil {
		t.Fatalf("RunTask: %v", errRun)
	}
	h.clock.Advance(20 * time.Minute)

	group, errGroup := h.service.Group("g")
	if errGroup != nil {
		t.Fatalf("Group: %v", errGroup)
	}
	group.Schedule = &Trigger{Type: TriggerFixedRate, Value: "15m"}
	if _, errPut := h.service.PutGroup("g", group); errPut != nil {
		t.Fatalf("PutGroup: %v", errPut)
	}

	runtime := h.taskRuntime("g", "a")
	if runtime.RunCount != 1 || len(runtime.Runs) != 1 {
		t.Fatalf("history was lost on a trigger edit: %+v", runtime)
	}
	want := h.clock.Now().Add(15 * time.Minute)
	if runtime.NextRun == nil || !runtime.NextRun.Equal(want) {
		t.Fatalf("NextRun = %v, want %s", runtime.NextRun, want)
	}
}

func TestGroupAndTaskCRUD(t *testing.T) {
	h := newHarness(t)

	if _, errCreate := h.service.CreateGroup(rateGroup("g", "1h", "a")); errCreate != nil {
		t.Fatalf("CreateGroup: %v", errCreate)
	}
	if _, errAgain := h.service.CreateGroup(rateGroup("g", "1h", "a")); !errors.Is(errAgain, ErrGroupExists) {
		t.Fatalf("second CreateGroup error = %v, want ErrGroupExists", errAgain)
	}
	if _, errPut := h.service.PutGroup("missing", rateGroup("missing", "1h")); !errors.Is(errPut, ErrGroupNotFound) {
		t.Fatalf("PutGroup on a missing group error = %v, want ErrGroupNotFound", errPut)
	}

	if _, errPut := h.service.PutTask("g", "a", Task{
		ID: "a", Name: "renamed", Enabled: true,
		Request: Request{Method: "POST", URL: "https://example.com/renamed"},
	}); errPut != nil {
		t.Fatalf("PutTask: %v", errPut)
	}
	group, errGroup := h.service.Group("g")
	if errGroup != nil {
		t.Fatalf("Group: %v", errGroup)
	}
	if group.Tasks[0].Name != "renamed" || group.Tasks[0].Request.Method != "POST" {
		t.Fatalf("PutTask did not replace the task: %+v", group.Tasks[0])
	}
	if _, errPut := h.service.PutTask("g", "absent", Task{
		ID: "absent", Name: "absent", Enabled: true, Request: Request{URL: "https://example.com/x"},
	}); !errors.Is(errPut, ErrTaskNotFound) {
		t.Fatalf("PutTask on a missing task error = %v, want ErrTaskNotFound", errPut)
	}

	if errDelete := h.service.DeleteTask("g", "a"); errDelete != nil {
		t.Fatalf("DeleteTask: %v", errDelete)
	}
	if errDelete := h.service.DeleteTask("g", "a"); !errors.Is(errDelete, ErrTaskNotFound) {
		t.Fatalf("second DeleteTask error = %v, want ErrTaskNotFound", errDelete)
	}
	if errDelete := h.service.DeleteGroup("g"); errDelete != nil {
		t.Fatalf("DeleteGroup: %v", errDelete)
	}
	if errDelete := h.service.DeleteGroup("g"); !errors.Is(errDelete, ErrGroupNotFound) {
		t.Fatalf("second DeleteGroup error = %v, want ErrGroupNotFound", errDelete)
	}
	// Deleting a group must drop its runtime state too.
	h.service.mu.Lock()
	remaining := len(h.service.state)
	h.service.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("runtime state survived the delete: %d entries", remaining)
	}
}

func TestApplyDirMovesTheStore(t *testing.T) {
	h := newHarness(t)
	h.put(rateGroup("g", "1h", "a"))
	if snapshot := h.service.Snapshot(); snapshot.GroupCount != 1 {
		t.Fatalf("GroupCount = %d", snapshot.GroupCount)
	}
	other := t.TempDir()
	h.service.ApplyDir(other)
	snapshot := h.service.Snapshot()
	if snapshot.GroupCount != 0 {
		t.Fatalf("GroupCount after moving = %d, want 0", snapshot.GroupCount)
	}
	if snapshot.Dir != other {
		t.Fatalf("Dir = %q, want %q", snapshot.Dir, other)
	}
	h.service.mu.Lock()
	remaining := len(h.service.state)
	h.service.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("runtime state survived the move: %d entries", remaining)
	}
}

// A scheduled run must actually be launched by the loop's due pass, not merely
// armed.
func TestDuePassLaunchesEveryDueTaskOnce(t *testing.T) {
	h := newHarness(t)
	h.put(rateGroup("g", "1h", "a", "b"))
	h.clock.Advance(time.Hour)

	h.service.fireDue(context.Background())
	h.waitRun()
	h.waitRun()

	if count := h.runner.count(); count != 2 {
		t.Fatalf("ran %d tasks, want 2", count)
	}
	// A second pass with nothing due must not launch anything.
	h.service.fireDue(context.Background())
	if count := h.runner.count(); count != 2 {
		t.Fatalf("second pass launched extra runs: %d", count)
	}
}
