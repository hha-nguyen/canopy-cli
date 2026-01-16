#!/bin/sh
set -e

REPO="hha-nguyen/canopy-cli"
BINARY_NAME="canopy"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

get_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64|amd64)
            echo "amd64"
            ;;
        arm64|aarch64)
            echo "arm64"
            ;;
        *)
            echo "Unsupported architecture: $ARCH" >&2
            exit 1
            ;;
    esac
}

get_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case $OS in
        linux)
            echo "linux"
            ;;
        darwin)
            echo "darwin"
            ;;
        mingw*|msys*|cygwin*)
            echo "windows"
            ;;
        *)
            echo "Unsupported OS: $OS" >&2
            exit 1
            ;;
    esac
}

get_latest_version() {
    curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | \
        grep '"tag_name":' | \
        sed -E 's/.*"([^"]+)".*/\1/'
}

download_and_install() {
    OS=$(get_os)
    ARCH=$(get_arch)
    VERSION=${VERSION:-$(get_latest_version)}

    if [ -z "$VERSION" ]; then
        echo "Error: Could not determine version" >&2
        exit 1
    fi

    echo "Installing Canopy CLI ${VERSION} for ${OS}/${ARCH}..."

    FILENAME="${BINARY_NAME}-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        FILENAME="${FILENAME}.exe"
        ARCHIVE_EXT="zip"
    else
        ARCHIVE_EXT="tar.gz"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}.${ARCHIVE_EXT}"
    CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

    TMP_DIR=$(mktemp -d)
    trap "rm -rf ${TMP_DIR}" EXIT

    echo "Downloading ${DOWNLOAD_URL}..."
    curl -sL "${DOWNLOAD_URL}" -o "${TMP_DIR}/archive.${ARCHIVE_EXT}"
    curl -sL "${CHECKSUM_URL}" -o "${TMP_DIR}/checksums.txt" 2>/dev/null || true

    if [ -f "${TMP_DIR}/checksums.txt" ]; then
        echo "Verifying checksum..."
        EXPECTED=$(grep "${FILENAME}.${ARCHIVE_EXT}" "${TMP_DIR}/checksums.txt" | awk '{print $1}')
        if [ -n "$EXPECTED" ]; then
            ACTUAL=$(shasum -a 256 "${TMP_DIR}/archive.${ARCHIVE_EXT}" | awk '{print $1}')
            if [ "$EXPECTED" != "$ACTUAL" ]; then
                echo "Error: Checksum mismatch" >&2
                echo "Expected: ${EXPECTED}" >&2
                echo "Actual:   ${ACTUAL}" >&2
                exit 1
            fi
            echo "Checksum verified."
        fi
    fi

    echo "Extracting..."
    cd "${TMP_DIR}"
    if [ "$ARCHIVE_EXT" = "zip" ]; then
        unzip -q "archive.${ARCHIVE_EXT}"
    else
        tar -xzf "archive.${ARCHIVE_EXT}"
    fi

    BINARY_PATH="${TMP_DIR}/${BINARY_NAME}"
    if [ "$OS" = "windows" ]; then
        BINARY_PATH="${BINARY_PATH}.exe"
    fi

    if [ ! -f "${BINARY_PATH}" ]; then
        BINARY_PATH=$(find "${TMP_DIR}" -name "${BINARY_NAME}*" -type f | head -n 1)
    fi

    if [ ! -f "${BINARY_PATH}" ]; then
        echo "Error: Binary not found in archive" >&2
        exit 1
    fi

    echo "Installing to ${INSTALL_DIR}..."
    if [ -w "${INSTALL_DIR}" ]; then
        mv "${BINARY_PATH}" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        echo "Need sudo to install to ${INSTALL_DIR}"
        sudo mv "${BINARY_PATH}" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    echo ""
    echo "✓ Canopy CLI installed successfully!"
    echo ""
    echo "Run 'canopy version' to verify the installation."
    echo "Run 'canopy --help' for usage information."
}

main() {
    if ! command -v curl >/dev/null 2>&1; then
        echo "Error: curl is required but not installed" >&2
        exit 1
    fi

    download_and_install
}

main "$@"
