package schedule

import "time"

// RunEntry records one execution attempt of one task.
type RunEntry struct {
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	// Status is StatusSuccess, StatusFailed, or StatusSkipped.
	Status string `json:"status"`
	// StatusCode is the upstream HTTP status, or 0 when the request never
	// produced a response.
	StatusCode int   `json:"status_code,omitempty"`
	DurationMS int64 `json:"duration_ms"`
	// Error explains a failure or a skip. It never carries request headers.
	Error string `json:"error,omitempty"`
	// ResponseSnippet is the first bytes of the response body, truncated.
	ResponseSnippet string `json:"response_snippet,omitempty"`
}

// TaskRuntime is the observable state of one task. It is held in memory and
// cleared by a restart; nothing here is persisted.
type TaskRuntime struct {
	Scheduled bool `json:"scheduled"`
	// TriggerSource is SourceGroup, SourceTask, or SourceNone.
	TriggerSource string     `json:"trigger_source"`
	NextRun       *time.Time `json:"next_run,omitempty"`
	LastRun       *time.Time `json:"last_run,omitempty"`
	// LastStatus is StatusNever before the first run.
	LastStatus     string `json:"last_status"`
	LastStatusCode int    `json:"last_status_code,omitempty"`
	LastDurationMS int64  `json:"last_duration_ms,omitempty"`
	LastError      string `json:"last_error,omitempty"`
	RunCount       int64  `json:"run_count"`
	FailureCount   int64  `json:"failure_count"`
	// Running reports whether a run is in flight right now.
	Running bool `json:"running"`
	// Runs holds the most recent entries, newest first.
	Runs []RunEntry `json:"runs"`
}

// GroupRuntime aggregates the state of a group's tasks. A group owns no runs of
// its own: a group trigger fires its tasks, and each of those records a run.
type GroupRuntime struct {
	Scheduled     bool       `json:"scheduled"`
	TriggerSource string     `json:"trigger_source"`
	NextRun       *time.Time `json:"next_run,omitempty"`
	LastRun       *time.Time `json:"last_run,omitempty"`
	LastStatus    string     `json:"last_status"`
	RunningTasks  int        `json:"running_tasks"`
	TaskCount     int        `json:"task_count"`
	EnabledTasks  int        `json:"enabled_tasks"`
	RunCount      int64      `json:"run_count"`
	FailureCount  int64      `json:"failure_count"`
}

// TaskSnapshot pairs a stored task with its runtime state.
type TaskSnapshot struct {
	Task
	Runtime TaskRuntime `json:"runtime"`
}

// GroupSnapshot pairs a stored group with its runtime state and its tasks.
type GroupSnapshot struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Enabled     bool           `json:"enabled"`
	Schedule    *Trigger       `json:"schedule,omitempty"`
	Runtime     GroupRuntime   `json:"runtime"`
	Tasks       []TaskSnapshot `json:"tasks"`
}

// Snapshot is the whole scheduler view the management API serves.
type Snapshot struct {
	Configured         bool            `json:"configured"`
	Dir                string          `json:"dir,omitempty"`
	SchedulerRunning   bool            `json:"scheduler_running"`
	ServerTime         time.Time       `json:"server_time"`
	GroupCount         int             `json:"group_count"`
	TaskCount          int             `json:"task_count"`
	EnabledTaskCount   int             `json:"enabled_task_count"`
	ScheduledTaskCount int             `json:"scheduled_task_count"`
	Problems           []LoadProblem   `json:"problems"`
	Groups             []GroupSnapshot `json:"groups"`
}
