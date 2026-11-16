package checkin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/rand/v2"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/checkin/template"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

const workerCount = 4
const queueCapacity = 64
const clockCalibration = 30 * time.Second

type schedulerTimer interface {
	Chan() <-chan time.Time
	Stop() bool
}
type schedulerClock interface {
	Now() time.Time
	NewTimer(time.Duration) schedulerTimer
}
type realClock struct{}
type realTimer struct{ *time.Timer }

func (realClock) Now() time.Time                          { return time.Now() }
func (realClock) NewTimer(d time.Duration) schedulerTimer { return realTimer{time.NewTimer(d)} }
func (t realTimer) Chan() <-chan time.Time                { return t.C }

type accountIdentity interface{ Identity(*coreauth.Auth) string }

// Metadata dependencies are repair hints, not mandatory-field validation rules.
type metadataDependencies interface{ RequiredMetadata() []string }

// externalDependency reports a value a credential depends on that does not live
// in its metadata, such as a session the operator refreshes on the local client.
//
// It is a repair hint like metadataDependencies: the returned string is hashed
// and compared between reconciliations, so a change replans a blocked account.
// Adapters that store everything in metadata omit it.
type externalDependency interface{ ExternalDependency(*coreauth.Auth) string }

type providerRuntime struct {
	config        config.CheckinProviderConfig
	fingerprint   string
	revision      uint64
	location      *time.Location
	headers       *template.Headers
	paths         []string
	problem       string
	calendarDate  string
	windowStart   time.Time
	windowEnd     time.Time
	calendarError error
	nextMidnight  time.Time
}

func (runtime *providerRuntime) window(date string) (time.Time, time.Time, error) {
	if runtime.calendarDate != date {
		runtime.calendarDate = date
		runtime.windowStart, runtime.windowEnd, runtime.calendarError = calendarWindow(date, runtime.config, runtime.location)
		runtime.nextMidnight, _ = resolveWall(adjacentDate(date, 1), 0, runtime.location)
	}
	return runtime.windowStart, runtime.windowEnd, runtime.calendarError
}

type accountObservation struct {
	provider string
	epoch    uint64
	identity string
	enabled  bool
	date     string
	revision uint64
}
type planKey struct{ provider, id, date string }
type accountPlan struct {
	status            AccountStatus
	epoch             uint64
	identity          string
	revision          uint64
	generation        uint64
	repair            string
	credentialVersion uint64
	scheduled         time.Time
}
type scheduledJob struct {
	key        planKey
	epoch      uint64
	identity   string
	revision   uint64
	generation uint64
	ctx        context.Context
	cancel     context.CancelFunc
}

// Service owns one scheduling loop and a fixed worker pool. Network operations never hold mu.
type Service struct {
	mu         sync.Mutex
	source     AccountSource
	adapters   map[string]Adapter
	providers  map[string]*providerRuntime
	accounts   map[string]accountObservation
	plans      map[planKey]*accountPlan
	active     map[string]*scheduledJob
	cfg        *config.Config
	clock      schedulerClock
	random     func(int64) int64
	wake       chan struct{}
	jobs       chan *scheduledJob
	cancel     context.CancelFunc
	ctx        context.Context
	done       chan struct{}
	started    bool
	stopped    bool
	reconciled bool
	sequence   uint64
	workers    sync.WaitGroup
}

func NewService(source AccountSource, adapters ...Adapter) *Service {
	return newService(source, realClock{}, rand.Int64N, adapters...)
}

func newService(source AccountSource, clock schedulerClock, random func(int64) int64, adapters ...Adapter) *Service {
	service := &Service{source: source, adapters: make(map[string]Adapter), providers: make(map[string]*providerRuntime), accounts: make(map[string]accountObservation), plans: make(map[planKey]*accountPlan), active: make(map[string]*scheduledJob), clock: clock, random: random, wake: make(chan struct{}, 1), jobs: make(chan *scheduledJob, queueCapacity), done: make(chan struct{})}
	for _, adapter := range adapters {
		if adapter != nil && adapter.Provider() != "" {
			service.adapters[adapter.Provider()] = adapter
		}
	}
	return service
}

// Start is nonblocking and may only be called once, even after Stop.
func (s *Service) Start(ctx context.Context, cfg *config.Config) error {
	if ctx == nil {
		return errors.New("checkin_context_required")
	}
	s.mu.Lock()
	if s.started || s.stopped {
		s.mu.Unlock()
		return errors.New("checkin_already_started")
	}
	s.started = true
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.mu.Unlock()
	s.ApplyConfig(cfg)
	s.workers.Add(workerCount)
	for range workerCount {
		go s.worker()
	}
	go s.loop()
	return nil
}

// ApplyConfig copies the caller's snapshot and invalidates only meaningful provider changes.
func (s *Service) ApplyConfig(cfg *config.Config) {
	if cfg == nil {
		cfg = &config.Config{}
	}
	snapshot := cfg.CloneForRuntime()
	s.mu.Lock()
	previousHome := s.cfg != nil && s.cfg.Home.Enabled
	s.cfg = snapshot
	for id, adapter := range s.adapters {
		effective := snapshot.AutoCheckin.Providers[id].Effective()
		fingerprint := Fingerprint(effective)
		previous := s.providers[id]
		if previous != nil && previous.fingerprint == fingerprint && previousHome == snapshot.Home.Enabled {
			continue
		}
		revision := uint64(1)
		if previous != nil {
			revision = previous.revision + 1
		}
		runtime := &providerRuntime{config: effective, fingerprint: fingerprint, revision: revision}
		runtime.location, _ = time.LoadLocation(effective.Timezone)
		if runtime.location == nil {
			runtime.problem = "invalid_timezone"
		}
		var err error
		runtime.headers, err = template.Compile(effective.Headers, id)
		if err != nil {
			runtime.problem = "invalid_header_template"
		} else {
			runtime.paths = runtime.headers.Paths()
		}
		if deps, ok := adapter.(metadataDependencies); ok {
			runtime.paths = append(runtime.paths, deps.RequiredMetadata()...)
		}
		s.providers[id] = runtime
		for _, job := range s.active {
			if job.key.provider == id {
				job.cancel()
			}
		}
	}
	s.mu.Unlock()
	s.Notify()
}

// Notify coalesces account and configuration wakeups without blocking their owners.
func (s *Service) Notify() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *Service) Stop() {
	s.mu.Lock()
	if !s.started {
		s.stopped = true
		s.mu.Unlock()
		return
	}
	cancel, done := s.cancel, s.done
	s.mu.Unlock()
	cancel()
	<-done
}

func (s *Service) loop() {
	defer close(s.done)
	defer func() {
		s.mu.Lock()
		s.stopped = true
		for _, job := range s.active {
			job.cancel()
		}
		s.mu.Unlock()
		s.workers.Wait()
		s.mu.Lock()
		for id, job := range s.active {
			if plan := s.plans[job.key]; plan != nil && plan.generation == job.generation {
				s.blockPlanLocked(plan, "task_cancelled", s.clock.Now())
			}
			delete(s.active, id)
		}
		s.mu.Unlock()
	}()
	for {
		if s.ctx.Err() != nil {
			return
		}
		s.reconcile()
		s.mu.Lock()
		delay := s.nextWakeLocked(s.clock.Now())
		s.mu.Unlock()
		timer := s.clock.NewTimer(delay)
		select {
		case <-s.ctx.Done():
			timer.Stop()
			return
		case <-s.wake:
			timer.Stop()
		case <-timer.Chan():
		}
	}
}

func (s *Service) reconcile() {
	s.mu.Lock()
	if s.cfg == nil {
		s.mu.Unlock()
		return
	}
	enabled := false
	for _, provider := range s.providers {
		if provider.config.Enabled {
			enabled = true
			break
		}
	}
	if s.cfg.Home.Enabled || !enabled {
		for _, job := range s.active {
			job.cancel()
		}
		for key, plan := range s.plans {
			if !plan.status.Completed && plan.status.Phase != "running" {
				delete(s.plans, key)
			}
		}
		for id, observed := range s.accounts {
			observed.enabled = false
			s.accounts[id] = observed
		}
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	var accounts []*coreauth.Auth
	if s.source != nil {
		accounts = s.source.List()
	}
	now := s.clock.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg == nil || s.cfg.Home.Enabled {
		return
	}
	seen := make(map[string]bool, len(accounts))
	for _, auth := range accounts {
		if auth == nil || auth.ID == "" {
			continue
		}
		runtime := s.providers[auth.Provider]
		if runtime == nil {
			continue
		}
		seen[auth.ID] = true
		identity := s.identity(auth)
		previous, known := s.accounts[auth.ID]
		replacement := known && (previous.provider != auth.Provider || previous.epoch != auth.RegistrationEpoch || (previous.identity != "" && identity != "" && previous.identity != identity))
		if replacement {
			s.removeAccountLocked(auth.ID)
			known = false
		}
		date := ""
		if runtime.location != nil {
			date = now.In(runtime.location).Format("2006-01-02")
		}
		enabled := !auth.Disabled && auth.Status != coreauth.StatusDisabled && runtime.config.Enabled && !s.cfg.Home.Enabled
		observation := accountObservation{provider: auth.Provider, epoch: auth.RegistrationEpoch, identity: identity, enabled: enabled, date: date, revision: runtime.revision}
		s.accounts[auth.ID] = observation
		key := planKey{auth.Provider, auth.ID, date}
		plan := s.plans[key]
		if !enabled {
			if job := s.active[auth.ID]; job != nil {
				job.cancel()
			}
			for otherKey, other := range s.plans {
				if otherKey.id == auth.ID && !other.status.Completed && other.status.Phase != "running" {
					delete(s.plans, otherKey)
				}
			}
			continue
		}
		trigger := "account_added"
		if !s.reconciled {
			trigger = "recovery"
		}
		daily := known && previous.enabled && previous.revision == runtime.revision && previous.date != "" && previous.date != date && !replacement
		if daily {
			trigger = "daily"
		}
		if known && (!previous.enabled || (plan != nil && plan.revision != runtime.revision)) {
			trigger = "config_changed"
		}
		replan := plan == nil || (!plan.status.Completed && plan.revision != runtime.revision)
		repair := ""
		if replan || (plan != nil && plan.status.Phase == "blocked") {
			repair = dependencyFingerprint(auth.Metadata, runtime.paths)
			if provider, ok := s.adapters[auth.Provider].(externalDependency); ok {
				// Fold the opaque value in so a refreshed session outside the
				// credential file still counts as a repair.
				if value := provider.ExternalDependency(auth); value != "" {
					repair = Fingerprint(repair + "\x00" + value)
				}
			}
		}
		credentialRepaired := plan != nil && plan.status.Error == "credential_unavailable" && plan.credentialVersion != auth.CredentialVersion
		if plan != nil && plan.status.Phase == "blocked" && (plan.repair != repair || credentialRepaired) {
			replan = true
			trigger = "config_changed"
		}
		if replan {
			s.sequence++
			plan = &accountPlan{status: AccountStatus{AuthID: auth.ID, Account: accountDisplay(auth.ID), Provider: auth.Provider, Date: date, Phase: "scheduled", Trigger: trigger}, epoch: auth.RegistrationEpoch, identity: identity, revision: runtime.revision, generation: s.sequence, repair: repair}
			plan.credentialVersion = auth.CredentialVersion
			s.plans[key] = plan
			if runtime.problem != "" {
				s.blockPlanLocked(plan, runtime.problem, now)
			} else {
				start, end, err := runtime.window(date)
				if err != nil {
					s.blockPlanLocked(plan, err.Error(), now)
				} else {
					lower := start
					if !daily {
						if now.Before(start) || !now.Before(end) {
							lower = now
							end = now
						} else {
							lower = now
						}
					} else if now.After(lower) {
						lower = now
					}
					plan.scheduled = lower
					if end.After(lower) {
						plan.scheduled = lower.Add(time.Duration(s.random(int64(end.Sub(lower)))))
					}
					plan.status.ScheduledAt = timestamp(plan.scheduled)
				}
			}
		}
	}
	s.reconciled = true
	for id := range s.accounts {
		if !seen[id] {
			s.removeAccountLocked(id)
		}
	}
	for key, plan := range s.plans {
		runtime := s.providers[key.provider]
		if runtime == nil || runtime.location == nil {
			continue
		}
		today := now.In(runtime.location).Format("2006-01-02")
		if key.date < adjacentDate(today, -1) && s.active[key.id] == nil {
			delete(s.plans, key)
			continue
		}
		if key.date != today || plan.status.Phase != "scheduled" || plan.scheduled.After(now) {
			continue
		}
		if s.active[key.id] != nil {
			continue
		}
		observed, exists := s.accounts[key.id]
		if !exists || !observed.enabled || plan.revision != runtime.revision {
			continue
		}
		ctx, cancel := context.WithCancel(s.ctx)
		job := &scheduledJob{key: key, epoch: plan.epoch, identity: plan.identity, revision: plan.revision, generation: plan.generation, ctx: ctx, cancel: cancel}
		select {
		case s.jobs <- job:
			s.active[key.id] = job
			plan.status.Phase = "queued"
		default:
			cancel()
		}
	}
}

func (s *Service) removeAccountLocked(id string) {
	if job := s.active[id]; job != nil {
		job.cancel()
	}
	for key := range s.plans {
		if key.id == id {
			delete(s.plans, key)
		}
	}
	delete(s.accounts, id)
}

func (s *Service) blockPlanLocked(plan *accountPlan, code string, now time.Time) {
	plan.status.Phase = "blocked"
	plan.status.Outcome = OutcomeBlocked
	plan.status.Error = safeCode(code, "checkin_blocked")
	plan.status.FinishedAt = timestamp(now)
}

func (s *Service) nextWakeLocked(now time.Time) time.Duration {
	next := now.Add(clockCalibration)
	for _, runtime := range s.providers {
		if !runtime.config.Enabled || runtime.location == nil {
			continue
		}
		today := now.In(runtime.location).Format("2006-01-02")
		runtime.window(today)
		if midnight := runtime.nextMidnight; midnight.After(now) && midnight.Before(next) {
			next = midnight
		}
	}
	for key, plan := range s.plans {
		if plan.status.Phase == "scheduled" && plan.scheduled.After(now) && plan.scheduled.Before(next) && s.active[key.id] == nil {
			next = plan.scheduled
		}
	}
	return next.Sub(now)
}

func (s *Service) worker() {
	defer s.workers.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		case job := <-s.jobs:
			s.execute(job)
		}
	}
}

// validJobLocked checks plan ownership without consulting mutable credential contents.
func (s *Service) validJobLocked(job *scheduledJob, auth *coreauth.Auth, now time.Time) (*accountPlan, *providerRuntime, bool) {
	plan := s.plans[job.key]
	runtime := s.providers[job.key.provider]
	if job.ctx.Err() != nil || plan == nil || runtime == nil || runtime.location == nil || s.cfg.Home.Enabled || !runtime.config.Enabled || plan.generation != job.generation || runtime.revision != job.revision || plan.revision != job.revision || s.active[job.key.id] != job {
		return nil, nil, false
	}
	if auth == nil || auth.Disabled || auth.Status == coreauth.StatusDisabled || auth.Provider != job.key.provider || auth.RegistrationEpoch != job.epoch || auth.ID != job.key.id {
		return nil, nil, false
	}
	identity := s.identity(auth)
	if job.identity != "" && identity != "" && identity != job.identity {
		return nil, nil, false
	}
	if now.In(runtime.location).Format("2006-01-02") != job.key.date {
		return nil, nil, false
	}
	return plan, runtime, true
}

func (s *Service) execute(job *scheduledJob) {
	defer func() {
		job.cancel()
		s.mu.Lock()
		if s.active[job.key.id] == job {
			delete(s.active, job.key.id)
		}
		if plan := s.plans[job.key]; plan != nil && plan.generation == job.generation && (plan.status.Phase == "queued" || plan.status.Phase == "running") {
			plan.status.Phase = "blocked"
			plan.status.Outcome = OutcomeBlocked
			plan.status.Error = "task_cancelled"
			plan.status.FinishedAt = timestamp(s.clock.Now())
		}
		s.mu.Unlock()
		s.Notify()
	}()
	if job.ctx.Err() != nil || s.source == nil {
		return
	}
	auth, ok := s.source.GetByID(job.key.id)
	if !ok {
		return
	}
	s.mu.Lock()
	plan, _, valid := s.validJobLocked(job, auth, s.clock.Now())
	if valid {
		plan.status.Phase = "running"
		plan.status.StartedAt = timestamp(s.clock.Now())
	}
	s.mu.Unlock()
	if !valid {
		return
	}
	auth, err := s.source.PrepareCredentialForBackground(job.ctx, job.key.id, job.epoch)
	if err != nil {
		s.finish(job, Result{Outcome: OutcomeBlocked, Message: "credential_unavailable"})
		return
	}
	// Refresh may race with removal, replacement, a config change, or midnight.
	latest, exists := s.source.GetByID(job.key.id)
	s.mu.Lock()
	_, runtime, valid := s.validJobLocked(job, auth, s.clock.Now())
	if valid {
		_, _, valid = s.validJobLocked(job, latest, s.clock.Now())
		valid = valid && exists
	}
	var cfg *config.Config
	if valid {
		cfg = s.cfg
	}
	s.mu.Unlock()
	if !valid {
		return
	}
	auth = latest
	headers, err := runtime.headers.Render(auth.Metadata)
	if err != nil {
		s.finish(job, Result{Outcome: OutcomeBlocked, Message: "header_template_unresolved"})
		return
	}
	if job.ctx.Err() != nil {
		return
	}
	result := s.adapters[job.key.provider].CheckIn(job.ctx, cfg, auth, headers, job.key.date)
	s.finish(job, result)
}

func (s *Service) finish(job *scheduledJob, result Result) {
	// Re-read the registration before accepting a result, including unnotified replacements.
	auth, exists := s.source.GetByID(job.key.id)
	now := s.clock.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	plan := s.plans[job.key]
	runtime := s.providers[job.key.provider]
	if !exists || auth == nil || auth.Disabled || auth.Status == coreauth.StatusDisabled || auth.RegistrationEpoch != job.epoch || auth.Provider != job.key.provider || plan == nil || runtime == nil || plan.generation != job.generation || plan.revision != job.revision || runtime.revision != job.revision || job.ctx.Err() != nil {
		return
	}
	identity := s.identity(auth)
	if job.identity != "" && identity != "" && identity != job.identity {
		return
	}
	plan.status.FinishedAt = timestamp(now)
	outcome := result.Outcome
	switch outcome {
	case OutcomeSuccess, OutcomeAlreadyDone, OutcomeUnavailable, OutcomeFailed, OutcomeBlocked:
	default:
		outcome = OutcomeFailed
	}
	completed := outcome == OutcomeSuccess || outcome == OutcomeAlreadyDone
	if completed && ((result.Date != "" && result.Date != job.key.date) || (result.Date == "" && now.In(runtime.location).Format("2006-01-02") != job.key.date)) {
		outcome = OutcomeFailed
		completed = false
		result.Message = "result_date_uncertain"
	}
	plan.status.Outcome = outcome
	plan.status.Completed = completed
	plan.status.Warning = safeCode(result.Warning, "")
	if completed {
		plan.status.Phase = "done"
		plan.status.Error = ""
	} else {
		plan.status.Phase = "failed"
		if outcome == OutcomeBlocked {
			plan.status.Phase = "blocked"
			plan.credentialVersion = auth.CredentialVersion
		}
		plan.status.Error = safeCode(result.Message, "checkin_"+outcome)
	}
}

func (s *Service) identity(auth *coreauth.Auth) string {
	if provider, ok := s.adapters[auth.Provider].(accountIdentity); ok {
		if value := provider.Identity(auth); value != "" {
			return Fingerprint(value)
		}
	}
	return ""
}

func dependencyFingerprint(metadata map[string]any, paths []string) string {
	values := make(map[string]any, len(paths))
	for _, path := range paths {
		var value any = metadata
		for _, segment := range strings.Split(path, ".") {
			object, ok := value.(map[string]any)
			if !ok {
				value = nil
				break
			}
			value = object[segment]
		}
		values[path] = value
	}
	return Fingerprint(values)
}

func accountDisplay(id string) string {
	sum := sha256.Sum256([]byte(id))
	return "Account " + hex.EncodeToString(sum[:4])
}
func timestamp(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339Nano)
}
func safeCode(value, fallback string) string {
	if len(value) == 0 || len(value) > 96 {
		return fallback
	}
	for _, char := range value {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '_' {
			return fallback
		}
	}
	return value
}

// Status returns independent, deterministic DTOs without labels, paths, headers, or metadata.
func (s *Service) Status(provider string) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot := Snapshot{Providers: []ProviderStatus{}, Accounts: []AccountStatus{}}
	effective := make(map[string]config.CheckinProviderConfig, len(s.providers))
	now := s.clock.Now()
	if s.cfg != nil && s.cfg.Home.Enabled {
		snapshot.DisabledReason = "home_mode"
	}
	for id, runtime := range s.providers {
		effective[id] = runtime.config
		if provider != "" && provider != id {
			continue
		}
		date := ""
		if runtime.location != nil {
			date = now.In(runtime.location).Format("2006-01-02")
		}
		snapshot.Providers = append(snapshot.Providers, ProviderStatus{ID: id, Enabled: runtime.config.Enabled, Timezone: runtime.config.Timezone, Date: date, ConfigFingerprint: runtime.fingerprint})
	}
	snapshot.ConfigFingerprint = Fingerprint(effective)
	for key, plan := range s.plans {
		if provider == "" || provider == key.provider {
			snapshot.Accounts = append(snapshot.Accounts, plan.status)
		}
	}
	sort.Slice(snapshot.Providers, func(i, j int) bool { return snapshot.Providers[i].ID < snapshot.Providers[j].ID })
	sort.Slice(snapshot.Accounts, func(i, j int) bool {
		a, b := snapshot.Accounts[i], snapshot.Accounts[j]
		if a.Provider != b.Provider {
			return a.Provider < b.Provider
		}
		if a.AuthID != b.AuthID {
			return a.AuthID < b.AuthID
		}
		return a.Date < b.Date
	})
	return snapshot
}
