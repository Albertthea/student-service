// Package mapper provides conversion helpers between proto, domain, and repository layers.
package mapper

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	d "example.com/student-service/domain"
	"example.com/student-service/proto"
	repo "example.com/student-service/repository/student"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// --- Create/Update ---

// CreateReqToRepo converts a proto.CreateStudentRequest into a repo.Student
// with the provided ID and timestamp.
func CreateReqToRepo(id string, now time.Time, req *proto.CreateStudentRequest) (repo.Student, error) {
	if req.Student == nil {
		return repo.Student{}, errors.New("student is required")
	}
	s := req.Student

	st := repo.Student{
		ID:           id,
		FirstName:    s.FirstName,
		LastName:     s.LastName,
		Grade:        s.Grade,
		CreatedAt:    now,
		MiddleName:   ptrToNull(s.MiddleName),
		Status:       enumToNull(s.Status),
		HomeAddress:  jsonToBytes(s.HomeAddress),
		CourseGrades: jsonToBytes(s.CourseGrades),
		Friends:      s.Friends,
	}

	switch v := s.StudentDetails.(type) {
	case *proto.Student_Local:
		st.Local = jsonToBytes(v.Local)
	case *proto.Student_Exchange:
		st.Exchange = jsonToBytes(v.Exchange)
	}
	return st, nil
}

// UpdateReqToRepo maps an UpdateStudentRequest and current repository.Student into a new repository.Student.
func UpdateReqToRepo(current *repo.Student, req *proto.UpdateStudentRequest) (repo.Student, error) {
	if req.Student == nil {
		return repo.Student{}, errors.New("student is required")
	}
	s := req.Student

	out := repo.Student{
		ID:           s.Id,
		FirstName:    s.FirstName,
		LastName:     s.LastName,
		Grade:        s.Grade,
		CreatedAt:    current.CreatedAt,
		MiddleName:   ptrToNull(s.MiddleName),
		Status:       enumToNull(s.Status),
		HomeAddress:  jsonToBytes(s.HomeAddress),
		CourseGrades: jsonToBytes(s.CourseGrades),
		Friends:      s.Friends,
	}

	switch v := s.StudentDetails.(type) {
	case *proto.Student_Local:
		out.Local = jsonToBytes(v.Local)
		out.Exchange = nil
	case *proto.Student_Exchange:
		out.Exchange = jsonToBytes(v.Exchange)
		out.Local = nil
	default:
		out.Local = current.Local
		out.Exchange = current.Exchange
	}
	return out, nil
}

// --- Repo → Proto ---

// RepoToProto converts a repo.Student into a proto.Student for gRPC responses.
func RepoToProto(st *repo.Student) *proto.Student {
	if st == nil {
		return nil
	}
	pb := &proto.Student{
		Id:           st.ID,
		FirstName:    st.FirstName,
		LastName:     st.LastName,
		Grade:        st.Grade,
		CreatedAt:    timestamppb.New(st.CreatedAt),
		MiddleName:   nullToPtr(st.MiddleName),
		Status:       parseStatus(st.Status),
		HomeAddress:  parseHome(st.HomeAddress),
		CourseGrades: parseCourse(st.CourseGrades),
		Friends:      st.Friends,
	}

	if len(st.Local) > 0 {
		var v proto.Student_LocalStudent
		if json.Unmarshal(st.Local, &v) == nil {
			pb.StudentDetails = &proto.Student_Local{Local: &v}
		}
	} else if len(st.Exchange) > 0 {
		var v proto.Student_ExchangeStudent
		if json.Unmarshal(st.Exchange, &v) == nil {
			pb.StudentDetails = &proto.Student_Exchange{Exchange: &v}
		}
	}
	return pb
}

// --- Domain → Repo ---

// DomainToRepo converts a domain.Student into a repo.Student for persistence.
func DomainToRepo(id string, now time.Time, dom d.Student) (repo.Student, error) {
	st := repo.Student{
		ID:           id,
		FirstName:    dom.FirstName,
		LastName:     dom.LastName,
		Grade:        dom.Grade,
		CreatedAt:    now,
		MiddleName:   ptrToNull(dom.MiddleName),
		Status:       sql.NullString{String: string(dom.Status), Valid: dom.Status != ""},
		CourseGrades: jsonToBytes(dom.Course),
		Friends:      dom.Friends,
	}

	if dom.Home != nil {
		st.HomeAddress = jsonToBytes(dom.Home)
	}

	if dom.Details.Local != nil {
		st.Local = jsonToBytes(dom.Details.Local)
	} else if dom.Details.Exchange != nil {
		st.Exchange = jsonToBytes(dom.Details.Exchange)
	}
	return st, nil
}

// --- helpers ---

func ptrToNull(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func enumToNull(s proto.Student_Status) sql.NullString {
	if s == proto.Student_STATUS_UNSPECIFIED {
		return sql.NullString{}
	}
	return sql.NullString{String: s.String(), Valid: true}
}

func jsonToBytes(v any) []byte {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Slice, reflect.Map:
		if rv.IsNil() {
			return nil
		}
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	if len(b) == 4 && string(b) == "null" {
		return nil
	}
	return b
}

func parseStatus(ns sql.NullString) proto.Student_Status {
	if !ns.Valid {
		return proto.Student_STATUS_UNSPECIFIED
	}
	if v, ok := proto.Student_Status_value[ns.String]; ok {
		return proto.Student_Status(v)
	}
	return proto.Student_STATUS_UNSPECIFIED
}

func parseHome(b []byte) *proto.Student_Address {
	if len(b) == 0 {
		return nil
	}
	var a proto.Student_Address
	if err := json.Unmarshal(b, &a); err != nil {
		return nil
	}
	return &a
}

func parseCourse(b []byte) map[string]string {
	if len(b) == 0 {
		return nil
	}
	var m map[string]string
	_ = json.Unmarshal(b, &m)
	return m
}
