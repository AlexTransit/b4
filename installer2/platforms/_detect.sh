#!/bin/sh
# Auto-detection: iterate registered platforms and find the best match
#
# Override with: B4_PLATFORM=<id> environment variable

platform_auto_detect() {
    # User override — most reliable
    if [ -n "$B4_PLATFORM" ]; then
        # Verify the platform exists
        for p in $REGISTERED_PLATFORMS; do
            if [ "$p" = "$B4_PLATFORM" ]; then
                log_ok "Using user-specified platform: $B4_PLATFORM"
                return 0
            fi
        done
        log_err "Unknown platform: $B4_PLATFORM"
        log_info "Available: $REGISTERED_PLATFORMS"
        exit 1
    fi

    # Try each registered platform's match function
    for p in $REGISTERED_PLATFORMS; do
        if platform_dispatch "$p" match 2>/dev/null; then
            B4_PLATFORM="$p"
            pname=$(platform_dispatch "$p" name)
            log_ok "Detected platform: ${pname}"
            return 0
        fi
    done

    # Fallback to generic_linux if registered
    for p in $REGISTERED_PLATFORMS; do
        if [ "$p" = "generic_linux" ]; then
            B4_PLATFORM="generic_linux"
            log_warn "No specific platform matched, using Generic Linux"
            return 0
        fi
    done

    return 1
}
