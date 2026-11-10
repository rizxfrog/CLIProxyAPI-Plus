package management

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/schedule"
)

// SetScheduleService installs the scheduled-request service. Handlers only read
// and mutate definitions; they never start timers themselves.
func (h *Handler) SetScheduleService(service *schedule.Service) {
	if h == nil {
		return
	}
	h.mu.Lock()
	h.scheduleService = service
	h.mu.Unlock()
}

// scheduleServiceOrError resolves the service, or answers 503.
func (h *Handler) scheduleServiceOrError(c *gin.Context) (*schedule.Service, bool) {
	h.mu.Lock()
	service := h.scheduleService
	h.mu.Unlock()
	if service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "schedule_unavailable"})
		return nil, false
	}
	return service, true
}

// writeScheduleError maps a service error onto a status code. A missing
// directory is a conflict rather than a bad request: the payload may be perfect
// while the feature is switched off.
func writeScheduleError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, schedule.ErrNotConfigured):
		status = http.StatusConflict
	case errors.Is(err, schedule.ErrGroupNotFound), errors.Is(err, schedule.ErrTaskNotFound):
		status = http.StatusNotFound
	case errors.Is(err, schedule.ErrGroupExists), errors.Is(err, schedule.ErrTaskRunning):
		status = http.StatusConflict
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

// decodeScheduleBody reads a JSON body into target, answering 400 on failure.
func decodeScheduleBody(c *gin.Context, target any) bool {
	body, errRead := c.GetRawData()
	if errRead != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unreadable request body"})
		return false
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body is required"})
		return false
	}
	if errDecode := json.Unmarshal(body, target); errDecode != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + errDecode.Error()})
		return false
	}
	return true
}

// GetScheduleStatus reports whether the feature is configured and how much is loaded.
func (h *Handler) GetScheduleStatus(c *gin.Context) {
	service, ok := h.scheduleServiceOrError(c)
	if !ok {
		return
	}
	snapshot := service.Snapshot()
	// The status endpoint answers "is it on, and how much is there"; the group
	// list is a separate read.
	snapshot.Groups = nil
	c.JSON(http.StatusOK, snapshot)
}

// ListScheduleGroups returns every group with its tasks and runtime state.
func (h *Handler) ListScheduleGroups(c *gin.Context) {
	service, ok := h.scheduleServiceOrError(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, service.Snapshot())
}

// CreateScheduleGroup stores a new group.
func (h *Handler) CreateScheduleGroup(c *gin.Context) {
	service, ok := h.scheduleServiceOrError(c)
	if !ok {
		return
	}
	var group schedule.Group
	if !decodeScheduleBody(c, &group) {
		return
	}
	saved, errCreate := service.CreateGroup(group)
	if errCreate != nil {
		writeScheduleError(c, errCreate)
		return
	}
	h.writeScheduleGroup(c, saved.ID, http.StatusCreated)
}

// GetScheduleGroup returns one group.
func (h *Handler) GetScheduleGroup(c *gin.Context) {
	if _, ok := h.scheduleServiceOrError(c); !ok {
		return
	}
	h.writeScheduleGroup(c, c.Param("id"), http.StatusOK)
}

// PutScheduleGroup replaces an existing group.
func (h *Handler) PutScheduleGroup(c *gin.Context) {
	service, ok := h.scheduleServiceOrError(c)
	if !ok {
		return
	}
	var group schedule.Group
	if !decodeScheduleBody(c, &group) {
		return
	}
	saved, errPut := service.PutGroup(c.Param("id"), group)
	if errPut != nil {
		writeScheduleError(c, errPut)
		return
	}
	h.writeScheduleGroup(c, saved.ID, http.StatusOK)
}

// DeleteScheduleGroup removes a group and everything in it.
func (h *Handler) DeleteScheduleGroup(c *gin.Context) {
	service, ok := h.scheduleServiceOrError(c)
	if !ok {
		return
	}
	if errDelete := service.DeleteGroup(c.Param("id")); errDelete != nil {
		writeScheduleError(c, errDelete)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "deleted": true})
}

// AddScheduleTasks appends tasks to a group. The body carries a list, so one
// paste of several curl commands becomes one request; entries that cannot be
// used are reported per index instead of failing the batch.
func (h *Handler) AddScheduleTasks(c *gin.Context) {
	service, ok := h.scheduleServiceOrError(c)
	if !ok {
		return
	}
	var body struct {
		Tasks []schedule.Task `json:"tasks"`
	}
	if !decodeScheduleBody(c, &body) {
		return
	}
	if len(body.Tasks) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tasks is required"})
		return
	}
	added, skipped, _, errAdd := service.AddTasks(c.Param("id"), body.Tasks)
	if errAdd != nil {
		writeScheduleError(c, errAdd)
		return
	}
	if len(added) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "no task was added",
			"added":   0,
			"skipped": skipped,
		})
		return
	}
	snapshot, errSnapshot := service.GroupSnapshotOf(c.Param("id"))
	response := gin.H{"status": "ok", "added": len(added), "tasks": added, "skipped": skipped}
	if errSnapshot == nil {
		response["group"] = snapshot
	}
	c.JSON(http.StatusCreated, response)
}

// PutScheduleTask replaces one task in place.
func (h *Handler) PutScheduleTask(c *gin.Context) {
	service, ok := h.scheduleServiceOrError(c)
	if !ok {
		return
	}
	var task schedule.Task
	if !decodeScheduleBody(c, &task) {
		return
	}
	saved, errPut := service.PutTask(c.Param("id"), c.Param("task"), task)
	if errPut != nil {
		writeScheduleError(c, errPut)
		return
	}
	h.writeScheduleGroup(c, c.Param("id"), http.StatusOK, saved)
}

// DeleteScheduleTask removes one task from a group.
func (h *Handler) DeleteScheduleTask(c *gin.Context) {
	service, ok := h.scheduleServiceOrError(c)
	if !ok {
		return
	}
	if errDelete := service.DeleteTask(c.Param("id"), c.Param("task")); errDelete != nil {
		writeScheduleError(c, errDelete)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "deleted": true})
}

// RunScheduleTask performs one task's request now, regardless of its trigger.
func (h *Handler) RunScheduleTask(c *gin.Context) {
	service, ok := h.scheduleServiceOrError(c)
	if !ok {
		return
	}
	entry, errRun := service.RunTask(c.Request.Context(), c.Param("id"), c.Param("task"))
	if errRun != nil {
		writeScheduleError(c, errRun)
		return
	}
	response := gin.H{"status": "ok", "run": entry}
	if snapshot, errSnapshot := service.GroupSnapshotOf(c.Param("id")); errSnapshot == nil {
		response["group"] = snapshot
	}
	c.JSON(http.StatusOK, response)
}

// ParseScheduleCurl converts pasted curl commands into task drafts. It never
// executes anything: the page shows the result for review before saving.
func (h *Handler) ParseScheduleCurl(c *gin.Context) {
	if _, ok := h.scheduleServiceOrError(c); !ok {
		return
	}
	var body struct {
		Curl string `json:"curl"`
	}
	if !decodeScheduleBody(c, &body) {
		return
	}
	if strings.TrimSpace(body.Curl) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curl is required"})
		return
	}
	commands, problems := schedule.ParseCurls(body.Curl)
	requests := make([]gin.H, 0, len(commands))
	for _, command := range commands {
		requests = append(requests, gin.H{"name": command.Name, "request": command.Request})
	}
	if problems == nil {
		problems = []schedule.CurlError{}
	}
	c.JSON(http.StatusOK, gin.H{"requests": requests, "errors": problems})
}

// writeScheduleGroup answers with one group's full view, so a mutation and the
// subsequent list read cannot disagree about what was stored.
func (h *Handler) writeScheduleGroup(c *gin.Context, id string, status int, extra ...schedule.Task) {
	service := h.scheduleService
	if service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "schedule_unavailable"})
		return
	}
	snapshot, errSnapshot := service.GroupSnapshotOf(id)
	if errSnapshot != nil {
		writeScheduleError(c, errSnapshot)
		return
	}
	response := gin.H{"status": "ok", "group": snapshot}
	if len(extra) > 0 {
		response["task"] = extra[0]
	}
	c.JSON(status, response)
}
