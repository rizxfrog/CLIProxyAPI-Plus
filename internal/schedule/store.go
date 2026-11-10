package schedule

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Store persists group documents under one directory.
//
// One file per group, named after a validated id, so a stored document can never
// address a path outside the directory. Writes are atomic and the files are not
// group/other readable, because request headers routinely carry bearer tokens.
//
// Typed decoding and encoding use encoding/json deliberately: ojson is the
// document path get/set library the repository requires for mutating payloads,
// and it exposes no struct codec.
type Store struct {
	dir string
}

const (
	groupFileSuffix = ".json"
	// maxGroupFileBytes bounds a definition file. A schedule is a small
	// document; anything larger is not one.
	maxGroupFileBytes = 1 << 20
	// groupFilePerm keeps tokens in request headers out of other users' reach.
	groupFilePerm = 0o600
	// groupDirPerm is the permission used when the directory must be created.
	groupDirPerm = 0o700
)

// LoadProblem reports one file that exists but could not be used. A single broken
// file must not hide the healthy schedules, and it must not be silently omitted
// either, so listing returns them alongside the groups that did load.
type LoadProblem struct {
	File    string `json:"file"`
	Message string `json:"message"`
}

// NewStore builds a store rooted at dir. An empty dir yields an unconfigured
// store, which reports no groups and refuses every write.
func NewStore(dir string) *Store {
	return &Store{dir: strings.TrimSpace(dir)}
}

// Dir is the configured directory, or empty when unconfigured.
func (s *Store) Dir() string {
	if s == nil {
		return ""
	}
	return s.dir
}

// Configured reports whether the store has a directory to work in.
func (s *Store) Configured() bool {
	return s != nil && s.dir != ""
}

// path resolves a group id to its file. The id is validated first, so the result
// is always a direct child of the store directory.
func (s *Store) path(id string) (string, error) {
	if !s.Configured() {
		return "", fmt.Errorf("schedule: automation.schedule.dir is not configured")
	}
	if errID := ValidateID(id); errID != nil {
		return "", errID
	}
	return filepath.Join(s.dir, id+groupFileSuffix), nil
}

// List returns every stored group sorted by id, plus a problem per file that
// exists but could not be read.
//
// A missing directory is not an error: the store is created on first write, so
// nothing has been scheduled yet.
func (s *Store) List() ([]Group, []LoadProblem, error) {
	if !s.Configured() {
		return nil, nil, nil
	}
	entries, errRead := os.ReadDir(s.dir)
	if errRead != nil {
		if errors.Is(errRead, fs.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("schedule: read %s: %w", s.dir, errRead)
	}
	groups := make([]Group, 0, len(entries))
	problems := make([]LoadProblem, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), groupFileSuffix) {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), groupFileSuffix)
		group, found, errLoad := s.Get(id)
		if errLoad != nil {
			problems = append(problems, LoadProblem{File: entry.Name(), Message: errLoad.Error()})
			continue
		}
		if !found {
			continue
		}
		groups = append(groups, group)
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].ID < groups[j].ID })
	sort.Slice(problems, func(i, j int) bool { return problems[i].File < problems[j].File })
	return groups, problems, nil
}

// Get loads one group. A missing file reports found=false and no error.
func (s *Store) Get(id string) (Group, bool, error) {
	path, errPath := s.path(id)
	if errPath != nil {
		return Group{}, false, errPath
	}
	info, errStat := os.Stat(path)
	if errStat != nil {
		if errors.Is(errStat, fs.ErrNotExist) {
			return Group{}, false, nil
		}
		return Group{}, false, fmt.Errorf("schedule: stat %s: %w", path, errStat)
	}
	if info.IsDir() {
		return Group{}, false, fmt.Errorf("schedule: %s is a directory", path)
	}
	if info.Size() > maxGroupFileBytes {
		return Group{}, false, fmt.Errorf("schedule: %s exceeds %d bytes", path, maxGroupFileBytes)
	}
	data, errRead := os.ReadFile(path)
	if errRead != nil {
		return Group{}, false, fmt.Errorf("schedule: read %s: %w", path, errRead)
	}
	var group Group
	if errDecode := json.Unmarshal(data, &group); errDecode != nil {
		return Group{}, false, fmt.Errorf("schedule: %s is not a valid group document: %w", path, errDecode)
	}
	// A file on disk may predate the current rules or have been edited by hand.
	// Report it rather than trusting it, but do not rewrite it here: reading must
	// stay side-effect free.
	normalized, errNormalize := NormalizeGroup(group)
	if errNormalize != nil {
		return Group{}, false, fmt.Errorf("schedule: %s: %w", path, errNormalize)
	}
	if normalized.ID != id {
		return Group{}, false, fmt.Errorf("schedule: %s declares id %q", path, normalized.ID)
	}
	return normalized, true, nil
}

// Put validates a group and writes it atomically, replacing any existing file.
func (s *Store) Put(group Group) (Group, error) {
	normalized, errNormalize := NormalizeGroup(group)
	if errNormalize != nil {
		return Group{}, errNormalize
	}
	path, errPath := s.path(normalized.ID)
	if errPath != nil {
		return Group{}, errPath
	}
	if errDir := os.MkdirAll(s.dir, groupDirPerm); errDir != nil {
		return Group{}, fmt.Errorf("schedule: create %s: %w", s.dir, errDir)
	}
	encoded, errEncode := json.MarshalIndent(normalized, "", "  ")
	if errEncode != nil {
		return Group{}, fmt.Errorf("schedule: encode group %s: %w", normalized.ID, errEncode)
	}
	encoded = append(encoded, '\n')
	if errWrite := writeFileAtomic(path, encoded, groupFilePerm); errWrite != nil {
		return Group{}, errWrite
	}
	return normalized, nil
}

// Delete removes a group file. It reports whether the file existed.
func (s *Store) Delete(id string) (bool, error) {
	path, errPath := s.path(id)
	if errPath != nil {
		return false, errPath
	}
	errRemove := os.Remove(path)
	if errRemove != nil {
		if errors.Is(errRemove, fs.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("schedule: remove %s: %w", path, errRemove)
	}
	return true, nil
}

// writeFileAtomic writes via a sibling temp file and a rename, so a reader never
// observes a half-written document and a crash cannot truncate an existing one.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	temp, errCreate := os.CreateTemp(dir, ".schedule-*.tmp")
	if errCreate != nil {
		return fmt.Errorf("schedule: create temp file in %s: %w", dir, errCreate)
	}
	tempName := temp.Name()
	// The rename below consumes the temp file on success; on any failure the
	// leftover is removed here.
	defer func() { _ = os.Remove(tempName) }()
	if errWrite := temp.Chmod(perm); errWrite != nil {
		_ = temp.Close()
		return fmt.Errorf("schedule: set mode on %s: %w", tempName, errWrite)
	}
	if _, errWrite := temp.Write(data); errWrite != nil {
		_ = temp.Close()
		return fmt.Errorf("schedule: write %s: %w", tempName, errWrite)
	}
	if errSync := temp.Sync(); errSync != nil {
		_ = temp.Close()
		return fmt.Errorf("schedule: sync %s: %w", tempName, errSync)
	}
	if errClose := temp.Close(); errClose != nil {
		return fmt.Errorf("schedule: close %s: %w", tempName, errClose)
	}
	if errRename := os.Rename(tempName, path); errRename != nil {
		return fmt.Errorf("schedule: replace %s: %w", path, errRename)
	}
	return nil
}
