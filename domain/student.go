package domain

import (
	"errors"
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

// CanUpdate checks if the student can be updated with the given new data.
func (s *Student) CanUpdate(updated Student) error {
	if !updated.CreatedAt.Equal(s.CreatedAt) {
		return ErrCreatedAtImmutable
	}
	if updated.Grade < s.Grade {
		return ErrGradeDecrease
	}
	return nil
}

// ValidateCreate validates the required fields for creating a new student.
func (s *Student) ValidateCreate() error {
	if strings.TrimSpace(s.FirstName) == "" || strings.TrimSpace(s.LastName) == "" {
		return errors.New("first_name and last_name are required")
	}
	if s.Details.Local == nil && s.Details.Exchange == nil {
		return ErrDetailsRequired
	}
	return nil
}
