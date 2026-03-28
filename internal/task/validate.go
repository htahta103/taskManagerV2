package task

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxTitleRunes       = 255
	maxDescriptionRunes = 10000
)

// ValidateTitle checks required title and max length (rune count).
func ValidateTitle(title string) error {
	t := strings.TrimSpace(title)
	if t == "" {
		return errors.New("title is required")
	}
	if utf8.RuneCountInString(t) > maxTitleRunes {
		return fmt.Errorf("title must be at most %d characters", maxTitleRunes)
	}
	return nil
}

// ValidateDescription checks optional description length when present.
func ValidateDescription(desc *string) error {
	if desc == nil {
		return nil
	}
	if utf8.RuneCountInString(*desc) > maxDescriptionRunes {
		return fmt.Errorf("description must be at most %d characters", maxDescriptionRunes)
	}
	return nil
}

// ValidateStatus checks allowed status when provided.
func ValidateStatus(s Status) error {
	switch s {
	case StatusTodo, StatusDoing, StatusDone:
		return nil
	default:
		return fmt.Errorf("invalid status %q, expected todo, doing, or done", s)
	}
}

// ValidatePriority checks allowed priority when provided.
func ValidatePriority(p Priority) error {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return nil
	default:
		return fmt.Errorf("invalid priority %q, expected low, medium, or high", p)
	}
}

// ValidateFocusBucket checks allowed focus_bucket when provided.
func ValidateFocusBucket(f FocusBucket) error {
	switch f {
	case FocusNone, FocusToday, FocusNext, FocusLater:
		return nil
	default:
		return fmt.Errorf("invalid focus_bucket %q, expected none, today, next, or later", f)
	}
}

// ValidateCreatePayload runs all validation for a create request after JSON decode.
func ValidateCreatePayload(p *CreatePayload) error {
	if p == nil {
		return errors.New("request body is required")
	}
	if err := ValidateTitle(p.Title); err != nil {
		return err
	}
	if err := ValidateDescription(p.Description); err != nil {
		return err
	}
	if p.Status != nil {
		if err := ValidateStatus(*p.Status); err != nil {
			return err
		}
	}
	if p.Priority != nil {
		if err := ValidatePriority(*p.Priority); err != nil {
			return err
		}
	}
	if p.FocusBucket != nil {
		if err := ValidateFocusBucket(*p.FocusBucket); err != nil {
			return err
		}
	}
	return nil
}
