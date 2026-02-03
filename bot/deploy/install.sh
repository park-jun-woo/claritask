#!/bin/bash
# Claribot Installation Script

set -e

echo "=== Claribot Installation ==="
echo

# Check if running as root for system installation
if [ "$EUID" -ne 0 ]; then
    echo "Note: Run with sudo for system-wide installation"
    echo
fi

# Create system user
echo "Creating claribot user..."
if id "claribot" &>/dev/null; then
    echo "  User 'claribot' already exists"
else
    sudo useradd -r -s /bin/false claribot
    echo "  User 'claribot' created"
fi

# Create config directory
echo "Creating config directory..."
sudo mkdir -p /etc/claribot
sudo chown claribot:claribot /etc/claribot
sudo chmod 750 /etc/claribot
echo "  /etc/claribot created"

# Copy service file
echo "Installing systemd service..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
sudo cp "$SCRIPT_DIR/claribot.service" /etc/systemd/system/
echo "  Service file installed"

# Reload systemd
echo "Reloading systemd..."
sudo systemctl daemon-reload
echo "  Systemd reloaded"

echo
echo "=== Installation Complete ==="
echo
echo "Next steps:"
echo "1. Copy environment file:"
echo "   sudo cp $SCRIPT_DIR/claribot.env.example /etc/claribot/claribot.env"
echo
echo "2. Edit configuration:"
echo "   sudo nano /etc/claribot/claribot.env"
echo "   - Set TELEGRAM_TOKEN from @BotFather"
echo "   - Set ALLOWED_USERS with your Telegram ID"
echo "   - Set CLARITASK_DB path"
echo
echo "3. Secure the config file:"
echo "   sudo chmod 600 /etc/claribot/claribot.env"
echo "   sudo chown claribot:claribot /etc/claribot/claribot.env"
echo
echo "4. Build and install binary:"
echo "   cd $(dirname "$SCRIPT_DIR") && make install"
echo
echo "5. Enable and start service:"
echo "   sudo systemctl enable claribot"
echo "   sudo systemctl start claribot"
echo
echo "6. Check status:"
echo "   sudo systemctl status claribot"
echo "   sudo journalctl -u claribot -f"
