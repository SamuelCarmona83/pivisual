# Pi Visual — Session Viewer for Pi Coding Agent

A modern web-based session viewer for [Pi Coding Agent](https://github.com/earendil-works/pi-mono). Browse, search, and read your coding sessions in a clean chat-style interface.

![Pi Visual](https://img.shields.io/badge/status-stable-success)

## Features

- **Chat-style interface** — Read sessions as conversations with user/assistant bubbles, thinking blocks, and collapsible tool outputs
- **Session browser** — Sidebar with search, project filter, and time-grouped history (Today, Yesterday, This Week, Earlier)
- **Pin sessions** — Star important sessions for quick access; pins persist in `localStorage`
- **Markdown rendering** — Code blocks with dark terminal styling, headings, lists, tables, blockquotes, and inline code
- **Diff highlighting** — Git diffs in tool outputs are auto-detected and colorized (green/red/purple/blue)
- **ANSI stripping** — Terminal escape codes and carriage returns are cleaned from tool output
- **Session metadata** — Token counts, cost breakdown, model/provider info at a glance
- **Dashboard** — Aggregate stats on load: total sessions, tokens, cost, most-used model, and recent sessions
- **Copy session IDs** — Click to select and use with `pi --session <id>` to resume any session

## Quick Start

### Simple (Python stdlib)

```bash
python3 server.py
# Open http://localhost:8765
```

### Modern (Go + Vue)

```bash
# Build & run backend
cd backend && go build -o pivisual-server . && ./pivisual-server

# Dev mode with hot reload
cd frontend && npm run dev   # → http://localhost:5173 (proxies /api to :8765)
```

Custom port:

```bash
./pivisual-server 3000
```

## Docker

```bash
docker compose up --build -d
# Open http://localhost:8765
```

The volume mount gives the container read-only access to your Pi session files.

## Architecture

```
pivisual/
├── server.py              # Python stdlib fallback (zero-dependency)
├── backend/
│   ├── main.go            # Go Gin API server + static file server
│   ├── go.mod / go.sum
│   └── pivisual-server    # Compiled binary
├── frontend/
│   ├── src/
│   │   ├── App.vue               # Root layout + sidebar toggle (Ctrl+B)
│   │   ├── components/
│   │   │   ├── AppSidebar.vue    # Session list with search, filter, pins
│   │   │   ├── AppDashboard.vue  # Aggregate stats + recent sessions
│   │   │   ├── ChatView.vue      # Chat rendering engine
│   │   │   └── ChatMessage.vue   # Single message (user/assistant/system)
│   │   ├── composables/
│   │   │   └── useSessions.ts    # API calls, state, pins, grouping
│   │   └── types.ts              # TypeScript interfaces
│   ├── index.html                # Vite entry with design tokens & fonts
│   ├── package.json
│   └── vite.config.ts            # Dev proxy to Go backend
├── Dockerfile              # Multi-stage: Node → Go → Alpine
├── docker-compose.yml
└── README.md
```

### Stack

| Component | Tech |
|-----------|------|
| Frontend  | Vue 3 + TypeScript + Vite |
| Icons     | Phosphor Icons (`@phosphor-icons/vue`) |
| Markdown  | `marked` |
| Backend   | Go + Gin |
| Fallback  | Python 3 stdlib (`server.py`) |

### API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/sessions` | Returns all sessions as JSON with summary stats |
| `GET /api/session?file=<path>` | Returns full session JSONL lines for detail view |

## Design System

The UI follows [Anthropic Claude's design language](DESING.md):
- Warm cream canvas (`#faf9f5`) with serif display typography (Tiempos Headline)
- Coral primary accent (`#cc785c`) for user messages and interactive elements
- Dark navy surfaces (`#181715`) for code blocks and terminal output
- Inter body font + JetBrains Mono for code

## Session Resume Reference

Click any session in the sidebar, copy its ID from the chat header, then:

```bash
pi --session <id>          # Resume specific session
pi --fork <id>             # Fork into new session
```

Or through Pi's interactive mode:

```
/resume                    # Browse and select sessions
/session                   # Show current session info
/tree                      # Navigate session tree in-place
```

## Requirements

- **Python version**: Python 3.9+
- **Go version** (optional): Go 1.26+
- **Node.js** (optional): Node 22+
- Pi Coding Agent sessions in `~/.pi/agent/sessions/`

## License

MIT
