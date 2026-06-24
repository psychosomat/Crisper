#!/usr/bin/env bash
set -euo pipefail

VERSION="$1"
SOURCE_URL="$2"
SHA256="$3"

PKGNAME="crisper-bin"
AUR_DIR="packaging/aur"
mkdir -p "$AUR_DIR"

cat > "${AUR_DIR}/PKGBUILD" << EOF
# Maintainer: psychosomat <hello@ddark.dev>
pkgname=${PKGNAME}
pkgver=${VERSION}
pkgrel=1
pkgdesc='Local video transcription with speaker diarization'
arch=('x86_64')
url='https://github.com/psychosomat/Crisper'
license=('MIT')
depends=('gtk3' 'webkit2gtk-4.1')
provides=('crisper')
conflicts=('crisper')
source=("${SOURCE_URL}")
sha256sums=('${SHA256}')

package() {
    install -Dm755 Crisper "\${pkgdir}/usr/bin/Crisper"
    install -Dm644 appicon.png "\${pkgdir}/usr/share/icons/hicolor/256x256/apps/crisper.png"

    mkdir -p "\${pkgdir}/usr/share/applications"
    cat > "\${pkgdir}/usr/share/applications/crisper.desktop" << EOF2
[Desktop Entry]
Type=Application
Name=Crisper
Exec=/usr/bin/Crisper
Icon=crisper
Categories=AudioVideo;Audio;
Terminal=false
EOF2
}
EOF

cat > "${AUR_DIR}/.SRCINFO" << EOF
pkgbase = ${PKGNAME}
	pkgver = ${VERSION}
	pkgrel = 1
	arch = x86_64
	depends = gtk3
	depends = webkit2gtk-4.1
	maintainer = psychosomat <hello@ddark.dev>
	pkgdesc = Local video transcription with speaker diarization
	url = https://github.com/aspect-apps/Crisper
	license = MIT
	provides = crisper
	conflicts = crisper
	source = ${SOURCE_URL}
	sha256sums = ${SHA256}

	pkgname = ${PKGNAME}
EOF

echo "Rendered AUR files in ${AUR_DIR}"
