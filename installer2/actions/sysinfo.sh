#!/bin/sh
# Action: Show system diagnostics

action_sysinfo() {
    log_header "B4 System Diagnostics"
    log_sep

    # OS info
    log_detail "OS" "$(uname -s) $(uname -r)"
    log_detail "Architecture" "$(uname -m)"
    [ -f /etc/os-release ] && log_detail "Distribution" "$(. /etc/os-release && echo "$PRETTY_NAME")"
    [ -f /etc/openwrt_release ] && log_detail "OpenWrt" "$(. /etc/openwrt_release && echo "$DISTRIB_DESCRIPTION")"

    # Platform detection
    platform_auto_detect 2>/dev/null || true
    if [ -n "$B4_PLATFORM" ]; then
        pname=$(platform_dispatch "$B4_PLATFORM" name 2>/dev/null)
        log_detail "Detected platform" "${pname} (${B4_PLATFORM})"
        platform_call info 2>/dev/null || true
        log_detail "Binary dir" "${B4_BIN_DIR}"
        log_detail "Data dir" "${B4_DATA_DIR}"
        log_detail "Service type" "${B4_SERVICE_TYPE}"
    fi

    log_sep

    # B4 installation status
    found_bin=""
    for dir in /usr/local/bin /usr/bin /usr/sbin /opt/bin /opt/sbin /tmp/b4; do
        if [ -f "${dir}/${BINARY_NAME}" ]; then
            found_bin="${dir}/${BINARY_NAME}"
            ver=$("$found_bin" --version 2>&1 | head -1) || ver="unknown"
            log_detail "B4 binary" "${found_bin} (${ver})"
            break
        fi
    done
    [ -z "$found_bin" ] && log_detail "B4 binary" "${RED}not found${NC}"

    if is_b4_running; then
        log_detail "B4 status" "${GREEN}running${NC}"
    else
        log_detail "B4 status" "${YELLOW}not running${NC}"
    fi

    # Config
    for cfg in /etc/b4/b4.json /opt/etc/b4/b4.json; do
        [ -f "$cfg" ] && log_detail "Config" "$cfg" && break
    done

    log_sep

    # Kernel modules
    echo ""
    log_info "Kernel modules:"
    for mod in xt_NFQUEUE nfnetlink_queue xt_connbytes xt_multiport nf_conntrack; do
        if lsmod 2>/dev/null | grep -q "^${mod}"; then
            printf "    ${GREEN}loaded${NC}  %s\n" "$mod" >&2
        else
            printf "    ${RED}missing${NC} %s\n" "$mod" >&2
        fi
    done

    # Network tools
    echo ""
    log_info "Network tools:"
    for tool in iptables nft curl wget jq tar sha256sum; do
        if command_exists "$tool"; then
            printf "    ${GREEN}found${NC}   %s\n" "$tool" >&2
        else
            printf "    ${YELLOW}missing${NC} %s\n" "$tool" >&2
        fi
    done

    # Package manager
    echo ""
    detect_pkg_manager
    log_detail "Package manager" "${B4_PKG_MANAGER:-none}"

    # HTTPS support
    if check_https_support 2>/dev/null; then
        log_detail "HTTPS support" "${GREEN}yes${NC}"
    else
        log_detail "HTTPS support" "${RED}no${NC}"
    fi

    # Storage
    echo ""
    log_info "Storage:"
    for dir in / /opt /tmp /jffs /mnt/sda1 /etc/storage; do
        if [ -d "$dir" ]; then
            avail=$(df -h "$dir" 2>/dev/null | tail -1 | awk '{print $4}')
            writable="rw"
            [ ! -w "$dir" ] && writable="ro"
            printf "    %-15s %s available (%s)\n" "$dir" "${avail:-?}" "$writable" >&2
        fi
    done

    echo ""
    log_sep
}
