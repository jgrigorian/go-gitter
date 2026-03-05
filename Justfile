# Justfile for go-gitter

set dotenv-load := true

name := "go-gitter"

# Default target
default: build

# Build the binary
build:
    go build -o {{name}} .

# Build with debug info
build-debug:
    go build -o {{name}} -gcflags=all=-d=symoff=true .

# Run tests
test:
    go test -v ./...

# Install binary to ~/.local/bin
install:
    go build -o ~/.local/bin/{{name}} .

# Clean build artifacts
clean:
    rm -f {{name}}*
    rm -rf ./release

# Run the CLI
run ARGS='':
    go run . {{ ARGS }}

# Add a test repository
add-test:
    ./{{name}} add ~/.dotfiles test

# List repositories
list:
    ./{{name}} list

# Sync all repositories
sync:
    ./{{name}} sync

# Cross-platform builds
darwin:
    @echo "Building {{name}}-darwin_x86_64 binary"
    GOOS=darwin GOARCH=amd64 go build -o {{name}}-darwin_x86_64 .
    @mkdir -p release
    @echo "Compressing {{name}}-darwin_x86_64 binary..."
    tar -czvf ./release/{{name}}-darwin_x86_64.tar.gz ./{{name}}-darwin_x86_64
    rm ./{{name}}-darwin_x86_64
    @echo " "

darwin_arm:
    @echo "Building {{name}}-darwin_arm64 binary..."
    GOOS=darwin GOARCH=arm64 go build -o {{name}}-darwin_arm64 .
    @mkdir -p release
    @echo "Compressing {{name}}-darwin_arm64 binary..."
    tar -czvf ./release/{{name}}-darwin_arm64.tar.gz ./{{name}}-darwin_arm64
    rm ./{{name}}-darwin_arm64
    @echo " "

linux:
    @echo "Building {{name}}-linux_x86_64 binary..."
    GOOS=linux GOARCH=amd64 go build -o {{name}}-linux_x86_64 .
    @mkdir -p release
    @echo "Compressing {{name}}-linux_x86_64 binary..."
    tar -czvf ./release/{{name}}-linux_x86_64.tar.gz ./{{name}}-linux_x86_64
    rm ./{{name}}-linux_x86_64
    @echo " "

linux_arm:
    @echo "Building {{name}}-linux_arm64 binary..."
    GOOS=linux GOARCH=arm64 go build -o {{name}}-linux_arm64 .
    @mkdir -p release
    @echo "Compressing {{name}}-linux_arm64 binary..."
    tar -czvf ./release/{{name}}-linux_arm64.tar.gz ./{{name}}-linux_arm64
    rm ./{{name}}-linux_arm64
    @echo " "

# Build all platforms
all: clean
    just darwin
    just darwin_arm
    just linux
    just linux_arm
