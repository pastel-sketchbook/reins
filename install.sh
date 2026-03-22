#!/bin/sh
# install.sh — Download and install the reins binary.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/pastel-sketchbook/reins/main/install.sh | sh
#   curl -fsSL ... | sh -s -- --version v0.2.0
#   curl -fsSL ... | sh -s -- --dir /usr/local/bin
#
# The script detects the current OS and architecture, downloads the
# matching binary from GitHub releases, verifies its SHA256 checksum,
# and installs it to the chosen directory.

set -eu

REPO="pastel-sketchbook/reins"
INSTALL_DIR="${HOME}/.local/bin"
VERSION=""

# -----------------------------------------------------------------------
# Argument parsing
# -----------------------------------------------------------------------
while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --dir)     INSTALL_DIR="$2"; shift 2 ;;
    -h|--help)
      echo "Usage: install.sh [--version VERSION] [--dir INSTALL_DIR]"
      echo ""
      echo "Options:"
      echo "  --version   Release tag to install (default: latest)"
      echo "  --dir       Installation directory (default: ~/.local/bin)"
      exit 0
      ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

# -----------------------------------------------------------------------
# Helpers
# -----------------------------------------------------------------------
die() { echo "ERROR: $*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || die "'$1' is required but not found in PATH"
}

# -----------------------------------------------------------------------
# Detect OS
# -----------------------------------------------------------------------
detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux"   ;;
    Darwin*) echo "darwin"  ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) die "Unsupported operating system: $(uname -s)" ;;
  esac
}

# -----------------------------------------------------------------------
# Detect architecture
# -----------------------------------------------------------------------
detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64)  echo "arm64" ;;
    *) die "Unsupported architecture: $(uname -m)" ;;
  esac
}

# -----------------------------------------------------------------------
# Resolve the latest release tag from GitHub
# -----------------------------------------------------------------------
resolve_version() {
  if [ -n "$VERSION" ]; then
    echo "$VERSION"
    return
  fi
  need curl
  # Do NOT follow redirects (-L) here; we need the 302 Location URL.
  tag=$(curl -fsS -o /dev/null -w '%{redirect_url}' \
    "https://github.com/${REPO}/releases/latest" 2>/dev/null \
    | grep -o '[^/]*$') || true
  [ -n "$tag" ] || die "Could not determine latest release"
  echo "$tag"
}

# -----------------------------------------------------------------------
# Main
# -----------------------------------------------------------------------
main() {
  need curl

  OS="$(detect_os)"
  ARCH="$(detect_arch)"
  VERSION="$(resolve_version)"

  # Strip leading 'v' for the archive filename (tag is v0.2.0, file uses 0.2.0)
  FILE_VERSION="${VERSION#v}"

  if [ "$OS" = "windows" ]; then
    ARCHIVE="reins-v${FILE_VERSION}-${OS}-${ARCH}.zip"
  else
    ARCHIVE="reins-v${FILE_VERSION}-${OS}-${ARCH}.tar.gz"
  fi
  CHECKSUMS="checksums-v${FILE_VERSION}.txt"

  BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
  ARCHIVE_URL="${BASE_URL}/${ARCHIVE}"
  CHECKSUMS_URL="${BASE_URL}/${CHECKSUMS}"

  echo "Installing reins ${VERSION} (${OS}/${ARCH})..."

  TMPDIR="$(mktemp -d)"
  trap 'rm -rf "$TMPDIR"' EXIT

  # Download archive and checksums
  echo "Downloading ${ARCHIVE}..."
  curl -fsSL -o "${TMPDIR}/${ARCHIVE}" "$ARCHIVE_URL" \
    || die "Failed to download ${ARCHIVE_URL}"

  echo "Downloading checksums..."
  curl -fsSL -o "${TMPDIR}/${CHECKSUMS}" "$CHECKSUMS_URL" \
    || die "Failed to download ${CHECKSUMS_URL}"

  # Verify checksum
  echo "Verifying checksum..."
  EXPECTED=$(grep "$ARCHIVE" "${TMPDIR}/${CHECKSUMS}" | awk '{print $1}')
  [ -n "$EXPECTED" ] || die "Archive not found in checksums file"

  if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL=$(sha256sum "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')
  elif command -v shasum >/dev/null 2>&1; then
    ACTUAL=$(shasum -a 256 "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')
  else
    die "No sha256sum or shasum found; cannot verify checksum"
  fi

  [ "$EXPECTED" = "$ACTUAL" ] || die "Checksum mismatch: expected ${EXPECTED}, got ${ACTUAL}"
  echo "Checksum verified."

  # Extract binary
  if [ "$OS" = "windows" ]; then
    need unzip
    unzip -qo "${TMPDIR}/${ARCHIVE}" -d "${TMPDIR}"
    BINARY="reins-${OS}-${ARCH}.exe"
  else
    tar -xzf "${TMPDIR}/${ARCHIVE}" -C "${TMPDIR}"
    BINARY="reins-${OS}-${ARCH}"
  fi

  # Install
  mkdir -p "$INSTALL_DIR"
  if [ "$OS" = "windows" ]; then
    DEST="${INSTALL_DIR}/reins.exe"
  else
    DEST="${INSTALL_DIR}/reins"
  fi

  mv "${TMPDIR}/${BINARY}" "$DEST"
  chmod +x "$DEST"

  echo "Installed reins to ${DEST}"

  # Check PATH
  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      echo ""
      echo "NOTE: ${INSTALL_DIR} is not in your PATH."
      echo "Add it with:  export PATH=\"${INSTALL_DIR}:\$PATH\""
      ;;
  esac

  echo ""
  "${DEST}" version 2>/dev/null || true

  # Offer skill installation for AI tool discovery.
  echo ""
  "${DEST}" skill || true
}

main
