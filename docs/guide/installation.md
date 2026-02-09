# Installation

Kassie can be installed in multiple ways. Choose the method that works best for you.

## Curl Install Script (Recommended)

Quick installation via curl for Linux and macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/KashifKhn/kassie/main/install.sh | sh
```

This script will:
1. Detect your OS and architecture
2. Download the latest release from GitHub
3. Install to `/usr/local/bin` or `~/.local/bin`
4. Configure your PATH automatically
5. Verify the installation

To install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/KashifKhn/kassie/main/install.sh | sh -s -- --version 0.1.1
```

To skip PATH modification:

```bash
curl -fsSL https://raw.githubusercontent.com/KashifKhn/kassie/main/install.sh | sh -s -- --no-modify-path
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
go install github.com/KashifKhn/kassie@v0.1.1
```

## Homebrew (Coming Soon)

::: info Coming Soon
Homebrew tap is planned for a future release. Use the curl install script or Go install for now.
:::

```bash
# Coming soon
brew tap KashifKhn/kassie
brew install kassie
```

## Pre-built Binaries

Download pre-built binaries from [GitHub Releases](https://github.com/KashifKhn/kassie/releases):

::: tip Asset Naming
Release assets use the format: `kassie_VERSION_OS_ARCH.tar.gz`  
Example: `kassie_0.1.1_linux_amd64.tar.gz`
:::

### Linux (amd64)

```bash
curl -L https://github.com/KashifKhn/kassie/releases/download/v0.1.1/kassie_0.1.1_linux_amd64.tar.gz | tar -xz
chmod +x kassie
sudo mv kassie /usr/local/bin/
```

### Linux (arm64)

```bash
curl -L https://github.com/KashifKhn/kassie/releases/download/v0.1.1/kassie_0.1.1_linux_arm64.tar.gz | tar -xz
chmod +x kassie
sudo mv kassie /usr/local/bin/
```

### macOS (Intel)

```bash
curl -L https://github.com/KashifKhn/kassie/releases/download/v0.1.1/kassie_0.1.1_darwin_amd64.tar.gz | tar -xz
chmod +x kassie
sudo mv kassie /usr/local/bin/
```

### macOS (Apple Silicon)

```bash
curl -L https://github.com/KashifKhn/kassie/releases/download/v0.1.1/kassie_0.1.1_darwin_arm64.tar.gz | tar -xz
chmod +x kassie
sudo mv kassie /usr/local/bin/
```

### Windows (amd64)

Download the `.zip` file from [releases](https://github.com/KashifKhn/kassie/releases/latest) and extract it.

Or using PowerShell:

```powershell
Invoke-WebRequest -Uri https://github.com/KashifKhn/kassie/releases/download/v0.1.1/kassie_0.1.1_windows_amd64.zip -OutFile kassie.zip
Expand-Archive -Path kassie.zip -DestinationPath .
Move-Item kassie.exe C:\Windows\System32\
```

## Docker

::: warning Web UI Under Development
Docker images currently support TUI and server modes. Web UI support is coming in Phase 5.
:::

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

<!-- Web UI mode coming soon
### Web UI

```bash
docker run -p 8080:8080 ghcr.io/kashifkhn/kassie:latest web
```

Visit `http://localhost:8080` in your browser.
-->

### Server Mode

```bash
docker run -p 50051:50051 -p 8080:8080 \
  ghcr.io/kashifkhn/kassie:latest server
```

## Building from Source

If you want to build Kassie yourself:

### Prerequisites

- Go 1.24+
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

<VersionInfo />

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
mv kassie ~/bin/kassie
export PATH=$PATH:~/bin
```

Add the export to your `~/.bashrc` or `~/.zshrc`.

### Windows: Add to PATH

Add Kassie to your PATH via System Properties:
1. Search for "Environment Variables"
2. Edit "Path" under User variables
3. Add the directory containing `kassie.exe`

## Upgrading Kassie

Kassie includes a built-in upgrade command that makes updating to newer versions easy and safe.

### Self-Upgrade Command

Upgrade to the latest version:

```bash
kassie upgrade
```

The upgrade process will:
1. Check for the latest release on GitHub
2. Download and verify the new version
3. Create a backup of your current binary
4. Install the new version
5. Verify the installation works
6. Automatically rollback if anything fails

### Check for Updates

To only check if an update is available without installing:

```bash
kassie upgrade --check
```

Output:
```
🎯 Kassie Upgrade
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Current version: v0.1.0
  Latest version:  v0.1.1
  ✓ Update available
```

### Upgrade to Specific Version

Install or downgrade to a specific version:

```bash
kassie upgrade --version v0.1.0
```

This is useful for:
- Downgrading if a new version has issues
- Installing a specific tested version in production
- Testing different versions

### Force Reinstall

Reinstall the current version (useful for fixing corrupted installations):

```bash
kassie upgrade --force
```

### JSON Output

For scripting and automation:

```bash
kassie upgrade --check --json
```

Output:
```json
{
  "current_version": "v0.1.0",
  "latest_version": "v0.1.1",
  "update_available": true,
  "platform": {
    "os": "linux",
    "arch": "amd64"
  }
}
```

### Upgrade via Curl Script

The install script also works for upgrades:

```bash
curl -fsSL https://raw.githubusercontent.com/KashifKhn/kassie/main/install.sh | sh
```

### Upgrade with Homebrew (Coming Soon)

Once Homebrew tap is available:

```bash
brew upgrade kassie
```

### Upgrade with Go Install

If installed via `go install`, simply run:

```bash
go install github.com/KashifKhn/kassie@latest
```

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

**Curl Install**: The script should automatically configure PATH. If not:
```bash
echo 'export PATH=$PATH:~/.local/bin' >> ~/.bashrc
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
