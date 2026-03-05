# Justfile for go-gitter

# Build the binary
build:
    go build -o go-gitter .

# Build with debug info
build-debug:
    go build -o go-gitter -gcflags=all=-d=symoff=true .

# Run tests
test:
    go test -v ./...

# Install binary to ~/.local/bin
install:
    go build -o ~/.local/bin/go-gitter .

# Clean build artifacts
clean:
    rm -f go-gitter

# Run the CLI
run ARGS='':
    go run . {{ ARGS }}

# Add a test repository
add-test:
    ./go-gitter add ~/.dotfiles test

# List repositories
list:
    ./go-gitter list

# Sync all repositories
sync:
    ./go-gitter sync

# Default target
default: build
