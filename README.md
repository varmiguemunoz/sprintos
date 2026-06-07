<div align="center">

# ⚡ SprintOS

### The terminal-first project manager for developer teams

[![Release](https://img.shields.io/github/v/release/varmiguemunoz/sprintos)](https://github.com/varmiguemunoz/sprintos/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22-blue)](https://golang.org)

**A fast, keyboard-driven project manager that lives entirely in your terminal.**  
No browser. No Electron. No configuration. Just install and run.

[Installation](#-installation) · [Quick Start](#-quick-start) · [TUI Guide](#%EF%B8%8F-complete-tui-guide) · [Menu Bar](#-macos-menu-bar) · [CLI Reference](#-cli-reference) · [REST API](#-rest-api) · [AI & MCP](#-ai--mcp-tools)

</div>

---

## ✨ Why SprintOS?

Every PM tool eventually becomes the thing you manage instead of the thing that helps you manage. SprintOS is different:

- 🖥️ **Lives in the terminal** — no context switching, no browser tabs, no distractions
- 🍎 **macOS Menu Bar app** — timer and task access without opening a terminal
- 🍅 **Built-in Pomodoro** — focus sessions of 15, 30, or 45 minutes with auto-restart
- ⏱️ **Time tracking** — start/stop timers per task, log time manually, export reports
- 📊 **PDF reports** — executive-grade reports with charts, KPIs, and team performance
- 🤖 **AI-native** — your AI agent can create tasks, generate sprints, and review the board via MCP
- 🔗 **GitHub-connected** — PRs automatically move tasks between states
- ⚡ **Zero config** — one install command, login with GitHub, done
- 🌐 **REST API** — every feature accessible via HTTP for Zapier, Make, or custom scripts

---

## 📦 Installation

### 🍎 macOS — Homebrew

```bash
brew install varmiguemunoz/sprintos/sprintos
```

### 🪟 Windows — Scoop

```powershell
scoop bucket add sprintos https://github.com/varmiguemunoz/scoop-sprintos
scoop install sprintos
```

### 🐧 Linux — Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/varmiguemunoz/sprintos/main/install.sh | sh
```

### ✅ Verify Installation

```bash
sprintos --version
```

---

## 🚀 Quick Start

```bash
# Launch the interactive TUI
sprintos start
```

On first run, SprintOS walks you through a guided setup:

```
Step 1 of 3  ●○○  Login with GitHub
Step 2 of 3  ●●○  Create your organization
Step 3 of 3  ●●●  Set up your first board
```

After setup you land on the **Executive Dashboard** — a live view of all your projects, KPIs, and team health. Everything is keyboard-driven from here.

> 💡 **First-time setup tip:** SprintOS uses PostgreSQL. Point it to any hosted instance (Supabase, Neon, Railway) by setting the `DATABASE_URL` environment variable before running.

> 🔧 **After installing a new version:** run `sprintos start --migrate` once to apply schema updates. Normal launches skip migrations entirely for fast startup.

---

## 🖥️ Complete TUI Guide

SprintOS is entirely keyboard-driven. Every screen has a hint bar at the bottom showing available shortcuts. Press `?` on any screen for the full help overlay.

---

### 📊 Executive Dashboard

The first screen you see after login. It shows a live snapshot of your entire organization's health across all projects.

**What you see:**
- 🏆 **Sprint completion rate** — percentage of current sprint tasks done
- ⏱️ **Avg cycle time** — mean days from task creation to completion
- 📈 **Weekly throughput** — tasks completed in the last 7 days
- 🎯 **On-time delivery rate** — % of tasks completed before their due date
- 🔴 **Overdue count** — tasks past their due date
- 📉 **Velocity trend** — bar chart of the last 4 sprints
- 🗂️ **State distribution** — how tasks are spread across board columns
- 👥 **Team workload** — assigned vs completed per team member
- ✅ **Recently completed** — last 8 finished tasks

| Key | Action |
|-----|--------|
| `f` | Filter metrics by project |
| `r` | Manual refresh |
| `p` | Go to projects list |
| `q` | Quit |

> 💡 The dashboard auto-refreshes every **10 minutes** in the background. Use `r` to refresh on demand.

---

### 📋 Projects List

Your organization's project directory.

| Key | Action |
|-----|--------|
| `↑ / ↓` or `j / k` | Navigate projects |
| `enter` | Open project kanban board |
| `d` | Go to Executive Dashboard |
| `n` | Create new project |
| `e` | Edit selected project (name, description) |
| `D` | Delete project (confirms with `y / n`) |
| `/` | Fuzzy search across all tasks |
| `s` | Organization settings |
| `?` | Toggle keyboard help |
| `q` | Quit |

---

### 📌 Kanban Board

The main workspace. Tasks are organized in columns by state (Backlog → In Progress → In Review → Done, or any custom workflow).

**Reading the board:**

| Symbol | Meaning |
|--------|---------|
| `!!` | Critical priority |
| `↑` | High priority |
| `✗` (red) | Task is overdue |
| `⚠` (yellow) | Due within 48 hours |
| `MA` | Assignee initials |

| Key | Action |
|-----|--------|
| `← / →` or `h / l` | Move between columns |
| `↑ / ↓` or `j / k` | Move between tasks in the current column |
| `enter` | Open task detail |
| `n` or `+` | Create new task in the current column |
| `m` | Move selected task to a different state (inline dialog) |
| `d` | Delete selected task (confirms with `y / n`) |
| `R` | **Reorder columns** — enter column drag mode |
| `v` | Switch to Sprint View |
| `b` | Edit board layout (add, remove, rename states) |
| `E` | Export PDF executive report |
| `/` | Fuzzy search tasks |
| `?` | Toggle keyboard help |
| `esc` | Back to Projects list |
| `q` | Quit |

#### 🔀 Column Reorder Mode

Press `R` to enter Reorder Mode. The selected column turns **amber** and you can drag it freely:

| Key | Action |
|-----|--------|
| `←` / `h` | Move column left |
| `→` / `l` | Move column right |
| `enter` | Save the new order to the database |
| `esc` | Cancel — discard changes |

> 💡 Column positions are persisted immediately after you press `enter`. Other team members will see the new order on their next board load.

---

### 🗒️ Task Detail

A full view of a single task with all metadata, time tracking, subtasks, and comments.

| Key | Action |
|-----|--------|
| `a` | Assign / reassign / unassign a team member |
| `c` | Add a comment |
| `e` | Edit task (title, description) |
| `m` | **Move task to a different state** (inline state picker) |
| `s` | Create a subtask |
| `T` | Start / stop time tracker for this task |
| `l` | Log time manually (enter minutes + optional note) |
| `↑ / ↓` or `j / k` | Navigate subtasks |
| `enter` | Open selected subtask detail |
| `d` | Delete selected subtask |
| `esc` | Back to Kanban board |
| `q` | Quit |

**Details shown:**
- State, assignee, priority, due date, start date
- Description (full multi-line text)
- Time tracked total + live running timer
- Subtask list with completion progress
- Comment thread with author and timestamp

> ⏱️ When a timer is running for this task, the **Task Detail** screen shows a live elapsed time counter (e.g., `● 01:23:45`) that updates every second.

---

### 📝 Create / Edit Task

Multi-field form for creating or editing a task.

| Field | Description |
|-------|-------------|
| **Title** | Required. Short, descriptive name for the task. |
| **Description** | Optional. Multi-line text area — press `enter` for new lines. |
| **Due Date** | Optional. Format: `YYYY-MM-DD` (e.g., `2026-06-30`). |

| Key | Action |
|-----|--------|
| `tab` | Move to next field |
| `shift+tab` | Move to previous field |
| `enter` | Advance to next field (or submit on Due Date field) |
| `ctrl+s` | Save from any field |
| `esc` | Cancel and go back |

---

### 🔍 Search

Global fuzzy search across all tasks in your organization.

| Key | Action |
|-----|--------|
| Type | Filter tasks in real time |
| `↑ / ↓` | Navigate results |
| `enter` | Open selected task detail |
| `esc` | Close search |

---

### 🏃 Sprint View

Plan, monitor, and manage all sprints for a project.

**Sprint list view:**

Each sprint shows: name, date range, task count, completion count, days remaining, and status badge (● active / upcoming / completed).

Selecting a sprint expands it to show all its tasks with completion checkmarks.

| Key | Action |
|-----|--------|
| `↑ / ↓` or `j / k` | Navigate sprints |
| `c` | Create a new sprint (opens form) |
| `e` | Edit selected sprint (name, goal, dates) |
| `D` | Delete selected sprint (confirms with `y / n`) |
| `p` | Enter planning mode |
| `esc` | Back to Kanban board |
| `q` | Quit |

**📋 Planning Mode** (`p`):

Move backlog tasks into the selected sprint.

| Key | Action |
|-----|--------|
| `↑ / ↓` or `j / k` | Navigate backlog tasks |
| `enter` | Add selected task to the sprint |
| `esc` | Exit planning mode |

> 💡 **Completing a sprint** via the CLI (`sprintos sprint complete`) automatically moves unfinished tasks back to Backlog.

---

### ✏️ Create / Edit Sprint

Form with 4 fields pre-populated when editing.

| Field | Description |
|-------|-------------|
| **Sprint name** | Required. E.g., "Sprint 3", "Q3 Auth Sprint". |
| **Goal** | Optional. One-line sprint objective. |
| **Start date** | Required. Format: `YYYY-MM-DD`. |
| **End date** | Required. Must be after start date. |

| Key | Action |
|-----|--------|
| `tab / ↓` | Next field |
| `shift+tab / ↑` | Previous field |
| `enter` | Advance / save on last field |
| `esc` | Cancel |

---

### 🧩 Subtask Detail

Full view of a subtask with its own comment thread and time tracking.

| Key | Action |
|-----|--------|
| `e` | Edit subtask (title, description) |
| `T` | Start / stop time tracker |
| `l` | Log time manually |
| `c` | Add comment |
| `esc` | Back to parent task |

---

### ⚙️ Organization Settings

Manage your organization profile, invite team members, and connect integrations.

| Key | Action |
|-----|--------|
| `tab / ↓` | Next field |
| `shift+tab / ↑` | Previous field |
| `enter` | Save changes |
| `i` | Invite a team member by email |
| `m` | Open MCP setup (connect AI tools) |
| `n` | Configure notification channels |
| `L` | Logout |
| `esc` | Back to dashboard |

---

### 🎯 Board Setup

Create, rename, reorder, and delete board columns (states) for a project. Choose from preset workflow templates or build your own.

**Preset templates:**
- **Standard** — Backlog → In Progress → In Review → Done
- **Simple** — Todo → Done

| Key | Action |
|-----|--------|
| `↑ / ↓` | Navigate states |
| `enter` | Select template / confirm |
| `esc` | Cancel |

---

### 📤 Export PDF Report

Press `E` from the Kanban board to open the 2-step export wizard.

**Step 1 — Project selection:**

| Key | Action |
|-----|--------|
| `↑ / ↓` | Navigate options |
| `space` | Toggle a project on/off |
| `enter` | Proceed to step 2 |
| `esc` | Cancel |

> Select "All projects" (pre-selected) or use `space` to pick individual projects.

**Step 2 — Time range:**

| Option | Description |
|--------|-------------|
| Last 7 days | Compact weekly snapshot |
| Last 30 days | Monthly view |
| Last 90 days | Quarterly review |
| Custom range | Enter `From` and `To` dates manually |

| Key | Action |
|-----|--------|
| `↑ / ↓` | Navigate options |
| `enter` | Generate PDF (or confirm custom dates) |
| `tab` | Switch between date fields (custom range) |
| `esc` | Cancel / go back |

**The generated PDF contains:**
1. 📄 **Cover** — org name, date range, generation timestamp
2. 📊 **Executive Summary** — 6 KPI numbers: tasks created, completed, on-time %, hours logged, avg cycle time, overdue count
3. 🏢 **Project Health table** — per-project: done, in-progress, backlog, overdue, hours, cycle time
4. 📈 **Weekly Velocity chart** — bar chart of tasks completed per week
5. 👥 **Team Performance table** — per member: tasks completed + hours logged
6. 🔴 **Priority Risk** — critical/high open tasks with age, and overdue task list

> 💾 The PDF is automatically saved to `~/Desktop/sprintos-report-YYYY-MM-DD.pdf`

---

## 🍎 macOS Menu Bar

SprintOS runs a lightweight menu bar app on macOS that lets you track time and monitor your active timer without opening a terminal.

### How to launch

The menu bar app starts automatically with `sprintos start`. It appears as the ⚡ icon in your menu bar.

### Timer section

| Action | Description |
|--------|-------------|
| **Select Task ▶** | Browse all non-completed tasks across all projects. Navigate pages with Previous / Next. |
| **▶ Start Timer** | Start the time tracker for the selected task. A macOS notification confirms it started. |
| **■ Stop Timer** | Stop the running timer. Time is saved automatically. |
| **Status label** | Shows elapsed time (e.g., `Active: 01:23 — Fix login bug`) while a timer runs. |

> 🔒 **One timer at a time:** If a timer is already running (from the TUI or menu bar), starting another is blocked with a notification: _"A timer is already running. Stop it first."_

> 🔄 **Real-time sync:** If you start a timer from the TUI terminal, the menu bar automatically detects it within 1 second, enables the Stop button, and shows the task name — no manual action needed.

> 🔔 **macOS notifications:** A notification fires whenever a timer starts, regardless of whether it was started from the TUI or the menu bar.

### Pomodoro section 🍅

| Action | Description |
|--------|-------------|
| **Start 15 minutes** | Start a 15-minute focus session |
| **Start 30 minutes** | Start a 30-minute focus session |
| **Start 45 minutes** | Start a 45-minute focus session |
| **■ Stop Pomodoro** | Cancel the current focus session |

**How Pomodoro works:**
1. Choose a session length. The menu bar title shows a live countdown (e.g., `🍅 24:35`).
2. When time runs out, a macOS notification fires: _"Time's up! Take a break."_
3. A **15-second grace period** starts — shown as `⚠ 00:12` in the menu bar.
4. If you don't press Stop within the grace period, the session **auto-restarts** with another notification.
5. Press **■ Stop Pomodoro** at any time to end the session.

> 💡 The Pomodoro timer and the task timer are **independent** — you can track time on a task AND run a Pomodoro session simultaneously.

---

## ⏱️ Time Tracking

SprintOS has built-in time tracking per task and subtask.

### Starting a timer

**From Task Detail (`T` key):**
- Press `T` on a task that has no subtasks to toggle the timer on/off.
- The timer shows as `● 01:23:45` in the Time Tracked section, updating live every second.
- A macOS notification fires when the timer starts.

**From the macOS Menu Bar:**
- Select a task from the "Select Task ▶" submenu.
- Click "▶ Start Timer".

### Stopping a timer

- Press `T` again in Task Detail to stop and save.
- Click "■ Stop Timer" in the menu bar.
- Time is automatically logged in minutes (rounded up to the nearest minute, minimum 1 minute).

### Logging time manually

Press `l` in Task Detail (or Subtask Detail) to open the manual log form:

| Field | Description |
|-------|-------------|
| **Minutes** | Time to log, e.g., `90` for 1h 30m |
| **Note** | Optional description of what was done |

---

## 📚 CLI Reference

### Core Commands

```bash
sprintos start              # Launch the interactive TUI + macOS menu bar
sprintos start --migrate    # Run DB migrations, then launch (use after updates)
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

# List tasks (reads .sprintos for project ID if no --project flag)
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
  --start 2026-06-01 \
  --end 2026-06-14

# Create with a goal
sprintos sprint create \
  --name "Sprint 2" \
  --project 1 \
  --start 2026-06-15 \
  --end 2026-06-28 \
  --goal "Ship the auth module"

# List sprints for a project
sprintos sprint list --project 1

# Assign a task to a sprint
sprintos sprint assign --sprint 1 --task 5

# View velocity and progress
sprintos sprint velocity --id 1

# Complete a sprint (unfinished tasks move back to backlog)
sprintos sprint complete --id 1 --backlog-state 1

# Take a daily burndown snapshot (recommended: add to cron)
sprintos sprint snapshot
sprintos sprint snapshot --project 1
sprintos sprint snapshot --id 1
```

**Recommended cron job** (saves a burndown snapshot every night at 11pm):
```bash
0 23 * * * /usr/local/bin/sprintos sprint snapshot
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

> 🖨️ For a full PDF executive report with charts and KPIs, use the interactive `E` shortcut from the Kanban board inside the TUI.

---

### 📅 Standup & Review

```bash
# Generate today's standup update
sprintos standup

# Run a board health review
sprintos review
sprintos review --days 5             # tasks stale for 5+ days
sprintos review --days 3 --notify   # also send a notification digest
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

### 👤 My Tasks

```bash
# Show all tasks assigned to you across every project
sprintos my-tasks
```

---

### 🔔 Notifications

```bash
# Configure a Slack or Discord channel
sprintos notify config

# List configured channels
sprintos notify list

# Send a test notification to all channels
sprintos notify test
```

**Events that trigger notifications:**

| Event | Description |
|-------|-------------|
| `task.created` | A new task was added |
| `task.moved` | A task changed state |
| `task.completed` | A task reached a Done state |
| `task.assigned` | Someone was assigned to a task |
| `comment.created` | A comment was added to a task |
| `@username` in comment | Direct email notification to that user |

---

### 🔗 GitHub Integration

Connect a GitHub repository to automatically move tasks when PRs are opened or merged.

```bash
# Connect a GitHub repo to a SprintOS project (one-time setup)
sprintos github setup

# List all connected repos
sprintos github list

# Start the webhook server (receives GitHub events)
sprintos serve --port 8090

# For local testing with ngrok
ngrok http 8090
# → use the ngrok HTTPS URL as your GitHub webhook URL
```

**Webhook event mapping:**

| GitHub Event | SprintOS Action |
|---|---|
| PR opened | Task moves to **"In Review"** |
| PR merged | Task moves to **"Done"** |
| PR closed (no merge) | No change |

**Naming convention** — include the task ID in your PR title or branch name:
```
feat: TSK-42 Add rate limiting
feature/TSK-42-add-rate-limiting
fix/TSK-7-fix-auth-bug
```

---

### 🌐 API Server

```bash
# Start the REST API + webhook server
sprintos serve
sprintos serve --port 8090

# Available after starting:
# REST API        →  http://localhost:8090/api
# GitHub webhook  →  http://localhost:8090/webhooks/github
# Health check    →  http://localhost:8090/api/health
# OpenAPI docs    →  http://localhost:8090/api/docs
```

---

### 🔑 API Keys

```bash
# Generate an API key
sprintos api-key create --name "my-ci-key"
sprintos api-key create --name "zapier"

# List all active keys
sprintos api-key list

# Revoke a key
sprintos api-key revoke --id 3
```

---

### 📁 Repository Init

```bash
# Initialize SprintOS in your current repository
cd ~/myproject
sprintos init
# → creates a .sprintos file with your project ID

# After init, no --project flag needed anywhere
sprintos task ls
sprintos task create "Fix auth"
```

The `.sprintos` file is searched upward from the current directory, so it works from any subdirectory.

```json
{
  "project_id": 1,
  "project_name": "TaoFlow",
  "org_id": 1
}
```

---

### 🤝 Team Invitations

```bash
# From the TUI: press 'i' on the Organization Settings screen
# From the CLI: send an invitation email

# Accept an invitation (when you receive an invite link)
sprintos join --token abc123def456...
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

# Use in all requests
curl -H "Authorization: Bearer sk_abc123..." \
     http://localhost:8090/api/projects
```

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/health` | Health check — no auth required |
| `GET` | `/api/docs` | OpenAPI specification — no auth required |
| `GET` | `/api/projects` | List all projects |
| `POST` | `/api/projects` | Create a project |
| `GET` | `/api/tasks?project_id=1` | List tasks |
| `GET` | `/api/tasks?project_id=1&state_id=2` | Filter by state |
| `GET` | `/api/tasks?project_id=1&assignee_id=3` | Filter by assignee |
| `POST` | `/api/tasks` | Create a task |
| `GET` | `/api/tasks/:id` | Get full task detail |
| `PATCH` | `/api/tasks/:id` | Update a task |
| `DELETE` | `/api/tasks/:id` | Delete a task |
| `POST` | `/api/tasks/:id/move` | Move task to a different state |
| `GET` | `/api/states?project_id=1` | List board states for a project |
| `GET` | `/api/members` | List organization members |
| `GET` | `/api/webhooks` | List outbound webhooks |
| `POST` | `/api/webhooks` | Register an outbound webhook |
| `DELETE` | `/api/webhooks/:id` | Delete an outbound webhook |

### Examples

```bash
BASE="http://localhost:8090"
TOKEN="sk_your_key_here"

# List all projects
curl -H "Authorization: Bearer $TOKEN" $BASE/api/projects

# Create a task
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Fix bug","project_id":1,"state_id":1,"priority":"high"}' \
  $BASE/api/tasks

# Move a task to a different state
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

# Export tasks as JSON and filter with jq
curl -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/tasks?project_id=1" | jq '.[].title'
```

---

## 🤖 AI & MCP Tools

SprintOS exposes an MCP (Model Context Protocol) server that any AI agent — Claude, GPT, Cursor, Windsurf — can connect to and control directly.

### Setup

```bash
# Start the MCP server
sprintos mcp

# Or configure automatically from inside the TUI:
# Organization Settings → press 'm' → select your AI tool → install
```

The MCP setup screen auto-detects and writes the correct config for:
- **Claude Desktop** — `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Cursor** — `~/.cursor/mcp.json`
- **Windsurf** — `~/.codeium/windsurf/mcp_config.json`
- **Zed** — `~/.config/zed/settings.json`

Restart your AI tool after setup to activate.

### Available MCP Tools

| Tool | Description |
|------|-------------|
| `list_projects` | List all projects in the organization |
| `create_project` | Create a project with states from a template |
| `list_states` | List board columns for a project |
| `list_tasks` | List all tasks with IDs, states, and assignees |
| `get_task_detail` | Full task detail including comments and subtasks |
| `create_task` | Create a task with title, state, priority, and due date |
| `update_task` | Edit title, description, or due date |
| `delete_task` | Permanently delete a task |
| `move_task` | Move a task to a different state |
| `assign_task` | Assign or unassign a team member |
| `add_comment` | Add a comment to a task |
| `list_members` | List all organization members |
| `list_overdue_tasks` | All tasks past their due date |
| `analyze_stale_tasks` | Find tasks stuck in a state with suggested actions |
| `summarize_project` | Project health: counts, overdue, workload |
| `generate_sprint` | Create multiple tasks from a structured list |
| `list_organizations` | List all organizations the user belongs to |

### Example AI Prompts

Once connected, your AI agent can manage your entire board through natural language:

```
"Generate a sprint for the TaoFlow project based on this PRD: [paste PRD]"

"Which tasks in project 1 have been in the same state for more than 5 days?"

"Create 5 tasks for the auth module and put them all in Backlog"

"Show me the health summary for project 2"

"Move task #15 to In Review and assign it to Miguel"

"What's the velocity trend for the last 3 sprints?"
```

---

## ⚙️ Configuration

### Environment Variables

Create a `.env` file in your working directory (or export these variables in your shell):

```env
DATABASE_URL=postgres://user:password@host:5432/sprintos?sslmode=require
GITHUB_CLIENT_ID=your_github_oauth_client_id
GITHUB_CLIENT_SECRET=your_github_oauth_client_secret
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_FROM=you@gmail.com
SMTP_PASSWORD=your_app_password
EVOLUTION_API_URL=https://your-evolution-api-instance.com
EVOLUTION_API_TOKEN=your_evolution_api_token
```

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | ✅ Yes | PostgreSQL connection string |
| `GITHUB_CLIENT_ID` | ✅ Yes | GitHub OAuth app client ID (for login) |
| `GITHUB_CLIENT_SECRET` | ✅ Yes | GitHub OAuth app client secret |
| `SMTP_HOST` | Optional | SMTP server for email notifications |
| `SMTP_PORT` | Optional | SMTP port (usually `587`) |
| `SMTP_FROM` | Optional | From address for notification emails |
| `SMTP_PASSWORD` | Optional | SMTP password or app password |
| `EVOLUTION_API_URL` | Optional | WhatsApp Evolution API URL |
| `EVOLUTION_API_TOKEN` | Optional | WhatsApp Evolution API token |

---

## 🏗️ Self-Hosting

SprintOS uses **PostgreSQL** as its only infrastructure dependency. Any hosted PostgreSQL provider works:

| Provider | Notes |
|---|---|
| [Supabase](https://supabase.com) | Generous free tier, easy setup |
| [Neon](https://neon.tech) | Serverless PostgreSQL, scales to zero |
| [Railway](https://railway.app) | Simple deploy with one click |
| Your own server | Any PostgreSQL 13+ instance |

```bash
# Connection string format
DATABASE_URL=postgres://username:password@host:5432/dbname?sslmode=require
```

**First launch:** SprintOS detects a missing schema and automatically runs migrations.

**After updates:** Run once with the `--migrate` flag to apply schema changes:
```bash
sprintos start --migrate
```

All subsequent launches skip migrations and boot instantly.

---

## 🧩 Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22 |
| TUI | Bubble Tea + Lip Gloss + Bubbles |
| CLI | Cobra |
| Database | PostgreSQL + GORM |
| Auth | GitHub OAuth via Goth |
| Menu Bar | systray |
| Notifications | beeep (macOS native) |
| PDF Generation | maroto v2 + go-chart |
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
./bin/sprintos start --migrate  # first run: apply schema
./bin/sprintos start            # subsequent runs: fast boot
```

### Make Commands

```bash
make build       # Compile binary to ./bin/sprintos
make run         # go run main.go start (dev mode)
make fmt         # Format all Go files with gofmt
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
