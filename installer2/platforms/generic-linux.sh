#!/bin/sh
# Platform: Generic Linux (Ubuntu, Debian, Fedora, Arch, Alpine, etc.)
# Covers any systemd-based or sysv-init desktop/server Linux

platform_generic_linux_name() {
    echo "Generic Linux (Ubuntu/Debian/Fedora/Arch/Alpine)"
}

platform_generic_linux_match() {
    # Match any Linux with systemd or standard init.d
    # This is the lowest-priority fallback — other platforms should match first
    [ "$(uname -s)" = "Linux" ] || return 1

    # Don't match if this looks like a router firmware
    [ -f /etc/openwrt_release ] && return 1
    [ -f /etc/merlinwrt_release ] && return 1
    [ -d /etc/storage ] && [ -d /etc_ro ] && return 1  # Padavan

    # Match systemd or standard init
    command_exists systemctl && return 0
    [ -d /etc/init.d ] && return 0

    return 0
}

platform_generic_linux_info() {
    B4_BIN_DIR="/usr/local/bin"
    B4_DATA_DIR="/etc/b4"
    B4_CONFIG_FILE="${B4_DATA_DIR}/b4.json"

    if command_exists systemctl && systemctl list-units >/dev/null 2>&1; then
        B4_SERVICE_TYPE="systemd"
        B4_SERVICE_DIR="/etc/systemd/system"
        B4_SERVICE_NAME="b4.service"
    elif [ -d /etc/init.d ]; then
        B4_SERVICE_TYPE="sysv"
        B4_SERVICE_DIR="/etc/init.d"
        B4_SERVICE_NAME="b4"
    else
        B4_SERVICE_TYPE="none"
    fi

    detect_pkg_manager
}

platform_generic_linux_check_deps() {
    missing=""

    # Check basic tools
    if ! command_exists curl && ! command_exists wget; then
        missing="${missing} wget"
    fi
    command_exists tar || missing="${missing} tar"

    if [ -n "$missing" ]; then
        log_warn "Missing required:${missing}"
        if confirm "Install missing packages?"; then
            pkg_install $missing || log_warn "Some packages failed to install"
        else
            log_err "Cannot continue without:${missing}"
            exit 1
        fi
    fi

    ensure_https_support || exit 1

    # Check kernel modules
    _generic_linux_check_kmods

    # Recommended packages
    _generic_linux_check_recommended
}

_generic_linux_check_kmods() {
    for mod in xt_NFQUEUE xt_connbytes xt_multiport nf_conntrack; do
        if ! lsmod 2>/dev/null | grep -q "^${mod}"; then
            modprobe "$mod" 2>/dev/null || true
        fi
    done

    # Verify at least NFQUEUE is available
    if ! lsmod 2>/dev/null | grep -q "xt_NFQUEUE\|nfnetlink_queue"; then
        log_warn "xt_NFQUEUE kernel module not loaded"
        case "$B4_PKG_MANAGER" in
        apt) log_info "Try: apt install xtables-addons-common" ;;
        dnf | yum) log_info "Try: dnf install xtables-addons" ;;
        pacman) log_info "Try: pacman -S xtables-addons" ;;
        esac
    fi
}

_generic_linux_check_recommended() {
    rec_missing=""
    command_exists jq || rec_missing="${rec_missing} jq"
    command_exists iptables || command_exists nft || rec_missing="${rec_missing} iptables"

    if [ -n "$rec_missing" ]; then
        log_warn "Recommended but missing:${rec_missing}"
        if confirm "Install recommended packages?"; then
            pkg_install $rec_missing || true
        fi
    fi
}

platform_generic_linux_install_service() {
    case "$B4_SERVICE_TYPE" in
    systemd) _generic_linux_install_systemd ;;
    sysv)    _generic_linux_install_sysv ;;
    none)    log_warn "No init system detected, skipping service setup" ;;
    esac
}

_generic_linux_install_systemd() {
    cat >"${B4_SERVICE_DIR}/${B4_SERVICE_NAME}" <<EOF
[Unit]
Description=B4 DPI Bypass Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=${B4_BIN_DIR}/${BINARY_NAME} --config ${B4_CONFIG_FILE}
Restart=on-failure
RestartSec=5
TimeoutStopSec=10

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    log_ok "Systemd service created: ${B4_SERVICE_NAME}"
    log_info "  systemctl start b4"
    log_info "  systemctl enable b4  # auto-start on boot"
}

_generic_linux_install_sysv() {
    cat >"${B4_SERVICE_DIR}/${B4_SERVICE_NAME}" <<EOF
#!/bin/sh
# B4 DPI Bypass Service
PROG="${B4_BIN_DIR}/${BINARY_NAME}"
CONFIG="${B4_CONFIG_FILE}"
PIDFILE="/var/run/b4.pid"

kernel_mod_load() {
    modprobe xt_connbytes 2>/dev/null || true
    modprobe xt_NFQUEUE 2>/dev/null || true
    modprobe xt_multiport 2>/dev/null || true
}

start() {
    echo "Starting b4..."
    [ -f "\$PIDFILE" ] && kill -0 \$(cat "\$PIDFILE") 2>/dev/null && echo "Already running" && return 1
    kernel_mod_load
    nohup \$PROG --config \$CONFIG >/var/log/b4.log 2>&1 &
    echo \$! >"\$PIDFILE"
    sleep 1
    if kill -0 \$(cat "\$PIDFILE") 2>/dev/null; then
        echo "b4 started (PID: \$(cat \$PIDFILE))"
    else
        echo "b4 failed to start, check /var/log/b4.log"
        rm -f "\$PIDFILE"
        return 1
    fi
}

stop() {
    echo "Stopping b4..."
    [ -f "\$PIDFILE" ] && kill \$(cat "\$PIDFILE") 2>/dev/null
    rm -f "\$PIDFILE"
    echo "b4 stopped"
}

case "\$1" in
    start)   start ;;
    stop)    stop ;;
    restart) stop; sleep 1; start ;;
    *)       echo "Usage: \$0 {start|stop|restart}"; exit 1 ;;
esac
EOF

    chmod +x "${B4_SERVICE_DIR}/${B4_SERVICE_NAME}"
    log_ok "Init script created: ${B4_SERVICE_DIR}/${B4_SERVICE_NAME}"
}

platform_generic_linux_remove_service() {
    case "$B4_SERVICE_TYPE" in
    systemd)
        systemctl stop b4 2>/dev/null || true
        systemctl disable b4 2>/dev/null || true
        rm -f "${B4_SERVICE_DIR}/${B4_SERVICE_NAME}"
        systemctl daemon-reload
        ;;
    sysv)
        "${B4_SERVICE_DIR}/${B4_SERVICE_NAME}" stop 2>/dev/null || true
        rm -f "${B4_SERVICE_DIR}/${B4_SERVICE_NAME}"
        ;;
    esac
}

platform_generic_linux_start_service() {
    case "$B4_SERVICE_TYPE" in
    systemd)
        systemctl restart b4 2>/dev/null && log_ok "Service started" && return 0
        ;;
    sysv)
        "${B4_SERVICE_DIR}/${B4_SERVICE_NAME}" start 2>/dev/null && log_ok "Service started" && return 0
        ;;
    esac
    log_warn "Could not start service"
    return 1
}

platform_generic_linux_stop_service() {
    case "$B4_SERVICE_TYPE" in
    systemd)  systemctl stop b4 2>/dev/null ;;
    sysv)     "${B4_SERVICE_DIR}/${B4_SERVICE_NAME}" stop 2>/dev/null ;;
    esac
}

platform_generic_linux_find_storage() {
    # Standard Linux — no special storage detection needed
    return 0
}

register_platform "generic_linux"
