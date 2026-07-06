# Pi Visual — Session Viewer for Pi Coding Agent

Web-based session viewer for [Pi Coding Agent](https://github.com/earendil-works/pi-mono). Browse, search, and read your coding sessions in a clean chat-style interface.

## Features

- **Chat-style interface** — conversations with user/assistant bubbles, thinking blocks, tool outputs
- **Smart compact mode** — consecutive tool calls and thinking are auto-grouped into expandable blocks
- **Session browser** — sidebar with search, project select, and time-grouped history
- **Pin sessions** — star important sessions (persisted in `localStorage`)
- **Markdown rendering** — code blocks, headings, lists, tables, blockquotes, inline code
- **Diff highlighting** — auto-detected and colorized (green/red/purple/blue)
- **ANSI stripping** — terminal escape codes cleaned from tool output
- **Dashboard** — total sessions, tokens, cost, most-used model, recent sessions
- **Copy session IDs** — click `pi --session <id>` in the header to copy

## Quick Start

```bash
mise run dev        # builds frontend + backend, starts server
# → http://localhost:8765
```

Or manually:

```bash
cd frontend && npm run build
cd ../backend && go build -o pivisual-server . && ./pivisual-server
```

Custom port via `.env`:

```env
PORT=3000
```

Or environment variable:

```bash
PORT=3000 ./pivisual-server
```

## Development

```bash
# Frontend hot reload (proxies /api to backend)
cd frontend && npm run dev      # → :5173

# Backend tests
mise run test                   # 12 tests, 0 deps
go test -v -count=1 ./backend/...

# Production build
mise run prod                   # CGO=0, stripped binary
```

## Docker

```bash
docker compose up --build -d
```

Volume mounts `~/.pi/agent/sessions` read-only.

## Architecture

```
pivisual/
├── server.py                   # Python stdlib fallback (0 deps)
├── backend/
│   ├── main.go                 # net/http API + static server (0 deps)
│   ├── main_test.go            # 12 tests against real session fixtures
│   ├── testdata/               # Real session files for testing
│   ├── go.mod / go.sum
│   └── pivisual-server         # Compiled binary (~8MB)
├── frontend/
│   ├── src/
│   │   ├── App.vue             # Root layout + sidebar toggle (Ctrl+B)
│   │   ├── components/
│   │   │   ├── AppSidebar.vue  # Session list, search, project filter, pins
│   │   │   ├── AppDashboard.vue # Stats + recent sessions
│   │   │   ├── ChatView.vue    # Message parsing + compact mode
│   │   │   └── ChatMessage.vue # Markdown, diffs, ANSI, tool grouping
│   │   ├── composables/
│   │   │   └── useSessions.ts  # State, API, pins, grouping
│   │   └── types.ts
│   ├── index.html              # Vite entry, design tokens, fonts
│   └── vite.config.ts
├── .github/workflows/test.yml  # CI: backend tests + frontend build
├── .mise.toml                  # Go version + tasks (test, build, dev, prod)
├── .env.example
├── Dockerfile                  # Multi-stage: Node → Go → Alpine
└── docker-compose.yml
```

### Stack

| Layer | Tech | Dependencies |
|-------|------|-------------|
| Backend | Go `net/http` | 0 |
| Frontend | Vue 3 + TypeScript + Vite | `marked`, `@phosphor-icons/vue` |
| Toolchain | mise | Go, Node |
| CI | GitHub Actions | `jdx/mise-action` |

### API

| Endpoint | Description |
|----------|-------------|
| `GET /api/sessions` | All sessions with summary stats (ID, project, tokens, cost, model) |
| `GET /api/session?file=<path>` | Full JSONL lines for a session |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8765` | HTTP port |
| `HOME` | `$HOME` | Base path for `~/.pi/agent/sessions/` |

`.env` is read from the binary directory, parent directory, or CWD.

## Testing

```bash
mise run test
```

12 tests covering:
- Session file parsing (real fixtures)
- Missing files, empty files, malformed JSON
- User text extraction, session info names
- Project path decoding
- Directory scanning

CI runs on every push to `main` via GitHub Actions.

## Design System

Anthropic Claude-inspired:
- Cream canvas `#faf9f5` · Coral primary `#cc785c`
- Serif display (Georgia) · Inter body · JetBrains Mono code
- Dark surfaces `#181715` for code blocks and sidebar

## Session Resume

```bash
pi --session <id>     # Resume session
pi --fork <id>        # Fork into new session
```

## License

MIT
