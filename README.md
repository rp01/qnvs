# QNVS - Quick Node Version Switcher

[![GitHub release](https://img.shields.io/github/release/qnvs/qnvs.svg)](https://github.com/qnvs/qnvs/releases)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-windows%20%7C%20macos%20%7C%20linux-lightgrey.svg)]()

```
 ██████╗ ███╗   ██╗██╗   ██╗███████╗
██╔═══██╗████╗  ██║██║   ██║██╔════╝
██║   ██║██╔██╗ ██║██║   ██║███████╗
██║▄▄ ██║██║╚██╗██║╚██╗ ██╔╝╚════██║
╚██████╔╝██║ ╚████║ ╚████╔╝ ███████║
 ╚══▀▀═╝ ╚═╝  ╚═══╝  ╚═══╝  ╚══════╝
```

**🚀 A fast, lightweight Node.js version manager that requires NO admin privileges!**

Built with Go and featuring a beautiful interactive TUI, QNVS is a single binary with zero dependencies.

## ✨ Features

- ✅ **No admin/root privileges required** - Works entirely in user space
- ✅ **Windows compatible** - Automatic shim fallback when junctions are blocked
- ✅ **Interactive TUI** - Beautiful terminal interface, just run `qnvs`
- ✅ **Single binary** - No dependencies, no installation scripts
- ✅ **Cross-platform** - Windows, macOS, and Linux
- ✅ **VPN/Proxy friendly** - Built-in TLS skip option for corporate networks
- ✅ **Fast switching** - Instant version changes
- ✅ **Lightweight** - ~10MB binary size

## 📸 Screenshots

### Interactive Mode
Just run `qnvs` with no arguments:

```
 ██████╗ ███╗   ██╗██╗   ██╗███████╗
██╔═══██╗████╗  ██║██║   ██║██╔════╝
██║   ██║██╔██╗ ██║██║   ██║███████╗
██║▄▄ ██║██║╚██╗██║╚██╗ ██╔╝╚════██║
╚██████╔╝██║ ╚████║ ╚████╔╝ ███████║
 ╚══▀▀═╝ ╚═╝  ╚═══╝  ╚═══╝  ╚══════╝
  Quick Node Version Switcher • No admin required

╭──────────────────────────────────────────────╮
│                                              │
│  ▸ 📦  Install Node.js                       │
│        Download and install a new version   │
│                                              │
│    📋  List/Switch                           │
│    🗑️   Uninstall                            │
│    🔧  Setup                                 │
│    🔓  Toggle TLS Skip                       │
│    ❓  Help                                  │
│    👋  Exit                                  │
│                                              │
│  ──────────────────────────────────────────  │
│   2 installed   Active: v22.22.0             │
│                                              │
╰──────────────────────────────────────────────╯

  ↑↓ navigate │ ⏎ select │ q quit
```

## 📦 Installation

### Download Binary

Download the latest binary for your platform from [Releases](https://github.com/qnvs/qnvs/releases):

| Platform | Download |
|----------|----------|
| Windows x64 | `qnvs-windows-amd64.exe` |
| Windows ARM64 | `qnvs-windows-arm64.exe` |
| macOS Intel | `qnvs-macos-amd64` |
| macOS Apple Silicon | `qnvs-macos-arm64` |
| Linux x64 | `qnvs-linux-amd64` |
| Linux ARM64 | `qnvs-linux-arm64` |

### Quick Install

**macOS / Linux:**
```bash
# Download (replace with your platform)
curl -L https://github.com/qnvs/qnvs/releases/latest/download/qnvs-linux-amd64 -o qnvs
chmod +x qnvs
./qnvs setup
```

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri "https://github.com/qnvs/qnvs/releases/latest/download/qnvs-windows-amd64.exe" -OutFile "qnvs.exe"
.\qnvs.exe setup
```

### Build from Source

```bash
git clone https://github.com/qnvs/qnvs.git
cd qnvs
go build -o qnvs .
./qnvs setup
```

## 🚀 Quick Start

### Interactive Mode (Recommended)

Just run `qnvs` with no arguments to launch the interactive TUI:

```bash
qnvs
```

Use arrow keys to navigate, Enter to select.

### CLI Mode

```bash
# Install a Node.js version
qnvs install 22
qnvs install lts
qnvs install 20.10.0

# Switch versions
qnvs use 22

# List installed versions
qnvs list

# Show current version
qnvs current

# Uninstall a version
qnvs uninstall 20
```

## 📖 Commands

| Command | Description |
|---------|-------------|
| `qnvs` | Launch interactive TUI |
| `qnvs install <version>` | Install a Node.js version |
| `qnvs use <version>` | Switch to a version |
| `qnvs list` | List installed versions |
| `qnvs current` | Show active version |
| `qnvs uninstall <version>` | Remove a version |
| `qnvs setup` | Initialize QNVS and configure PATH |
| `qnvs help` | Show help message |

### Version Formats

| Format | Example | Description |
|--------|---------|-------------|
| Major | `22` | Latest 22.x.x |
| Full | `22.10.0` | Specific version |
| LTS | `lts` | Latest LTS release |
| Latest | `latest` | Latest available |

## 🔐 Corporate VPN / Proxy Support

If you're behind a corporate VPN (Cato, Zscaler, etc.) that does TLS inspection, you may encounter certificate errors.

### CLI Flag
```bash
qnvs install 22 --insecure
qnvs --insecure install lts
```

### Interactive Mode
Select **"🔓 Toggle TLS Skip"** from the menu before installing.

## 🏢 Windows Corporate Environments

QNVS is designed to work in restricted Windows environments where:
- Developer Mode is disabled
- `mklink` is blocked by Group Policy
- Users don't have admin privileges

### How It Works

QNVS uses a **hybrid approach**:

1. **First, it tries directory junctions** (`mklink /J`) - these are fast and work on most systems without admin rights
2. **If junctions fail, it automatically falls back to shim mode** - creates small `.cmd` wrapper scripts that redirect to the active Node.js version

This fallback is automatic and transparent - you don't need to configure anything!

### Shim Mode Details

When shim mode is active, QNVS creates these files in `~/.qnvs/bin/`:
- `node.cmd` - redirects to the active Node.js executable
- `npm.cmd` - redirects to npm
- `npx.cmd` - redirects to npx

These shims cannot be blocked by Group Policy since they're just regular batch files.

## 📁 Directory Structure

QNVS stores data in the following locations:

| Platform | Default Location |
|----------|------------------|
| **Windows** | `%LOCALAPPDATA%\qnvs` (e.g., `C:\Users\xxx\AppData\Local\qnvs`) |
| **macOS/Linux** | `~/.qnvs` |

### Custom Location

You can override the default location by setting the `QNVS_HOME` environment variable:

**Windows (PowerShell):**
```powershell
$env:QNVS_HOME = "D:\tools\qnvs"
```

**Windows (CMD):**
```cmd
set QNVS_HOME=D:\tools\qnvs
```

**macOS/Linux:**
```bash
export QNVS_HOME=/opt/qnvs
```

### Directory Contents

```
qnvs/
├── bin/           # QNVS binary and shims
│   ├── qnvs.exe   # (or qnvs on Unix)
│   ├── node.cmd   # (Windows shim mode only)
│   ├── npm.cmd
│   └── npx.cmd
├── versions/      # Installed Node.js versions
│   ├── v20.10.0/
│   └── v22.22.0/
├── current        # Symlink/junction to active version
└── current_version # (Windows shim mode only) stores active version
```

## ⚙️ PATH Configuration

After running `qnvs setup`, add to your shell config:

**Bash (~/.bashrc):**
```bash
export PATH="$HOME/.qnvs/bin:$HOME/.qnvs/current/bin:$PATH"
```

**Zsh (~/.zshrc):**
```bash
export PATH="$HOME/.qnvs/bin:$HOME/.qnvs/current/bin:$PATH"
```

**PowerShell:**
```powershell
$env:Path = "$env:LOCALAPPDATA\qnvs\bin;$env:LOCALAPPDATA\qnvs\current;$env:Path"
```

**Windows CMD:**
```cmd
set PATH=%LOCALAPPDATA%\qnvs\bin;%LOCALAPPDATA%\qnvs\current;%PATH%
```

> 💡 **Note:** If you set a custom `QNVS_HOME`, use that path instead.

## 🛠️ Development

### Prerequisites
- Go 1.24+

### Build

```bash
# Build for current platform
go build -o qnvs .

# Cross-compile
GOOS=windows GOARCH=amd64 go build -o qnvs-windows.exe .
GOOS=darwin GOARCH=arm64 go build -o qnvs-macos-arm64 .
GOOS=linux GOARCH=amd64 go build -o qnvs-linux .
```

### Dependencies

QNVS uses only the [Charm](https://charm.sh/) ecosystem for the TUI:
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/charmbracelet/lipgloss` - Styling

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -am 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

## 🔧 Troubleshooting

### "Access is denied" when creating directory

If you see an error like:
```
Init failed: failed to create directory C:\Users\xxx\.qnvs: Access is denied.
```

**Solutions:**

1. **Set a custom location** using `QNVS_HOME`:
   ```powershell
   $env:QNVS_HOME = "C:\tools\qnvs"
   qnvs setup
   ```

2. **Use a different drive** if your user profile is restricted:
   ```cmd
   set QNVS_HOME=D:\qnvs
   qnvs setup
   ```

3. **Contact IT** to request write access to `%LOCALAPPDATA%`

### "Unknown Publisher" warning on Windows

This appears because the binary isn't code-signed. You can:
- Click "More info" → "Run anyway"
- Or right-click the file → Properties → Unblock

### Junction/symlink errors

If you see junction-related errors, QNVS will automatically fall back to shim mode. No action needed!

If shim mode also fails, set a custom `QNVS_HOME` to a location where you have full write access.

### Node.js not found after switching versions

Make sure your PATH is configured correctly:
```powershell
# Check if qnvs directories are in PATH
$env:Path -split ';' | Select-String 'qnvs'
```

Run `qnvs setup` to see the correct PATH configuration for your system.

## 🐛 Issues & Support

- **Bug Reports**: [GitHub Issues](https://github.com/qnvs/qnvs/issues)
- **Feature Requests**: [GitHub Issues](https://github.com/qnvs/qnvs/issues)

---

**Made with ❤️ in Go**