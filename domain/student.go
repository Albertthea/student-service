package domain

import (
	"fmt"
	"strings"
	"time"
)

// Status represents the lifecycle status of a student.
type Status string

const (
	// StatusUnspecified is the default zero-value status.
	StatusUnspecified Status = "STATUS_UNSPECIFIED"

	// StatusActive means the student is currently active.
	StatusActive Status = "ACTIVE"

	// StatusGraduated means the student has graduated.
	StatusGraduated Status = "GRADUATED"

	// StatusSuspended means the student has been suspended.
	StatusSuspended Status = "SUSPENDED"

	// StatusExchange means the student is an exchange student.
	StatusExchange Status = "EXCHANGE"
)

// Valid returns true if the status is within the allowed set.
func (st Status) Valid() bool {
	switch st {
	case StatusUnspecified, StatusActive, StatusGraduated, StatusSuspended, StatusExchange:
		return true
	default:
		return false
	}
}

// ParseStatus trims and uppercases the input string and returns a valid Status.
func ParseStatus(s string) (Status, error) {
	st := Status(strings.ToUpper(strings.TrimSpace(s)))
	if st.Valid() {
		return st, nil
	}
	return "", fmt.Errorf("%w: unknown status %q", ErrInvalidArgument, s)
}

// Address holds the student's home address.
type Address struct {
	Street  string
	City    string
	Country string
}

// LocalStudent represents details for a local student.
type LocalStudent struct {
	NationalID  string
	Scholarship bool
}

// ExchangeStudent represents details for an exchange student.
type ExchangeStudent struct {
	HomeUniversity  string
	CountryOfOrigin string
	ProgramStart    time.Time
	ProgramEnd      time.Time
}

// Details holds either local or exchange student details.
type Details struct {
	Local    *LocalStudent
	Exchange *ExchangeStudent
}

// Student is the main domain entity representing a student.
type Student struct {
	ID         string
	FirstName  string
	LastName   string
	MiddleName *string
	Grade      int32
	Status     Status
	Home       *Address
	Course     map[string]string
	Friends    []string
	CreatedAt  time.Time
	Details    Details
}

const (
	// minGrade defines the minimum allowed grade.
	minGrade = 1
	// maxGrade defines the maximum allowed grade.
	maxGrade = 12
)

// normalizeTime converts a time to UTC and truncates to microsecond precision.
// This makes equality checks deterministic across database round-trips.
func normalizeTime(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.UTC().Truncate(time.Microsecond)
}

// Normalize cleans and normalizes fields of the Student entity (idempotent).
func (s *Student) Normalize() {
	s.CreatedAt = normalizeTime(s.CreatedAt)

	s.FirstName = strings.TrimSpace(s.FirstName)
	s.LastName = strings.TrimSpace(s.LastName)

	if s.MiddleName != nil {
		m := strings.TrimSpace(*s.MiddleName)
		if m == "" {
			s.MiddleName = nil
		} else {
			s.MiddleName = &m
		}
	}

	// Deduplicate and trim friends.
	if len(s.Friends) > 0 {
		seen := make(map[string]struct{}, len(s.Friends))
		out := s.Friends[:0]
		for _, f := range s.Friends {
			f = strings.TrimSpace(f)
			if f == "" {
				continue
			}
			if _, ok := seen[f]; ok {
				continue
			}
			seen[f] = struct{}{}
			out = append(out, f)
		}
		s.Friends = out
	}

	// Normalize course map: trim keys/values, drop empty keys.
	if s.Course != nil {
		for k, v := range s.Course {
			kt := strings.TrimSpace(k)
			vt := strings.TrimSpace(v)
			if kt == "" {
				delete(s.Course, k)
				continue
			}
			if kt != k || vt != v {
				delete(s.Course, k)
				s.Course[kt] = vt
			}
		}
	}
}

// ValidateCoreInvariants checks invariants common to create and update.
func (s *Student) ValidateCoreInvariants() error {
	hasLocal := s.Details.Local != nil
	hasExch := s.Details.Exchange != nil
	if !hasLocal && !hasExch {
		return ErrDetailsRequired
	}
	if hasLocal && hasExch {
		return fmt.Errorf("%w: exactly one of local or exchange must be set", ErrInvalidArgument)
	}

	if !s.Status.Valid() {
		return fmt.Errorf("%w: invalid status %q", ErrInvalidArgument, s.Status)
	}
	if s.Status == StatusExchange && !hasExch {
		return fmt.Errorf("%w: status EXCHANGE requires exchange details", ErrInvalidArgument)
	}

	if s.Grade < minGrade || s.Grade > maxGrade {
		return fmt.Errorf("%w: grade must be in [%d..%d]", ErrInvalidArgument, minGrade, maxGrade)
	}

	if ex := s.Details.Exchange; ex != nil && ex.ProgramEnd.Before(ex.ProgramStart) {
		return fmt.Errorf("%w: program_end before program_start", ErrInvalidArgument)
	}

	return nil
}

// ValidateCreate validates rules specific to creating a new student.
func (s *Student) ValidateCreate() error {
	s.Normalize()

	if s.ID == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidArgument)
	}
	if strings.TrimSpace(s.FirstName) == "" || strings.TrimSpace(s.LastName) == "" {
		return fmt.Errorf("%w: first_name and last_name are required", ErrInvalidArgument)
	}

	return s.ValidateCoreInvariants()
}

// CanUpdate checks if the student can be updated with the given new data.
func (s *Student) CanUpdate(updated Student) error {
	s.Normalize()
	updated.Normalize()

	if !updated.CreatedAt.Equal(s.CreatedAt) {
		return ErrCreatedAtImmutable
	}
	if updated.Grade < s.Grade {
		return ErrGradeDecrease
	}

	return updated.ValidateCoreInvariants()
}
