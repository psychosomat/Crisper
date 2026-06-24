#!/usr/bin/env bash
set -euo pipefail

VERSION="$1"
SOURCE_URL="$2"
SHA256="$3"

PKGNAME="crisper-bin"
AUR_REPO="ssh://aur@aur.archlinux.org/${PKGNAME}.git"
WORKDIR=$(mktemp -d)
trap 'rm -rf "$WORKDIR"' EXIT

git clone "$AUR_REPO" "$WORKDIR"
cp packaging/aur/PKGBUILD "$WORKDIR/"
cp packaging/aur/.SRCINFO "$WORKDIR/"

cd "$WORKDIR"
git config user.name "psychosomat"
git config user.email "hello@ddark.dev"
git add PKGBUILD .SRCINFO
git commit -m "Update to ${VERSION}"
git push origin HEAD
