# Claribot Passwordless Reinstall Guide

## Problem

Currently, `make install` uses `sudo` at every step:
- Binary copy (`sudo cp` → `/usr/local/bin/`)
- Service file move (`sudo mv` → `/etc/systemd/system/`)
- systemctl commands (`sudo systemctl daemon-reload/enable/start/restart`)

The Claude Code agent cannot enter a `sudo` password, so in this setup the agent cannot reinstall or restart the service.

## Prerequisites

- **Go**: 1.24+ (bot), 1.22+ (cli)
- **Node.js**: 18+ (for GUI build with Vite)
- **npm**: 8+
- **systemd**: Required for service management
- **CGO**: Enabled (required for SQLite - `mattn/go-sqlite3`)

## Solution: sudoers NOPASSWD Configuration

Register sudoers rules that allow passwordless sudo for specific commands only. This opens system security minimally while enabling the agent to perform necessary operations.

## Initial Setup (One-time, performed manually by the user)

### Step 1: Create the sudoers file

```bash
sudo visudo -f /etc/sudoers.d/claribot
```

Enter the following content. Replace `<username>` with your actual username:

```
# Claribot service management - no password required
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl daemon-reload
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl start claribot.service
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl stop claribot.service
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl restart claribot.service
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl enable claribot.service
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl disable claribot.service
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl status claribot.service
<username> ALL=(root) NOPASSWD: /bin/cp * /usr/local/bin/claribot
<username> ALL=(root) NOPASSWD: /bin/cp * /usr/local/bin/clari
<username> ALL=(root) NOPASSWD: /bin/chmod +x /usr/local/bin/claribot
<username> ALL=(root) NOPASSWD: /bin/chmod +x /usr/local/bin/clari
<username> ALL=(root) NOPASSWD: /bin/mv /tmp/claribot.service /etc/systemd/system/claribot.service
<username> ALL=(root) NOPASSWD: /bin/rm -f /usr/local/bin/claribot
<username> ALL=(root) NOPASSWD: /bin/rm -f /usr/local/bin/clari
<username> ALL=(root) NOPASSWD: /bin/rm -f /etc/systemd/system/claribot.service
<username> ALL=(root) NOPASSWD: /usr/bin/journalctl -u claribot.service *
```

### Step 2: Verify permissions

```bash
# Verify that file permissions are correctly set
ls -la /etc/sudoers.d/claribot
# Should be: -r--r----- 1 root root ...

# Validate syntax with visudo (already validated when using visudo -f, but double-check)
sudo visudo -c
```

### Step 3: Test

```bash
# Test that commands run without a password prompt
sudo systemctl status claribot.service
sudo cp /dev/null /dev/null  # Simple cp test (not an actual copy)
```

If the commands execute without a password prompt, the setup is complete.

## Operations Available to the Agent After Setup

Once the sudoers configuration is complete, the Claude Code agent can execute the following without a password:

| Operation | Command |
|-----------|---------|
| Build | `make build` (no sudo required) |
| CLI install | `make install-cli` |
| Bot install | `make install-bot` |
| Full install | `make install` |
| Restart service | `make restart` |
| Service status | `make status` |
| Service logs | `make logs` |
| Full uninstall | `make uninstall` |

No modifications to the existing Makefile are needed. The `sudo` commands in the Makefile will pass through without a password thanks to the sudoers rules.

### Development Targets

| Operation | Command |
|-----------|---------|
| GUI dev server | `make dev-gui` (Vite HMR + API proxy → 127.0.0.1:9847) |
| Bot local run | `make run-bot` |
| CLI local run | `make run-cli` |
| Run tests | `make test` |
| Clean build artifacts | `make clean` |
| Show help | `make help` |

## Build Process

The `make build` command executes three stages in order:

```
1. build-gui   → cd gui && npm install && npm run build (tsc -b && vite build → dist/)
                → rm -rf bot/internal/webui/dist
                → cp -r gui/dist bot/internal/webui/dist (copy to Go embed dir)
2. build-cli   → cd cli && go build -o ../bin/clari ./cmd/clari
3. build-bot   → cd bot && go build -o ../bin/claribot ./cmd/claribot (embeds GUI dist)
```

The GUI is embedded into the bot binary via Go's `embed` package (`bot/internal/webui/webui.go` with `//go:embed dist/*`), so the final `claribot` binary serves the Web UI without external files.

## Deploy Script (Self-Deploy)

For deployments triggered by the agent itself, the deploy script (`deploy/claribot-deploy.sh`) handles the stop→copy→start cycle via `nohup` so the process survives the parent (claribot) being stopped:

```bash
make build && nohup deploy/claribot-deploy.sh > /tmp/deploy.log 2>&1 &
```

The script:
1. Waits 2 seconds (for claribot to finish sending its response)
2. Stops the claribot service (`systemctl stop`)
3. Copies both binaries (`claribot` and `clari`) to `/usr/local/bin/`
4. Starts the service (`systemctl start`)
5. Waits 2 seconds, then verifies the service is active

Deploy log: `/tmp/claribot-deploy.log`

## Configuration

Copy the example config to `~/.claribot/`:

```bash
cp deploy/config.example.yaml ~/.claribot/config.yaml
```

Key settings in `config.yaml`:

| Section | Key | Default | Description |
|---------|-----|---------|-------------|
| service | host | 127.0.0.1 | HTTP listen address |
| service | port | 9847 | HTTP listen port |
| telegram | token | - | Bot token from @BotFather |
| telegram | allowed_users | [] | Allowed Telegram user IDs |
| claude | timeout | 1200 | Idle timeout (seconds) |
| claude | max_timeout | 1800 | Absolute timeout (seconds, range: 60-7200) |
| claude | max | 10 | Max concurrent Claude instances |
| project | path | ~/projects | Default path for new projects |
| pagination | page_size | 10 | Items per page (max: 100) |
| log | level | info | Log level (debug, info, warn, error) |
| log | file | ~/.claribot/claribot.log | Log file path (empty = stdout only) |

## Service Template

The systemd service file is generated from `deploy/claribot.service.template`:

```ini
[Unit]
Description=Claribot - LLM Project Automation Service
After=network.target

[Service]
Type=simple
User=__USER__
WorkingDirectory=__HOME__/.claribot
ExecStart=/usr/local/bin/claribot
Restart=on-failure
RestartSec=5
Environment=HOME=__HOME__
Environment=PATH=__HOME__/.local/bin:/usr/local/bin:/usr/bin:/bin

[Install]
WantedBy=multi-user.target
```

During `make install-bot`, `__USER__` and `__HOME__` are replaced with actual values via `sed`.

## Security Considerations

### Allowed Scope

- Only systemctl commands related to the claribot service are allowed (cannot control other services)
- Binary copy paths are limited to `/usr/local/bin/claribot` and `/usr/local/bin/clari`
- Service file move path is limited to `/etc/systemd/system/claribot.service`

### Not Allowed

- Controlling other system services
- Copying files to arbitrary paths
- Installing/removing packages (apt, yum, etc.)
- User management
- Other system administration commands

### Removal

When no longer needed:

```bash
sudo rm /etc/sudoers.d/claribot
```

## Note: Verifying systemctl Paths

Paths for systemctl, cp, chmod, mv, rm may differ depending on the distribution:

```bash
which systemctl  # /usr/bin/systemctl or /bin/systemctl
which cp         # /usr/bin/cp or /bin/cp
which chmod      # /usr/bin/chmod or /bin/chmod
which mv         # /usr/bin/mv or /bin/mv
which rm         # /usr/bin/rm or /bin/rm
which journalctl # /usr/bin/journalctl or /bin/journalctl
```

The paths in the sudoers file must be adjusted to match your actual system. This document is written based on Ubuntu/Debian.

## Automated Setup Script

A script is provided that performs all the above steps at once. **The user only needs to run this script once manually**:

```bash
# Usage: sudo bash deploy/setup-sudoers.sh
# Remove: sudo rm /etc/sudoers.d/claribot
```

The script auto-detects system command paths and generates the sudoers file. See `deploy/setup-sudoers.sh` for details.

## Deploy Directory Contents

| File | Description |
|------|-------------|
| `claribot-deploy.sh` | Self-deploy script (nohup, stop→copy→start) |
| `claribot.service.template` | systemd service template |
| `setup-sudoers.sh` | Passwordless sudo setup script |
| `config.example.yaml` | Example configuration file |
| `logo.png` | Telegram bot profile image |

## Operational Flow

```
[One-time] User runs: sudo bash deploy/setup-sudoers.sh
                      cp deploy/config.example.yaml ~/.claribot/config.yaml
                      (edit config.yaml with telegram token, etc.)
  |
[First install] make install
  |
[Thereafter] Claude Code agent freely executes:
  make build && nohup deploy/claribot-deploy.sh > /tmp/deploy.log 2>&1 &
  or
  make build -> make install -> make restart
  or
  make uninstall -> make install
```
