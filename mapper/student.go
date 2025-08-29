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

//
// Create / Update (proto -> repo)
//

// CreateReqToRepo converts proto CreateStudentRequest into repo.Student,
// using the provided id and timestamp as authoritative values.
func CreateReqToRepo(id string, now time.Time, req *proto.CreateStudentRequest) (repo.Student, error) {
	if req == nil || req.GetStudent() == nil {
		return repo.Student{}, errors.New("student is required")
	}
	s := req.GetStudent()

	out := repo.Student{
		ID:           id,
		FirstName:    s.GetFirstName(),
		LastName:     s.GetLastName(),
		Grade:        s.GetGrade(),
		CreatedAt:    now,
		MiddleName:   strToNull(s.GetMiddleName()),
		Status:       enumToNull(s.GetStatus()),
		HomeAddress:  jsonToBytes(s.GetHomeAddress()),
		CourseGrades: jsonToBytes(s.GetCourseGrades()),
		Friends:      s.GetFriends(),
	}

	if v := s.GetLocal(); v != nil {
		out.Local = jsonToBytes(v)
	}
	if v := s.GetExchange(); v != nil {
		out.Exchange = jsonToBytes(v)
	}
	return out, nil
}

// UpdateReqToRepo maps UpdateStudentRequest and current repo.Student into a new repo.Student.
func UpdateReqToRepo(current *repo.Student, req *proto.UpdateStudentRequest) (repo.Student, error) {
	if req == nil || req.GetStudent() == nil {
		return repo.Student{}, errors.New("student is required")
	}
	s := req.GetStudent()

	out := repo.Student{
		ID:           s.GetId(),
		FirstName:    s.GetFirstName(),
		LastName:     s.GetLastName(),
		Grade:        s.GetGrade(),
		CreatedAt:    current.CreatedAt, // immutable
		MiddleName:   strToNull(s.GetMiddleName()),
		Status:       enumToNull(s.GetStatus()),
		HomeAddress:  jsonToBytes(s.GetHomeAddress()),
		CourseGrades: jsonToBytes(s.GetCourseGrades()),
		Friends:      s.GetFriends(),
	}

	switch {
	case s.GetLocal() != nil:
		out.Local = jsonToBytes(s.GetLocal())
		out.Exchange = nil
	case s.GetExchange() != nil:
		out.Exchange = jsonToBytes(s.GetExchange())
		out.Local = nil
	default:
		out.Local = current.Local
		out.Exchange = current.Exchange
	}
	return out, nil
}

// RepoToProto converts a repository.Student into a proto.Student.
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

// DomainToRepo converts a domain.Student into a repository.Student for persistence in the DB.
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

//
// helpers
//

// ptrToNull converts *string to sql.NullString.
func ptrToNull(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// strToNull converts string (from proto getter) to sql.NullString.
func strToNull(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// nullToPtr converts sql.NullString to *string.
func nullToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

// enumToNull converts enum to sql.NullString (unspecified -> NULL).
func enumToNull(s proto.Student_Status) sql.NullString {
	if s == proto.Student_STATUS_UNSPECIFIED {
		return sql.NullString{}
	}
	return sql.NullString{String: s.String(), Valid: true}
}

// jsonToBytes marshals value to JSON; nil (or JSON "null") -> nil slice.
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

// parseStatus converts sql.NullString to proto enum.
func parseStatus(ns sql.NullString) proto.Student_Status {
	if !ns.Valid {
		return proto.Student_STATUS_UNSPECIFIED
	}
	if v, ok := proto.Student_Status_value[ns.String]; ok {
		return proto.Student_Status(v)
	}
	return proto.Student_STATUS_UNSPECIFIED
}

// parseHome unmarshals JSON into *proto.Student_Address.
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

// parseCourse unmarshals JSON into map[string]string.
func parseCourse(b []byte) map[string]string {
	if len(b) == 0 {
		return nil
	}
	var m map[string]string
	_ = json.Unmarshal(b, &m)
	return m
}
