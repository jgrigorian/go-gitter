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

Show all tracked repositories:

```bash
go-gitter list
go-gitter list -g work    # filter by group
```

### Sync repositories

Fetch updates for all tracked repositories:

```bash
go-gitter sync
go-gitter sync -g work    # sync only a specific group
go-gitter sync -p         # also pull (not just fetch)
```

### Remove a repository

Stop tracking a repository:

```bash
go-gitter rm my-repo
go-gitter rm /path/to/repo
```

## Configuration

Configuration is stored in `~/.config/go-gitter/config.yaml` (or `$XDG_CONFIG_HOME/go-gitter/config.yaml` if set).

Example config:

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

## Features

- Track multiple git repositories
- Organize repositories into groups
- Parallel syncing with progress indicator
- Colorized output
- Persistent configuration
