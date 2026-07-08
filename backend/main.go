package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var sessionsDir = filepath.Join(os.Getenv("HOME"), ".pi/agent/sessions")

func decodeProject(dirname string) string {
	inner := strings.Trim(dirname, "-")
	project := strings.ReplaceAll(inner, "-", "/")
	if strings.HasPrefix(project, "Users/") {
		project = "~/" + project[6:]
	}
	return project
}

type SessionSummary struct {
	ID            string  `json:"id"`
	File          string  `json:"file"`
	Project       string  `json:"project"`
	Timestamp     string  `json:"timestamp"`
	Name          *string `json:"name"`
	FirstUser     string  `json:"first_user"`
	Model         string  `json:"model"`
	Provider      string  `json:"provider"`
	UserMsgs      int     `json:"user_msgs"`
	AssistantMsgs int     `json:"assistant_msgs"`
	TotalTokens   int     `json:"total_tokens"`
	TotalCost     float64 `json:"total_cost"`
	ParentSession string  `json:"parent_session,omitempty"`
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if line := sc.Text(); strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func firstUserText(content any) string {
	switch c := content.(type) {
	case string:
		return c
	case []any:
		var parts []string
		for _, item := range c {
			if m, ok := item.(map[string]any); ok && m["type"] == "text" {
				if t, ok := m["text"].(string); ok {
					parts = append(parts, t)
				}
			}
		}
		return strings.Join(parts, " ")
	}
	return ""
}

func parseSessionFile(fpath, project string) *SessionSummary {
	lines, err := readLines(fpath)
	if err != nil || len(lines) == 0 {
		return nil
	}

	var header map[string]any
	if json.Unmarshal([]byte(lines[0]), &header) != nil || header["type"] != "session" {
		return nil
	}

	var sessionName *string
	firstUser := ""
	userMsgs, assistantMsgs := 0, 0
	totalInput, totalOutput, totalCost := 0.0, 0.0, 0.0
	model, provider := "?", "?"

	for _, line := range lines[1:] {
		var entry map[string]any
		if json.Unmarshal([]byte(line), &entry) != nil {
			continue
		}
		switch entry["type"].(string) {
		case "session_info":
			if name, ok := entry["name"].(string); ok {
				n := name
				sessionName = &n
			}
		case "message":
			msg, _ := entry["message"].(map[string]any)
			switch msg["role"].(string) {
			case "user":
				userMsgs++
				if firstUser == "" {
					firstUser = firstUserText(msg["content"])
					if len(firstUser) > 200 {
						firstUser = firstUser[:200]
					}
				}
			case "assistant":
				assistantMsgs++
				if m, ok := msg["model"].(string); ok {
					model = m
				}
				if p, ok := msg["provider"].(string); ok {
					provider = p
				}
				if usage, ok := msg["usage"].(map[string]any); ok {
					if inp, ok := usage["input"].(float64); ok {
						totalInput += inp
					}
					if out, ok := usage["output"].(float64); ok {
						totalOutput += out
					}
					if cost, ok := usage["cost"].(map[string]any); ok {
						if t, ok := cost["total"].(float64); ok {
							totalCost += t
						}
					}
				}
			}
		}
	}

	if firstUser == "" {
		firstUser = "(empty)"
	}

	id, _ := header["id"].(string)
	ts, _ := header["timestamp"].(string)
	parentSession, _ := header["parentSession"].(string)

	return &SessionSummary{
		ID: id, File: fpath, Project: project, Timestamp: ts,
		Name: sessionName, FirstUser: firstUser, Model: model, Provider: provider,
		UserMsgs: userMsgs, AssistantMsgs: assistantMsgs,
		TotalTokens: int(totalInput + totalOutput), TotalCost: totalCost,
		ParentSession: parentSession,
	}
}

func parseSessionDir() []SessionSummary {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil
	}
	var sessions []SessionSummary
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		project := decodeProject(entry.Name())
		projPath := filepath.Join(sessionsDir, entry.Name())
		files, err := os.ReadDir(projPath)
		if err != nil {
			continue
		}
		sort.Slice(files, func(i, j int) bool { return files[i].Name() > files[j].Name() })
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
				continue
			}
			if s := parseSessionFile(filepath.Join(projPath, f.Name()), project); s != nil {
				sessions = append(sessions, *s)
			}
		}
	}
	return sessions
}

func main() {
	// Resolve frontend dist path relative to the binary
	exe, _ := os.Executable()
	distPath := filepath.Join(filepath.Dir(exe), "frontend", "dist")
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		distPath = filepath.Join("..", "frontend", "dist") // dev: running from backend/
	}

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(parseSessionDir())
	})

	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		q := strings.ToLower(r.URL.Query().Get("q"))
		w.Header().Set("Content-Type", "application/json")
		if q == "" { json.NewEncoder(w).Encode([]any{}); return }
		type Match struct {
			SessionID string `json:"session_id"`
			File      string `json:"file"`
			Line      int    `json:"line"`
			Role      string `json:"role"`
			Context   string `json:"context"`
		}
		var results []Match
		entries, _ := os.ReadDir(sessionsDir)
		for _, entry := range entries {
			if !entry.IsDir() { continue }
			projPath := filepath.Join(sessionsDir, entry.Name())
			files, _ := os.ReadDir(projPath)
			for _, f := range files {
				if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") { continue }
				fpath := filepath.Join(projPath, f.Name())
				lines, err := readLines(fpath)
				if err != nil { continue }
				var header map[string]any
				sid := ""
				if len(lines) > 0 && json.Unmarshal([]byte(lines[0]), &header) == nil {
					sid, _ = header["id"].(string)
				}
				for i, line := range lines {
					if !strings.Contains(strings.ToLower(line), q) { continue }
					var entry map[string]any
					if json.Unmarshal([]byte(line), &entry) != nil { continue }
					role := ""
					var ctx string
					if msg, ok := entry["message"].(map[string]any); ok {
						role, _ = msg["role"].(string)
						if c, ok := msg["content"].(string); ok {
							ctx = c
						} else if arr, ok := msg["content"].([]any); ok {
							for _, b := range arr {
								if m, ok := b.(map[string]any); ok && m["type"] == "text" {
									if t, ok := m["text"].(string); ok { ctx += t + " " }
								}
							}
						}
					}
					if len(ctx) > 200 {
						idx := strings.Index(strings.ToLower(ctx), q)
						start := idx - 40
						if start < 0 { start = 0 }
						end := idx + len(q) + 40
						if end > len(ctx) { end = len(ctx) }
						ctx = "..." + ctx[start:end] + "..."
					}
					results = append(results, Match{SessionID: sid, File: fpath, Line: i, Role: role, Context: strings.TrimSpace(ctx)})
				}
			}
		}
		json.NewEncoder(w).Encode(results)
	})

	mux.HandleFunc("/api/session", func(w http.ResponseWriter, r *http.Request) {
		fpath := r.URL.Query().Get("file")
		if fpath == "" || !strings.HasPrefix(fpath, sessionsDir) {
			http.NotFound(w, r)
			return
		}
		lines, err := readLines(fpath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(lines)
	})

	// Static files
	if _, err := os.Stat(distPath); err == nil {
		mux.Handle("/", http.FileServer(http.Dir(distPath)))
		log.Println("Serving frontend from", distPath)
	} else {
		log.Println("Frontend dist not found, API-only mode")
	}

	port := os.Getenv("PORT")
	if port == "" {
		// Read from .env file if it exists (simple dotenv, no dependency)
		// Try binary dir first (Docker), then parent dir (dev), then CWD
		dotenvPath := filepath.Join(filepath.Dir(exe), ".env")
		if _, err := os.Stat(dotenvPath); os.IsNotExist(err) {
			dotenvPath = filepath.Join(filepath.Dir(exe), "..", ".env")
		}
		if _, err := os.Stat(dotenvPath); os.IsNotExist(err) {
			dotenvPath = ".env"
		}
		if data, err := os.ReadFile(dotenvPath); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				if before, after, ok := strings.Cut(line, "="); ok {
					os.Setenv(strings.TrimSpace(before), strings.TrimSpace(after))
				}
			}
			port = os.Getenv("PORT")
		}
		if port == "" {
			port = "4000"
		}
	}
	fmt.Printf("Serving at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, mux))
}
