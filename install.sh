#!/usr/bin/env bash
# NeuronFS Installer for macOS/Linux
# Usage: curl -sL https://raw.githubusercontent.com/rhino-acoustic/NeuronFS/main/install.sh | bash

set -e

echo "🧠 Installing NeuronFS..."

# 1. Check requirements
if ! command -v go &> /dev/null; then
    echo "❌ Error: 'go' is not installed. Please install Go (1.22+) first."
    exit 1
fi
if ! command -v git &> /dev/null; then
    echo "❌ Error: 'git' is not installed."
    exit 1
fi

INSTALL_DIR="$HOME/.neuronfs"
BIN_DIR="$HOME/.local/bin"

# 2. Setup directories
mkdir -p "$INSTALL_DIR"
mkdir -p "$BIN_DIR"

# 3. Clone or Update
if [ -d "$INSTALL_DIR/repo" ]; then
    echo "🔄 Updating existing installation..."
    cd "$INSTALL_DIR/repo"
    git pull origin main
else
    echo "📦 Cloning NeuronFS repository..."
    git clone https://github.com/rhino-acoustic/NeuronFS.git "$INSTALL_DIR/repo"
    cd "$INSTALL_DIR/repo"
fi

# 4. Build
echo "🔨 Building core engine..."
cd runtime
go build -o "$BIN_DIR/neuronfs" .
chmod +x "$BIN_DIR/neuronfs"

echo "✅ Installed successfully to $BIN_DIR/neuronfs!"

# 5. Check PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo "⚠️  Please add $BIN_DIR to your PATH."
    echo "Run: export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

echo "🚀 Run 'neuronfs --init ./my_brain' to create your first autonomous brain."
