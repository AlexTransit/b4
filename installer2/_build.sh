#!/bin/sh
# Builds install.sh from installer2/ components
# Auto-discovers platforms and features — just add files and rebuild

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUTPUT="${SCRIPT_DIR}/../install.sh"

>"$OUTPUT"

# Append a file to output. First arg: file path, second: "header" to keep shebang
append() {
    file="$1"
    mode="$2"
    basename=$(echo "$file" | sed "s|${SCRIPT_DIR}/||")

    if [ ! -f "$file" ]; then
        echo "Missing: $basename" >&2
        exit 1
    fi

    if [ "$mode" = "header" ]; then
        # Header goes first — no prefix comment before shebang
        cat "$file" >>"$OUTPUT"
    else
        echo "" >>"$OUTPUT"
        echo "# ======== ${basename} ========" >>"$OUTPUT"
        # Skip shebang line, strip leading blank lines
        tail -n +2 "$file" | sed '/./,$!d' >>"$OUTPUT"
    fi

    echo "" >>"$OUTPUT"
}

# 1. Header (keep shebang)
append "${SCRIPT_DIR}/header.sh" "header"

# 2. Lib modules (fixed order: colors first, then log, then rest)
append "${SCRIPT_DIR}/lib/colors.sh"
append "${SCRIPT_DIR}/lib/log.sh"
append "${SCRIPT_DIR}/lib/utils.sh"
append "${SCRIPT_DIR}/lib/wizard.sh"

# 3. Platform system: interface, detect, then auto-discover platform files
append "${SCRIPT_DIR}/platforms/_interface.sh"
append "${SCRIPT_DIR}/platforms/_detect.sh"
for f in "${SCRIPT_DIR}"/platforms/*.sh; do
    case "$(basename "$f")" in
    _*) continue ;; # skip _interface.sh, _detect.sh
    esac
    append "$f"
done

# 4. Feature system: interface, then auto-discover feature files
append "${SCRIPT_DIR}/features/_interface.sh"
for f in "${SCRIPT_DIR}"/features/*.sh; do
    case "$(basename "$f")" in
    _*) continue ;; # skip _interface.sh
    esac
    append "$f"
done

# 5. Actions (fixed order)
for action in install remove update sysinfo; do
    file="${SCRIPT_DIR}/actions/${action}.sh"
    [ -f "$file" ] && append "$file"
done

# 6. Main entry point
append "${SCRIPT_DIR}/main.sh"

chmod +x "$OUTPUT"
echo "Built: $OUTPUT ($(wc -l <"$OUTPUT") lines)"
