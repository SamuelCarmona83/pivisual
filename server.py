#!/usr/bin/env python3
"""Static server + /api/sessions endpoint for pi-sessions.html"""
import json
import os
from http.server import HTTPServer, SimpleHTTPRequestHandler
from urllib.parse import urlparse, parse_qs

SESSION_DIR = os.path.expanduser("~/.pi/agent/sessions")

def parse_session_dir():
    """Scan SESSION_DIR and return all sessions with summary info."""
    sessions = []
    if not os.path.isdir(SESSION_DIR):
        return sessions
    for project_dir in sorted(os.listdir(SESSION_DIR)):
        proj_path = os.path.join(SESSION_DIR, project_dir)
        if not os.path.isdir(proj_path):
            continue
        # Decode path: --<encoded>-- where - encodes /
        inner = project_dir.strip("-")  # remove leading/trailing --
        project = inner.replace("-", "/")
        # Fix home dir
        if project.startswith("Users/"):
            project = "~/" + project[6:]
        for fname in sorted(os.listdir(proj_path), reverse=True):
            if not fname.endswith(".jsonl"):
                continue
            fpath = os.path.join(proj_path, fname)
            try:
                with open(fpath) as f:
                    lines = [line for line in f if line.strip()]
                if not lines:
                    continue
                header = json.loads(lines[0])
            except Exception:
                continue
            if header.get("type") != "session":
                continue

            # Gather stats from messages
            user_msgs = 0
            assistant_msgs = 0
            total_input = 0
            total_output = 0
            total_cost = 0
            first_user = None
            model = "?"
            provider = "?"
            session_name = None

            for line in lines[1:]:
                try:
                    entry = json.loads(line)
                except Exception:
                    continue
                t = entry.get("type")

                if t == "session_info" and entry.get("name"):
                    session_name = entry["name"]

                if t != "message":
                    continue
                msg = entry.get("message", {})
                role = msg.get("role")
                if role == "user":
                    user_msgs += 1
                    content = msg.get("content", "")
                    if isinstance(content, list):
                        text = " ".join(c.get("text","") for c in content if c.get("type")=="text")
                    else:
                        text = str(content)
                    if first_user is None:
                        first_user = text
                elif role == "assistant":
                    assistant_msgs += 1
                    model = msg.get("model", model)
                    provider = msg.get("provider", provider)
                    usage = msg.get("usage", {})
                    total_input += usage.get("input", 0)
                    total_output += usage.get("output", 0)
                    total_cost += usage.get("cost", {}).get("total", 0)

            sessions.append({
                "id": header.get("id", fname),
                "file": fpath,
                "project": project,
                "timestamp": header.get("timestamp", ""),
                "name": session_name,
                "first_user": (first_user or "(empty)")[:200],
                "model": model,
                "provider": provider,
                "user_msgs": user_msgs,
                "assistant_msgs": assistant_msgs,
                "total_tokens": total_input + total_output,
                "total_cost": round(total_cost, 6),
            })
    return sessions


class Handler(SimpleHTTPRequestHandler):
    def do_GET(self):
        parsed = urlparse(self.path)
        path = parsed.path
        if path == "/api/sessions":
            data = parse_session_dir()
            self.send_response(200)
            self.send_header("Content-Type", "application/json; charset=utf-8")
            self.end_headers()
            self.wfile.write(json.dumps(data, indent=2).encode())
        elif path == "/api/session":
            qs = parse_qs(parsed.query)
            fpath = qs.get("file", [None])[0]
            if fpath and os.path.isfile(fpath) and fpath.startswith(SESSION_DIR):
                with open(fpath) as f:
                    lines = [line for line in f if line.strip()]
                self.send_response(200)
                self.send_header("Content-Type", "application/json; charset=utf-8")
                self.end_headers()
                self.wfile.write(json.dumps(lines).encode())
            else:
                self.send_response(404)
                self.end_headers()
        else:
            super().do_GET()

if __name__ == "__main__":
    import sys
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8765
    print(f"Serving at http://localhost:{port}")
    HTTPServer(("0.0.0.0", port), Handler).serve_forever()
