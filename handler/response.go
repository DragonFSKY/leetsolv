package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/eannchen/leetsolv/core"
	"github.com/eannchen/leetsolv/internal/errs"
)

// CLIResponse is the standard JSON envelope for all CLI output.
type CLIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Error   string `json:"error"`
}

// QuestionDTO is the JSON-serializable representation of a Question
// with user-friendly values (1-indexed).
type QuestionDTO struct {
	ID           int     `json:"id"`
	URL          string  `json:"url"`
	Note         string  `json:"note"`
	Familiarity  int     `json:"familiarity"`   // 1-5 (internal 0-4 + 1)
	Importance   int     `json:"importance"`    // 1-4 (internal 0-3 + 1)
	Memory       int     `json:"memory,omitempty"` // 1-3 (internal 0-2 + 1), omitted when 0
	LastReviewed string  `json:"last_reviewed"` // YYYY-MM-DD
	NextReview   string  `json:"next_review"`   // YYYY-MM-DD
	ReviewCount  int     `json:"review_count"`
	EaseFactor   float64 `json:"ease_factor"`
}

// DeltaDTO is the JSON-serializable representation of a Delta.
type DeltaDTO struct {
	Action     string       `json:"action"`
	QuestionID int          `json:"question_id"`
	OldState   *QuestionDTO `json:"old_state"`
	NewState   *QuestionDTO `json:"new_state"`
	CreatedAt  string       `json:"created_at"` // RFC3339
}

// ToQuestionDTO converts a core.Question to a QuestionDTO with user-friendly values.
func ToQuestionDTO(q *core.Question) QuestionDTO {
	return QuestionDTO{
		ID:           q.ID,
		URL:          q.URL,
		Note:         q.Note,
		Familiarity:  int(q.Familiarity) + 1,
		Importance:   int(q.Importance) + 1,
		Memory:       0, // Not stored in Question; set by caller if needed
		LastReviewed: q.LastReviewed.Local().Format("2006-01-02"),
		NextReview:   q.NextReview.Local().Format("2006-01-02"),
		ReviewCount:  q.ReviewCount,
		EaseFactor:   q.EaseFactor,
	}
}

// ToDeltaDTO converts a core.Delta to a DeltaDTO with user-friendly values.
func ToDeltaDTO(d *core.Delta) DeltaDTO {
	dto := DeltaDTO{
		Action:     string(d.Action),
		QuestionID: d.QuestionID,
		CreatedAt:  d.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if d.OldState != nil {
		old := ToQuestionDTO(d.OldState)
		dto.OldState = &old
	}
	if d.NewState != nil {
		ns := ToQuestionDTO(d.NewState)
		dto.NewState = &ns
	}
	return dto
}

// WriteJSONSuccess writes a successful JSON envelope to the writer.
func WriteJSONSuccess(w io.Writer, data any) {
	resp := CLIResponse{
		Success: true,
		Data:    data,
		Error:   "",
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(resp)
}

// WriteJSONError writes an error JSON envelope to the writer.
func WriteJSONError(w io.Writer, err error) {
	msg := extractUserMessage(err)
	resp := CLIResponse{
		Success: false,
		Data:    nil,
		Error:   msg,
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(resp)
}

// WriteJSONErrorMsg writes an error JSON envelope with a plain string message.
func WriteJSONErrorMsg(w io.Writer, msg string) {
	resp := CLIResponse{
		Success: false,
		Data:    nil,
		Error:   msg,
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(resp)
}

// extractUserMessage returns the user-friendly message from an error.
func extractUserMessage(err error) string {
	if err == nil {
		return ""
	}
	var codedErr *errs.CodedError
	if errors.As(err, &codedErr) {
		return codedErr.UserMessage()
	}
	return fmt.Sprintf("Error: %s", err.Error())
}
