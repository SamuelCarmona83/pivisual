# Pi Visual — Session Viewer for Pi Coding Agent

A lightweight web-based session viewer for [Pi Coding Agent](https://github.com/earendil-works/pi-mono) that lets you browse, search, and read your coding sessions in a clean chat-style interface.

![Pi Visual](https://img.shields.io/badge/status-stable-success)

## Features

- **Chat-style interface** — Read sessions as conversations with user/assistant bubbles, thinking blocks, and collapsible tool outputs
- **Session browser** — Sidebar with search, project filter, and time-grouped history (Today, Yesterday, This Week, Earlier)
- **Pin sessions** — Star important sessions for quick access; pins persist in `localStorage`
- **Markdown rendering** — Code blocks with dark terminal styling, headings, lists, tables, blockquotes, and inline code
- **Session metadata** — Token counts, cost breakdown, model/provider info at a glance
- **Copy session IDs** — Click to select and use with `pi --session <id>` to resume any session
- **Zero dependencies** — Python stdlib server + vanilla HTML/CSS/JS; no npm, no build step

## Quick Start

```bash
python3 server.py
# Open http://localhost:8765
```

Custom port:

```bash
python3 server.py 3000
```

## Docker

```bash
docker compose up
```

Or manually:

```bash
docker build -t pivisual .
docker run -p 8765:8765 -v ~/.pi/agent/sessions:/root/.pi/agent/sessions:ro pivisual
```

The volume mount gives the container read-only access to your Pi session files.

## Architecture

```
pivisual/
├── server.py       # Python stdlib HTTP server + /api/sessions & /api/session endpoints
├── index.html      # Single-page app: sidebar + chat viewer
├── Dockerfile      # Python 3.12 slim image
├── docker-compose.yml
└── README.md
```

### API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/sessions` | Returns all sessions as JSON with summary stats (ID, project, timestamp, tokens, cost, model) |
| `GET /api/session?file=<path>` | Returns full session JSONL lines for detail view |

### Design System

The UI follows [Anthropic Claude's design language](DESING.md):
- Warm cream canvas (`#faf9f5`) with serif display typography
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

- Python 3.9+
- Pi Coding Agent sessions in `~/.pi/agent/sessions/`

## License

MIT
