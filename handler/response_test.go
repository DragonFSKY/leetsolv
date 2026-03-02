package handler

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/eannchen/leetsolv/core"
	"github.com/eannchen/leetsolv/internal/errs"
)

func TestToQuestionDTO(t *testing.T) {
	q := &core.Question{
		ID:           1,
		URL:          "https://leetcode.com/problems/two-sum",
		Note:         "hash map approach",
		Familiarity:  core.Medium,    // internal 2
		Importance:   core.HighImportance, // internal 2
		LastReviewed: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC),
		NextReview:   time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC),
		ReviewCount:  3,
		EaseFactor:   2.5,
	}

	dto := ToQuestionDTO(q)

	if dto.ID != 1 {
		t.Errorf("expected ID=1, got %d", dto.ID)
	}
	if dto.Familiarity != 3 { // internal 2 + 1 = 3
		t.Errorf("expected familiarity=3, got %d", dto.Familiarity)
	}
	if dto.Importance != 3 { // internal 2 + 1 = 3
		t.Errorf("expected importance=3, got %d", dto.Importance)
	}
	if dto.ReviewCount != 3 {
		t.Errorf("expected review_count=3, got %d", dto.ReviewCount)
	}
	if dto.EaseFactor != 2.5 {
		t.Errorf("expected ease_factor=2.5, got %f", dto.EaseFactor)
	}
	if dto.LastReviewed != "2026-02-25" {
		t.Errorf("expected last_reviewed=2026-02-25, got %s", dto.LastReviewed)
	}
	if dto.NextReview != "2026-03-04" {
		t.Errorf("expected next_review=2026-03-04, got %s", dto.NextReview)
	}
}

func TestToQuestionDTO_Boundaries(t *testing.T) {
	tests := []struct {
		name          string
		familiarity   core.Familiarity
		importance    core.Importance
		expectedFam   int
		expectedImp   int
	}{
		{"min values", core.VeryHard, core.LowImportance, 1, 1},
		{"max values", core.VeryEasy, core.CriticalImportance, 5, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &core.Question{
				Familiarity:  tt.familiarity,
				Importance:   tt.importance,
				LastReviewed: time.Now(),
				NextReview:   time.Now(),
			}
			dto := ToQuestionDTO(q)
			if dto.Familiarity != tt.expectedFam {
				t.Errorf("expected familiarity=%d, got %d", tt.expectedFam, dto.Familiarity)
			}
			if dto.Importance != tt.expectedImp {
				t.Errorf("expected importance=%d, got %d", tt.expectedImp, dto.Importance)
			}
		})
	}
}

func TestToDeltaDTO(t *testing.T) {
	now := time.Date(2026, 2, 25, 10, 30, 0, 0, time.UTC)
	d := &core.Delta{
		Action:     core.ActionAdd,
		QuestionID: 1,
		OldState:   nil,
		NewState: &core.Question{
			ID:           1,
			URL:          "https://leetcode.com/problems/two-sum",
			Familiarity:  core.Hard,
			Importance:   core.MediumImportance,
			LastReviewed: now,
			NextReview:   now.AddDate(0, 0, 3),
		},
		CreatedAt: now,
	}

	dto := ToDeltaDTO(d)

	if dto.Action != "add" {
		t.Errorf("expected action=add, got %s", dto.Action)
	}
	if dto.OldState != nil {
		t.Error("expected old_state=nil")
	}
	if dto.NewState == nil {
		t.Fatal("expected new_state to be non-nil")
	}
	if dto.NewState.Familiarity != 2 { // Hard=1, +1=2
		t.Errorf("expected familiarity=2, got %d", dto.NewState.Familiarity)
	}
}

func TestWriteJSONSuccess(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"total": 5,
	}

	WriteJSONSuccess(&buf, data)

	var resp CLIResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Error != "" {
		t.Errorf("expected empty error, got %q", resp.Error)
	}
}

func TestWriteJSONError(t *testing.T) {
	var buf bytes.Buffer
	err := errs.WrapValidationError(nil, "bad input")

	WriteJSONError(&buf, err)

	var resp CLIResponse
	if jsonErr := json.Unmarshal(buf.Bytes(), &resp); jsonErr != nil {
		t.Fatalf("failed to unmarshal: %v", jsonErr)
	}
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error != "bad input" {
		t.Errorf("expected error='bad input', got %q", resp.Error)
	}
	if resp.Data != nil {
		t.Errorf("expected data=nil, got %v", resp.Data)
	}
}

func TestWriteJSONErrorMsg(t *testing.T) {
	var buf bytes.Buffer
	WriteJSONErrorMsg(&buf, "something failed")

	var resp CLIResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error != "something failed" {
		t.Errorf("expected error='something failed', got %q", resp.Error)
	}
}

func TestQuestionDTO_MemoryOmitempty(t *testing.T) {
	q := &core.Question{
		ID:           1,
		URL:          "https://leetcode.com/problems/two-sum",
		Familiarity:  core.Medium,
		Importance:   core.HighImportance,
		LastReviewed: time.Now(),
		NextReview:   time.Now(),
	}

	dto := ToQuestionDTO(q)
	// Memory should be 0 (default, not set)
	if dto.Memory != 0 {
		t.Errorf("expected memory=0 for non-upsert DTO, got %d", dto.Memory)
	}

	// When marshaled to JSON, memory should be omitted
	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	jsonStr := string(data)
	if strings.Contains(jsonStr, `"memory"`) {
		t.Errorf("expected memory field to be omitted in JSON, got: %s", jsonStr)
	}

	// When memory is set (upsert context), it should appear
	dto.Memory = 2
	data, err = json.Marshal(dto)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	jsonStr = string(data)
	if !strings.Contains(jsonStr, `"memory":2`) {
		t.Errorf("expected memory=2 in JSON, got: %s", jsonStr)
	}
}
