#!/bin/bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*"; exit 1; }

require_brew() {
    if command -v brew &>/dev/null; then
        info "Homebrew found: $(brew --prefix)"
        return
    fi

    warn "Homebrew not found. Installing..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

    if [[ "$(uname -m)" == "arm64" ]]; then
        eval "$(/opt/homebrew/bin/brew shellenv)"
    else
        eval "$(/usr/local/bin/brew shellenv)"
    fi

    if ! command -v brew &>/dev/null; then
        error "Homebrew installation failed. Install manually: https://brew.sh"
    fi
    info "Homebrew installed successfully"
}

require_cmd() {
    local cmd="$1"
    local pkg="${2:-$1}"

    if command -v "$cmd" &>/dev/null; then
        info "$cmd found: $(command -v "$cmd")"
        return
    fi

    warn "$cmd not found. Installing $pkg..."
    brew install "$pkg"

    if ! command -v "$cmd" &>/dev/null; then
        error "Failed to install $cmd via 'brew install $pkg'"
    fi
    info "$cmd installed successfully"
}

main() {
    if [[ "$(uname)" != "Darwin" ]]; then
        error "This script is for macOS only."
    fi

    info "=== Crisper macOS dependency setup ==="

    require_brew
    require_cmd whisper-cli whisper-cpp
    require_cmd ffmpeg

    info "=== All dependencies installed ==="
    info "whisper-cli: $(command -v whisper-cli)"
    info "ffmpeg:      $(command -v ffmpeg)"

    echo ""
    info "You can now build Crisper:"
    info "  wails build"
}

main "$@"
