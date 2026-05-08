package confy

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

type mockFS struct {
	files     map[string]string
	statCalls []string
}

func (m *mockFS) Open(name string) (io.ReadCloser, error) {
	if data, ok := m.files[name]; ok {
		return io.NopCloser(strings.NewReader(data)), nil
	}
	return nil, os.ErrNotExist
}

func (m *mockFS) Stat(name string) (os.FileInfo, error) {
	m.statCalls = append(m.statCalls, name)
	if _, ok := m.files[name]; ok {
		return &mockFileInfo{name: name}, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockFS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return nil, os.ErrNotExist
}

type mockFileInfo struct{ name string }

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() any           { return nil }

func TestSetFS(t *testing.T) {
	v := New()
	m := &mockFS{files: map[string]string{}}
	v.SetFS(m)
	if v.fs == nil {
		t.Fatal("fs should not be nil after SetFS")
	}
}

func TestFSUsage(t *testing.T) {
	v := New()
	m := &mockFS{files: map[string]string{"/mock/file": "hello"}}
	v.SetFS(m)

	exists, err := exists(v.fs, "/mock/file")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatal("expected /mock/file to exist")
	}
	if len(m.statCalls) != 1 || m.statCalls[0] != "/mock/file" {
		t.Fatalf("expected Stat to be called once with /mock/file, got %v", m.statCalls)
	}
}
