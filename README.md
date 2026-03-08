# linear-cli

Command-line interface for the Linear project management API.

## Installation

### From source

Requirements: Go 1.25+

```
git clone https://github.com/your-org/linear-cli
cd linear-cli
make build
```

The binary is placed at `./linear-cli`. Move it to a directory in your PATH:

```
mv linear-cli /usr/local/bin/linear
```

## Authentication

linear-cli requires a Linear personal API key.

Generate one at: Linear Settings -> API -> Personal API keys

### Save key to config

```
linear auth
```

Prompts for your API key and saves it to the config file.

### Check auth status

```
linear auth status
```

Shows whether an API key is configured (masked).

## Configuration

Config file location: `~/.config/linear-cli/config.yaml`

Format:

```yaml
api_key: lin_api_xxxx
default_team: ""
```

### Environment variable override

Set `LINEAR_API_KEY` to override the config file value without modifying it:

```
export LINEAR_API_KEY=lin_api_xxxx
linear auth status
```

The environment variable takes precedence over the config file.

## Usage

```
linear [command] [flags]

Flags:
  --json        Output in JSON format
  --version     Show version
  -h, --help    Show help
```

## Issue Commands

Manage Linear issues with the `issue` subcommand.

### List issues

```
linear issue list [flags]
```

Flags:
```
  --team string        filter by team key (e.g. ENG)
  --assignee string    filter by assignee display name
  --state string       filter by workflow state name
  --priority int       filter by priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low
  --limit int          maximum number of issues to return (default 50)
  --include-archived   include archived issues
  --order-by string    sort order: createdAt|updatedAt (default "updatedAt")
  --json               output as JSON array
```

Output columns: ID, Title, Status, Priority, Assignee

### Show issue

```
linear issue show <identifier> [flags]
```

Displays all fields for an issue: identifier, title, status, priority, team, assignee, due date, estimate, labels, URL, created/updated timestamps, and description.

Flags:
```
  --json    output as JSON object
```

Example:
```
linear issue show ENG-42
```

### Create issue

```
linear issue create --title <title> --team <team> [flags]
```

Flags:
```
  --title string         issue title (required)
  --team string          team key or ID (required)
  --description string   issue description in markdown
  --assignee string      assignee name, email, UUID, or "me"
  --state string         workflow state name or ID
  --priority int         priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low
  --label stringArray    label name or ID (repeatable)
  --due-date string      due date (YYYY-MM-DD)
  --estimate int         complexity estimate (integer)
  --cycle string         cycle ID
  --project string       project name or ID
  --parent string        parent issue identifier or ID
  --json                 output created issue as JSON
```

Example:
```
linear issue create --title "Fix login bug" --team ENG --priority 1 --assignee me
```

### Update issue

```
linear issue update <identifier> [flags]
```

Only flags explicitly provided are sent to the API - omitted flags leave fields unchanged.

Flags:
```
  --title string            issue title
  --description string      issue description in markdown
  --assignee string         assignee name, email, UUID, or "me"
  --state string            workflow state name or ID
  --priority int            priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low
  --label stringArray       set labels, replacing all existing (repeatable)
  --add-label stringArray   add label by name or ID (repeatable)
  --remove-label stringArray remove label by name or ID (repeatable)
  --due-date string         due date (YYYY-MM-DD)
  --estimate int            complexity estimate (integer)
  --cycle string            cycle ID
  --project string          project name or ID
  --parent string           parent issue identifier or ID
  --json                    output updated issue as JSON
```

Example:
```
linear issue update ENG-42 --state Done --assignee me
```

### Delete issue

```
linear issue delete <identifier> [flags]
```

By default, moves the issue to trash (30-day grace period for recovery). Use `--archive` to archive instead.

Flags:
```
  --archive   archive the issue instead of trashing it
  --yes       skip confirmation prompt
```

Example:
```
linear issue delete ENG-42 --yes
linear issue delete ENG-42 --archive
```

## Me Command

Show the currently authenticated user.

```
linear me [flags]
```

Flags:
```
  --assigned   show issues assigned to the current user
  --created    show issues created by the current user
  --json       output as JSON
```

Examples:
```
linear me
linear me --assigned
linear me --created
linear me --json
```

Output (default): Name, Email, Role, Active status, and team memberships.
Output (--assigned/--created): table of issues with columns ID, Title, Status.

## Team Commands

Manage Linear teams with the `team` subcommand.

### List teams

```
linear team list [flags]
```

Flags:
```
  --json   output as JSON array
```

Output columns: Name, Key, Description, Cycles

### Show team

```
linear team show <id|key> [flags]
```

Accepts a team UUID or team key (e.g. ENG).

Flags:
```
  --json   output as JSON object
```

Output fields: Name, Key, Cycles, Estimation, Description, Created, Updated.

Example:
```
linear team show ENG
```

## Project Commands

Manage Linear projects with the `project` subcommand.

### List projects

```
linear project list [flags]
```

Flags:
```
  --team string      filter by team key (e.g. ENG)
  --status string    filter by status (backlog|planned|started|paused|completed|canceled)
  --health string    filter by health (onTrack|atRisk|offTrack)
  --limit int        maximum number of projects to return (default 50)
  --include-archived include archived projects
  --order-by string  sort order: createdAt|updatedAt (default "updatedAt")
  --json             output as JSON array
```

Output columns: Name | Status | Health | Progress% | Target Date

### Show project

```
linear project show <id> [flags]
```

Flags:
```
  --json   output as JSON object
```

Output fields: Name, Status, Health, Progress, Teams, Creator, Start Date, Target Date, URL, Created, Updated, and description.

### Create project

```
linear project create --name <name> --team <team> [flags]
```

Flags:
```
  --name string         project name (required)
  --team stringArray    team key or ID (repeatable, required)
  --description string  project description
  --color string        project color (hex)
  --target-date string  target date (YYYY-MM-DD)
  --start-date string   start date (YYYY-MM-DD)
  --json                output created project as JSON
```

Example:
```
linear project create --name "Q2 Launch" --team ENG --target-date 2026-06-30
```

### Update project

```
linear project update <id> [flags]
```

Only flags explicitly provided are sent to the API - omitted flags leave fields unchanged.

Flags:
```
  --name string         project name
  --description string  project description
  --state string        project state (backlog|planned|started|paused|completed|canceled)
  --target-date string  target date (YYYY-MM-DD)
  --start-date string   start date (YYYY-MM-DD)
  --health string       project health (onTrack|atRisk|offTrack)
  --json                output updated project as JSON
```

Example:
```
linear project update abc123 --state started --health onTrack
```

### Delete project

```
linear project delete <id> [flags]
```

Flags:
```
  --yes   skip confirmation prompt
```

Example:
```
linear project delete abc123 --yes
```

## Cycle Commands

Manage Linear cycles with the `cycle` subcommand.

### List cycles

```
linear cycle list [flags]
```

Flags:
```
  --team string      filter by team key (e.g. ENG)
  --limit int        maximum number of cycles to return (default 50)
  --include-archived include archived cycles
  --order-by string  sort order: createdAt|updatedAt (default "updatedAt")
  --json             output as JSON array
```

Output columns: # | Name | Start | End | Progress% | Status (Active/Past/Future)

### Show cycle

```
linear cycle show <id> [flags]
```

Flags:
```
  --json   output as JSON object
```

Output fields: Number, Name, Status, Progress, Start, End, Team, and description.

### Active cycle

```
linear cycle active --team <key> [flags]
```

Shows the currently active cycle for a team.

Flags:
```
  --team string   team key (required)
  --json          output as JSON object
```

Example:
```
linear cycle active --team ENG
```

## User Commands

Manage Linear users with the `user` subcommand.

### List users

```
linear user list [flags]
```

Flags:
```
  --include-guests     include guest users (excluded by default)
  --include-disabled   include disabled/deactivated users
  --json               output as JSON array
```

Output columns: Name, Email, Role, Active

### Show user

```
linear user show <id> [flags]
```

Accepts a user UUID.

Flags:
```
  --json   output as JSON object
```

Output fields: Name, Email, Role, Active, Created, Updated.

Example:
```
linear user show abc123de-f456-7890-abcd-ef1234567890
```
