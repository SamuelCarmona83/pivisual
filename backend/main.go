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

	"github.com/gin-gonic/gin"
)

var sessionsDir = filepath.Join(os.Getenv("HOME"), ".pi/agent/sessions")

func decodeProject(dirname string) string {
	// Dirname is --<path>-- where - encodes /
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
		// Sort by name descending (newest first)
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() > files[j].Name()
		})

		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
				continue
			}
			fpath := filepath.Join(projPath, f.Name())
			summary := parseSessionFile(fpath, project)
			if summary != nil {
				sessions = append(sessions, *summary)
			}
		}
	}
	return sessions
}

func parseSessionFile(fpath, project string) *SessionSummary {
	f, err := os.Open(fpath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return nil
	}

	var header map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &header); err != nil {
		return nil
	}
	if header["type"] != "session" {
		return nil
	}

	var sessionName *string
	firstUser := ""
	userMsgs, assistantMsgs := 0, 0
	totalInput, totalOutput := 0.0, 0.0
	totalCost := 0.0
	model, provider := "?", "?"

	for _, line := range lines[1:] {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		t, _ := entry["type"].(string)
		if t == "session_info" {
			if name, ok := entry["name"].(string); ok {
				n := name
				sessionName = &n
			}
			continue
		}
		if t != "message" {
			continue
		}

		msg, _ := entry["message"].(map[string]any)
		role, _ := msg["role"].(string)

		if role == "user" {
			userMsgs++
			if firstUser == "" {
				content := msg["content"]
				switch c := content.(type) {
				case string:
					firstUser = c
				case []any:
					var parts []string
					for _, item := range c {
						if m, ok := item.(map[string]any); ok {
							if m["type"] == "text" {
								if t, ok := m["text"].(string); ok {
									parts = append(parts, t)
								}
							}
						}
					}
					firstUser = strings.Join(parts, " ")
				}
				if len(firstUser) > 200 {
					firstUser = firstUser[:200]
				}
			}
		} else if role == "assistant" {
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

	if firstUser == "" {
		firstUser = "(empty)"
	}

	id, _ := header["id"].(string)
	ts, _ := header["timestamp"].(string)

	return &SessionSummary{
		ID:            id,
		File:          fpath,
		Project:       project,
		Timestamp:     ts,
		Name:          sessionName,
		FirstUser:     firstUser,
		Model:         model,
		Provider:      provider,
		UserMsgs:      userMsgs,
		AssistantMsgs: assistantMsgs,
		TotalTokens:   int(totalInput + totalOutput),
		TotalCost:     totalCost,
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// API: list sessions
	r.GET("/api/sessions", func(c *gin.Context) {
		c.JSON(http.StatusOK, parseSessionDir())
	})

	// API: get session lines
	r.GET("/api/session", func(c *gin.Context) {
		fpath := c.Query("file")
		if fpath == "" || !strings.HasPrefix(fpath, sessionsDir) {
			c.Status(http.StatusNotFound)
			return
		}
		if _, err := os.Stat(fpath); os.IsNotExist(err) {
			c.Status(http.StatusNotFound)
			return
		}

		f, err := os.Open(fpath)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		defer f.Close()

		var lines []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) != "" {
				lines = append(lines, line)
			}
		}
		c.JSON(http.StatusOK, lines)
	})

	// Static files: serve frontend/dist if it exists
	// Try relative to binary location first (Docker), then relative to CWD (dev)
	distPath := filepath.Join("frontend", "dist")
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		distPath = filepath.Join("..", "frontend", "dist")
	}
	absDist, _ := filepath.Abs(distPath)
	if _, err := os.Stat(absDist); err == nil {
		r.Static("/assets", filepath.Join(absDist, "assets"))
		r.StaticFile("/favicon.ico", filepath.Join(absDist, "favicon.ico"))
		r.GET("/", func(c *gin.Context) {
			c.File(filepath.Join(absDist, "index.html"))
		})
		r.NoRoute(func(c *gin.Context) {
			c.File(filepath.Join(absDist, "index.html"))
		})
		log.Println("Serving frontend from", absDist)
	} else {
		log.Println("Frontend dist not found at", absDist, ", API-only mode")
	}

	port := "8765"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	fmt.Printf("Serving at http://localhost:%s\n", port)
	log.Fatal(r.Run("0.0.0.0:" + port))
}
