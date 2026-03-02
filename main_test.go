package main

import (
	"errors"
	"testing"

	"github.com/eannchen/leetsolv/handler"
	"github.com/eannchen/leetsolv/internal/errs"
)

func TestHasJSONFlag(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		expect bool
	}{
		{"present", []string{"leetsolv", "status", "--json"}, true},
		{"absent", []string{"leetsolv", "status"}, false},
		{"with other flags", []string{"leetsolv", "list", "--no-color", "--json"}, true},
		{"empty", []string{}, false},
		{"only program name", []string{"leetsolv"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasJSONFlag(tt.args)
			if got != tt.expect {
				t.Errorf("hasJSONFlag(%v) = %v, want %v", tt.args, got, tt.expect)
			}
		})
	}
}

func TestParseGlobalFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantCommand  string
		wantFiltered []string
		wantOpt      handler.CLIOptions
		wantErr      bool
	}{
		{
			name:         "command only",
			args:         []string{"status"},
			wantCommand:  "status",
			wantFiltered: nil,
			wantOpt:      handler.CLIOptions{},
		},
		{
			name:         "command with args",
			args:         []string{"upsert", "https://leetcode.com/problems/two-sum", "--familiarity=3"},
			wantCommand:  "upsert",
			wantFiltered: []string{"https://leetcode.com/problems/two-sum", "--familiarity=3"},
			wantOpt:      handler.CLIOptions{},
		},
		{
			name:         "json before command",
			args:         []string{"--json", "status"},
			wantCommand:  "status",
			wantFiltered: nil,
			wantOpt:      handler.CLIOptions{JSON: true, NoColor: true, NoPager: true},
		},
		{
			name:         "json after command",
			args:         []string{"status", "--json"},
			wantCommand:  "status",
			wantFiltered: nil,
			wantOpt:      handler.CLIOptions{JSON: true, NoColor: true, NoPager: true},
		},
		{
			name:         "no-color before command",
			args:         []string{"--no-color", "list"},
			wantCommand:  "list",
			wantFiltered: nil,
			wantOpt:      handler.CLIOptions{NoColor: true},
		},
		{
			name:         "no-pager after command",
			args:         []string{"list", "--no-pager"},
			wantCommand:  "list",
			wantFiltered: nil,
			wantOpt:      handler.CLIOptions{NoPager: true},
		},
		{
			name:         "all global flags before command",
			args:         []string{"--json", "--no-color", "--no-pager", "status"},
			wantCommand:  "status",
			wantFiltered: nil,
			wantOpt:      handler.CLIOptions{JSON: true, NoColor: true, NoPager: true},
		},
		{
			name:         "mixed: global flags + command + command-specific flags",
			args:         []string{"--json", "upsert", "--familiarity=3", "--importance=2"},
			wantCommand:  "upsert",
			wantFiltered: []string{"--familiarity=3", "--importance=2"},
			wantOpt:      handler.CLIOptions{JSON: true, NoColor: true, NoPager: true},
		},
		{
			name:        "unknown flag before command → error",
			args:        []string{"--unknown", "status"},
			wantCommand: "",
			wantErr:     true,
		},
		{
			name:         "no command (only global flags)",
			args:         []string{"--json"},
			wantCommand:  "",
			wantFiltered: nil,
			wantOpt:      handler.CLIOptions{JSON: true, NoColor: true, NoPager: true},
		},
		{
			name:         "empty args",
			args:         []string{},
			wantCommand:  "",
			wantFiltered: nil,
			wantOpt:      handler.CLIOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, filtered, opt, err := parseGlobalFlags(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cmd != tt.wantCommand {
				t.Errorf("command = %q, want %q", cmd, tt.wantCommand)
			}

			// Compare filtered
			if len(filtered) != len(tt.wantFiltered) {
				t.Errorf("filtered length = %d, want %d", len(filtered), len(tt.wantFiltered))
			} else {
				for i := range filtered {
					if filtered[i] != tt.wantFiltered[i] {
						t.Errorf("filtered[%d] = %q, want %q", i, filtered[i], tt.wantFiltered[i])
					}
				}
			}

			// Compare options
			if opt.JSON != tt.wantOpt.JSON {
				t.Errorf("opt.JSON = %v, want %v", opt.JSON, tt.wantOpt.JSON)
			}
			if opt.NoColor != tt.wantOpt.NoColor {
				t.Errorf("opt.NoColor = %v, want %v", opt.NoColor, tt.wantOpt.NoColor)
			}
			if opt.NoPager != tt.wantOpt.NoPager {
				t.Errorf("opt.NoPager = %v, want %v", opt.NoPager, tt.wantOpt.NoPager)
			}
		})
	}
}

func TestMapExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"validation error", errs.ErrInvalidFamiliarityLevel, 1},
		{"business error", errs.ErrQuestionNotFound, 1},
		{"system error", errs.WrapInternalError(errors.New("disk full"), "save failed"), 2},
		{"plain error", errors.New("unknown"), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapExitCode(tt.err)
			if got != tt.want {
				t.Errorf("mapExitCode(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}
