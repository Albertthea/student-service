// Package app contains the repository adapter that converts between
// SQL repository models and domain models for the student service.
package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	d "example.com/student-service/domain"
	repostu "example.com/student-service/repository/student"
	"github.com/lib/pq"
)

// newRepoAdapter wraps SQL repository to implement service.Repository over domain types.
func newRepoAdapter(r *repostu.Repository) *repoAdapter { return &repoAdapter{r: r} }

type repoAdapter struct{ r *repostu.Repository }

// Create maps domain.Student -> repo.Student and delegates to SQL repository.
func (a *repoAdapter) Create(ctx context.Context, st d.Student) (string, error) {
	rs, err := domainToRepo(st)
	if err != nil {
		return "", err
	}
	return a.r.Create(ctx, rs)
}

// GetByID loads repo.Student and maps it to domain.Student.
func (a *repoAdapter) GetByID(ctx context.Context, id string) (*d.Student, error) {
	rs, err := a.r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ds, err := repoToDomain(*rs)
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

// Update maps domain.Student -> repo.Student and updates in SQL repository.
func (a *repoAdapter) Update(ctx context.Context, st d.Student) error {
	rs, err := domainToRepo(st)
	if err != nil {
		return err
	}
	return a.r.Update(ctx, rs)
}

// Delete delegates to SQL repository.
func (a *repoAdapter) Delete(ctx context.Context, id string) error {
	return a.r.Delete(ctx, id)
}

// List returns all students mapped to domain.
func (a *repoAdapter) List(ctx context.Context) ([]d.Student, error) {
	items, err := a.r.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]d.Student, 0, len(items))
	for _, it := range items {
		ds, err := repoToDomain(it)
		if err != nil {
			return nil, err
		}
		out = append(out, ds)
	}
	return out, nil
}

// ListByGrade returns students by grade mapped to domain.
func (a *repoAdapter) ListByGrade(ctx context.Context, grade int32) ([]d.Student, error) {
	items, err := a.r.ListByGrade(ctx, grade)
	if err != nil {
		return nil, err
	}
	out := make([]d.Student, 0, len(items))
	for _, it := range items {
		ds, err := repoToDomain(it)
		if err != nil {
			return nil, err
		}
		out = append(out, ds)
	}
	return out, nil
}

// --- mappers domain <-> repo ---

func domainToRepo(s d.Student) (repostu.Student, error) {
	var (
		home, course, local, exch []byte
		err                       error
	)
	if s.Home != nil {
		home, err = json.Marshal(struct {
			Street, City, Country string
		}{s.Home.Street, s.Home.City, s.Home.Country})
		if err != nil {
			return repostu.Student{}, err
		}
	}
	if s.Course != nil {
		course, err = json.Marshal(s.Course)
		if err != nil {
			return repostu.Student{}, err
		}
	}
	if s.Details.Local != nil {
		local, err = json.Marshal(struct {
			NationalID  string `json:"national_id"`
			Scholarship bool   `json:"scholarship"`
		}{s.Details.Local.NationalID, s.Details.Local.Scholarship})
		if err != nil {
			return repostu.Student{}, err
		}
	}
	if s.Details.Exchange != nil {
		exch, err = json.Marshal(struct {
			HomeUniversity, CountryOfOrigin string
			ProgramStart, ProgramEnd        time.Time
		}{
			s.Details.Exchange.HomeUniversity,
			s.Details.Exchange.CountryOfOrigin,
			s.Details.Exchange.ProgramStart,
			s.Details.Exchange.ProgramEnd,
		})
		if err != nil {
			return repostu.Student{}, err
		}
	}
	var middle sql.NullString
	if s.MiddleName != nil {
		middle = sql.NullString{String: *s.MiddleName, Valid: true}
	}
	var status sql.NullString
	if s.Status != "" {
		status = sql.NullString{String: string(s.Status), Valid: true}
	}

	return repostu.Student{
		ID:           s.ID,
		FirstName:    s.FirstName,
		LastName:     s.LastName,
		Grade:        s.Grade,
		CreatedAt:    s.CreatedAt,
		MiddleName:   middle,
		Status:       status,
		HomeAddress:  home,
		CourseGrades: course,
		Friends:      pq.StringArray(s.Friends),
		Local:        local,
		Exchange:     exch,
	}, nil
}

func repoToDomain(s repostu.Student) (d.Student, error) {
	var (
		home   *d.Address
		course map[string]string
		loc    *d.LocalStudent
		exch   *d.ExchangeStudent
	)
	if len(s.HomeAddress) > 0 {
		var tmp struct{ Street, City, Country string }
		if err := json.Unmarshal(s.HomeAddress, &tmp); err != nil {
			return d.Student{}, err
		}
		home = &d.Address{Street: tmp.Street, City: tmp.City, Country: tmp.Country}
	}
	if len(s.CourseGrades) > 0 {
		if err := json.Unmarshal(s.CourseGrades, &course); err != nil {
			return d.Student{}, err
		}
	}
	if len(s.Local) > 0 {
		var tmp struct {
			NationalID  string `json:"national_id"`
			Scholarship bool   `json:"scholarship"`
		}
		if err := json.Unmarshal(s.Local, &tmp); err != nil {
			return d.Student{}, err
		}
		loc = &d.LocalStudent{NationalID: tmp.NationalID, Scholarship: tmp.Scholarship}
	}
	if len(s.Exchange) > 0 {
		var tmp struct {
			HomeUniversity, CountryOfOrigin string
			ProgramStart, ProgramEnd        time.Time
		}
		if err := json.Unmarshal(s.Exchange, &tmp); err != nil {
			return d.Student{}, err
		}
		exch = &d.ExchangeStudent{
			HomeUniversity:  tmp.HomeUniversity,
			CountryOfOrigin: tmp.CountryOfOrigin,
			ProgramStart:    tmp.ProgramStart,
			ProgramEnd:      tmp.ProgramEnd,
		}
	}
	var middle *string
	if s.MiddleName.Valid {
		middle = &s.MiddleName.String
	}
	var status d.Status
	if s.Status.Valid {
		status = d.Status(s.Status.String)
	}

	return d.Student{
		ID:         s.ID,
		FirstName:  s.FirstName,
		LastName:   s.LastName,
		MiddleName: middle,
		Grade:      s.Grade,
		Status:     status,
		Home:       home,
		Course:     course,
		Friends:    []string(s.Friends),
		CreatedAt:  s.CreatedAt,
		Details:    d.Details{Local: loc, Exchange: exch},
	}, nil
}
