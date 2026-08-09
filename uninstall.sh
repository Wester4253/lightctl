#!/usr/bin/env bash
set -e

# lightctl uninstaller

echo "Uninstalling lightctl..."

# Remove binary from common installation locations
for location in /usr/local/bin/lightctl /usr/bin/lightctl; do
    if [ -f "$location" ]; then
        echo "Found lightctl at $location"
        if [ -w "$(dirname "$location")" ]; then
            rm "$location"
            echo "✓ Removed $location"
        else
            sudo rm "$location"
            echo "✓ Removed $location (with sudo)"
        fi
    fi
done

# Ask about config removal
echo ""
read -p "Remove configuration file at ~/.config/govee/config.json? (y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -f ~/.config/govee/config.json ]; then
        rm ~/.config/govee/config.json
        echo "✓ Removed configuration"
    fi
    if [ -d ~/.config/govee ]; then
        rmdir ~/.config/govee 2>/dev/null || true
    fi
fi

echo ""
echo "✓ lightctl uninstalled successfully"
