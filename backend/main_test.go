package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadLines(t *testing.T) {
	lines, err := readLines("testdata/session_pivisual.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) < 10 {
		t.Fatalf("expected >10 lines, got %d", len(lines))
	}
	// First line must be a session header
	var header map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &header); err != nil {
		t.Fatal("first line not valid JSON:", err)
	}
	if header["type"] != "session" {
		t.Fatalf("expected session header, got type=%v", header["type"])
	}
}

func TestParseSessionFile_Pivisual(t *testing.T) {
	s := parseSessionFile("testdata/session_pivisual.jsonl", "~/test/pivisual")
	if s == nil {
		t.Fatal("parseSessionFile returned nil")
	}

	if s.ID != "019f33b1-bdf2-7631-b679-6b9db05cf4f8" {
		t.Errorf("wrong ID: %s", s.ID)
	}
	if s.Project != "~/test/pivisual" {
		t.Errorf("wrong project: %s", s.Project)
	}
	if s.UserMsgs == 0 {
		t.Error("expected user messages")
	}
	if s.AssistantMsgs == 0 {
		t.Error("expected assistant messages")
	}
	if s.TotalTokens == 0 {
		t.Error("expected tokens")
	}
	if s.FirstUser == "" {
		t.Error("expected first user message")
	}
	if s.Model == "?" {
		t.Error("expected model")
	}
	if s.Provider == "?" {
		t.Error("expected provider")
	}

	t.Logf("ID: %s", s.ID)
	t.Logf("First user: %.60s", s.FirstUser)
	t.Logf("Model: %s/%s", s.Provider, s.Model)
	t.Logf("Messages: %d user, %d assistant", s.UserMsgs, s.AssistantMsgs)
	t.Logf("Tokens: %d, Cost: $%.6f", s.TotalTokens, s.TotalCost)
}

func TestParseSessionFile_Sampler(t *testing.T) {
	s := parseSessionFile("testdata/session_sampler.jsonl", "~/test/sampler")
	if s == nil {
		t.Fatal("parseSessionFile returned nil")
	}
	if s.ID != "019f29dc-213c-7a83-a84f-0150d0444da5" {
		t.Errorf("wrong ID: %s", s.ID)
	}
	if s.UserMsgs == 0 {
		t.Error("expected user messages")
	}
	// This session has many assistant messages with tool calls
	if s.AssistantMsgs < 50 {
		t.Errorf("expected >50 assistant msgs, got %d", s.AssistantMsgs)
	}
	t.Logf("ID: %s, Messages: %d/%d, Tokens: %d", s.ID, s.UserMsgs, s.AssistantMsgs, s.TotalTokens)
}

func TestParseSessionFile_NotFound(t *testing.T) {
	s := parseSessionFile("testdata/nonexistent.jsonl", "~/test")
	if s != nil {
		t.Error("expected nil for missing file")
	}
}

func TestParseSessionFile_Empty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.jsonl")
	os.WriteFile(path, []byte("\n\n"), 0644)
	s := parseSessionFile(path, "~/test")
	if s != nil {
		t.Error("expected nil for empty file")
	}
}

func TestParseSessionFile_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.jsonl")
	os.WriteFile(path, []byte(`{"type":"session","id":"x"}`+"\nnot json\n"), 0644)
	s := parseSessionFile(path, "~/test")
	if s == nil {
		t.Fatal("should parse valid header + skip bad line")
	}
	if s.ID != "x" {
		t.Errorf("wrong ID: %s", s.ID)
	}
}

func TestParseSessionFile_MissingFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "minimal.jsonl")
	os.WriteFile(path, []byte(`{"type":"session","id":"min","timestamp":"2024-01-01T00:00:00Z"}`), 0644)
	s := parseSessionFile(path, "~/test")
	if s == nil {
		t.Fatal("should parse minimal session")
	}
	if s.ID != "min" {
		t.Errorf("wrong ID: %s", s.ID)
	}
	if s.FirstUser != "(empty)" {
		t.Errorf("expected '(empty)', got '%s'", s.FirstUser)
	}
	if s.Model != "?" {
		t.Errorf("expected '?' model, got '%s'", s.Model)
	}
}

func TestParseSessionFile_UserTextExtraction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "user_text.jsonl")
	content := `{"type":"session","id":"u1"}
{"type":"message","message":{"role":"user","content":[{"type":"text","text":"Hello"},{"type":"text","text":"world"}]}}
{"type":"message","message":{"role":"assistant","content":"ok"}}`
	os.WriteFile(path, []byte(content), 0644)

	s := parseSessionFile(path, "~/test")
	if s == nil {
		t.Fatal("should parse")
	}
	if s.FirstUser != "Hello world" {
		t.Errorf("expected 'Hello world', got '%s'", s.FirstUser)
	}
}

func TestParseSessionFile_SessionInfo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "named.jsonl")
	content := `{"type":"session","id":"n1"}
{"type":"session_info","name":"My Session"}
{"type":"message","message":{"role":"user","content":"hi"}}
{"type":"message","message":{"role":"assistant","content":"hey"}}`
	os.WriteFile(path, []byte(content), 0644)

	s := parseSessionFile(path, "~/test")
	if s == nil {
		t.Fatal("should parse")
	}
	if s.Name == nil || *s.Name != "My Session" {
		t.Errorf("expected 'My Session', got %v", s.Name)
	}
}

func TestDecodeProject(t *testing.T) {
	tests := []struct {
		dirname  string
		expected string
	}{
		{"--Users-x--", "~/x"},
		{"--Users-x-Documents-Repositories-foo-bar--", "~/x/Documents/Repositories/foo/bar"},
	}
	for _, tt := range tests {
		got := decodeProject(tt.dirname)
		if !strings.HasPrefix(got, "~/") && !strings.HasPrefix(got, "home") {
			t.Errorf("decodeProject(%q) = %q, expected prefix ~/", tt.dirname, got)
		}
		if got != tt.expected {
			t.Errorf("decodeProject(%q) = %q, expected %q", tt.dirname, got, tt.expected)
		}
	}
}

func TestParseSessionDir(t *testing.T) {
	// Set up a mock sessions directory
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	// Override the global sessionsDir for this test
	oldDir := sessionsDir
	sessionsDir = filepath.Join(tmp, ".pi", "agent", "sessions")
	defer func() { sessionsDir = oldDir }()

	// Create project dir with encoded name
	projDir := filepath.Join(sessionsDir, "--Users-x-Documents-Repositories-test--")
	os.MkdirAll(projDir, 0755)

	// Copy a test session into it
	data, err := os.ReadFile("testdata/session_pivisual.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(projDir, "2026-07-05T19-11-54_test.jsonl"), data, 0644)

	sessions := parseSessionDir()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != "019f33b1-bdf2-7631-b679-6b9db05cf4f8" {
		t.Errorf("wrong ID: %s", sessions[0].ID)
	}
	if sessions[0].Project != "~/x/Documents/Repositories/test" {
		t.Errorf("wrong project: %s", sessions[0].Project)
	}
}

func TestFirstUserText(t *testing.T) {
	// String content
	if got := firstUserText("hello"); got != "hello" {
		t.Errorf("string: got %q", got)
	}
	// Array content
	arr := []any{
		map[string]any{"type": "text", "text": "hi"},
		map[string]any{"type": "text", "text": "there"},
	}
	if got := firstUserText(arr); got != "hi there" {
		t.Errorf("array: got %q", got)
	}
	// Nil
	if got := firstUserText(nil); got != "" {
		t.Errorf("nil: got %q", got)
	}
}
