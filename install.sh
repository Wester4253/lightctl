#!/usr/bin/env bash
set -e

# lightctl installer
# Usage: curl -fsSL https://raw.githubusercontent.com/Wester4253/lightctl/main/install.sh | bash

echo "╔═══════════════════════════════════════════════╗"
echo "║         lightctl Installation Script          ║"
echo "║   Home Assistant Light Control for Terminal  ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="Wester4253/lightctl"
BRANCH="main"
GITHUB_URL="https://github.com/${REPO}"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="lightctl"
BUILD_DIR="/tmp/lightctl-install-$$"
LEGACY_PATH="/usr/bin/$BINARY_NAME"

# Functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

check_command() {
    if command -v "$1" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

# Check for required commands
log_info "Checking prerequisites..."

if ! check_command git; then
    log_error "git is not installed. Please install git first."
    exit 1
fi

if ! check_command go; then
    log_error "Go is not installed. Please install Go 1.22 or newer."
    log_info "Visit: https://go.dev/doc/install"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+' || echo "0.0")
GO_MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
GO_MINOR=$(echo "$GO_VERSION" | cut -d. -f2)

if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 22 ]); then
    log_error "Go version $GO_VERSION found, but 1.22 or newer is required."
    exit 1
fi

log_success "Go $GO_VERSION detected"

# Clone repository
log_info "Cloning lightctl from GitHub (branch: $BRANCH)..."
rm -rf "$BUILD_DIR"
git clone --depth 1 --branch "$BRANCH" "$GITHUB_URL" "$BUILD_DIR" > /dev/null 2>&1

if [ $? -ne 0 ]; then
    log_error "Failed to clone repository from $GITHUB_URL"
    exit 1
fi

log_success "Repository cloned"

# Build the Go binary
log_info "Building lightctl..."
cd "$BUILD_DIR"

# Main branch has Go files at root (no subdirectory needed)
if [ ! -f "go.mod" ]; then
    log_error "Could not find Go project in repository"
    rm -rf "$BUILD_DIR"
    exit 1
fi

go mod download > /dev/null 2>&1
go build -ldflags="-s -w" -o "$BINARY_NAME" . > /dev/null 2>&1

if [ $? -ne 0 ]; then
    log_error "Failed to build lightctl"
    rm -rf "$BUILD_DIR"
    exit 1
fi

log_success "Build complete"

# Install binary
log_info "Installing to $INSTALL_DIR/$BINARY_NAME..."

if [ -w "$INSTALL_DIR" ]; then
    # Can write without sudo
    cp "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    # Need sudo
    log_warning "Need sudo to install to $INSTALL_DIR"
    sudo cp "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

if [ $? -ne 0 ]; then
    log_error "Failed to install binary"
    rm -rf "$BUILD_DIR"
    exit 1
fi

log_success "Installed to $INSTALL_DIR/$BINARY_NAME"

# Remove the Python executable left by older installations. This matters on
# systems where /usr/bin appears before /usr/local/bin in PATH.
if [ -e "$LEGACY_PATH" ] && [ "$INSTALL_DIR" != "$(dirname "$LEGACY_PATH")" ]; then
    log_warning "Removing older lightctl installation at $LEGACY_PATH"
    if rm -f "$LEGACY_PATH" 2>/dev/null; then
        log_success "Removed older installation"
    elif [ -e "$LEGACY_PATH" ] && sudo rm -f "$LEGACY_PATH"; then
        log_success "Removed older installation"
    else
        log_warning "Could not remove $LEGACY_PATH; run 'sudo rm $LEGACY_PATH' to avoid invoking the older version."
    fi
fi

# Clean up
rm -rf "$BUILD_DIR"
log_success "Cleanup complete"

echo ""
echo "╔═══════════════════════════════════════════════╗"
echo "║           Installation Complete! 🎉           ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""
log_info "lightctl has been installed to $INSTALL_DIR/$BINARY_NAME"
log_info "If your shell still finds an older version, run 'hash -r' and start a new shell."
echo ""
echo "Next steps:"
echo "  1. Generate a Home Assistant long-lived access token"
echo "     (Settings → Your Profile → Long-Lived Access Tokens)"
echo ""
echo "  2. Run 'lightctl' to create the config file"
echo ""
echo "  3. Edit ~/.config/govee/config.json and add your token"
echo ""
echo "  4. Run 'lightctl' again to launch the TUI!"
echo ""
echo "Commands:"
echo "  lightctl              - Launch interactive TUI"
echo "  lightctl status       - Show device status"
echo "  lightctl night        - Apply night profile"
echo "  lightctl brightness 50 - Set brightness"
echo "  lightctl --help       - Show all commands"
echo ""
log_success "Enjoy controlling your lights! ✨"
