<div align="center">

# ⚡ SprintOS

### The terminal-first project manager for developer teams

[![Release](https://img.shields.io/github/v/release/varmiguemunoz/sprintos)](https://github.com/varmiguemunoz/sprintos/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22-blue)](https://golang.org)

**A fast, keyboard-driven project manager that lives entirely in your terminal.**
No browser. No Electron. No configuration. Just install and run.

[Install](#-installation) · [Quick Start](#-quick-start) · [Commands](#-cli-reference) · [API](#-rest-api) · [MCP](#-ai--mcp-tools)

</div>

---

## ✨ Why SprintOS?

Every PM tool eventually becomes the thing you manage instead of the thing that helps you manage. SprintOS is different:

- 🖥️ **Lives in the terminal** — no context switching, no browser tabs
- 🤖 **AI-native** — your AI agent can create tasks, generate sprints, and review the board via MCP
- 🔗 **GitHub-connected** — PRs automatically move tasks between states
- ⚡ **Zero config** — one install command, login with GitHub, done
- 🌐 **REST API** — every feature is accessible via HTTP for Zapier, Make, or custom scripts

---

## 📦 Installation

### macOS

```bash
brew install varmiguemunoz/sprintos/sprintos
```

### Windows

```powershell
scoop bucket add sprintos https://github.com/varmiguemunoz/scoop-sprintos
scoop install sprintos
```

### Linux

```bash
curl -fsSL https://raw.githubusercontent.com/varmiguemunoz/sprintos/main/install.sh | sh
```

### Verify

```bash
sprintos --version
```

---

## 🚀 Quick Start

```bash
# Launch the interactive TUI
sprintos start
```

On first run, SprintOS walks you through a 3-step wizard:

```
Step 1 of 3  ●○○  Login with GitHub
Step 2 of 3  ●●○  Create your organization
Step 3 of 3  ●●●  Set up your first board
```

After setup, you land on the project dashboard — fully keyboard-driven.

---

## 🖥️ TUI — Keyboard Reference

### 📋 Project Dashboard

| Key | Action |
|-----|--------|
| `↑ / ↓` or `j / k` | Navigate projects |
| `enter` | Open project kanban board |
| `n` | Create new project |
| `e` | Edit selected project |
| `D` | Delete project (asks confirmation) |
| `/` | Fuzzy search across all tasks |
| `s` | Organization settings |
| `L` | Logout |
| `?` | Show keyboard shortcut help |
| `q` | Quit |

### 📌 Kanban Board

| Key | Action |
|-----|--------|
| `← / →` or `h / l` | Move between columns |
| `↑ / ↓` or `j / k` | Move between tasks |
| `enter` | View task detail |
| `n` or `+` | Create new task in current column |
| `m` | Move task to another state (inline picker) |
| `d` | Delete task (asks confirmation with `y/n`) |
| `v` | Switch to sprint view |
| `b` | Edit board layout (add/remove/rename states) |
| `/` | Fuzzy search tasks |
| `?` | Show keyboard shortcut help |
| `esc` | Back to projects |
| `q` | Quit |

> 💡 **Task IDs** like `#1`, `#2` are shown on every card. Use them in PR titles or branch names to trigger automatic state changes.

> 🔴 **Red `✗`** = task is overdue. **Yellow `⚠`** = due within 48 hours.
> **`↑`** = high priority. **`!!`** = critical priority.

### 🔍 Task Detail

| Key | Action |
|-----|--------|
| `a` | Assign / reassign / unassign user |
| `c` | Add a comment |
| `e` | Edit task (title, description) |
| `esc` | Back to kanban board |
| `q` | Quit |

### 🏃 Sprint View

| Key | Action |
|-----|--------|
| `↑ / ↓` | Navigate sprints |
| `p` | Enter planning mode (move tasks into sprint) |
| `esc` | Back to kanban |

### ⚙️ Organization Settings

| Key | Action |
|-----|--------|
| `tab / ↓` | Next field |
| `shift+tab / ↑` | Previous field |
| `enter` | Save changes |
| `i` | Invite a team member |
| `m` | MCP setup (connect to AI tools) |
| `L` | Logout |
| `esc` | Back to dashboard |

---

## 📚 CLI Reference

### Core

```bash
sprintos start              # Launch the interactive TUI
sprintos --help             # Show all available commands
sprintos --version          # Show current version
```

---

### 📋 Task Commands

```bash
# Create a task
sprintos task create "Fix login bug"
sprintos task create "Fix login bug" --state backlog
sprintos task create "Add rate limiting" --state backlog --priority high
sprintos task create "Deploy v2" --project 1 --state "In Progress"

# List tasks (reads .sprintos for project if no --project flag)
sprintos task ls
sprintos task ls --state backlog
sprintos task ls --state "In Review"
sprintos task ls --format json
sprintos task ls --format json | jq '.[].Title'

# Move a task to another state
sprintos task move 5 "In Review"
sprintos task move 5 done
sprintos task move 5 3          # also accepts state ID

# Show full task detail
sprintos task show 5
sprintos task show 5 --format json
```

**Priority values:** `low` · `medium` · `high` · `critical`

---

### 🏃 Sprint Commands

```bash
# Create a sprint
sprintos sprint create \
  --name "Sprint 1" \
  --project 1 \
  --start 2025-06-01 \
  --end 2025-06-14

# Create with a goal
sprintos sprint create \
  --name "Sprint 2" \
  --project 1 \
  --start 2025-06-15 \
  --end 2025-06-28 \
  --goal "Ship the auth module"

# List sprints for a project
sprintos sprint list --project 1

# Assign a task to a sprint
sprintos sprint assign --sprint 1 --task 5

# View velocity + progress
sprintos sprint velocity --id 1

# Complete a sprint (unfinished tasks move to backlog)
sprintos sprint complete --id 1 --backlog-state 1

# Take a daily burndown snapshot (add to cron)
sprintos sprint snapshot
sprintos sprint snapshot --project 1
sprintos sprint snapshot --id 1

# Cron example (every night at 11pm)
# 0 23 * * * /usr/local/bin/sprintos sprint snapshot
```

---

### 📊 Reporting & Export

```bash
# Full health report for all projects
sprintos report

# Report for a specific project
sprintos report --project 1

# Include completed tasks from last 30 days
sprintos report --project 1 --completed 30

# Export to CSV
sprintos export --project 1 --format csv --output tasks.csv

# Export to JSON
sprintos export --project 1 --format json --output tasks.json

# Pipe-friendly stdout export
sprintos export --format json | jq '.[].Title'
sprintos export --format csv > sprint-data.csv
```

The report shows: task counts by state, overdue tasks, unassigned tasks, average lead time (creation → done), average cycle time per state, and team workload chart.

---

### 📅 Standup & Review

```bash
# Generate today's standup update
sprintos standup

# Run a board health review
sprintos review
sprintos review --days 5             # tasks stale for 5+ days
sprintos review --days 3 --notify   # also send email digest

# Cron example (every morning at 9am)
# 0 9 * * * /usr/local/bin/sprintos review --notify
```

**Example standup output:**
```
── Standup — Monday Jun 2 ─────────────────

Yesterday:
  ✓ [TaoFlow] #4 Set up CI pipeline

Today:
  → [TaoFlow] #5 Write unit tests for auth module

Blockers:
  none
```

---

### 👥 My Tasks

```bash
# Show all tasks assigned to you across every project
sprintos my-tasks
```

---

### 🔔 Notifications

```bash
# Configure a Slack or Discord channel
sprintos notify config
# → enter: slack
# → enter: https://hooks.slack.com/services/T.../B.../xxx

# List configured channels
sprintos notify list

# Send a test notification to all channels
sprintos notify test
```

**Events that trigger notifications:**
- `task.created` — new task added
- `task.moved` — task changes state
- `task.completed` — task reaches a Done state
- `task.assigned` — someone gets assigned
- `comment.created` — comment added to a task
- `@username` in comment — direct email to that user

---

### 🔗 GitHub Integration

```bash
# Connect a GitHub repo to a SprintOS project (one-time setup)
sprintos github setup

# List all connected repos
sprintos github list

# Start the webhook server (receives GitHub events)
sprintos serve --port 8090

# For local testing with ngrok
ngrok http 8090
# → use the ngrok URL as GitHub webhook URL
```

**GitHub webhook events handled:**

| GitHub Event | SprintOS Action |
|---|---|
| PR opened | Task moves to "In Review" |
| PR merged | Task moves to "Done" |
| PR closed (no merge) | Task stays, no change |

**Naming convention:** Include the task ID in your PR title or branch name:
```
"feat: TSK-42 Add rate limiting"
feature/TSK-42-add-rate-limiting
```

---

### 🌐 API Server

```bash
# Start the REST API + webhook server
sprintos serve
sprintos serve --port 8090

# Generate an API key
sprintos api-key create --name "my-ci-key"
sprintos api-key create --name "zapier"

# List API keys
sprintos api-key list

# Revoke an API key
sprintos api-key revoke --id 3
```

---

### 📁 Repository Init

```bash
# Initialize SprintOS in your current repo
cd ~/myproject
sprintos init
# → creates .sprintos with your project ID

# After init, no --project flag needed in any command
sprintos task ls
sprintos task create "Fix auth"
```

The `.sprintos` file is searched from the current directory upward, so it works from any subdirectory of your project.

---

### 🤝 Invitations & Team

```bash
# Invite a teammate (from org settings in TUI, or via email)
# Press 'i' in org settings screen

# Accept an invitation
sprintos join --token abc123def456...
```

---

### 🔧 Server & Webhooks

```bash
# Start the full server (REST API + GitHub webhooks)
sprintos serve
sprintos serve --port 8090

# Endpoints available after starting:
# REST API:        http://localhost:8090/api
# GitHub webhook:  http://localhost:8090/webhooks/github
# Health check:    http://localhost:8090/api/health
# API docs:        http://localhost:8090/api/docs
```

---

### 🐚 Shell Completion

```bash
# Bash
sprintos completion bash > /usr/local/etc/bash_completion.d/sprintos

# Zsh
sprintos completion zsh > "${fpath[1]}/_sprintos"

# After setup, tab completion works everywhere:
sprintos <TAB>
sprintos task <TAB>
sprintos sprint <TAB>
```

---

## 🌐 REST API

All endpoints require `Authorization: Bearer <api-key>`.

Rate limit: **60 requests per minute** per API key.

### Authentication

```bash
# Generate a key
sprintos api-key create --name "my-key"

# Use in requests
curl -H "Authorization: Bearer sk_abc123..." \
     http://localhost:8090/api/projects
```

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/health` | Health check (no auth) |
| `GET` | `/api/docs` | OpenAPI spec (no auth) |
| `GET` | `/api/projects` | List all projects |
| `POST` | `/api/projects` | Create a project |
| `GET` | `/api/tasks?project_id=1` | List tasks |
| `GET` | `/api/tasks?project_id=1&state_id=2` | Filter by state |
| `GET` | `/api/tasks?project_id=1&assignee_id=3` | Filter by assignee |
| `POST` | `/api/tasks` | Create a task |
| `GET` | `/api/tasks/:id` | Get task detail |
| `PATCH` | `/api/tasks/:id` | Update a task |
| `DELETE` | `/api/tasks/:id` | Delete a task |
| `POST` | `/api/tasks/:id/move` | Move task to state |
| `GET` | `/api/states?project_id=1` | List states for project |
| `GET` | `/api/members` | List org members |
| `GET` | `/api/webhooks` | List outbound webhooks |
| `POST` | `/api/webhooks` | Register outbound webhook |
| `DELETE` | `/api/webhooks/:id` | Delete webhook |

### Examples

```bash
BASE=http://localhost:8090
TOKEN="sk_your_key_here"
AUTH="-H 'Authorization: Bearer $TOKEN'"

# List projects
curl -H "Authorization: Bearer $TOKEN" $BASE/api/projects

# Create a task
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Fix bug","project_id":1,"state_id":1}' \
  $BASE/api/tasks

# Move a task
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"state_id":3}' \
  $BASE/api/tasks/5/move

# Register a Zapier webhook
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://hooks.zapier.com/hooks/catch/xyz","events":["task.moved","task.created"]}' \
  $BASE/api/webhooks

# Export tasks as JSON via API
curl -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/tasks?project_id=1" | jq '.[].Title'
```

---

## 🤖 AI & MCP Tools

SprintOS exposes an MCP server that any AI agent (Claude, GPT, etc.) can connect to.

### Setup

```bash
# Start the MCP server
sprintos mcp

# Or configure it in your AI tool automatically
# In org settings → press 'm' → select your AI tool → install
```

### Available Tools

| Tool | Description |
|------|-------------|
| `list_projects` | List all projects in the organization |
| `create_project` | Create a project with states from a template |
| `list_states` | List board columns for a project |
| `list_tasks` | List all tasks including task IDs |
| `get_task_detail` | Full task detail with comments |
| `create_task` | Create a task with title, state, due date |
| `update_task` | Edit title, description, or due date |
| `delete_task` | Delete a task permanently |
| `move_task` | Move task to a different state |
| `assign_task` | Assign or unassign a team member |
| `add_comment` | Add a comment to a task |
| `list_members` | List all organization members |
| `list_overdue_tasks` | All tasks past their due date |
| `analyze_stale_tasks` | Find tasks stuck in a state with suggested actions |
| `summarize_project` | Project health: counts, overdue, workload |
| `generate_sprint` | Create multiple tasks from a JSON list |
| `list_organizations` | List orgs the current user belongs to |

### Example AI Prompts

Once connected, your AI agent can:

```
"Generate a sprint for the TaoFlow project based on this PRD: [paste PRD]"

"Which tasks in project 1 have been in the same state for more than 5 days?"

"Create 5 tasks for the auth module and put them in backlog"

"Show me the health summary for project 2"

"Move task #15 to In Review and assign it to Miguel"
```

### Configuring in Claude Desktop

The MCP setup screen (press `m` in org settings) auto-detects:
- **Claude Desktop** — `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Cursor** — `~/.cursor/mcp.json`
- **Windsurf** — `~/.codeium/windsurf/mcp_config.json`
- **Zed** — `~/.config/zed/settings.json`

It writes the correct config automatically. Restart your AI tool to activate.

---

## ⚙️ Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `GITHUB_CLIENT_ID` | GitHub OAuth app client ID |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth app client secret |
| `SMTP_HOST` | SMTP server hostname |
| `SMTP_PORT` | SMTP port (usually 587) |
| `SMTP_FROM` | From email address |
| `SMTP_PASSWORD` | SMTP password or app password |
| `EVOLUTION_API_URL` | WhatsApp Evolution API URL |
| `EVOLUTION_API_TOKEN` | WhatsApp Evolution API token |

Create a `.env` file in your project root:

```env
DATABASE_URL=postgres://user:password@host:5432/sprintos?sslmode=require
GITHUB_CLIENT_ID=your_client_id
GITHUB_CLIENT_SECRET=your_client_secret
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_FROM=you@gmail.com
SMTP_PASSWORD=your_app_password
```

### `.sprintos` Per-Repo Config

Run `sprintos init` in any repository to create a `.sprintos` file:

```json
{
  "project_id": 1,
  "project_name": "TaoFlow",
  "org_id": 1
}
```

Once this file exists, all task commands work without `--project` flags.

---

## 🏗️ Self-Hosting

SprintOS uses PostgreSQL as its database. You can use any hosted PostgreSQL provider:

- [Supabase](https://supabase.com) — generous free tier
- [Neon](https://neon.tech) — serverless PostgreSQL
- [Railway](https://railway.app) — easy deploy
- Your own server

```bash
# Connection string format
DATABASE_URL=postgres://username:password@host:5432/dbname?sslmode=require
```

The schema is created automatically on first run via GORM AutoMigrate.

---

## 🧩 Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22 |
| TUI | Bubble Tea + Lip Gloss + Bubbles |
| CLI | Cobra |
| Database | PostgreSQL + GORM |
| Auth | GitHub OAuth via Goth |
| MCP | mark3labs/mcp-go |
| Email | net/smtp |
| Config | Environment variables + ldflags |

---

## 🤝 Contributing

```bash
git clone https://github.com/varmiguemunoz/sprintos
cd sprintos
cp .env.example .env   # fill in your values
make build
go run main.go start
```

### Available Make Commands

```bash
make build       # Compile for local development
make fmt         # Format all Go files
make lint        # Run golangci-lint
make tidy        # Clean go.mod and go.sum
make check       # fmt + lint + build
```

---

## 📄 License

MIT © [Miguel Angel Muñoz](https://github.com/varmiguemunoz)

---

<div align="center">

**Built with ❤️ for developers who live in the terminal**

[⬆ Back to top](#-sprintos)

</div>
