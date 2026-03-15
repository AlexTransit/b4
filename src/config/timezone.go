package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ApplyTimezone sets time.Local to the given timezone name.
// Supports IANA names (e.g. "Europe/Oslo") and fixed UTC offsets
// (e.g. "UTC+1", "UTC-5", "UTC+5:30").
func ApplyTimezone(tzName string) {
	if tzName == "" {
		return
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil {
		loc = parseFixedOffset(tzName)
		if loc == nil {
			fmt.Fprintf(os.Stderr, "[WARN] Failed to load timezone %s: %v, using UTC\n", tzName, err)
			loc, _ = time.LoadLocation("UTC")
		}
	}

	time.Local = loc
	fmt.Fprintf(os.Stderr, "[INIT] Timezone set to %s\n", loc.String())
}

func parseFixedOffset(name string) *time.Location {
	if !strings.HasPrefix(name, "UTC") {
		return nil
	}
	rest := name[3:]
	if rest == "" {
		loc, _ := time.LoadLocation("UTC")
		return loc
	}

	sign := 1
	switch rest[0] {
	case '+':
		rest = rest[1:]
	case '-':
		sign = -1
		rest = rest[1:]
	default:
		return nil
	}

	hours, minutes := 0, 0
	if _, err := fmt.Sscanf(rest, "%d:%d", &hours, &minutes); err != nil {
		if _, err := fmt.Sscanf(rest, "%d", &hours); err != nil {
			return nil
		}
	}

	offset := sign * (hours*3600 + minutes*60)
	return time.FixedZone(name, offset)
}
