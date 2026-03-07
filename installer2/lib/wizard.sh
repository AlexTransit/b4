#!/bin/sh
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
