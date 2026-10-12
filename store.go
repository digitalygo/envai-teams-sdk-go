package envai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type CallState string

const (
	CallStarted   CallState = "started"
	CallSubmitted CallState = "submitted"
	CallResolved  CallState = "resolved"
	CallAmbiguous CallState = "ambiguous"
)

type CallIdentity struct {
	Namespace string `json:"namespace"`
	ChatID    string `json:"chat_id"`
	CallID    string `json:"call_id"`
	TurnID    string `json:"turn_id"`
	ToolName  string `json:"tool_name"`
}
type CallStore interface {
	Claim(context.Context, CallIdentity) (bool, error)
	Record(context.Context, CallIdentity, CallState) error
	State(context.Context, CallIdentity) (CallState, bool, error)
}

func identityKey(id CallIdentity) string {
	b, _ := json.Marshal([]string{id.Namespace, id.ChatID, id.CallID})
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

type MemoryCallStore struct {
	mu    sync.Mutex
	calls map[string]CallState
}

func NewMemoryCallStore() *MemoryCallStore { return &MemoryCallStore{calls: map[string]CallState{}} }
func (s *MemoryCallStore) Claim(ctx context.Context, id CallIdentity) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := identityKey(id)
	if _, ok := s.calls[key]; ok {
		return false, nil
	}
	s.calls[key] = CallStarted
	return true, nil
}
func (s *MemoryCallStore) Record(ctx context.Context, id CallIdentity, state CallState) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := identityKey(id)
	if _, ok := s.calls[key]; !ok {
		return errors.New("envai: cannot update unclaimed call")
	}
	s.calls[key] = state
	return nil
}

type FileCallStore struct{ root string }
type callRecord struct {
	Identity CallIdentity `json:"identity"`
	State    CallState    `json:"state"`
}

func NewFileCallStore(directory string) (*FileCallStore, error) {
	absolute, err := filepath.Abs(directory)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(absolute, 0700); err != nil {
		return nil, err
	}
	st, err := os.Lstat(absolute)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("envai: call store must be a real directory")
	}
	return &FileCallStore{root: absolute}, nil
}
func (s *FileCallStore) Claim(ctx context.Context, id CallIdentity) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	dir := filepath.Join(s.root, identityKey(id))
	err := os.Mkdir(dir, 0700)
	if errors.Is(err, os.ErrExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err = s.write(dir, id, CallStarted); err != nil {
		return false, err
	}
	root, err := os.Open(s.root)
	if err != nil {
		return false, err
	}
	defer root.Close()
	if err = root.Sync(); err != nil {
		return false, err
	}
	return true, nil
}
func (s *FileCallStore) Record(ctx context.Context, id CallIdentity, state CallState) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dir := filepath.Join(s.root, identityKey(id))
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("envai: invalid call record directory")
	}
	return s.write(dir, id, state)
}
func (s *FileCallStore) write(dir string, id CallIdentity, state CallState) error {
	b, err := json.Marshal(callRecord{id, state})
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".record-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(name, filepath.Join(dir, "state.json")); err != nil {
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

func (s *MemoryCallStore) State(ctx context.Context, id CallIdentity) (CallState, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.calls[identityKey(id)]
	return state, ok, nil
}
func (s *FileCallStore) State(ctx context.Context, id CallIdentity) (CallState, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	dir := filepath.Join(s.root, identityKey(id))
	info, err := os.Lstat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", true, errors.New("envai: invalid call state directory")
	}
	b, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if errors.Is(err, os.ErrNotExist) {
		return CallAmbiguous, true, nil
	}
	if err != nil {
		return "", true, err
	}
	var record callRecord
	if err = json.Unmarshal(b, &record); err != nil {
		return "", true, errors.New("envai: corrupt call state; reconciliation required")
	}
	if identityKey(record.Identity) != identityKey(id) {
		return "", true, errors.New("envai: mismatched call state")
	}
	switch record.State {
	case CallStarted, CallSubmitted, CallResolved, CallAmbiguous:
		return record.State, true, nil
	}
	return "", true, errors.New("envai: invalid call state")
}
