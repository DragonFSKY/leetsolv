package command

import (
	"fmt"
	"strconv"
	"strings"
)

// UpsertArgSet holds parsed non-interactive upsert arguments.
type UpsertArgSet struct {
	URL         string
	Familiarity int // 1-5 (user-friendly)
	Importance  int // 1-4 (user-friendly)
	Memory      int // 1-3 (user-friendly), 0 means not provided
	Note        string
	HasFlags    bool // true if any --flag was present (triggers non-interactive mode)
}

// parseUpsertArgs parses upsert command arguments into an UpsertArgSet.
// Returns hasFlags=false if no non-interactive flags were found (use interactive mode).
func parseUpsertArgs(args []string) (UpsertArgSet, error) {
	var set UpsertArgSet

	hasFamiliarity := false
	hasImportance := false
	hasMemory := false
	hasNote := false

	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--familiarity="):
			val, err := parseIntFlag(arg, "--familiarity=")
			if err != nil {
				return set, fmt.Errorf("invalid --familiarity value: %s", arg)
			}
			set.Familiarity = val
			hasFamiliarity = true
		case strings.HasPrefix(arg, "--importance="):
			val, err := parseIntFlag(arg, "--importance=")
			if err != nil {
				return set, fmt.Errorf("invalid --importance value: %s", arg)
			}
			set.Importance = val
			hasImportance = true
		case strings.HasPrefix(arg, "--memory="):
			val, err := parseIntFlag(arg, "--memory=")
			if err != nil {
				return set, fmt.Errorf("invalid --memory value: %s", arg)
			}
			set.Memory = val
			hasMemory = true
		case strings.HasPrefix(arg, "--note="):
			set.Note = strings.TrimPrefix(arg, "--note=")
			// Remove surrounding quotes if present
			set.Note = strings.Trim(set.Note, "\"'")
			hasNote = true
		default:
			// Position argument: URL (first non-flag arg)
			if set.URL == "" && !strings.HasPrefix(arg, "--") {
				set.URL = arg
			}
		}
	}

	set.HasFlags = hasFamiliarity || hasImportance || hasMemory || hasNote

	if !set.HasFlags {
		return set, nil
	}

	// Validate required fields for non-interactive mode
	if set.URL == "" {
		return set, fmt.Errorf("URL is required for non-interactive upsert")
	}
	if !hasFamiliarity {
		return set, fmt.Errorf("--familiarity is required for non-interactive upsert")
	}
	if !hasImportance {
		return set, fmt.Errorf("--importance is required for non-interactive upsert")
	}

	// Validate ranges
	if set.Familiarity < 1 || set.Familiarity > 5 {
		return set, fmt.Errorf("--familiarity must be between 1 and 5, got %d", set.Familiarity)
	}
	if set.Importance < 1 || set.Importance > 4 {
		return set, fmt.Errorf("--importance must be between 1 and 4, got %d", set.Importance)
	}

	// Memory validation: required when familiarity >= 3
	if set.Familiarity >= 3 && !hasMemory {
		return set, fmt.Errorf("--memory is required when familiarity >= 3")
	}
	if hasMemory {
		if set.Memory < 1 || set.Memory > 3 {
			return set, fmt.Errorf("--memory must be between 1 and 3, got %d", set.Memory)
		}
	} else {
		// Default: Reasoned (internal 0, user-friendly 1)
		set.Memory = 1
	}

	return set, nil
}

func parseIntFlag(arg, prefix string) (int, error) {
	valStr := strings.TrimPrefix(arg, prefix)
	return strconv.Atoi(valStr)
}
