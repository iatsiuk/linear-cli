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
  --team string             filter by team key (e.g. ENG)
  --assignee string         filter by assignee display name
  --state string            filter by workflow state name
  --priority int            filter by priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low
  --limit int               maximum number of issues to return (default 50)
  --include-archived        include archived issues
  --order-by string         sort order: createdAt|updatedAt (default "updatedAt")
  --json                    output as JSON array

  --created-after string    issues created after date
  --created-before string   issues created before date
  --updated-after string    issues updated after date
  --updated-before string   issues updated before date
  --due-after string        issues with due date after
  --due-before string       issues with due date before
  --completed-after string  issues completed after date
  --completed-before string issues completed before date
  --priority-gte int        minimum priority value
  --priority-lte int        maximum priority value
  --my                      only issues assigned to me
  --no-assignee             only issues with no assignee
  --no-project              only issues with no project
  --no-cycle                only issues with no cycle
  --or                      combine filters with OR logic (default is AND)
```

Output columns: ID, Title, Status, Priority, Assignee

#### Filter date syntax

Date flags accept:
```
  7d, 14d        N days ago (-P7D, -P14D)
  2w, 4w         N weeks ago (-P2W, -P4W)
  1m, 3m         N months ago (-P1M, -P3M)
  today          current date (ISO 8601)
  yesterday      yesterday's date (ISO 8601)
  2026-03-01     exact ISO 8601 date
  -P30D          ISO 8601 duration (passed directly to API)
```

Examples:
```
linear issue list --created-after 7d --priority-gte 2 --no-assignee
linear issue list --team ENG --my --state "In Progress"
linear issue list --due-before today --priority-lte 2
linear issue list --updated-after 2w --or --no-project --no-cycle
```

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

### Batch update issues

```
linear issue batch update [<id1> <id2> ...] [flags]
```

Updates multiple issues in a single API call. Accepts issue identifiers (e.g. ENG-42) as arguments or from stdin (one per line). Maximum 50 issues per batch.

Flags:
```
  --assignee string          assignee name, email, UUID, or "me"
  --state string             workflow state name or ID
  --priority int             priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low
  --label stringArray        set labels, replacing all existing (repeatable)
  --add-label stringArray    add label by name or ID (repeatable)
  --remove-label stringArray remove label by name or ID (repeatable)
  --project string           project name or ID
  --cycle string             cycle ID
  --json                     output updated issues as JSON array
```

At least one change flag is required. `--label` cannot be combined with `--add-label` or `--remove-label`.

Note: `--state` resolves the state name workspace-wide. If multiple teams have states with the same name, the first match is used.

Examples:
```
linear issue batch update ENG-1 ENG-2 ENG-3 --state Done
linear issue batch update ENG-10 ENG-11 --assignee me --priority 2
linear issue batch update ENG-5 ENG-6 --add-label bug --remove-label "needs triage"
echo -e "ENG-1\nENG-2\nENG-3" | linear issue batch update --state "In Review"
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
  --state string        project state type or UUID (backlog|planned|started|paused|completed|canceled)
  --target-date string  target date (YYYY-MM-DD)
  --start-date string   start date (YYYY-MM-DD)
  --json                output updated project as JSON
```

Example:
```
linear project update abc123 --state started
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
linear cycle list --team <key> [flags]
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

## Label Commands

Manage Linear issue labels with the `label` subcommand.

### List labels

```
linear label list [flags]
```

Flags:
```
  --team string   filter by team key (e.g. ENG); omit to list all labels
  --json          output as JSON array
```

Output columns: Name | Color | Description | Team | Group

Example:
```
linear label list
linear label list --team ENG
```

### Create label

```
linear label create --name <name> --color <hex> [flags]
```

Flags:
```
  --name string          label name (required)
  --color string         label color in hex format, e.g. #ff0000 (required)
  --team string          team key or ID (omit for workspace-level label)
  --description string   label description
  --json                 output created label as JSON
```

Example:
```
linear label create --name "bug" --color "#e11d48" --team ENG
```

### Update label

```
linear label update <id> [flags]
```

Only flags explicitly provided are sent to the API - omitted flags leave fields unchanged.

Flags:
```
  --name string          label name
  --color string         label color in hex format
  --description string   label description
  --json                 output updated label as JSON
```

Example:
```
linear label update abc123 --color "#f97316"
```

## State Commands

Manage Linear workflow states with the `state` subcommand.

### List states

```
linear state list --team <key> [flags]
```

Output is grouped by state type: Triage, Backlog, Unstarted, Started, Completed, Canceled.

Flags:
```
  --team string   team key (required)
  --json          output as JSON array
```

Output columns: Name | Type | Color

Example:
```
linear state list --team ENG
```

## Comment Commands

Manage Linear issue comments with the `comment` subcommand.

### List comments

```
linear comment list <issue-identifier> [flags]
```

Flags:
```
  --json   output as JSON array
```

Output columns: Author | Date | Body (truncated). Replies (threaded comments) are prefixed with `> `.

Example:
```
linear comment list ENG-42
```

### Create comment

```
linear comment create <issue-identifier> --body <text> [flags]
```

Flags:
```
  --body string     comment body in markdown (required)
  --parent string   parent comment ID for threading (reply to a comment)
  --json            output created comment as JSON
```

Example:
```
linear comment create ENG-42 --body "Looks good, merging soon."
linear comment create ENG-42 --body "Agreed." --parent abc123
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

## Search Command

Full-text search across all issues.

```
linear search <query> [flags]
```

Flags:
```
  --team string   boost results for a specific team (team key, e.g. ENG)
  --limit int     maximum number of results to return (default 25)
  --json          output as JSON array
```

Output columns: ID | Title | Status | Team

Examples:
```
linear search "login bug"
linear search "payment timeout" --team ENG
linear search "auth" --limit 10 --json
```

## Document Commands

Manage Linear documents with the `doc` subcommand.

### List documents

```
linear doc list [flags]
```

Flags:
```
  --project string       filter by project name or ID
  --limit int            maximum number of documents to return (default 50)
  --include-archived     include archived documents
  --json                 output as JSON array
```

Output columns: Title | Project | Creator | Updated

### Show document

```
linear doc show <id> [flags]
```

Flags:
```
  --json   output as JSON object
```

Output fields: Title, Project, Creator, Created, Updated, URL, and content body.

Example:
```
linear doc show abc123
```

### Create document

```
linear doc create --title <title> [flags]
```

Flags:
```
  --title string         document title (required)
  --content string       document content in markdown
  --content-file string  read content from a file (mutually exclusive with --content)
  --project string       project name or ID
  --json                 output created document as JSON
```

Examples:
```
linear doc create --title "Architecture Overview" --content "# Overview\n..."
linear doc create --title "Meeting Notes" --content-file notes.md --project "Q2 Launch"
```

### Update document

```
linear doc update <id> [flags]
```

Only flags explicitly provided are sent to the API - omitted flags leave fields unchanged.

Flags:
```
  --title string         new document title
  --content string       new document content in markdown
  --content-file string  read content from a file (mutually exclusive with --content)
  --json                 output updated document as JSON
```

Example:
```
linear doc update abc123 --title "Updated Title"
```

### Delete document

```
linear doc delete <id> [flags]
```

Moves the document to trash with a 30-day grace period. Use `--restore` to restore a trashed document.

Flags:
```
  --restore   restore document from trash instead of deleting
  --yes       skip confirmation prompt
```

Examples:
```
linear doc delete abc123 --yes
linear doc delete abc123 --restore
```

## Attachment Commands

Manage Linear issue attachments with the `attachment` subcommand. Attachment creation is idempotent: using the same URL and issue produces an update rather than a duplicate.

### List attachments

```
linear attachment list <issue-identifier> [flags]
```

Flags:
```
  --json   output as JSON array
```

Output columns: Title | URL | Created

Example:
```
linear attachment list ENG-42
```

### Create attachment

```
linear attachment create <issue-identifier> [flags]
```

Flags:
```
  --title string   attachment title (required)
  --url string     attachment URL (mutually exclusive with --file)
  --file string    local file to upload and attach (mutually exclusive with --url)
  --json           output created attachment as JSON
```

Prints the created attachment ID on success.

Examples:
```
linear attachment create ENG-42 --url "https://example.com/spec.pdf" --title "Spec"
linear attachment create ENG-42 --file ./screenshot.png --title "Screenshot"
```

When `--file` is used, the file is uploaded to Linear's storage first (see File Upload Workflow below), and the returned asset URL is used for the attachment.

### Delete attachment

```
linear attachment delete <id> [flags]
```

Flags:
```
  --yes   skip confirmation prompt
```

Example:
```
linear attachment delete abc123 --yes
```

## File Upload Workflow

When using `attachment create --file`, the CLI performs a two-step upload:

1. Calls the `fileUpload` mutation with the file's content type, name, and size to get a pre-signed upload URL and a final asset URL.
2. HTTP PUT the file to the pre-signed upload URL with the headers returned in step 1.
3. Uses the asset URL as the attachment URL.

The asset URL is a permanent Linear-hosted URL that can be embedded in issue descriptions or other content.

## Shell Completion

Generate tab-completion scripts for your shell.

```
linear completion <shell>
```

Supported shells: bash, zsh, fish, powershell

Setup:
```
# bash
linear completion bash > /etc/bash_completion.d/linear

# zsh
linear completion zsh > "${fpath[1]}/_linear"

# fish
linear completion fish > ~/.config/fish/completions/linear.fish

# powershell
linear completion powershell >> $PROFILE
```

## Pipe-friendly Workflows

All commands support `--json` output for use in pipelines. The `issue batch update` command reads identifiers from stdin when no arguments are given.

Get all high-priority unassigned issues and assign them to me:
```
linear issue list --priority-gte 1 --no-assignee --json | jq -r '.[].identifier' | linear issue batch update --assignee me
```

Close all issues in a specific state:
```
linear issue list --state "Cancelled" --json | jq -r '.[].identifier' | linear issue batch update --state Done
```

Search and update by priority:
```
linear search "login" --json | jq -r '.[] | select(.priority == "Urgent") | .identifier' | linear issue batch update --assignee me
```

Export issues to CSV (using jq):
```
linear issue list --team ENG --json | jq -r '.[] | [.identifier, .title, .state.name] | @csv'
```
