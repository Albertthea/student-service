// Package handler provides gRPC handlers and helpers for converting between
// proto-generated types and domain models.
package handler

import (
	"time"

	d "example.com/student-service/domain"
	pb "example.com/student-service/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// --- Status ---

func protoStatusToDomain(s pb.Student_Status) d.Status {
	switch s {
	case pb.Student_ACTIVE:
		return d.StatusActive
	case pb.Student_GRADUATED:
		return d.StatusGraduated
	case pb.Student_SUSPENDED:
		return d.StatusSuspended
	case pb.Student_EXCHANGE:
		return d.StatusExchange
	default:
		return d.StatusUnspecified
	}
}

func domainStatusToProto(s d.Status) pb.Student_Status {
	switch s {
	case d.StatusActive:
		return pb.Student_ACTIVE
	case d.StatusGraduated:
		return pb.Student_GRADUATED
	case d.StatusSuspended:
		return pb.Student_SUSPENDED
	case d.StatusExchange:
		return pb.Student_EXCHANGE
	default:
		return pb.Student_STATUS_UNSPECIFIED
	}
}

// --- Address ---

func protoToDomainAddress(a *pb.Student_Address) *d.Address {
	if a == nil {
		return nil
	}
	return &d.Address{Street: a.GetStreet(), City: a.GetCity(), Country: a.GetCountry()}
}

func domainToProtoAddress(a *d.Address) *pb.Student_Address {
	if a == nil {
		return nil
	}
	return &pb.Student_Address{Street: a.Street, City: a.City, Country: a.Country}
}

// --- Details (Local / Exchange via oneof) ---

func protoLocalToDomain(l *pb.Student_LocalStudent) *d.LocalStudent {
	if l == nil {
		return nil
	}
	return &d.LocalStudent{
		NationalID:  l.GetNationalId(),
		Scholarship: l.GetScholarship(),
	}
}

func protoExchangeToDomain(e *pb.Student_ExchangeStudent) *d.ExchangeStudent {
	if e == nil {
		return nil
	}
	var start, end time.Time
	if ts := e.GetProgramStart(); ts != nil {
		start = ts.AsTime()
	}
	if ts := e.GetProgramEnd(); ts != nil {
		end = ts.AsTime()
	}
	return &d.ExchangeStudent{
		HomeUniversity:  e.GetHomeUniversity(),
		CountryOfOrigin: e.GetCountryOfOrigin(),
		ProgramStart:    start,
		ProgramEnd:      end,
	}
}

func domainLocalToProto(l *d.LocalStudent) *pb.Student_LocalStudent {
	if l == nil {
		return nil
	}
	return &pb.Student_LocalStudent{
		NationalId:  l.NationalID,
		Scholarship: l.Scholarship,
	}
}

func domainExchangeToProto(e *d.ExchangeStudent) *pb.Student_ExchangeStudent {
	if e == nil {
		return nil
	}
	return &pb.Student_ExchangeStudent{
		HomeUniversity:  e.HomeUniversity,
		CountryOfOrigin: e.CountryOfOrigin,
		ProgramStart:    timestamppb.New(e.ProgramStart),
		ProgramEnd:      timestamppb.New(e.ProgramEnd),
	}
}

// ProtoToDomainStudent converts a proto Student message into a domain Student entity.
func ProtoToDomainStudent(p *pb.Student) d.Student {
	if p == nil {
		return d.Student{}
	}

	// MiddleName: в proto это *string (oneof), геттер вернёт "" если nil.
	var middle *string
	if mn := p.GetMiddleName(); mn != "" {
		m := mn
		middle = &m
	}

	// Map копируем, чтобы не тащить ссылку из proto.
	course := map[string]string{}
	for k, v := range p.GetCourseGrades() {
		course[k] = v
	}

	var createdAt time.Time
	if ts := p.GetCreatedAt(); ts != nil {
		createdAt = ts.AsTime()
	}

	// oneof: берем либо Local, либо Exchange (или оба nil).
	local := protoLocalToDomain(p.GetLocal())
	exch := protoExchangeToDomain(p.GetExchange())

	return d.Student{
		ID:         p.GetId(),
		FirstName:  p.GetFirstName(),
		LastName:   p.GetLastName(),
		MiddleName: middle,
		Grade:      p.GetGrade(),
		Status:     protoStatusToDomain(p.GetStatus()),
		Home:       protoToDomainAddress(p.GetHomeAddress()),
		Course:     course,
		Friends:    append([]string(nil), p.GetFriends()...),
		CreatedAt:  createdAt,
		Details:    d.Details{Local: local, Exchange: exch},
	}
}

// DomainToProtoStudent converts a domain Student entity into a proto Student message.
func DomainToProtoStudent(s d.Student) *pb.Student {
	out := &pb.Student{
		Id:          s.ID,
		FirstName:   s.FirstName,
		LastName:    s.LastName,
		Grade:       s.Grade,
		Status:      domainStatusToProto(s.Status),
		HomeAddress: domainToProtoAddress(s.Home),
		Friends:     append([]string(nil), s.Friends...),
		CreatedAt:   timestamppb.New(s.CreatedAt),
	}

	// MiddleName: нужно положить *string.
	if s.MiddleName != nil {
		out.MiddleName = s.MiddleName
	}

	// CourseGrades map.
	if len(s.Course) > 0 {
		out.CourseGrades = make(map[string]string, len(s.Course))
		for k, v := range s.Course {
			out.CourseGrades[k] = v
		}
	}

	// oneof StudentDetails: положим либо Local, либо Exchange.
	if s.Details.Local != nil {
		out.StudentDetails = &pb.Student_Local{
			Local: domainLocalToProto(s.Details.Local),
		}
	} else if s.Details.Exchange != nil {
		out.StudentDetails = &pb.Student_Exchange{
			Exchange: domainExchangeToProto(s.Details.Exchange),
		}
	}

	return out
}
