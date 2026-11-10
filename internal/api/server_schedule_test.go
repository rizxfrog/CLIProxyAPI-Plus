package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/api/handlers/management"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/schedule"
)

const scheduleTestPassword = "schedule-test-password"

// scheduleTestServer builds a server with the schedule routes registered over a
// real store directory, mirroring how the service wires itself up.
func scheduleTestServer(t *testing.T, dir string, routesEnabled bool) (*Server, *schedule.Service) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := fmt.Sprintf(`server: {port: 8317}
management: {secret-key: %s}
automation:
  schedule:
    dir: '%s'
`, scheduleTestPassword, filepath.ToSlash(dir))
	if errWrite := os.WriteFile(path, []byte(content), 0o600); errWrite != nil {
		t.Fatalf("write config: %v", errWrite)
	}
	cfg, errLoad := config.LoadConfig(path)
	if errLoad != nil {
		t.Fatalf("LoadConfig: %v", errLoad)
	}
	if !cfg.Schedule.Enabled() {
		t.Fatalf("schedule directory did not survive the config round trip: %+v", cfg.Schedule)
	}
	service := schedule.NewService(schedule.NewStore(dir))
	handler := management.NewHandler(cfg, path, nil)
	handler.SetLocalPassword(scheduleTestPassword)
	handler.SetScheduleService(service)
	server := &Server{cfg: cfg, engine: gin.New(), mgmt: handler}
	server.managementRoutesEnabled.Store(routesEnabled)
	server.registerManagementRoutes()
	return server, service
}

// scheduleRequest builds a request shaped like the one the panel sends. The
// management anti-bot middleware risk-scores state-changing requests, so a bare
// request with no User-Agent or Sec-Fetch headers is rejected with 403 before
// authentication runs; tests must exercise the same path the panel does.
func scheduleRequest(method, route, body string, authorized bool) *http.Request {
	request := httptest.NewRequest(method, route, strings.NewReader(body))
	request.RemoteAddr = "127.0.0.1:1234"
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("Sec-Fetch-Mode", "cors")
	request.Header.Set("Sec-Fetch-Dest", "empty")
	request.Header.Set("Origin", "http://127.0.0.1:8317")
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if authorized {
		request.Header.Set("Authorization", "Bearer "+scheduleTestPassword)
	}
	return request
}

// doSchedule issues an authenticated request against the v8 management API.
func doSchedule(t *testing.T, server *Server, method, route, body string) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	server.engine.ServeHTTP(response, scheduleRequest(method, route, body, true))
	return response
}

// Every new route must sit behind the same management authorization as the rest
// of the v8 contract, and must disappear when management is turned off. Each
// route is exercised with its own method, so a route registered for the wrong
// verb fails here instead of silently 404ing.
func TestScheduleManagementRoutesRequireManagementAuthorization(t *testing.T) {
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/v8/management/automation/schedule/status"},
		{http.MethodGet, "/v8/management/automation/schedule/groups"},
		{http.MethodPost, "/v8/management/automation/schedule/groups"},
		{http.MethodPost, "/v8/management/automation/schedule/parse-curl"},
		{http.MethodGet, "/v8/management/automation/schedule/groups/demo"},
		{http.MethodPut, "/v8/management/automation/schedule/groups/demo"},
		{http.MethodDelete, "/v8/management/automation/schedule/groups/demo"},
		{http.MethodPost, "/v8/management/automation/schedule/groups/demo/tasks"},
		{http.MethodPut, "/v8/management/automation/schedule/groups/demo/tasks/t1"},
		{http.MethodDelete, "/v8/management/automation/schedule/groups/demo/tasks/t1"},
		{http.MethodPost, "/v8/management/automation/schedule/groups/demo/tasks/t1/run"},
	}
	for _, route := range routes {
		// A fresh server per route: the management handler bans an IP for 30
		// minutes after five failed attempts, which is intended behaviour and
		// would otherwise turn the sixth probe here into a 403.
		server, _ := scheduleTestServer(t, t.TempDir(), true)
		response := httptest.NewRecorder()
		server.engine.ServeHTTP(response, scheduleRequest(route.method, route.path, "", false))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s without a key = %d, want 401 (is the route registered?): %s",
				route.method, route.path, response.Code, response.Body.String())
		}
	}

	disabled, _ := scheduleTestServer(t, t.TempDir(), false)
	for _, route := range routes {
		response := doSchedule(t, disabled, route.method, route.path, "")
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s %s with management disabled = %d, want 404",
				route.method, route.path, response.Code)
		}
	}
}

// The full lifecycle the panel drives, over real HTTP.
func TestScheduleManagementLifecycle(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer target.Close()

	server, _ := scheduleTestServer(t, t.TempDir(), true)

	// Nothing is stored yet, but the feature is configured.
	status := doSchedule(t, server, http.MethodGet, "/v8/management/automation/schedule/status", "")
	if status.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", status.Code, status.Body.String())
	}
	var statusBody struct {
		Configured  bool `json:"configured"`
		GroupCount  int  `json:"group_count"`
		TaskCount   int  `json:"task_count"`
		SchedulerRu bool `json:"scheduler_running"`
	}
	if errDecode := json.Unmarshal(status.Body.Bytes(), &statusBody); errDecode != nil {
		t.Fatalf("decode status: %v", errDecode)
	}
	if !statusBody.Configured || statusBody.GroupCount != 0 || statusBody.TaskCount != 0 {
		t.Fatalf("status = %+v", statusBody)
	}

	// Create a group whose tasks schedule themselves.
	create := doSchedule(t, server, http.MethodPost, "/v8/management/automation/schedule/groups", `{
		"id": "cli-smoke",
		"name": "CLI smoke",
		"enabled": true,
		"tasks": []
	}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("create group = %d: %s", create.Code, create.Body.String())
	}
	if conflict := doSchedule(t, server, http.MethodPost, "/v8/management/automation/schedule/groups", `{
		"id": "cli-smoke", "name": "CLI smoke", "enabled": true, "tasks": []
	}`); conflict.Code != http.StatusConflict {
		t.Fatalf("duplicate create = %d, want 409", conflict.Code)
	}

	// Batch add: one acceptable task, one duplicate, one unusable URL.
	add := doSchedule(t, server, http.MethodPost,
		"/v8/management/automation/schedule/groups/cli-smoke/tasks",
		fmt.Sprintf(`{"tasks":[
			{"id":"t1","name":"First","enabled":true,
			 "schedule":{"type":"cron","value":"0 9 * * *"},
			 "request":{"method":"POST","url":"%s/hook","headers":{"Authorization":"Bearer z"},"body":"{\"a\":1}"}},
			{"id":"t1","name":"Duplicate","enabled":true,"request":{"url":"%s"}},
			{"id":"t3","name":"Broken","enabled":true,"request":{"url":"not-a-url"}}
		]}`, target.URL, target.URL))
	if add.Code != http.StatusCreated {
		t.Fatalf("add tasks = %d: %s", add.Code, add.Body.String())
	}
	var addBody struct {
		Added   int                    `json:"added"`
		Skipped []schedule.SkippedTask `json:"skipped"`
	}
	if errDecode := json.Unmarshal(add.Body.Bytes(), &addBody); errDecode != nil {
		t.Fatalf("decode add: %v", errDecode)
	}
	if addBody.Added != 1 || len(addBody.Skipped) != 2 {
		t.Fatalf("add = %+v, want 1 added and 2 skipped", addBody)
	}
	if addBody.Skipped[0].Index != 0 && addBody.Skipped[0].Index != 1 {
		t.Fatalf("skipped indexes = %+v", addBody.Skipped)
	}
	for _, problem := range addBody.Skipped {
		if problem.Reason == "" {
			t.Fatalf("skipped entry %d has no reason", problem.Index)
		}
	}

	// The group read reflects the stored task and its resolved trigger.
	read := doSchedule(t, server, http.MethodGet, "/v8/management/automation/schedule/groups/cli-smoke", "")
	if read.Code != http.StatusOK {
		t.Fatalf("read group = %d: %s", read.Code, read.Body.String())
	}
	var readBody struct {
		Group struct {
			Tasks []struct {
				ID      string               `json:"id"`
				Runtime schedule.TaskRuntime `json:"runtime"`
			} `json:"tasks"`
		} `json:"group"`
	}
	if errDecode := json.Unmarshal(read.Body.Bytes(), &readBody); errDecode != nil {
		t.Fatalf("decode group: %v", errDecode)
	}
	if len(readBody.Group.Tasks) != 1 || readBody.Group.Tasks[0].ID != "t1" {
		t.Fatalf("group tasks = %+v", readBody.Group.Tasks)
	}
	if readBody.Group.Tasks[0].Runtime.TriggerSource != schedule.SourceTask {
		t.Fatalf("trigger source = %q, want %q", readBody.Group.Tasks[0].Runtime.TriggerSource, schedule.SourceTask)
	}
	if readBody.Group.Tasks[0].Runtime.NextRun == nil {
		t.Fatal("a cron task has no next run")
	}

	// A curl paste converts without saving anything.
	parsed := doSchedule(t, server, http.MethodPost, "/v8/management/automation/schedule/parse-curl",
		`{"curl":"curl -X POST 'https://api.example.com/v1/hook' -H 'X-A: 1' && curl https://api.example.com/v1/other"}`)
	if parsed.Code != http.StatusOK {
		t.Fatalf("parse-curl = %d: %s", parsed.Code, parsed.Body.String())
	}
	var parsedBody struct {
		Requests []struct {
			Name    string           `json:"name"`
			Request schedule.Request `json:"request"`
		} `json:"requests"`
		Errors []schedule.CurlError `json:"errors"`
	}
	if errDecode := json.Unmarshal(parsed.Body.Bytes(), &parsedBody); errDecode != nil {
		t.Fatalf("decode parse-curl: %v", errDecode)
	}
	if len(parsedBody.Requests) != 2 || len(parsedBody.Errors) != 0 {
		t.Fatalf("parse-curl returned %+v / %+v", parsedBody.Requests, parsedBody.Errors)
	}
	if parsedBody.Requests[0].Request.Method != "POST" {
		t.Fatalf("first request method = %q", parsedBody.Requests[0].Request.Method)
	}
	if parsedBody.Requests[1].Request.Method != "GET" {
		t.Fatalf("second request method = %q", parsedBody.Requests[1].Request.Method)
	}
	if parsedBody.Requests[0].Name == "" {
		t.Fatal("a parsed request has no suggested name")
	}

	// Run now: the request really goes out, and the result is reported.
	run := doSchedule(t, server, http.MethodPost,
		"/v8/management/automation/schedule/groups/cli-smoke/tasks/t1/run", "")
	if run.Code != http.StatusOK {
		t.Fatalf("run = %d: %s", run.Code, run.Body.String())
	}
	var runBody struct {
		Run schedule.RunEntry `json:"run"`
	}
	if errDecode := json.Unmarshal(run.Body.Bytes(), &runBody); errDecode != nil {
		t.Fatalf("decode run: %v", errDecode)
	}
	if runBody.Run.Status != schedule.StatusSuccess || runBody.Run.StatusCode != http.StatusOK {
		t.Fatalf("run = %+v", runBody.Run)
	}
	if !strings.Contains(runBody.Run.ResponseSnippet, "ok") {
		t.Fatalf("run snippet = %q", runBody.Run.ResponseSnippet)
	}

	// Replace the task, then delete it, then the group.
	put := doSchedule(t, server, http.MethodPut,
		"/v8/management/automation/schedule/groups/cli-smoke/tasks/t1",
		`{"name":"Renamed","enabled":false,"request":{"method":"GET","url":"https://example.com/renamed"}}`)
	if put.Code != http.StatusOK {
		t.Fatalf("put task = %d: %s", put.Code, put.Body.String())
	}
	missing := doSchedule(t, server, http.MethodPut,
		"/v8/management/automation/schedule/groups/cli-smoke/tasks/absent",
		`{"name":"Absent","enabled":true,"request":{"url":"https://example.com/x"}}`)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("put of a missing task = %d, want 404", missing.Code)
	}
	if deleted := doSchedule(t, server, http.MethodDelete,
		"/v8/management/automation/schedule/groups/cli-smoke/tasks/t1", ""); deleted.Code != http.StatusOK {
		t.Fatalf("delete task = %d: %s", deleted.Code, deleted.Body.String())
	}
	if gone := doSchedule(t, server, http.MethodDelete,
		"/v8/management/automation/schedule/groups/cli-smoke", ""); gone.Code != http.StatusOK {
		t.Fatalf("delete group = %d: %s", gone.Code, gone.Body.String())
	}
	if absent := doSchedule(t, server, http.MethodGet,
		"/v8/management/automation/schedule/groups/cli-smoke", ""); absent.Code != http.StatusNotFound {
		t.Fatalf("read of a deleted group = %d, want 404", absent.Code)
	}
}

// An invalid payload must be rejected without touching the store.
func TestScheduleManagementRejectsInvalidPayloads(t *testing.T) {
	dir := t.TempDir()
	server, _ := scheduleTestServer(t, dir, true)

	cases := []struct {
		name, method, route, body string
		want                      int
	}{
		{"malformed json", http.MethodPost, "/v8/management/automation/schedule/groups", "{", http.StatusBadRequest},
		{"empty body", http.MethodPost, "/v8/management/automation/schedule/groups", "", http.StatusBadRequest},
		{"bad id", http.MethodPost, "/v8/management/automation/schedule/groups",
			`{"id":"../escape","name":"x","enabled":true,"tasks":[]}`, http.StatusBadRequest},
		{"bad trigger", http.MethodPost, "/v8/management/automation/schedule/groups",
			`{"id":"ok","name":"x","enabled":true,"schedule":{"type":"cron","value":"nope"},"tasks":[]}`, http.StatusBadRequest},
		{"empty task batch", http.MethodPost, "/v8/management/automation/schedule/groups/ok/tasks",
			`{"tasks":[]}`, http.StatusBadRequest},
		{"parse with no text", http.MethodPost, "/v8/management/automation/schedule/parse-curl",
			`{"curl":"   "}`, http.StatusBadRequest},
		{"unknown group", http.MethodGet, "/v8/management/automation/schedule/groups/absent", "", http.StatusNotFound},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			response := doSchedule(t, server, testCase.method, testCase.route, testCase.body)
			if response.Code != testCase.want {
				t.Fatalf("%s %s = %d, want %d: %s",
					testCase.method, testCase.route, response.Code, testCase.want, response.Body.String())
			}
			var body map[string]any
			if errDecode := json.Unmarshal(response.Body.Bytes(), &body); errDecode != nil {
				t.Fatalf("error response is not JSON: %s", response.Body.String())
			}
			if response.Code >= 400 {
				if message, ok := body["error"].(string); !ok || strings.TrimSpace(message) == "" {
					t.Fatalf("error response has no message: %s", response.Body.String())
				}
			}
		})
	}

	entries, errRead := os.ReadDir(dir)
	if errRead != nil && !os.IsNotExist(errRead) {
		t.Fatalf("ReadDir: %v", errRead)
	}
	if len(entries) != 0 {
		t.Fatalf("rejected payloads wrote files: %v", entries)
	}
}

// Every group file is a real file in the configured directory, so an operator can
// version or hand-edit it.
func TestScheduleManagementPersistsJsonFiles(t *testing.T) {
	dir := t.TempDir()
	server, _ := scheduleTestServer(t, dir, true)
	create := doSchedule(t, server, http.MethodPost, "/v8/management/automation/schedule/groups",
		`{"id":"on-disk","name":"On disk","enabled":true,"tasks":[
			{"id":"t1","name":"T","enabled":true,"request":{"url":"https://example.com/x"}}
		]}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("create = %d: %s", create.Code, create.Body.String())
	}
	raw, errRead := os.ReadFile(filepath.Join(dir, "on-disk.json"))
	if errRead != nil {
		t.Fatalf("ReadFile: %v", errRead)
	}
	var stored schedule.Group
	if errDecode := json.Unmarshal(raw, &stored); errDecode != nil {
		t.Fatalf("stored file is not a group document: %v\n%s", errDecode, raw)
	}
	if stored.ID != "on-disk" || len(stored.Tasks) != 1 || stored.Tasks[0].ID != "t1" {
		t.Fatalf("stored document = %+v", stored)
	}
	if !strings.HasSuffix(strings.TrimSpace(string(raw)), "}") {
		t.Fatal("stored file is not complete JSON")
	}
}
