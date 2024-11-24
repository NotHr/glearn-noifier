#!/bin/bash
set -e

# Function to output error messages
error() {
    echo "Error: $1" >&2
    exit 1
}

# Function to detect OS and architecture
detect_platform() {
    local os arch

    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        MINGW*|MSYS*|CYGWIN*) os="windows";;
        *)          error "Unsupported operating system: $(uname -s)";;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64";;
        arm64|aarch64) arch="arm64";;
        i386|i686)    arch="386";;
        *)            error "Unsupported architecture: $(uname -m)";;
    esac

    echo "${os}_${arch}"
}

# Function to download and install
install_glearn() {
    local platform="$1"
    local repo="nothr/glearn-notifier"
    local tmp_dir
    tmp_dir=$(mktemp -d)

    echo "Detecting latest version..."
    # Get the latest release URL
    if ! command -v curl &> /dev/null; then
        error "curl is required but not installed"
    fi

    if [ "$platform" = "windows_amd64" ] || [ "$platform" = "windows_386" ]; then
        ext="zip"
    else
        ext="tar.gz"
    fi

    echo "Downloading latest release for $platform..."
    release_url=$(curl -s "https://api.github.com/repos/$repo/releases/latest" | \
                 grep "browser_download_url.*${platform}.${ext}" | \
                 cut -d '"' -f 4)

    if [ -z "$release_url" ]; then
        error "Could not find release for platform: $platform"
    fi

    # Download the release
    echo "Downloading from $release_url..."
    curl -L "$release_url" -o "$tmp_dir/glearn-notifier.$ext"

    # Extract the archive
    echo "Extracting archive..."
    cd "$tmp_dir"
    case "$ext" in
        "tar.gz")
            tar xzf "glearn-notifier.$ext"
            ;;
        "zip")
            if ! command -v unzip &> /dev/null; then
                error "unzip is required but not installed"
            fi
            unzip "glearn-notifier.$ext"
            ;;
    esac

    # Install the binary
    if [ "$platform" = "windows_amd64" ] || [ "$platform" = "windows_386" ]; then
        echo "Binary extracted to $tmp_dir/glearn-notifier.exe"
        echo "Please move it to a directory in your PATH"
    else
        echo "Installing to /usr/local/bin/glearn-notifier..."
        sudo mv glearn-notifier /usr/local/bin/
        sudo chmod +x /usr/local/bin/glearn-notifier
    fi

    # Cleanup
    cd - > /dev/null
    rm -rf "$tmp_dir"
    echo "Installation completed successfully!"
}

# Main script
main() {
    echo "Detecting platform..."
    platform=$(detect_platform)
    echo "Detected platform: $platform"
    install_glearn "$platform"
}

main
