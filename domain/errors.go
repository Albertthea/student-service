// Package domain contains the core business entities and domain-specific errors
// used across the student-service application.
package domain

import "errors"

var (
	// ErrNotFound is returned when a student record cannot be found.
	ErrNotFound = errors.New("student not found")

	// ErrCreatedAtImmutable is returned when trying to modify the immutable CreatedAt field.
	ErrCreatedAtImmutable = errors.New("created_at field cannot be modified")

	// ErrGradeDecrease is returned when an update attempts to decrease the student's grade.
	ErrGradeDecrease = errors.New("grade cannot be decreased")

	// ErrDetailsRequired is returned when creating a student without specifying details.
	ErrDetailsRequired = errors.New("student details must be provided")
)
