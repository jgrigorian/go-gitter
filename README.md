# go-gitter

A CLI tool to manage and sync multiple local git repositories.

## Installation

```bash
go install github.com/jgrigorian/go-gitter@latest
```

Or build from source using [just](https://github.com/casey/just):

```bash
just build        # Build the binary
just install      # Install to ~/.local/bin
just run ARGS=""  # Run the CLI
just test         # Run tests
just clean        # Clean build artifacts
```

## Usage

### Add a repository

Add a git repository to track:

```bash
go-gitter add /path/to/repo
go-gitter add /path/to/repo my-repo        # with custom name
go-gitter add /path/to/repo -g work        # add to a group
```

### List repositories

Show all tracked repositories with their current branch:

```bash
go-gitter list
go-gitter list -g work    # filter by group
```

Example output:
```
NAME     PATH                    GROUP    BRANCH    BEHIND  LAST SYNC
myrepo   /home/user/projects/foo  work     main      -       2026-03-05 10:30
myrepo2  /home/user/projects/bar  dev      feature   3       2026-03-05 10:25
```

### Sync repositories

Fetch or pull updates for all tracked repositories:

```bash
go-gitter sync              # fetch updates only
go-gitter sync -g work     # sync only a specific group
go-gitter sync -p          # pull from current branch
```

When syncing with `-p`, the tool will pull from the current branch's upstream. Non-standard branches (not main/master) are flagged with a warning.

Example output:
```
Syncing 2 repository(s) in parallel...

✓ myrepo (main) pulled changes
⚠ myrepo2 (feature-branch) 3 update(s) available
```

### Remove a repository

Stop tracking a repository:

```bash
go-gitter rm my-repo
go-gitter rm /path/to/repo
```

### Configuration management

View and manage your configuration:

```bash
go-gitter config path        # Show config file location
go-gitter config edit        # Open config in your editor ($EDITOR)
go-gitter config validate    # Validate config and check repositories
```

## Configuration

Configuration is stored in `~/.config/go-gitter/config.yaml` (or `$XDG_CONFIG_HOME/go-gitter/config.yaml` if set).

Both YAML and JSON formats are supported:

```yaml
repositories:
  - path: /home/user/projects/foo
    name: foo
    group: personal
    last_sync: "2024-01-15T10:30:00Z"
  - path: /home/user/projects/bar
    name: bar
    group: work
settings:
  auto_fetch: false
  sync_timeout: 300
```

```json
{
  "repositories": [
    {
      "path": "/home/user/projects/foo",
      "name": "foo",
      "group": "personal",
      "last_sync": "2024-01-15T10:30:00Z"
    },
    {
      "path": "/home/user/projects/bar",
      "name": "bar",
      "group": "work"
    }
  ],
  "settings": {
    "auto_fetch": false,
    "sync_timeout": 300
  }
}
```

## Features

- Track multiple git repositories
- Organize repositories into groups
- Parallel syncing with progress indicator
- Current branch display in list view
- Pull from current branch during sync
- Non-standard branch warnings
- Colorized output
- Persistent configuration
- Input validation (validates git repos before adding)
- Config validation and diagnostics
- JSON and YAML config support
- Thread-safe config updates during sync
