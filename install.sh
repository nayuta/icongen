#!/usr/bin/env bash
set -euo pipefail

# Install generate-icons.sh to /usr/local/bin
# Usage: curl -fsSL https://raw.githubusercontent.com/yourusername/generate-app-icons/main/install.sh | bash

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
SCRIPT_NAME="generate-icons"
REPO_URL="https://raw.githubusercontent.com/yourusername/generate-app-icons/main/generate_icons.sh"

echo "Installing generate-icons to $INSTALL_DIR..."

# Check if ImageMagick is installed
if ! command -v convert >/dev/null 2>&1 && ! command -v magick >/dev/null 2>&1; then
    echo "Warning: ImageMagick not found. Install it first:"
    echo "  macOS: brew install imagemagick"
    echo "  Ubuntu/Debian: sudo apt-get install imagemagick"
    echo "  CentOS/RHEL: sudo yum install ImageMagick"
fi

# Download and install
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$REPO_URL" -o "$INSTALL_DIR/$SCRIPT_NAME"
elif command -v wget >/dev/null 2>&1; then
    wget -qO "$INSTALL_DIR/$SCRIPT_NAME" "$REPO_URL"
else
    echo "Error: curl or wget required for installation" >&2
    exit 1
fi

chmod +x "$INSTALL_DIR/$SCRIPT_NAME"

echo "✅ Installation complete!"
echo "Usage: $SCRIPT_NAME /path/to/source-image.png"
echo "Help:  $SCRIPT_NAME --help"