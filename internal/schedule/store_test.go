package schedule

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func sampleGroup(id string) Group {
	return Group{
		ID:      id,
		Name:    "Group " + id,
		Enabled: true,
		Schedule: &Trigger{
			Type:  TriggerCron,
			Value: "0 9 * * *",
		},
		Tasks: []Task{{
			ID:      "t1",
			Name:    "First",
			Enabled: true,
			Request: Request{Method: "GET", URL: "https://example.com/a"},
		}},
	}
}

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	if !store.Configured() {
		t.Fatal("store with a directory reported unconfigured")
	}
	if _, errPut := store.Put(sampleGroup("alpha")); errPut != nil {
		t.Fatalf("Put: %v", errPut)
	}
	if _, errPut := store.Put(sampleGroup("beta")); errPut != nil {
		t.Fatalf("Put: %v", errPut)
	}
	groups, problems, errList := store.List()
	if errList != nil {
		t.Fatalf("List: %v", errList)
	}
	if len(problems) != 0 {
		t.Fatalf("List reported problems for a healthy store: %v", problems)
	}
	if len(groups) != 2 || groups[0].ID != "alpha" || groups[1].ID != "beta" {
		t.Fatalf("List returned %v, want alpha then beta", groupIDs(groups))
	}
	if groups[0].Tasks[0].Request.TimeoutMS != DefaultTimeoutMS {
		t.Fatalf("timeout default not applied: %d", groups[0].Tasks[0].Request.TimeoutMS)
	}
	deleted, errDelete := store.Delete("alpha")
	if errDelete != nil || !deleted {
		t.Fatalf("Delete = %v, %v; want true, nil", deleted, errDelete)
	}
	if again, errAgain := store.Delete("alpha"); errAgain != nil || again {
		t.Fatalf("second Delete = %v, %v; want false, nil", again, errAgain)
	}
}

func groupIDs(groups []Group) []string {
	ids := make([]string, 0, len(groups))
	for _, group := range groups {
		ids = append(ids, group.ID)
	}
	return ids
}

// An id is also a file name, so anything that could escape the directory must be
// refused rather than sanitized.
func TestStoreRejectsPathTraversalIDs(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	for _, id := range []string{"../outside", "..", ".", "a/b", `a\b`, "", "with space", "-leading"} {
		if _, errPut := store.Put(sampleGroup(id)); errPut == nil {
			t.Fatalf("Put accepted id %q", id)
		}
		if _, _, errGet := store.Get(id); errGet == nil {
			t.Fatalf("Get accepted id %q", id)
		}
	}
	entries, errRead := os.ReadDir(dir)
	if errRead != nil {
		t.Fatalf("ReadDir: %v", errRead)
	}
	if len(entries) != 0 {
		t.Fatalf("rejected writes left files behind: %v", entries)
	}
}

// One broken file must not hide the rest, and must not be silently dropped.
func TestStoreReportsUnusableFileWithoutHidingOthers(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	if _, errPut := store.Put(sampleGroup("good")); errPut != nil {
		t.Fatalf("Put: %v", errPut)
	}
	if errWrite := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{not json"), 0o600); errWrite != nil {
		t.Fatalf("WriteFile: %v", errWrite)
	}
	// A document whose id disagrees with its file name is unusable too.
	mismatched := `{"id":"other","name":"Mismatch","enabled":true,"tasks":[]}`
	if errWrite := os.WriteFile(filepath.Join(dir, "mismatch.json"), []byte(mismatched), 0o600); errWrite != nil {
		t.Fatalf("WriteFile: %v", errWrite)
	}
	groups, problems, errList := store.List()
	if errList != nil {
		t.Fatalf("List: %v", errList)
	}
	if len(groups) != 1 || groups[0].ID != "good" {
		t.Fatalf("healthy group lost: %v", groupIDs(groups))
	}
	if len(problems) != 2 {
		t.Fatalf("problems = %+v, want one per unusable file", problems)
	}
	for _, problem := range problems {
		if strings.TrimSpace(problem.Message) == "" {
			t.Fatalf("problem %q has no message", problem.File)
		}
	}
}

func TestStoreWriteIsAtomicAndNotWorldReadable(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	if _, errPut := store.Put(sampleGroup("atomic")); errPut != nil {
		t.Fatalf("Put: %v", errPut)
	}
	entries, errRead := os.ReadDir(dir)
	if errRead != nil {
		t.Fatalf("ReadDir: %v", errRead)
	}
	if len(entries) != 1 || entries[0].Name() != "atomic.json" {
		t.Fatalf("left temp files behind: %v", entries)
	}
	if runtime.GOOS == "windows" {
		// Windows has no POSIX mode bits; the check is meaningful on CI only.
		return
	}
	info, errStat := os.Stat(filepath.Join(dir, "atomic.json"))
	if errStat != nil {
		t.Fatalf("Stat: %v", errStat)
	}
	if perm := info.Mode().Perm(); perm != groupFilePerm {
		t.Fatalf("perm = %v, want %v (stored headers carry tokens)", perm, groupFilePerm)
	}
}

func TestStoreGetMissingIsNotAnError(t *testing.T) {
	store := NewStore(t.TempDir())
	group, found, errGet := store.Get("absent")
	if errGet != nil || found {
		t.Fatalf("Get = %+v, %v, %v; want zero, false, nil", group, found, errGet)
	}
}

func TestStoreUnconfiguredRefusesWrites(t *testing.T) {
	store := NewStore("   ")
	if store.Configured() {
		t.Fatal("blank directory reported configured")
	}
	if _, errPut := store.Put(sampleGroup("x")); errPut == nil {
		t.Fatal("Put succeeded without a directory")
	}
	groups, problems, errList := store.List()
	if errList != nil || groups != nil || problems != nil {
		t.Fatalf("List = %v, %v, %v; want all empty", groups, problems, errList)
	}
}

func TestNormalizeGroupRejectsInvalidDocuments(t *testing.T) {
	base := func(mutate func(*Group)) Group {
		group := sampleGroup("ok")
		mutate(&group)
		return group
	}
	cases := []struct {
		name  string
		group Group
		want  string
	}{
		{"blank name", base(func(g *Group) { g.Name = "  " }), "name is required"},
		{"bad url scheme", base(func(g *Group) { g.Tasks[0].Request.URL = "ftp://example.com" }), "scheme"},
		{"relative url", base(func(g *Group) { g.Tasks[0].Request.URL = "/only/path" }), "absolute"},
		{"body on get", base(func(g *Group) { g.Tasks[0].Request.Body = "x" }), "cannot carry a body"},
		{"body and form", base(func(g *Group) {
			g.Tasks[0].Request.Method = "POST"
			g.Tasks[0].Request.Body = "a=1"
			g.Tasks[0].Request.Form = []FormField{{Name: "a", Value: "1"}}
		}), "mutually exclusive"},
		{"bad cron", base(func(g *Group) { g.Schedule = &Trigger{Type: TriggerCron, Value: "not a cron"} }), "cron"},
		{"zero interval", base(func(g *Group) { g.Schedule = &Trigger{Type: TriggerFixedRate, Value: "0s"} }), "greater than zero"},
		{"unknown trigger", base(func(g *Group) { g.Schedule = &Trigger{Type: "hourly", Value: "1h"} }), "schedule.type"},
		{"duplicate task ids", base(func(g *Group) {
			g.Tasks = append(g.Tasks, g.Tasks[0])
		}), "duplicate id"},
		{"timeout out of range", base(func(g *Group) { g.Tasks[0].Request.TimeoutMS = MaxTimeoutMS + 1 }), "timeout_ms"},
		{"empty header name", base(func(g *Group) {
			g.Tasks[0].Request.Headers = map[string]string{"  ": "v"}
		}), "empty name"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, errNormalize := NormalizeGroup(testCase.group); errNormalize == nil {
				t.Fatalf("NormalizeGroup accepted an invalid document")
			} else if !strings.Contains(errNormalize.Error(), testCase.want) {
				t.Fatalf("error %q does not mention %q", errNormalize.Error(), testCase.want)
			}
		})
	}
}

func TestNormalizeGroupAppliesDefaults(t *testing.T) {
	normalized, errNormalize := NormalizeGroup(Group{
		ID:      "defaults",
		Name:    "  Defaults  ",
		Enabled: true,
		Tasks: []Task{{
			ID:      "t1",
			Name:    "  Lowercase method  ",
			Enabled: true,
			Request: Request{Method: "post", URL: "  https://example.com/x  ", Body: "a=1\n"},
		}},
	})
	if errNormalize != nil {
		t.Fatalf("NormalizeGroup: %v", errNormalize)
	}
	if normalized.Name != "Defaults" {
		t.Fatalf("Name = %q", normalized.Name)
	}
	task := normalized.Tasks[0]
	if task.Request.Method != "POST" {
		t.Fatalf("Method = %q, want POST", task.Request.Method)
	}
	if task.Request.URL != "https://example.com/x" {
		t.Fatalf("URL = %q", task.Request.URL)
	}
	if task.Request.Body != "a=1" {
		t.Fatalf("Body = %q, want the trailing newline trimmed", task.Request.Body)
	}
	if task.Request.TimeoutMS != DefaultTimeoutMS {
		t.Fatalf("TimeoutMS = %d", task.Request.TimeoutMS)
	}
}

// A missing directory is normal before the first write, and must not be created
// just by reading.
func TestStoreDoesNotCreateDirectoryOnRead(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "not-yet")
	store := NewStore(dir)
	if _, _, errList := store.List(); errList != nil {
		t.Fatalf("List: %v", errList)
	}
	if _, errStat := os.Stat(dir); !os.IsNotExist(errStat) {
		t.Fatalf("List created %s", dir)
	}
	if _, errPut := store.Put(sampleGroup("first")); errPut != nil {
		t.Fatalf("Put: %v", errPut)
	}
	if _, errStat := os.Stat(dir); errStat != nil {
		t.Fatalf("Put did not create %s: %v", dir, errStat)
	}
}
