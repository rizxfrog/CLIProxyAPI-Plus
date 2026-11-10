// Package schedule runs operator-authored HTTP requests on a timer.
//
// Definitions are stored as one JSON file per group under the directory named by
// automation.schedule.dir, and are managed through the v8 management API. A group
// carries an optional trigger that applies to every task in it; a task may carry its
// own trigger, which overrides the group's.
//
// This package is a general-purpose scheduler: it knows nothing about providers,
// routing, model catalogues, or credential state.
package schedule

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Trigger types. Each type interprets Trigger.Value differently.
const (
	// TriggerFixedRate fires every Value interval, measured from the schedule
	// activation anchor, regardless of how long a run takes.
	TriggerFixedRate = "fixed_rate"
	// TriggerFixedDelay waits Value after the previous run finishes.
	TriggerFixedDelay = "fixed_delay"
	// TriggerCron fires at each wall-clock time matched by the Value expression.
	TriggerCron = "cron"
)

// Trigger sources reported alongside runtime state.
const (
	SourceGroup = "group"
	SourceTask  = "task"
	SourceNone  = "none"
)

// Run statuses.
const (
	StatusNever   = "never"
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"
)

const (
	// DefaultTimeoutMS bounds one scheduled request when the task does not say.
	DefaultTimeoutMS = 30000
	// MinTimeoutMS and MaxTimeoutMS bound request.timeout_ms.
	MinTimeoutMS = 1
	MaxTimeoutMS = 600000
	// MaxNameLength bounds group and task display names.
	MaxNameLength = 200
	// MaxDescriptionLength bounds a group description.
	MaxDescriptionLength = 1000
	// MaxBodyBytes bounds a stored request body.
	MaxBodyBytes = 1 << 20
	// MaxHeaderCount bounds the header map size.
	MaxHeaderCount = 64
	// MaxTasksPerGroup bounds a group's task list.
	MaxTasksPerGroup = 200
)

// idPattern keeps an id usable both as a URL path segment and as a file name.
// The leading character is alphanumeric, so "." and ".." cannot be expressed.
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

// methodPattern is the RFC 7230 token production, restricted to what a request line accepts.
var methodPattern = regexp.MustCompile(`^[A-Z][A-Z0-9-]{0,31}$`)

// Trigger is a schedule attached to a group or a task.
type Trigger struct {
	// Type is one of TriggerFixedRate, TriggerFixedDelay, or TriggerCron.
	Type string `json:"type"`
	// Value is a Go duration string for the interval types, or a five-field
	// cron expression.
	Value string `json:"value"`
}

// FormField is one application/x-www-form-urlencoded entry, built per run.
type FormField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Request is the HTTP request a task performs.
type Request struct {
	Method  string            `json:"method,omitempty"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	// Body is a raw request body. Mutually exclusive with Form.
	Body string `json:"body,omitempty"`
	// Form holds urlencoded fields, encoded at send time so a run never reuses a
	// stale encoding. Mutually exclusive with Body.
	Form            []FormField `json:"form,omitempty"`
	TimeoutMS       int         `json:"timeout_ms,omitempty"`
	FollowRedirects bool        `json:"follow_redirects,omitempty"`
}

// Task is one scheduled HTTP request inside a group.
type Task struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	// Schedule overrides the group trigger when set.
	Schedule *Trigger `json:"schedule,omitempty"`
	Request  Request  `json:"request"`
}

// Group is the unit of storage: one file, one optional trigger, many tasks.
type Group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Enabled false disables every task in the group.
	Enabled bool `json:"enabled"`
	// Schedule applies to every task without its own trigger. Null means each
	// task schedules itself.
	Schedule *Trigger `json:"schedule,omitempty"`
	Tasks    []Task   `json:"tasks"`
}

// EffectiveTrigger resolves which trigger drives a task, and where it came from.
// A task's own trigger wins; the group's applies otherwise; a task with neither is
// never scheduled and runs only on demand.
func EffectiveTrigger(group *Group, task *Task) (*Trigger, string) {
	if task != nil && task.Schedule != nil {
		return task.Schedule, SourceTask
	}
	if group != nil && group.Schedule != nil {
		return group.Schedule, SourceGroup
	}
	return nil, SourceNone
}

// ValidateID reports whether an id is safe as a path segment and a file name.
func ValidateID(id string) error {
	if !idPattern.MatchString(id) {
		return fmt.Errorf("id %q must match %s", id, idPattern.String())
	}
	return nil
}

// ValidateTrigger checks a trigger's type, its value grammar, and that the result
// can be scheduled at all.
func ValidateTrigger(trigger *Trigger) error {
	if trigger == nil {
		return nil
	}
	switch trigger.Type {
	case TriggerFixedRate, TriggerFixedDelay:
		interval, errInterval := time.ParseDuration(strings.TrimSpace(trigger.Value))
		if errInterval != nil {
			return fmt.Errorf("%s: %q is not a duration such as 30s or 5m", trigger.Type, trigger.Value)
		}
		if interval <= 0 {
			return fmt.Errorf("%s: interval must be greater than zero", trigger.Type)
		}
		return nil
	case TriggerCron:
		if _, errCron := ParseCron(trigger.Value); errCron != nil {
			return errCron
		}
		return nil
	case "":
		return fmt.Errorf("schedule.type is required")
	default:
		return fmt.Errorf("schedule.type %q must be %s, %s, or %s",
			trigger.Type, TriggerFixedRate, TriggerFixedDelay, TriggerCron)
	}
}

// Interval returns the parsed interval for the interval trigger types.
func (t *Trigger) Interval() (time.Duration, error) {
	if t == nil {
		return 0, fmt.Errorf("schedule is not set")
	}
	if t.Type != TriggerFixedRate && t.Type != TriggerFixedDelay {
		return 0, fmt.Errorf("schedule.type %q is not an interval type", t.Type)
	}
	interval, errParse := time.ParseDuration(strings.TrimSpace(t.Value))
	if errParse != nil {
		return 0, fmt.Errorf("%s: %q is not a duration: %w", t.Type, t.Value, errParse)
	}
	if interval <= 0 {
		return 0, fmt.Errorf("%s: interval must be greater than zero", t.Type)
	}
	return interval, nil
}

// ValidateRequest checks a request without mutating it. Callers that accept
// operator input should run NormalizeRequest first so defaults are applied.
func ValidateRequest(request *Request) error {
	if request == nil {
		return fmt.Errorf("request is required")
	}
	if errMethod := validateMethod(request.Method); errMethod != nil {
		return errMethod
	}
	trimmed := strings.TrimSpace(request.URL)
	if trimmed == "" {
		return fmt.Errorf("request.url is required")
	}
	parsed, errParse := url.Parse(trimmed)
	if errParse != nil {
		return fmt.Errorf("request.url is not a valid URL: %w", errParse)
	}
	if parsed.Scheme == "" {
		return fmt.Errorf("request.url must be an absolute http or https URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("request.url scheme %q must be http or https", parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("request.url must include a host")
	}
	if request.Body != "" && len(request.Form) > 0 {
		return fmt.Errorf("request.body and request.form are mutually exclusive")
	}
	if request.Body != "" && !methodAllowsBody(request.Method) {
		return fmt.Errorf("request.method %q cannot carry a body", request.Method)
	}
	if len(request.Body) > MaxBodyBytes {
		return fmt.Errorf("request.body exceeds %d bytes", MaxBodyBytes)
	}
	if len(request.Headers) > MaxHeaderCount {
		return fmt.Errorf("request.headers exceeds %d entries", MaxHeaderCount)
	}
	for name := range request.Headers {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("request.headers contains an empty name")
		}
	}
	for index, field := range request.Form {
		if strings.TrimSpace(field.Name) == "" {
			return fmt.Errorf("request.form[%d].name is required", index)
		}
	}
	if request.TimeoutMS != 0 && (request.TimeoutMS < MinTimeoutMS || request.TimeoutMS > MaxTimeoutMS) {
		return fmt.Errorf("request.timeout_ms must be between %d and %d", MinTimeoutMS, MaxTimeoutMS)
	}
	return nil
}

func validateMethod(method string) error {
	trimmed := strings.TrimSpace(method)
	if trimmed == "" {
		return nil
	}
	if !methodPattern.MatchString(trimmed) {
		return fmt.Errorf("request.method %q must be an uppercase HTTP method token", method)
	}
	return nil
}

func methodAllowsBody(method string) bool {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS", "TRACE":
		return false
	default:
		return true
	}
}

// NormalizeRequest applies defaults and trims, then validates. It returns a value
// safe to store and to send.
func NormalizeRequest(request Request) (Request, error) {
	request.Method = strings.ToUpper(strings.TrimSpace(request.Method))
	if request.Method == "" {
		request.Method = "GET"
	}
	request.URL = strings.TrimSpace(request.URL)
	request.Body = strings.TrimRight(request.Body, "\r\n")
	if request.TimeoutMS == 0 {
		request.TimeoutMS = DefaultTimeoutMS
	}
	if len(request.Headers) > 0 {
		headers := make(map[string]string, len(request.Headers))
		for name, value := range request.Headers {
			trimmed := strings.TrimSpace(name)
			if trimmed == "" {
				// Drop nothing silently: a header with no name is a mistake in
				// the definition, not something to quietly discard.
				return Request{}, fmt.Errorf("request.headers contains an empty name")
			}
			headers[trimmed] = strings.TrimSpace(value)
		}
		request.Headers = headers
	}
	if errValidate := ValidateRequest(&request); errValidate != nil {
		return Request{}, errValidate
	}
	return request, nil
}

// NormalizeTask applies request defaults and validates the task itself.
func NormalizeTask(task Task) (Task, error) {
	task.ID = strings.TrimSpace(task.ID)
	task.Name = strings.TrimSpace(task.Name)
	if errID := ValidateID(task.ID); errID != nil {
		return Task{}, errID
	}
	if task.Name == "" {
		return Task{}, fmt.Errorf("name is required")
	}
	if len(task.Name) > MaxNameLength {
		return Task{}, fmt.Errorf("name exceeds %d characters", MaxNameLength)
	}
	if errTrigger := ValidateTrigger(task.Schedule); errTrigger != nil {
		return Task{}, errTrigger
	}
	request, errRequest := NormalizeRequest(task.Request)
	if errRequest != nil {
		return Task{}, errRequest
	}
	task.Request = request
	return task, nil
}

// NormalizeGroup applies defaults and validates the whole document, including that
// task ids are unique within the group. It is the single entry point every writer
// must use, so an invalid document never reaches disk.
func NormalizeGroup(group Group) (Group, error) {
	group.ID = strings.TrimSpace(group.ID)
	group.Name = strings.TrimSpace(group.Name)
	group.Description = strings.TrimSpace(group.Description)
	if errID := ValidateID(group.ID); errID != nil {
		return Group{}, errID
	}
	if group.Name == "" {
		return Group{}, fmt.Errorf("name is required")
	}
	if len(group.Name) > MaxNameLength {
		return Group{}, fmt.Errorf("name exceeds %d characters", MaxNameLength)
	}
	if len(group.Description) > MaxDescriptionLength {
		return Group{}, fmt.Errorf("description exceeds %d characters", MaxDescriptionLength)
	}
	if errTrigger := ValidateTrigger(group.Schedule); errTrigger != nil {
		return Group{}, errTrigger
	}
	if len(group.Tasks) > MaxTasksPerGroup {
		return Group{}, fmt.Errorf("tasks exceeds %d entries", MaxTasksPerGroup)
	}
	tasks := make([]Task, 0, len(group.Tasks))
	seen := make(map[string]struct{}, len(group.Tasks))
	for index, task := range group.Tasks {
		normalized, errTask := NormalizeTask(task)
		if errTask != nil {
			return Group{}, fmt.Errorf("tasks[%d]: %w", index, errTask)
		}
		if _, duplicate := seen[normalized.ID]; duplicate {
			return Group{}, fmt.Errorf("tasks[%d]: duplicate id %q", index, normalized.ID)
		}
		seen[normalized.ID] = struct{}{}
		tasks = append(tasks, normalized)
	}
	group.Tasks = tasks
	return group, nil
}

// FindTask returns the task with the given id, or nil.
func (g *Group) FindTask(id string) *Task {
	if g == nil {
		return nil
	}
	for index := range g.Tasks {
		if g.Tasks[index].ID == id {
			return &g.Tasks[index]
		}
	}
	return nil
}

// taskScheduled reports whether a task will ever fire on a timer.
func taskScheduled(group *Group, task *Task) bool {
	trigger, source := EffectiveTrigger(group, task)
	return trigger != nil && source != SourceNone
}
