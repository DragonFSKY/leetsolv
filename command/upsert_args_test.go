package command

import (
	"testing"
)

func TestParseUpsertArgs_AllFlags(t *testing.T) {
	args := []string{
		"https://leetcode.com/problems/two-sum",
		"--familiarity=4",
		"--importance=3",
		"--memory=2",
		"--note=dynamic programming",
	}

	set, err := parseUpsertArgs(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !set.HasFlags {
		t.Error("expected HasFlags=true")
	}
	if set.URL != "https://leetcode.com/problems/two-sum" {
		t.Errorf("expected URL, got %q", set.URL)
	}
	if set.Familiarity != 4 {
		t.Errorf("expected familiarity=4, got %d", set.Familiarity)
	}
	if set.Importance != 3 {
		t.Errorf("expected importance=3, got %d", set.Importance)
	}
	if set.Memory != 2 {
		t.Errorf("expected memory=2, got %d", set.Memory)
	}
	if set.Note != "dynamic programming" {
		t.Errorf("expected note='dynamic programming', got %q", set.Note)
	}
}

func TestParseUpsertArgs_NoFlags(t *testing.T) {
	args := []string{"https://leetcode.com/problems/two-sum"}

	set, err := parseUpsertArgs(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if set.HasFlags {
		t.Error("expected HasFlags=false for URL-only")
	}
}

func TestParseUpsertArgs_EmptyArgs(t *testing.T) {
	set, err := parseUpsertArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if set.HasFlags {
		t.Error("expected HasFlags=false for empty args")
	}
}

func TestParseUpsertArgs_MissingURL(t *testing.T) {
	args := []string{"--familiarity=3", "--importance=2"}
	_, err := parseUpsertArgs(args)
	if err == nil {
		t.Error("expected error for missing URL")
	}
}

func TestParseUpsertArgs_MissingFamiliarity(t *testing.T) {
	args := []string{"https://leetcode.com/problems/two-sum", "--importance=2", "--note=test"}
	_, err := parseUpsertArgs(args)
	if err == nil {
		t.Error("expected error for missing --familiarity")
	}
}

func TestParseUpsertArgs_MissingImportance(t *testing.T) {
	args := []string{"https://leetcode.com/problems/two-sum", "--familiarity=2", "--note=test"}
	_, err := parseUpsertArgs(args)
	if err == nil {
		t.Error("expected error for missing --importance")
	}
}

func TestParseUpsertArgs_FamiliarityOutOfRange(t *testing.T) {
	tests := []struct {
		name string
		val  string
	}{
		{"too low", "--familiarity=0"},
		{"too high", "--familiarity=6"},
		{"negative", "--familiarity=-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"https://leetcode.com/problems/two-sum", tt.val, "--importance=2"}
			_, err := parseUpsertArgs(args)
			if err == nil {
				t.Errorf("expected error for %s", tt.val)
			}
		})
	}
}

func TestParseUpsertArgs_ImportanceOutOfRange(t *testing.T) {
	tests := []struct {
		name string
		val  string
	}{
		{"too low", "--importance=0"},
		{"too high", "--importance=5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"https://leetcode.com/problems/two-sum", "--familiarity=1", tt.val}
			_, err := parseUpsertArgs(args)
			if err == nil {
				t.Errorf("expected error for %s", tt.val)
			}
		})
	}
}

func TestParseUpsertArgs_MemoryRequiredWhenFamiliarityGe3(t *testing.T) {
	args := []string{"https://leetcode.com/problems/two-sum", "--familiarity=3", "--importance=2"}
	_, err := parseUpsertArgs(args)
	if err == nil {
		t.Error("expected error: --memory required when familiarity >= 3")
	}
}

func TestParseUpsertArgs_MemoryDefaultWhenFamiliarityLt3(t *testing.T) {
	args := []string{"https://leetcode.com/problems/two-sum", "--familiarity=2", "--importance=2"}
	set, err := parseUpsertArgs(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if set.Memory != 1 {
		t.Errorf("expected default memory=1 (Reasoned), got %d", set.Memory)
	}
}

func TestParseUpsertArgs_MemoryOutOfRange(t *testing.T) {
	tests := []struct {
		name string
		val  string
	}{
		{"too low", "--memory=0"},
		{"too high", "--memory=4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"https://leetcode.com/problems/two-sum", "--familiarity=3", "--importance=2", tt.val}
			_, err := parseUpsertArgs(args)
			if err == nil {
				t.Errorf("expected error for %s", tt.val)
			}
		})
	}
}

func TestParseUpsertArgs_QuotedNote(t *testing.T) {
	args := []string{"https://leetcode.com/problems/two-sum", "--familiarity=1", "--importance=1", "--note=\"hello world\""}
	set, err := parseUpsertArgs(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if set.Note != "hello world" {
		t.Errorf("expected note='hello world', got %q", set.Note)
	}
}

func TestParseUpsertArgs_InvalidFamiliarityFormat(t *testing.T) {
	args := []string{"https://leetcode.com/problems/two-sum", "--familiarity=abc", "--importance=2"}
	_, err := parseUpsertArgs(args)
	if err == nil {
		t.Error("expected error for non-numeric familiarity")
	}
}
