#!/bin/sh
# Action: Update b4 to latest version

action_update() {
    force_arch="$1"

    check_root

    log_header "Updating B4"

    # Detect platform
    if [ -z "$B4_PLATFORM" ]; then
        platform_auto_detect || true
        if [ -n "$B4_PLATFORM" ]; then
            platform_call info
        fi
    fi

    # Find existing binary
    existing_bin=""
    for dir in "$B4_BIN_DIR" /usr/local/bin /usr/bin /usr/sbin /opt/bin /opt/sbin; do
        if [ -f "${dir}/${BINARY_NAME}" ]; then
            existing_bin="${dir}/${BINARY_NAME}"
            B4_BIN_DIR="$dir"
            break
        fi
    done

    if [ -z "$existing_bin" ]; then
        log_err "B4 is not installed. Use install mode instead."
        exit 1
    fi

    # Get current version
    current_ver=$("$existing_bin" --version 2>&1 | head -1) || current_ver="unknown"
    log_info "Current: ${current_ver}"

    # Detect arch from existing binary or system
    if [ -n "$force_arch" ]; then
        B4_ARCH="$force_arch"
    else
        B4_ARCH=$(detect_architecture)
    fi

    # Get latest version
    log_info "Checking for updates..."
    latest_ver=$(get_latest_version)
    log_info "Latest: ${latest_ver}"

    if [ "$current_ver" = "$latest_ver" ] || echo "$current_ver" | grep -q "$latest_ver"; then
        log_ok "Already up to date"
        return 0
    fi

    if [ "$QUIET_MODE" -eq 0 ]; then
        if ! confirm "Update to ${latest_ver}?"; then
            log_info "Update cancelled"
            return 0
        fi
    fi

    # Download and install
    setup_temp

    file_name="${BINARY_NAME}-linux-${B4_ARCH}.tar.gz"
    download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${latest_ver}/${file_name}"
    archive_path="${TEMP_DIR}/${file_name}"

    log_info "Downloading ${latest_ver}..."
    fetch_file "$download_url" "$archive_path" || { log_err "Download failed"; exit 1; }

    # Verify
    sha_url="${download_url}.sha256"
    verify_checksum "$archive_path" "$sha_url" || true

    # Extract
    cd "$TEMP_DIR"
    tar -xzf "$archive_path" || { log_err "Extraction failed"; exit 1; }

    # Stop, backup, replace
    stop_b4

    ts=$(date '+%Y%m%d_%H%M%S')
    cp "$existing_bin" "${existing_bin}.backup.${ts}"

    mv "${TEMP_DIR}/${BINARY_NAME}" "$existing_bin" 2>/dev/null || \
        cp "${TEMP_DIR}/${BINARY_NAME}" "$existing_bin" || \
        { log_err "Failed to replace binary"; exit 1; }
    chmod +x "$existing_bin"

    # Verify
    if "$existing_bin" --version >/dev/null 2>&1; then
        new_ver=$("$existing_bin" --version 2>&1 | head -1)
        log_ok "Updated to: ${new_ver}"
        rm -f "${existing_bin}".backup.* 2>/dev/null || true
    else
        log_warn "Updated binary failed version check"
    fi

    # Restart service if it was running
    if [ -n "$B4_SERVICE_TYPE" ] && [ "$B4_SERVICE_TYPE" != "none" ]; then
        log_info "Restarting service..."
        service_call start 2>/dev/null || true
    fi

    echo ""
    log_ok "Update complete"
    echo ""
}
