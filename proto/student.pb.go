// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: proto/student.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Student struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	FirstName     string                 `protobuf:"bytes,2,opt,name=first_name,json=firstName,proto3" json:"first_name,omitempty"`
	LastName      string                 `protobuf:"bytes,3,opt,name=last_name,json=lastName,proto3" json:"last_name,omitempty"`
	Grade         int32                  `protobuf:"varint,4,opt,name=grade,proto3" json:"grade,omitempty"`
	CreatedAt     *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Student) Reset() {
	*x = Student{}
	mi := &file_proto_student_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Student) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Student) ProtoMessage() {}

func (x *Student) ProtoReflect() protoreflect.Message {
	mi := &file_proto_student_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Student.ProtoReflect.Descriptor instead.
func (*Student) Descriptor() ([]byte, []int) {
	return file_proto_student_proto_rawDescGZIP(), []int{0}
}

func (x *Student) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Student) GetFirstName() string {
	if x != nil {
		return x.FirstName
	}
	return ""
}

func (x *Student) GetLastName() string {
	if x != nil {
		return x.LastName
	}
	return ""
}

func (x *Student) GetGrade() int32 {
	if x != nil {
		return x.Grade
	}
	return 0
}

func (x *Student) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

type CreateStudentRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	FirstName     string                 `protobuf:"bytes,1,opt,name=first_name,json=firstName,proto3" json:"first_name,omitempty"`
	LastName      string                 `protobuf:"bytes,2,opt,name=last_name,json=lastName,proto3" json:"last_name,omitempty"`
	Grade         int32                  `protobuf:"varint,3,opt,name=grade,proto3" json:"grade,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateStudentRequest) Reset() {
	*x = CreateStudentRequest{}
	mi := &file_proto_student_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateStudentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateStudentRequest) ProtoMessage() {}

func (x *CreateStudentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_student_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateStudentRequest.ProtoReflect.Descriptor instead.
func (*CreateStudentRequest) Descriptor() ([]byte, []int) {
	return file_proto_student_proto_rawDescGZIP(), []int{1}
}

func (x *CreateStudentRequest) GetFirstName() string {
	if x != nil {
		return x.FirstName
	}
	return ""
}

func (x *CreateStudentRequest) GetLastName() string {
	if x != nil {
		return x.LastName
	}
	return ""
}

func (x *CreateStudentRequest) GetGrade() int32 {
	if x != nil {
		return x.Grade
	}
	return 0
}

type CreateStudentResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateStudentResponse) Reset() {
	*x = CreateStudentResponse{}
	mi := &file_proto_student_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateStudentResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateStudentResponse) ProtoMessage() {}

func (x *CreateStudentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_student_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateStudentResponse.ProtoReflect.Descriptor instead.
func (*CreateStudentResponse) Descriptor() ([]byte, []int) {
	return file_proto_student_proto_rawDescGZIP(), []int{2}
}

func (x *CreateStudentResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetStudentRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetStudentRequest) Reset() {
	*x = GetStudentRequest{}
	mi := &file_proto_student_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetStudentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStudentRequest) ProtoMessage() {}

func (x *GetStudentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_student_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStudentRequest.ProtoReflect.Descriptor instead.
func (*GetStudentRequest) Descriptor() ([]byte, []int) {
	return file_proto_student_proto_rawDescGZIP(), []int{3}
}

func (x *GetStudentRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetStudentResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Student       *Student               `protobuf:"bytes,1,opt,name=student,proto3" json:"student,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetStudentResponse) Reset() {
	*x = GetStudentResponse{}
	mi := &file_proto_student_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetStudentResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStudentResponse) ProtoMessage() {}

func (x *GetStudentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_student_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStudentResponse.ProtoReflect.Descriptor instead.
func (*GetStudentResponse) Descriptor() ([]byte, []int) {
	return file_proto_student_proto_rawDescGZIP(), []int{4}
}

func (x *GetStudentResponse) GetStudent() *Student {
	if x != nil {
		return x.Student
	}
	return nil
}

type UpdateStudentRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Student       *Student               `protobuf:"bytes,1,opt,name=student,proto3" json:"student,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateStudentRequest) Reset() {
	*x = UpdateStudentRequest{}
	mi := &file_proto_student_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateStudentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateStudentRequest) ProtoMessage() {}

func (x *UpdateStudentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_student_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateStudentRequest.ProtoReflect.Descriptor instead.
func (*UpdateStudentRequest) Descriptor() ([]byte, []int) {
	return file_proto_student_proto_rawDescGZIP(), []int{5}
}

func (x *UpdateStudentRequest) GetStudent() *Student {
	if x != nil {
		return x.Student
	}
	return nil
}

type DeleteStudentRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteStudentRequest) Reset() {
	*x = DeleteStudentRequest{}
	mi := &file_proto_student_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteStudentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteStudentRequest) ProtoMessage() {}

func (x *DeleteStudentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_student_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteStudentRequest.ProtoReflect.Descriptor instead.
func (*DeleteStudentRequest) Descriptor() ([]byte, []int) {
	return file_proto_student_proto_rawDescGZIP(), []int{6}
}

func (x *DeleteStudentRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type ListStudentsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Grade         int32                  `protobuf:"varint,1,opt,name=grade,proto3" json:"grade,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListStudentsRequest) Reset() {
	*x = ListStudentsRequest{}
	mi := &file_proto_student_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListStudentsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListStudentsRequest) ProtoMessage() {}

func (x *ListStudentsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_student_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListStudentsRequest.ProtoReflect.Descriptor instead.
func (*ListStudentsRequest) Descriptor() ([]byte, []int) {
	return file_proto_student_proto_rawDescGZIP(), []int{7}
}

func (x *ListStudentsRequest) GetGrade() int32 {
	if x != nil {
		return x.Grade
	}
	return 0
}

type ListStudentsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Students      []*Student             `protobuf:"bytes,1,rep,name=students,proto3" json:"students,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListStudentsResponse) Reset() {
	*x = ListStudentsResponse{}
	mi := &file_proto_student_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListStudentsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListStudentsResponse) ProtoMessage() {}

func (x *ListStudentsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_student_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListStudentsResponse.ProtoReflect.Descriptor instead.
func (*ListStudentsResponse) Descriptor() ([]byte, []int) {
	return file_proto_student_proto_rawDescGZIP(), []int{8}
}

func (x *ListStudentsResponse) GetStudents() []*Student {
	if x != nil {
		return x.Students
	}
	return nil
}

var File_proto_student_proto protoreflect.FileDescriptor

const file_proto_student_proto_rawDesc = "" +
	"\n" +
	"\x13proto/student.proto\x12\astudent\x1a\x1bgoogle/protobuf/empty.proto\x1a\x1fgoogle/protobuf/timestamp.proto\"\xa6\x01\n" +
	"\aStudent\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x1d\n" +
	"\n" +
	"first_name\x18\x02 \x01(\tR\tfirstName\x12\x1b\n" +
	"\tlast_name\x18\x03 \x01(\tR\blastName\x12\x14\n" +
	"\x05grade\x18\x04 \x01(\x05R\x05grade\x129\n" +
	"\n" +
	"created_at\x18\x05 \x01(\v2\x1a.google.protobuf.TimestampR\tcreatedAt\"h\n" +
	"\x14CreateStudentRequest\x12\x1d\n" +
	"\n" +
	"first_name\x18\x01 \x01(\tR\tfirstName\x12\x1b\n" +
	"\tlast_name\x18\x02 \x01(\tR\blastName\x12\x14\n" +
	"\x05grade\x18\x03 \x01(\x05R\x05grade\"'\n" +
	"\x15CreateStudentResponse\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\"#\n" +
	"\x11GetStudentRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\"@\n" +
	"\x12GetStudentResponse\x12*\n" +
	"\astudent\x18\x01 \x01(\v2\x10.student.StudentR\astudent\"B\n" +
	"\x14UpdateStudentRequest\x12*\n" +
	"\astudent\x18\x01 \x01(\v2\x10.student.StudentR\astudent\"&\n" +
	"\x14DeleteStudentRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\"+\n" +
	"\x13ListStudentsRequest\x12\x14\n" +
	"\x05grade\x18\x01 \x01(\x05R\x05grade\"D\n" +
	"\x14ListStudentsResponse\x12,\n" +
	"\bstudents\x18\x01 \x03(\v2\x10.student.StudentR\bstudents2\x84\x03\n" +
	"\x0eStudentService\x12N\n" +
	"\rCreateStudent\x12\x1d.student.CreateStudentRequest\x1a\x1e.student.CreateStudentResponse\x12E\n" +
	"\n" +
	"GetStudent\x12\x1a.student.GetStudentRequest\x1a\x1b.student.GetStudentResponse\x12F\n" +
	"\rUpdateStudent\x12\x1d.student.UpdateStudentRequest\x1a\x16.google.protobuf.Empty\x12F\n" +
	"\rDeleteStudent\x12\x1d.student.DeleteStudentRequest\x1a\x16.google.protobuf.Empty\x12K\n" +
	"\fListStudents\x12\x1c.student.ListStudentsRequest\x1a\x1d.student.ListStudentsResponseB)Z'example.com/student-service/proto;protob\x06proto3"

var (
	file_proto_student_proto_rawDescOnce sync.Once
	file_proto_student_proto_rawDescData []byte
)

func file_proto_student_proto_rawDescGZIP() []byte {
	file_proto_student_proto_rawDescOnce.Do(func() {
		file_proto_student_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_student_proto_rawDesc), len(file_proto_student_proto_rawDesc)))
	})
	return file_proto_student_proto_rawDescData
}

var file_proto_student_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_proto_student_proto_goTypes = []any{
	(*Student)(nil),               // 0: student.Student
	(*CreateStudentRequest)(nil),  // 1: student.CreateStudentRequest
	(*CreateStudentResponse)(nil), // 2: student.CreateStudentResponse
	(*GetStudentRequest)(nil),     // 3: student.GetStudentRequest
	(*GetStudentResponse)(nil),    // 4: student.GetStudentResponse
	(*UpdateStudentRequest)(nil),  // 5: student.UpdateStudentRequest
	(*DeleteStudentRequest)(nil),  // 6: student.DeleteStudentRequest
	(*ListStudentsRequest)(nil),   // 7: student.ListStudentsRequest
	(*ListStudentsResponse)(nil),  // 8: student.ListStudentsResponse
	(*timestamppb.Timestamp)(nil), // 9: google.protobuf.Timestamp
	(*emptypb.Empty)(nil),         // 10: google.protobuf.Empty
}
var file_proto_student_proto_depIdxs = []int32{
	9,  // 0: student.Student.created_at:type_name -> google.protobuf.Timestamp
	0,  // 1: student.GetStudentResponse.student:type_name -> student.Student
	0,  // 2: student.UpdateStudentRequest.student:type_name -> student.Student
	0,  // 3: student.ListStudentsResponse.students:type_name -> student.Student
	1,  // 4: student.StudentService.CreateStudent:input_type -> student.CreateStudentRequest
	3,  // 5: student.StudentService.GetStudent:input_type -> student.GetStudentRequest
	5,  // 6: student.StudentService.UpdateStudent:input_type -> student.UpdateStudentRequest
	6,  // 7: student.StudentService.DeleteStudent:input_type -> student.DeleteStudentRequest
	7,  // 8: student.StudentService.ListStudents:input_type -> student.ListStudentsRequest
	2,  // 9: student.StudentService.CreateStudent:output_type -> student.CreateStudentResponse
	4,  // 10: student.StudentService.GetStudent:output_type -> student.GetStudentResponse
	10, // 11: student.StudentService.UpdateStudent:output_type -> google.protobuf.Empty
	10, // 12: student.StudentService.DeleteStudent:output_type -> google.protobuf.Empty
	8,  // 13: student.StudentService.ListStudents:output_type -> student.ListStudentsResponse
	9,  // [9:14] is the sub-list for method output_type
	4,  // [4:9] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_proto_student_proto_init() }
func file_proto_student_proto_init() {
	if File_proto_student_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_student_proto_rawDesc), len(file_proto_student_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_student_proto_goTypes,
		DependencyIndexes: file_proto_student_proto_depIdxs,
		MessageInfos:      file_proto_student_proto_msgTypes,
	}.Build()
	File_proto_student_proto = out.File
	file_proto_student_proto_goTypes = nil
	file_proto_student_proto_depIdxs = nil
}
