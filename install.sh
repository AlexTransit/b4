#!/bin/sh
# B4 Installer — Universal Linux installer with wizard interface
# Supports desktop Linux, OpenWRT, MerlinWRT, Keenetic, Mikrotik, Docker, and more
#
# AUTO-GENERATED — Do not edit directly
# Edit files in installer2/ and run: make build-installer
#

set -e

# Ensure sane PATH (Entware paths first for wget-ssl/curl from /opt/bin)
export PATH="/opt/bin:/opt/sbin:$HOME/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin:$PATH"


# ======== lib/colors.sh ========
# Terminal colors (disabled when not a TTY)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    CYAN='\033[0;36m'
    MAGENTA='\033[0;35m'
    BOLD='\033[1m'
    DIM='\033[2m'
    NC='\033[0m'
else
    RED='' GREEN='' YELLOW='' BLUE='' CYAN='' MAGENTA='' BOLD='' DIM='' NC=''
fi


# ======== lib/log.sh ========
# Logging functions

QUIET_MODE=0

log_info() {
    [ "$QUIET_MODE" -eq 1 ] && return
    printf "${BLUE}[INFO]${NC} %s\n" "$1" >&2
}

log_ok() {
    [ "$QUIET_MODE" -eq 1 ] && return
    printf "${GREEN}[ OK ]${NC} %s\n" "$1" >&2
}

log_warn() {
    [ "$QUIET_MODE" -eq 1 ] && return
    printf "${YELLOW}[WARN]${NC} %s\n" "$1" >&2
}

log_err() {
    printf "${RED}[ERR ]${NC} %s\n" "$1" >&2
}

log_header() {
    [ "$QUIET_MODE" -eq 1 ] && return
    printf "\n${MAGENTA}${BOLD}%s${NC}\n" "$1" >&2
}

log_detail() {
    [ "$QUIET_MODE" -eq 1 ] && return
    printf "  ${CYAN}%-22s${NC}: %b\n" "$1" "$2" >&2
}

# Print a separator line
log_sep() {
    [ "$QUIET_MODE" -eq 1 ] && return
    printf "${DIM}%s${NC}\n" "─────────────────────────────────────────" >&2
}


# ======== lib/utils.sh ========
# Core utility functions

# --- Configuration ---
REPO_OWNER="DanielLavrushin"
REPO_NAME="b4"
BINARY_NAME="b4"
TEMP_DIR="/tmp/b4_install_$$"
WGET_INSECURE=""
PROXY_BASE_URL="https://proxy.lavrush.in/github"

# --- Runtime state (set by platform/wizard) ---
B4_BIN_DIR=""
B4_DATA_DIR=""
B4_CONFIG_FILE=""
B4_SERVICE_TYPE=""
B4_SERVICE_DIR=""
B4_SERVICE_NAME=""
B4_PKG_MANAGER=""
B4_PLATFORM=""

# --- Command existence check (works on BusyBox/minimal shells) ---
command_exists() {
    command -v "$1" >/dev/null 2>&1 || which "$1" >/dev/null 2>&1
}

# --- Root check ---
check_root() {
    if [ "$(id -u 2>/dev/null)" = "0" ]; then
        return 0
    fi
    if [ "$USER" = "root" ]; then
        return 0
    fi
    # Fallback: try writing to /etc
    if touch /etc/.b4_root_test 2>/dev/null; then
        rm -f /etc/.b4_root_test
        return 0
    fi
    log_err "This script must be run as root"
    exit 1
}

# --- Temp directory management ---
setup_temp() {
    rm -rf "$TEMP_DIR" 2>/dev/null || true
    mkdir -p "$TEMP_DIR" || { log_err "Cannot create temp dir"; exit 1; }
}

cleanup_temp() {
    rm -rf "$TEMP_DIR" 2>/dev/null || true
}

trap cleanup_temp EXIT INT TERM

# --- Package manager detection ---
detect_pkg_manager() {
    if [ -n "$B4_PKG_MANAGER" ]; then
        return 0
    fi
    if command_exists apt-get; then
        B4_PKG_MANAGER="apt"
    elif command_exists dnf; then
        B4_PKG_MANAGER="dnf"
    elif command_exists yum; then
        B4_PKG_MANAGER="yum"
    elif command_exists pacman; then
        B4_PKG_MANAGER="pacman"
    elif command_exists apk; then
        B4_PKG_MANAGER="apk"
    elif command_exists opkg; then
        B4_PKG_MANAGER="opkg"
    fi
}

pkg_install() {
    detect_pkg_manager
    case "$B4_PKG_MANAGER" in
    apt)    apt-get update -qq >/dev/null 2>&1; apt-get install -y -qq "$@" >/dev/null 2>&1 ;;
    dnf)    dnf install -y -q "$@" >/dev/null 2>&1 ;;
    yum)    yum install -y -q "$@" >/dev/null 2>&1 ;;
    pacman) pacman -S --noconfirm --needed "$@" >/dev/null 2>&1 ;;
    apk)    apk add --quiet "$@" >/dev/null 2>&1 ;;
    opkg)   opkg update >/dev/null 2>&1; opkg install "$@" >/dev/null 2>&1 ;;
    *)      log_warn "No package manager detected"; return 1 ;;
    esac
}

# --- Architecture detection ---
detect_architecture() {
    arch=$(uname -m)

    case "$arch" in
    x86_64 | amd64)         echo "amd64" ;;
    i386 | i486 | i586 | i686) echo "386" ;;
    aarch64 | arm64)        echo "arm64" ;;
    armv7 | armv7l)
        # Check for full ARMv7 VFP support, otherwise use armv5 for safety
        if [ -f /proc/cpuinfo ] &&
           grep -qE "(vfpv[3-9])" /proc/cpuinfo 2>/dev/null &&
           grep -qE "CPU architecture:\s*7" /proc/cpuinfo 2>/dev/null; then
            echo "armv7"
        else
            echo "armv5"
        fi
        ;;
    armv6*)                 echo "armv6" ;;
    armv5*)                 echo "armv5" ;;
    arm*)
        if [ -f /proc/cpuinfo ]; then
            if grep -qE "CPU architecture:\s*7" /proc/cpuinfo 2>/dev/null; then echo "armv7"
            elif grep -qE "CPU architecture:\s*6" /proc/cpuinfo 2>/dev/null; then echo "armv6"
            else echo "armv5"
            fi
        else
            echo "armv5"
        fi
        ;;
    mips64*)
        variant="mips64"
        if is_little_endian; then variant="mips64le"; fi
        if is_softfloat; then variant="${variant}_softfloat"; fi
        echo "$variant"
        ;;
    mips*)
        variant="mips"
        if is_little_endian; then variant="mipsle"; fi
        if is_softfloat; then variant="${variant}_softfloat"; fi
        echo "$variant"
        ;;
    ppc64le)    echo "ppc64le" ;;
    ppc64)      echo "ppc64" ;;
    riscv64)    echo "riscv64" ;;
    s390x)      echo "s390x" ;;
    loongarch64) echo "loong64" ;;
    *) log_err "Unsupported architecture: $arch"; exit 1 ;;
    esac
}

is_little_endian() {
    uname -m | grep -qi "el" && return 0
    [ -f /proc/cpuinfo ] && grep -qi "little.endian\|byteorder.*little" /proc/cpuinfo 2>/dev/null && return 0
    command_exists opkg && opkg print-architecture 2>/dev/null | grep -qi "mipsel\|mips64el" && return 0
    # ELF header check
    [ "$(dd if=/bin/sh bs=1 skip=5 count=1 2>/dev/null | od -An -tx1 | tr -d ' ')" = "01" ] && return 0
    return 1
}

is_softfloat() {
    [ -f /proc/cpuinfo ] || return 1
    ! grep -qi "fpu" /proc/cpuinfo 2>/dev/null && return 0
    grep -qi "nofpu\|no fpu" /proc/cpuinfo 2>/dev/null && return 0
    return 1
}

# --- HTTPS support ---
check_https_support() {
    if command_exists curl && curl -sI --max-time 5 "https://github.com" >/dev/null 2>&1; then
        return 0
    fi
    if command_exists wget && wget --spider -q --timeout=5 "https://github.com" 2>/dev/null; then
        return 0
    fi
    # Try with --no-check-certificate
    if command_exists wget && wget --spider -q --timeout=5 --no-check-certificate "https://github.com" 2>/dev/null; then
        WGET_INSECURE="--no-check-certificate"
        log_warn "HTTPS works only with --no-check-certificate (CA certs missing)"
        return 0
    fi
    return 1
}

ensure_https_support() {
    if check_https_support; then
        return 0
    fi
    log_warn "HTTPS not available — trying to install SSL support"
    if command_exists opkg; then
        opkg update >/dev/null 2>&1 || true
        opkg install ca-certificates >/dev/null 2>&1 || true
        opkg install wget-ssl >/dev/null 2>&1 || true
        hash -r 2>/dev/null || true
        if check_https_support; then return 0; fi
    fi
    log_err "HTTPS not available. Cannot download from GitHub."
    log_info "On Entware/OpenWrt: opkg install wget-ssl ca-certificates"
    return 1
}

# --- Download helpers ---
convert_to_proxy_url() {
    url="$1"
    case "$url" in
    https://raw.githubusercontent.com/${REPO_OWNER}/* | \
    https://github.com/${REPO_OWNER}/* | \
    https://api.github.com/repos/${REPO_OWNER}/*)
        echo "${PROXY_BASE_URL}/${url}" ;;
    *) echo "$url" ;;
    esac
}

fetch_file() {
    url="$1"
    output="$2"

    if ! command_exists curl && ! command_exists wget; then
        log_err "Neither curl nor wget found"
        return 1
    fi

    # Try direct
    if command_exists curl && curl -sfL --max-time 30 -o "$output" "$url" 2>/dev/null; then return 0; fi
    if command_exists wget && wget -q $WGET_INSECURE --timeout=30 -O "$output" "$url" 2>/dev/null; then return 0; fi

    # Try proxy fallback
    proxy_url=$(convert_to_proxy_url "$url")
    if [ "$proxy_url" != "$url" ]; then
        log_warn "Direct download failed, trying proxy..."
        if command_exists curl && curl -sfL --max-time 30 -o "$output" "$proxy_url" 2>/dev/null; then return 0; fi
        if command_exists wget && wget -q $WGET_INSECURE --timeout=30 -O "$output" "$proxy_url" 2>/dev/null; then return 0; fi
    fi

    log_err "Failed to download: $url"
    return 1
}

fetch_stdout() {
    url="$1"

    if command_exists curl; then
        result=$(curl -sfL --max-time 15 "$url" 2>/dev/null) && [ -n "$result" ] && echo "$result" && return 0
    fi
    if command_exists wget; then
        result=$(wget -qO- $WGET_INSECURE --timeout=15 "$url" 2>/dev/null) && [ -n "$result" ] && echo "$result" && return 0
    fi

    # Proxy fallback
    proxy_url=$(convert_to_proxy_url "$url")
    if [ "$proxy_url" != "$url" ]; then
        if command_exists curl; then
            result=$(curl -sfL --max-time 15 "$proxy_url" 2>/dev/null) && [ -n "$result" ] && echo "$result" && return 0
        fi
        if command_exists wget; then
            result=$(wget -qO- $WGET_INSECURE --timeout=15 "$proxy_url" 2>/dev/null) && [ -n "$result" ] && echo "$result" && return 0
        fi
    fi

    return 1
}

# --- GitHub release helpers ---
get_latest_version() {
    api_url="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
    version=$(fetch_stdout "$api_url" | grep -o '"tag_name": *"[^"]*"' | head -1 | cut -d'"' -f4)
    if [ -z "$version" ]; then
        log_err "Failed to fetch latest version"
        exit 1
    fi
    echo "$version"
}

verify_checksum() {
    file="$1"
    checksum_url="$2"
    checksum_file="${file}.sha256"

    if ! fetch_file "$checksum_url" "$checksum_file"; then
        rm -f "$checksum_file"
        return 1
    fi

    expected=$(awk '{print $1}' "$checksum_file")
    rm -f "$checksum_file"
    [ -z "$expected" ] && return 1

    if ! command_exists sha256sum; then
        log_warn "sha256sum not found, skipping verification"
        return 1
    fi

    actual=$(sha256sum "$file" | awk '{print $1}')
    if [ "$expected" = "$actual" ]; then
        log_ok "SHA256 verified: $actual"
        return 0
    else
        log_err "SHA256 mismatch! Expected: $expected Got: $actual"
        return 2
    fi
}

# --- Process management ---
is_b4_running() {
    if command_exists pgrep; then
        pgrep -x "$BINARY_NAME" >/dev/null 2>&1
    else
        ps 2>/dev/null | grep -v grep | grep -q "[/]${BINARY_NAME}\$\|[/]${BINARY_NAME}[[:space:]]"
    fi
}

stop_b4() {
    if ! is_b4_running; then return 0; fi
    log_info "Stopping running b4 process..."
    if command_exists pkill; then
        pkill -x "$BINARY_NAME" 2>/dev/null || true
    else
        ps 2>/dev/null | grep -v grep | grep "${BINARY_NAME}" | awk '{print $1}' | while read pid; do
            kill "$pid" 2>/dev/null || true
        done
    fi
    sleep 2
}

# --- Directory helpers ---
is_writable_dir() {
    dir="$1"
    [ -d "$dir" ] && [ -w "$dir" ] && return 0
    # Try to create and test
    mkdir -p "$dir" 2>/dev/null && [ -w "$dir" ] && return 0
    return 1
}

ensure_dir() {
    dir="$1"
    label="$2"
    if ! mkdir -p "$dir" 2>/dev/null; then
        log_err "Cannot create ${label}: ${dir}"
        return 1
    fi
    if [ ! -w "$dir" ]; then
        log_err "${label} not writable: ${dir}"
        return 1
    fi
    return 0
}

# --- Read user input (works even when stdin is piped) ---
read_input() {
    prompt="$1"
    default="$2"
    printf "${CYAN}%b${NC}" "$prompt" >&2
    read answer </dev/tty 2>/dev/null || answer="$default"
    [ -z "$answer" ] && answer="$default"
    echo "$answer"
}

# --- Yes/No prompt ---
confirm() {
    prompt="$1"
    default="${2:-y}" # default yes

    if [ "$default" = "y" ]; then
        hint="Y/n"
    else
        hint="y/N"
    fi

    answer=$(read_input "${prompt} (${hint}): " "$default")

    case "$answer" in
    [yY] | [yY][eE][sS]) return 0 ;;
    [nN] | [nN][oO]) return 1 ;;
    *) [ "$default" = "y" ] && return 0 || return 1 ;;
    esac
}


# ======== lib/wizard.sh ========
# Interactive wizard — handles both manual and automatic modes

WIZARD_MODE="" # "auto" or "manual"

# Show the welcome banner and mode selection
wizard_start() {
    echo ""
    printf "${BOLD}"
    echo "  ╔═══════════════════════════════════════╗"
    echo "  ║       B4 Universal Installer          ║"
    echo "  ╚═══════════════════════════════════════╝"
    printf "${NC}"
    echo ""

    log_sep
    echo ""
    printf "  ${BOLD}1${NC}) Automatic detection ${DIM}(recommended)${NC}\n"
    printf "  ${BOLD}2${NC}) Manual configuration\n"
    echo ""

    choice=$(read_input "Select mode [1]: " "1")

    case "$choice" in
    2) WIZARD_MODE="manual" ;;
    *) WIZARD_MODE="auto" ;;
    esac
}

# Run automatic detection and show results for review
wizard_auto_detect() {
    log_header "Detecting system..."
    echo ""

    # 1. Detect platform
    platform_auto_detect
    if [ -z "$B4_PLATFORM" ]; then
        log_err "Could not detect platform"
        log_info "Try manual mode or set B4_PLATFORM environment variable"
        exit 1
    fi

    # 2. Load platform defaults
    platform_call info

    # 3. Detect architecture
    B4_ARCH=$(detect_architecture)

    # 4. Detect package manager
    detect_pkg_manager

    # 5. Show what was detected
    wizard_show_config

    echo ""
    if ! confirm "Proceed with these settings?"; then
        log_info "Switching to manual mode..."
        WIZARD_MODE="manual"
        wizard_manual_configure
    fi
}

# Manual configuration — ask for every setting
wizard_manual_configure() {
    log_header "Manual configuration"
    echo ""

    # 1. Platform selection
    echo "  Available platforms:"
    idx=1
    for p in $REGISTERED_PLATFORMS; do
        pname=$(platform_dispatch "$p" name)
        printf "    ${BOLD}%d${NC}) %s\n" "$idx" "$pname"
        idx=$((idx + 1))
    done
    echo ""

    choice=$(read_input "Select platform [1]: " "1")
    idx=1
    for p in $REGISTERED_PLATFORMS; do
        if [ "$idx" = "$choice" ]; then
            B4_PLATFORM="$p"
            break
        fi
        idx=$((idx + 1))
    done

    # Load platform defaults first
    platform_call info

    # 2. Binary directory
    B4_BIN_DIR=$(read_input "Binary directory [${B4_BIN_DIR}]: " "$B4_BIN_DIR")

    # 3. Data/config directory
    B4_DATA_DIR=$(read_input "Data directory [${B4_DATA_DIR}]: " "$B4_DATA_DIR")
    B4_CONFIG_FILE="${B4_DATA_DIR}/b4.json"

    # 4. Service type
    echo ""
    echo "  Service types: systemd, procd, sysv, entware, none"
    B4_SERVICE_TYPE=$(read_input "Service type [${B4_SERVICE_TYPE}]: " "$B4_SERVICE_TYPE")

    # 5. Architecture
    auto_arch=$(detect_architecture)
    B4_ARCH=$(read_input "Architecture [${auto_arch}]: " "$auto_arch")

    # 6. Package manager
    detect_pkg_manager
    B4_PKG_MANAGER=$(read_input "Package manager [${B4_PKG_MANAGER:-none}]: " "$B4_PKG_MANAGER")

    echo ""
    wizard_show_config
    echo ""
    if ! confirm "Proceed with these settings?"; then
        log_info "Aborted."
        exit 0
    fi
}

# Display current configuration
wizard_show_config() {
    log_sep
    pname=""
    if [ -n "$B4_PLATFORM" ]; then
        pname=$(platform_dispatch "$B4_PLATFORM" name)
    fi
    log_detail "Platform" "${BOLD}${pname}${NC} (${B4_PLATFORM})"
    log_detail "Architecture" "${B4_ARCH}"
    log_detail "Binary directory" "${B4_BIN_DIR}"
    log_detail "Data directory" "${B4_DATA_DIR}"
    log_detail "Config file" "${B4_CONFIG_FILE}"
    log_detail "Service type" "${B4_SERVICE_TYPE}"
    log_detail "Package manager" "${B4_PKG_MANAGER:-none}"

    # Show enabled features
    if [ -n "$REGISTERED_FEATURES" ]; then
        echo ""
        log_detail "Features" ""
        for f in $REGISTERED_FEATURES; do
            fname=$(feature_dispatch "$f" name)
            fdesc=$(feature_dispatch "$f" description)
            printf "    ${GREEN}+${NC} %s ${DIM}— %s${NC}\n" "$fname" "$fdesc" >&2
        done
    fi
    log_sep
}

# Feature selection wizard (called during install)
wizard_select_features() {
    if [ -z "$REGISTERED_FEATURES" ]; then
        return 0
    fi

    log_header "Optional features"
    echo ""

    for f in $REGISTERED_FEATURES; do
        fname=$(feature_dispatch "$f" name)
        fdesc=$(feature_dispatch "$f" description)
        fdefault=$(feature_dispatch "$f" default_enabled)

        if [ "$fdefault" = "yes" ]; then
            def="y"
        else
            def="n"
        fi

        if confirm "  Enable ${BOLD}${fname}${NC}? ${DIM}(${fdesc})${NC}" "$def"; then
            ENABLED_FEATURES="${ENABLED_FEATURES} ${f}"
        fi
    done
}


# ======== platforms/_interface.sh ========
# Platform registration and dispatch system
#
# Each platform file must define these functions (prefixed with platform_<id>_):
#   name             — Human-readable name
#   match            — Return 0 if this platform is detected, 1 otherwise
#   info             — Set B4_BIN_DIR, B4_DATA_DIR, B4_SERVICE_TYPE, etc.
#   check_deps       — Verify/install kernel modules and dependencies
#   install_service  — Write the service/init script
#   remove_service   — Remove the service/init script
#   start_service    — Start the b4 service
#   stop_service     — Stop the b4 service
#   find_storage     — Find writable storage (for routers with limited rootfs)
#
# Then register with: register_platform "<id>"

REGISTERED_PLATFORMS=""

register_platform() {
    id="$1"
    REGISTERED_PLATFORMS="${REGISTERED_PLATFORMS} ${id}"
}

# Dispatch a call to the active platform
# Usage: platform_call <function> [args...]
platform_call() {
    func="$1"
    shift
    platform_dispatch "$B4_PLATFORM" "$func" "$@"
}

# Dispatch to a specific platform
# Usage: platform_dispatch <platform_id> <function> [args...]
platform_dispatch() {
    pid="$1"
    func="$2"
    shift 2
    # Build function name: platform_<id>_<func>
    fn="platform_${pid}_${func}"
    if type "$fn" >/dev/null 2>&1; then
        "$fn" "$@"
    else
        log_warn "Platform '${pid}' does not implement '${func}'"
        return 1
    fi
}


# ======== platforms/_detect.sh ========
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


# ======== platforms/generic-linux.sh ========
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


# ======== features/_interface.sh ========
# Feature registration and dispatch system
#
# Each feature file must define these functions (prefixed with feature_<id>_):
#   name             — Human-readable name
#   description      — Short description
#   default_enabled  — "yes" or "no"
#   run              — Execute the feature (install/configure)
#   remove           — Undo/clean up the feature
#
# Then register with: register_feature "<id>"

REGISTERED_FEATURES=""
ENABLED_FEATURES=""

register_feature() {
    id="$1"
    REGISTERED_FEATURES="${REGISTERED_FEATURES} ${id}"
}

# Dispatch to a specific feature
feature_dispatch() {
    fid="$1"
    func="$2"
    shift 2
    fn="feature_${fid}_${func}"
    if type "$fn" >/dev/null 2>&1; then
        "$fn" "$@"
    else
        log_warn "Feature '${fid}' does not implement '${func}'"
        return 1
    fi
}

# Run all enabled features
features_run() {
    for f in $ENABLED_FEATURES; do
        fname=$(feature_dispatch "$f" name)
        log_header "Feature: ${fname}"
        feature_dispatch "$f" run || log_warn "Feature '${fname}' had issues"
    done
}

# Remove all registered features
features_remove() {
    for f in $REGISTERED_FEATURES; do
        feature_dispatch "$f" remove 2>/dev/null || true
    done
}


# ======== features/geodat.sh ========
# Feature: GeoData files (geosite.dat + geoip.dat)
# Downloads v2ray-format geo databases for domain/IP categorization

GEODAT_SOURCES="1|Loyalsoldier|https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download
2|RUNET Freedom (recommended)|https://raw.githubusercontent.com/runetfreedom/russia-v2ray-rules-dat/release
3|Nidelon|https://github.com/Nidelon/ru-block-v2ray-rules/releases/latest/download
4|DustinWin|https://github.com/DustinWin/ruleset_geodata/releases/download/mihomo
5|Chocolate4U|https://raw.githubusercontent.com/Chocolate4U/Iran-v2ray-rules/release"

feature_geodat_name() {
    echo "GeoData files"
}

feature_geodat_description() {
    echo "Download geosite.dat & geoip.dat for domain/IP filtering"
}

feature_geodat_default_enabled() {
    echo "yes"
}

feature_geodat_run() {
    log_sep
    echo ""

    # Select source
    echo "  Available geodata sources:"
    echo "$GEODAT_SOURCES" | while IFS='|' read -r num name _url; do
        [ -n "$num" ] && printf "    ${BOLD}%s${NC}) %s\n" "$num" "$name"
    done
    echo ""

    choice=$(read_input "Select source [2]: " "2")

    base_url=$(echo "$GEODAT_SOURCES" | grep "^${choice}|" | cut -d'|' -f3)
    if [ -z "$base_url" ]; then
        log_warn "Invalid selection, using default"
        base_url=$(echo "$GEODAT_SOURCES" | grep "^2|" | cut -d'|' -f3)
    fi

    # Destination directory
    save_dir="$B4_DATA_DIR"

    # Check if config already has a geodat path
    if [ -f "$B4_CONFIG_FILE" ] && command_exists jq; then
        existing=$(jq -r '.system.geo.sitedat_path // empty' "$B4_CONFIG_FILE" 2>/dev/null)
        if [ -n "$existing" ] && [ "$existing" != "null" ]; then
            save_dir=$(dirname "$existing")
            log_info "Found existing geodat path: $save_dir"
        fi
    fi

    save_dir=$(read_input "Save directory [${save_dir}]: " "$save_dir")

    ensure_dir "$save_dir" "Geodat directory" || return 1

    # Download files
    log_info "Downloading geosite.dat..."
    if ! fetch_file "${base_url}/geosite.dat" "${save_dir}/geosite.dat"; then
        log_err "Failed to download geosite.dat"
        return 1
    fi
    [ ! -s "${save_dir}/geosite.dat" ] && log_err "geosite.dat is empty" && return 1

    log_info "Downloading geoip.dat..."
    if ! fetch_file "${base_url}/geoip.dat" "${save_dir}/geoip.dat"; then
        log_err "Failed to download geoip.dat"
        return 1
    fi
    [ ! -s "${save_dir}/geoip.dat" ] && log_err "geoip.dat is empty" && return 1

    log_ok "GeoData downloaded to ${save_dir}"

    # Update config
    _geodat_update_config "${save_dir}/geosite.dat" "${save_dir}/geoip.dat" "$base_url"
}

_geodat_update_config() {
    sitedat_path="$1"
    ipdat_path="$2"
    base_url="$3"

    if ! command_exists jq; then
        log_warn "jq not found — please update config manually:"
        log_info "  Set system.geo.sitedat_path = $sitedat_path"
        log_info "  Set system.geo.ipdat_path = $ipdat_path"
        return 0
    fi

    if [ ! -f "$B4_CONFIG_FILE" ]; then
        # Create minimal config
        jq -n \
            --arg sp "$sitedat_path" \
            --arg su "${base_url}/geosite.dat" \
            --arg ip "$ipdat_path" \
            --arg iu "${base_url}/geoip.dat" \
            '{ system: { geo: { sitedat_path: $sp, sitedat_url: $su, ipdat_path: $ip, ipdat_url: $iu } } }' \
            >"$B4_CONFIG_FILE"
        log_ok "Created config with geodat paths"
        return 0
    fi

    # Update existing config
    tmp="${B4_CONFIG_FILE}.tmp"
    if jq \
        --arg sp "$sitedat_path" \
        --arg su "${base_url}/geosite.dat" \
        --arg ip "$ipdat_path" \
        --arg iu "${base_url}/geoip.dat" \
        '.system.geo = (.system.geo // {}) + { sitedat_path: $sp, sitedat_url: $su, ipdat_path: $ip, ipdat_url: $iu }' \
        "$B4_CONFIG_FILE" >"$tmp" 2>/dev/null; then
        mv "$tmp" "$B4_CONFIG_FILE"
        log_ok "Config updated with geodat paths"
    else
        rm -f "$tmp"
        log_warn "Failed to update config, please set paths manually"
    fi
}

feature_geodat_remove() {
    # Don't remove geodata files on uninstall — user may want them
    return 0
}

register_feature "geodat"


# ======== features/https.sh ========
# Feature: HTTPS for B4 web interface
# Detects existing TLS certificates on the system and configures b4 to use them

feature_https_name() {
    echo "HTTPS web interface"
}

feature_https_description() {
    echo "Enable HTTPS for B4 web UI using detected TLS certificates"
}

feature_https_default_enabled() {
    # Only suggest if certificates exist
    _https_detect_certs >/dev/null 2>&1 && echo "yes" || echo "no"
}

feature_https_run() {
    cert_info=$(_https_detect_certs)
    if [ -z "$cert_info" ]; then
        log_info "No TLS certificates found on this system"
        log_info "You can configure HTTPS later in B4 Web UI > Settings > Web Server"
        return 0
    fi

    cert_path=$(echo "$cert_info" | cut -d'|' -f1)
    key_path=$(echo "$cert_info" | cut -d'|' -f2)
    cert_source=$(echo "$cert_info" | cut -d'|' -f3)

    log_info "Found TLS certificate: ${cert_source}"
    log_detail "Certificate" "$cert_path"
    log_detail "Key" "$key_path"

    if ! confirm "Enable HTTPS with this certificate?"; then
        return 0
    fi

    if ! command_exists jq; then
        log_warn "jq not found — please update config manually:"
        log_info "  Set system.web_server.tls_cert = $cert_path"
        log_info "  Set system.web_server.tls_key = $key_path"
        return 0
    fi

    if [ ! -f "$B4_CONFIG_FILE" ]; then
        ensure_dir "$(dirname "$B4_CONFIG_FILE")" "Config directory" || return 1
        jq -n \
            --arg cert "$cert_path" \
            --arg key "$key_path" \
            '{ system: { web_server: { tls_cert: $cert, tls_key: $key } } }' \
            >"$B4_CONFIG_FILE"
    else
        tmp="${B4_CONFIG_FILE}.tmp"
        if jq --arg cert "$cert_path" --arg key "$key_path" \
            '.system.web_server.tls_cert = $cert | .system.web_server.tls_key = $key' \
            "$B4_CONFIG_FILE" >"$tmp" 2>/dev/null; then
            mv "$tmp" "$B4_CONFIG_FILE"
        else
            rm -f "$tmp"
            log_warn "Failed to update config"
            return 1
        fi
    fi

    log_ok "HTTPS enabled"
}

_https_detect_certs() {
    # Common certificate locations on various systems
    if [ -f "/etc/uhttpd.crt" ] && [ -f "/etc/uhttpd.key" ]; then
        echo "/etc/uhttpd.crt|/etc/uhttpd.key|OpenWrt uhttpd"
        return 0
    fi
    if [ -f "/etc/cert.pem" ] && [ -f "/etc/key.pem" ]; then
        echo "/etc/cert.pem|/etc/key.pem|System default"
        return 0
    fi
    if [ -f "/etc/ssl/certs/server.crt" ] && [ -f "/etc/ssl/private/server.key" ]; then
        echo "/etc/ssl/certs/server.crt|/etc/ssl/private/server.key|System SSL"
        return 0
    fi
    return 1
}

feature_https_remove() {
    return 0
}

register_feature "https"


# ======== actions/install.sh ========
# Action: Install b4

action_install() {
    version="$1"
    force_arch="$2"

    check_root

    # --- Wizard ---
    if [ "$QUIET_MODE" -eq 1 ]; then
        WIZARD_MODE="auto"
        platform_auto_detect
        platform_call info
        B4_ARCH="${force_arch:-$(detect_architecture)}"
        detect_pkg_manager
        # Enable all default features in quiet mode
        for f in $REGISTERED_FEATURES; do
            fdefault=$(feature_dispatch "$f" default_enabled)
            [ "$fdefault" = "yes" ] && ENABLED_FEATURES="${ENABLED_FEATURES} ${f}"
        done
    else
        wizard_start

        case "$WIZARD_MODE" in
        auto)
            wizard_auto_detect
            ;;
        manual)
            wizard_manual_configure
            ;;
        esac

        # Override arch if user forced it
        [ -n "$force_arch" ] && B4_ARCH="$force_arch"

        # Feature selection
        wizard_select_features
    fi

    echo ""
    log_header "Installing B4"

    # --- Check dependencies ---
    log_info "Checking dependencies..."
    platform_call check_deps

    # --- Resolve version ---
    if [ -z "$version" ]; then
        log_info "Fetching latest version..."
        version=$(get_latest_version)
    fi
    log_ok "Version: ${version}"
    log_ok "Architecture: ${B4_ARCH}"

    # --- Prepare directories ---
    ensure_dir "$B4_BIN_DIR" "Binary directory" || exit 1
    ensure_dir "$B4_DATA_DIR" "Data directory" || exit 1
    setup_temp

    # --- Download & install binary ---
    file_name="${BINARY_NAME}-linux-${B4_ARCH}.tar.gz"
    download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${version}/${file_name}"
    archive_path="${TEMP_DIR}/${file_name}"

    log_info "Downloading b4..."
    if ! fetch_file "$download_url" "$archive_path"; then
        log_err "Download failed for architecture: ${B4_ARCH}"
        exit 1
    fi

    # Verify checksum
    sha_url="${download_url}.sha256"
    verify_checksum "$archive_path" "$sha_url" || true

    # Extract
    log_info "Extracting..."
    cd "$TEMP_DIR"
    tar -xzf "$archive_path" || { log_err "Failed to extract archive"; exit 1; }
    rm -f "$archive_path"

    if [ ! -f "${BINARY_NAME}" ]; then
        log_err "Binary not found in archive"
        exit 1
    fi

    # Stop running instance
    stop_b4

    # Backup existing binary
    if [ -f "${B4_BIN_DIR}/${BINARY_NAME}" ]; then
        ts=$(date '+%Y%m%d_%H%M%S')
        mv "${B4_BIN_DIR}/${BINARY_NAME}" "${B4_BIN_DIR}/${BINARY_NAME}.backup.${ts}"
        log_info "Existing binary backed up"
    fi

    # Install
    mv "${BINARY_NAME}" "${B4_BIN_DIR}/" 2>/dev/null || cp "${BINARY_NAME}" "${B4_BIN_DIR}/" || {
        log_err "Failed to install binary to ${B4_BIN_DIR}"
        exit 1
    }
    chmod +x "${B4_BIN_DIR}/${BINARY_NAME}"

    # Verify
    if "${B4_BIN_DIR}/${BINARY_NAME}" --version >/dev/null 2>&1; then
        installed_ver=$("${B4_BIN_DIR}/${BINARY_NAME}" --version 2>&1 | head -1)
        log_ok "Binary installed: ${installed_ver}"
        # Clean old backups
        rm -f "${B4_BIN_DIR}/${BINARY_NAME}".backup.* 2>/dev/null || true
    else
        log_warn "Binary installed but version check failed"
    fi

    # --- Install service ---
    log_info "Setting up service..."
    platform_call install_service

    # --- Run enabled features ---
    if [ -n "$ENABLED_FEATURES" ]; then
        features_run
    fi

    # --- Summary ---
    _install_summary "$version"
}

_install_summary() {
    version="$1"

    echo ""
    log_header "Installation Complete"
    log_sep
    log_detail "Version" "$version"
    log_detail "Binary" "${B4_BIN_DIR}/${BINARY_NAME}"
    log_detail "Config" "${B4_CONFIG_FILE}"
    log_detail "Service" "${B4_SERVICE_TYPE}"
    log_sep

    # Check if binary is in PATH
    if ! echo "$PATH" | grep -q "$B4_BIN_DIR"; then
        log_warn "$B4_BIN_DIR is not in PATH"
        log_info "Consider: ln -s ${B4_BIN_DIR}/${BINARY_NAME} /usr/bin/${BINARY_NAME}"
    fi

    # Show web interface info
    _show_web_info

    echo ""
    log_info "To see all options: ${B4_BIN_DIR}/${BINARY_NAME} --help"
    echo ""

    # Offer to start service
    if [ "$QUIET_MODE" -eq 0 ] && [ "$B4_SERVICE_TYPE" != "none" ]; then
        if confirm "Start B4 service now?"; then
            platform_call start_service || true
        fi
    fi

    echo ""
    printf "${GREEN}${BOLD}  B4 installation finished!${NC}\n"
    echo ""
}

_show_web_info() {
    web_port="7000"
    protocol="http"

    if [ -f "$B4_CONFIG_FILE" ] && command_exists jq; then
        web_port=$(jq -r '.system.web_server.port // 7000' "$B4_CONFIG_FILE" 2>/dev/null)
        tls=$(jq -r '.system.web_server.tls_cert // ""' "$B4_CONFIG_FILE" 2>/dev/null)
        [ -n "$tls" ] && protocol="https"
    fi

    # Try to get LAN IP
    lan_ip=""
    if command_exists ip; then
        lan_ip=$(ip -4 addr show br0 2>/dev/null | grep 'inet ' | awk '{print $2}' | cut -d'/' -f1)
        [ -z "$lan_ip" ] && lan_ip=$(ip -4 addr 2>/dev/null | grep 'inet 192.168' | head -1 | awk '{print $2}' | cut -d'/' -f1)
    fi

    if [ -n "$lan_ip" ]; then
        echo ""
        log_info "Web interface: ${protocol}://${lan_ip}:${web_port}"
    fi
}


# ======== actions/remove.sh ========
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


# ======== actions/update.sh ========
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
    if [ -n "$B4_PLATFORM" ]; then
        log_info "Restarting service..."
        platform_call start_service 2>/dev/null || true
    fi

    echo ""
    log_ok "Update complete"
    echo ""
}


# ======== actions/sysinfo.sh ========
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


# ======== main.sh ========
# Main entry point — argument parsing and dispatch

main() {
    ACTION="install"
    VERSION=""
    FORCE_ARCH=""

    # Parse arguments
    for arg in "$@"; do
        case "$arg" in
        --remove | --uninstall | -r)
            ACTION="remove" ;;
        --update | -u)
            ACTION="update" ;;
        --sysinfo | --info | -i)
            ACTION="sysinfo" ;;
        --quiet | -q)
            QUIET_MODE=1 ;;
        --arch=*)
            FORCE_ARCH="${arg#*=}" ;;
        --platform=*)
            B4_PLATFORM="${arg#*=}" ;;
        --bin-dir=*)
            B4_BIN_DIR="${arg#*=}" ;;
        --data-dir=*)
            B4_DATA_DIR="${arg#*=}" ;;
        --dry-run)
            DRY_RUN=1 ;;
        --help | -h)
            _show_help
            exit 0 ;;
        v* | V*)
            VERSION="$arg" ;;
        esac
    done

    # Dispatch
    case "$ACTION" in
    install) action_install "$VERSION" "$FORCE_ARCH" ;;
    remove)  action_remove ;;
    update)  action_update "$FORCE_ARCH" ;;
    sysinfo) action_sysinfo ;;
    esac
}

_show_help() {
    echo "B4 Universal Installer"
    echo ""
    echo "Usage: $0 [OPTIONS] [VERSION]"
    echo ""
    echo "Actions:"
    echo "  (default)           Install b4 (interactive wizard)"
    echo "  --update, -u        Update b4 to latest version"
    echo "  --remove, -r        Uninstall b4"
    echo "  --sysinfo, -i       Show system diagnostics"
    echo ""
    echo "Options:"
    echo "  --arch=ARCH         Force architecture (skip detection)"
    echo "  --platform=ID       Force platform (skip detection)"
    echo "  --bin-dir=DIR       Override binary directory"
    echo "  --data-dir=DIR      Override data/config directory"
    echo "  --quiet, -q         Non-interactive mode with defaults"
    echo "  --help, -h          Show this help"
    echo ""
    echo "Environment overrides:"
    echo "  B4_PLATFORM         Platform ID (generic_linux, openwrt, merlinwrt, ...)"
    echo "  B4_BIN_DIR          Binary install directory"
    echo "  B4_DATA_DIR         Data/config directory"
    echo "  B4_PKG_MANAGER      Package manager (apt, dnf, pacman, opkg, ...)"
    echo ""
    echo "Architectures:"
    echo "  amd64, 386, arm64, armv5, armv6, armv7,"
    echo "  mips, mipsle, mips_softfloat, mipsle_softfloat,"
    echo "  mips64, mips64le, loong64, ppc64, ppc64le, riscv64, s390x"
    echo ""
    echo "Examples:"
    echo "  $0                            Interactive install"
    echo "  $0 v1.4.0                     Install specific version"
    echo "  $0 --arch=mipsle_softfloat    Force architecture"
    echo "  $0 --platform=openwrt         Force platform"
    echo "  $0 --quiet                    Non-interactive with defaults"
    echo "  $0 --update                   Update to latest"
    echo "  $0 --remove                   Uninstall"
    echo "  $0 --sysinfo                  Show diagnostics"
}

main "$@"

