# Claribot Passwordless Reinstall Guide

## Problem

Currently, `make install` uses `sudo` at every step:
- Binary copy (`sudo cp` → `/usr/local/bin/`)
- Service file move (`sudo mv` → `/etc/systemd/system/`)
- systemctl commands (`sudo systemctl daemon-reload/enable/start/restart`)

The Claude Code agent cannot enter a `sudo` password, so in this setup the agent cannot reinstall or restart the service.

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
#!/bin/bash
# deploy/setup-sudoers.sh
# Usage: sudo bash deploy/setup-sudoers.sh

set -e

if [ "$EUID" -ne 0 ]; then
    echo "Error: Run with sudo: sudo bash $0"
    exit 1
fi

# Actual user (sudo caller)
REAL_USER="${SUDO_USER:-$(whoami)}"

if [ "$REAL_USER" = "root" ]; then
    echo "Error: Run with sudo from a regular user account"
    exit 1
fi

# Detect command paths
SYSTEMCTL=$(which systemctl)
CP=$(which cp)
CHMOD=$(which chmod)
MV=$(which mv)
RM=$(which rm)
JOURNALCTL=$(which journalctl)

SUDOERS_FILE="/etc/sudoers.d/claribot"

cat > "$SUDOERS_FILE" << EOF
# Claribot service management - no password required
# Generated by setup-sudoers.sh for user: ${REAL_USER}

${REAL_USER} ALL=(root) NOPASSWD: ${SYSTEMCTL} daemon-reload
${REAL_USER} ALL=(root) NOPASSWD: ${SYSTEMCTL} start claribot.service
${REAL_USER} ALL=(root) NOPASSWD: ${SYSTEMCTL} stop claribot.service
${REAL_USER} ALL=(root) NOPASSWD: ${SYSTEMCTL} restart claribot.service
${REAL_USER} ALL=(root) NOPASSWD: ${SYSTEMCTL} enable claribot.service
${REAL_USER} ALL=(root) NOPASSWD: ${SYSTEMCTL} disable claribot.service
${REAL_USER} ALL=(root) NOPASSWD: ${SYSTEMCTL} status claribot.service
${REAL_USER} ALL=(root) NOPASSWD: ${CP} * /usr/local/bin/claribot
${REAL_USER} ALL=(root) NOPASSWD: ${CP} * /usr/local/bin/clari
${REAL_USER} ALL=(root) NOPASSWD: ${CHMOD} +x /usr/local/bin/claribot
${REAL_USER} ALL=(root) NOPASSWD: ${CHMOD} +x /usr/local/bin/clari
${REAL_USER} ALL=(root) NOPASSWD: ${MV} /tmp/claribot.service /etc/systemd/system/claribot.service
${REAL_USER} ALL=(root) NOPASSWD: ${RM} -f /usr/local/bin/claribot
${REAL_USER} ALL=(root) NOPASSWD: ${RM} -f /usr/local/bin/clari
${REAL_USER} ALL=(root) NOPASSWD: ${RM} -f /etc/systemd/system/claribot.service
${REAL_USER} ALL=(root) NOPASSWD: ${JOURNALCTL} -u claribot.service *
EOF

chmod 0440 "$SUDOERS_FILE"

# Validate syntax
if visudo -c -f "$SUDOERS_FILE" > /dev/null 2>&1; then
    echo "sudoers configuration complete: ${SUDOERS_FILE}"
    echo "User: ${REAL_USER}"
    echo ""
    echo "The Claude Code agent can now execute the following without a password:"
    echo "  make install    - Full build and install"
    echo "  make restart    - Restart service"
    echo "  make uninstall  - Full uninstall"
    echo ""
    echo "To remove: sudo rm ${SUDOERS_FILE}"
else
    echo "Error: sudoers syntax error detected. Removing the file."
    rm -f "$SUDOERS_FILE"
    exit 1
fi
```

## Operational Flow

```
[One-time] User runs: sudo bash deploy/setup-sudoers.sh
  |
[Thereafter] Claude Code agent freely executes:
  make build -> make install -> make restart
  or
  make uninstall -> make install
```
