# Installation

Kassie can be installed in multiple ways. Choose the method that works best for you.

## Homebrew (macOS & Linux)

The easiest way to install on macOS and Linux:

```bash
brew tap KashifKhn/kassie
brew install kassie
```

Verify the installation:

```bash
kassie version
```

To update:

```bash
brew upgrade kassie
```

## Go Install

If you have Go installed (1.24+):

```bash
go install github.com/KashifKhn/kassie@latest
```

Make sure `$GOPATH/bin` is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

To install a specific version:

```bash
go install github.com/KashifKhn/kassie@v1.0.0
```

## Curl Install Script

Quick installation via curl:

```bash
curl -sSL https://kassie.dev/install.sh | bash
```

This script will:
1. Detect your OS and architecture
2. Download the appropriate binary
3. Install it to `/usr/local/bin`
4. Verify the installation

To specify a custom install location:

```bash
curl -sSL https://kassie.dev/install.sh | bash -s -- --prefix=/custom/path
```

## Pre-built Binaries

Download pre-built binaries from [GitHub Releases](https://github.com/KashifKhn/kassie/releases):

### Linux (amd64)

```bash
wget https://github.com/KashifKhn/kassie/releases/download/v1.0.0/kassie-linux-amd64
chmod +x kassie-linux-amd64
sudo mv kassie-linux-amd64 /usr/local/bin/kassie
```

### Linux (arm64)

```bash
wget https://github.com/KashifKhn/kassie/releases/download/v1.0.0/kassie-linux-arm64
chmod +x kassie-linux-arm64
sudo mv kassie-linux-arm64 /usr/local/bin/kassie
```

### macOS (Intel)

```bash
wget https://github.com/KashifKhn/kassie/releases/download/v1.0.0/kassie-darwin-amd64
chmod +x kassie-darwin-amd64
sudo mv kassie-darwin-amd64 /usr/local/bin/kassie
```

### macOS (Apple Silicon)

```bash
wget https://github.com/KashifKhn/kassie/releases/download/v1.0.0/kassie-darwin-arm64
chmod +x kassie-darwin-arm64
sudo mv kassie-darwin-arm64 /usr/local/bin/kassie
```

### Windows (amd64)

Download `kassie-windows-amd64.exe` from releases and add it to your PATH.

Or using PowerShell:

```powershell
Invoke-WebRequest -Uri https://github.com/KashifKhn/kassie/releases/download/v1.0.0/kassie-windows-amd64.exe -OutFile kassie.exe
Move-Item kassie.exe C:\Windows\System32\
```

## Docker

Run Kassie in a container:

### TUI

```bash
docker run -it ghcr.io/kashifkhn/kassie:latest tui
```

With a custom config file:

```bash
docker run -it \
  -v ~/.config/kassie:/root/.config/kassie \
  ghcr.io/kashifkhn/kassie:latest tui
```

### Web UI

```bash
docker run -p 8080:8080 ghcr.io/kashifkhn/kassie:latest web
```

Visit `http://localhost:8080` in your browser.

### Server Mode

```bash
docker run -p 50051:50051 -p 8080:8080 \
  ghcr.io/kashifkhn/kassie:latest server
```

## Building from Source

If you want to build Kassie yourself:

### Prerequisites

- Go 1.24 or later
- Node.js 20+ (for web UI)
- protoc (Protocol Buffer compiler)
- Make

### Clone and Build

```bash
# Clone repository
git clone https://github.com/KashifKhn/kassie.git
cd kassie

# Setup dependencies
make setup

# Generate protobuf code
make proto

# Build full binary (includes web UI)
make build

# Binary will be at ./kassie
./kassie version
```

### Build Options

Build server only (no web assets):

```bash
make build-server
```

Build web UI only:

```bash
make web
```

Development build (faster, no optimization):

```bash
go build -o kassie cmd/kassie/main.go
```

## Verification

After installation, verify that Kassie is working:

```bash
kassie version
```

Expected output:

```
Kassie v1.0.0
Commit: abc1234
Built: 2024-01-15T10:30:00Z
```

Test connectivity:

```bash
kassie tui --help
```

## Platform-Specific Notes

### macOS Gatekeeper

If you download a binary directly, macOS may block it:

```bash
xattr -d com.apple.quarantine /usr/local/bin/kassie
```

Or right-click the binary, select "Open", and confirm.

### Linux: No sudo Required

Install to user directory without sudo:

```bash
mkdir -p ~/bin
mv kassie-linux-amd64 ~/bin/kassie
export PATH=$PATH:~/bin
```

Add the export to your `~/.bashrc` or `~/.zshrc`.

### Windows: Add to PATH

Add Kassie to your PATH via System Properties:
1. Search for "Environment Variables"
2. Edit "Path" under User variables
3. Add the directory containing `kassie.exe`

## Next Steps

Now that Kassie is installed:

- [Getting Started](/guide/getting-started) - Quick tutorial
- [Configuration](/guide/configuration) - Set up your profiles
- [TUI Usage](/guide/tui-usage) - Learn the terminal interface

## Troubleshooting

### Command not found

If you get "command not found" after installation:

**Go Install**: Ensure `$GOPATH/bin` is in your PATH:
```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

**Homebrew**: Ensure Homebrew's bin is in PATH:
```bash
echo 'export PATH=/usr/local/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
```

### Permission Denied

If you get permission errors:

```bash
chmod +x /path/to/kassie
```

### SSL Certificate Errors (Go Install)

If `go install` fails with SSL errors:

```bash
go env -w GOPRIVATE=github.com/KashifKhn/kassie
go install github.com/KashifKhn/kassie@latest
```

For more help, see [Troubleshooting](/guide/troubleshooting).
