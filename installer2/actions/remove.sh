#!/bin/sh
# Action: Remove b4

action_remove() {
    check_root

    log_header "Removing B4"

    # Detect platform if not set
    if [ -z "$B4_PLATFORM" ]; then
        platform_auto_detect || true
        if [ -n "$B4_PLATFORM" ]; then
            platform_call info
        fi
    fi

    # Stop running process
    stop_b4

    # Remove service
    if [ -n "$B4_PLATFORM" ]; then
        log_info "Removing service..."
        platform_call remove_service 2>/dev/null || true
    else
        # Manual cleanup of known service locations
        for svc in \
            /etc/systemd/system/b4.service \
            /etc/init.d/b4 \
            /opt/etc/init.d/S99b4; do
            if [ -f "$svc" ]; then
                rm -f "$svc"
                log_info "Removed: $svc"
            fi
        done
        command_exists systemctl && systemctl daemon-reload 2>/dev/null || true
    fi

    # Remove features
    features_remove

    # Remove binary from known locations
    for dir in /usr/local/bin /usr/bin /usr/sbin /opt/bin /opt/sbin /tmp/b4; do
        if [ -f "${dir}/${BINARY_NAME}" ]; then
            rm -f "${dir}/${BINARY_NAME}"
            rm -f "${dir}/${BINARY_NAME}".backup.* 2>/dev/null || true
            log_info "Removed binary from: ${dir}"
        fi
    done

    # Ask about config
    for cfg in /etc/b4 /opt/etc/b4; do
        if [ -d "$cfg" ]; then
            if [ "$QUIET_MODE" -eq 1 ] || confirm "Remove config directory ${cfg}?" "n"; then
                rm -rf "$cfg"
                log_info "Removed: ${cfg}"
            else
                log_info "Keeping: ${cfg}"
            fi
        fi
    done

    # Cleanup
    rm -f /var/run/b4.pid 2>/dev/null || true
    rm -f /var/log/b4.log 2>/dev/null || true

    echo ""
    log_ok "B4 has been removed"
    echo ""
}
