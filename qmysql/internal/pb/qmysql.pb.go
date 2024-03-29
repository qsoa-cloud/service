// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.12.4
// source: qmysql.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetDsnReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *GetDsnReq) Reset() {
	*x = GetDsnReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_qmysql_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDsnReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDsnReq) ProtoMessage() {}

func (x *GetDsnReq) ProtoReflect() protoreflect.Message {
	mi := &file_qmysql_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDsnReq.ProtoReflect.Descriptor instead.
func (*GetDsnReq) Descriptor() ([]byte, []int) {
	return file_qmysql_proto_rawDescGZIP(), []int{0}
}

func (x *GetDsnReq) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type GetDsnResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Dsn string `protobuf:"bytes,1,opt,name=dsn,proto3" json:"dsn,omitempty"`
}

func (x *GetDsnResp) Reset() {
	*x = GetDsnResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_qmysql_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDsnResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDsnResp) ProtoMessage() {}

func (x *GetDsnResp) ProtoReflect() protoreflect.Message {
	mi := &file_qmysql_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDsnResp.ProtoReflect.Descriptor instead.
func (*GetDsnResp) Descriptor() ([]byte, []int) {
	return file_qmysql_proto_rawDescGZIP(), []int{1}
}

func (x *GetDsnResp) GetDsn() string {
	if x != nil {
		return x.Dsn
	}
	return ""
}

var File_qmysql_proto protoreflect.FileDescriptor

var file_qmysql_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x71, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x1f,
	0x0a, 0x09, 0x47, 0x65, 0x74, 0x44, 0x73, 0x6e, 0x52, 0x65, 0x71, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22,
	0x1e, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x44, 0x73, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x12, 0x10, 0x0a,
	0x03, 0x64, 0x73, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x64, 0x73, 0x6e, 0x32,
	0x2a, 0x0a, 0x05, 0x4d, 0x79, 0x53, 0x71, 0x6c, 0x12, 0x21, 0x0a, 0x06, 0x47, 0x65, 0x74, 0x44,
	0x73, 0x6e, 0x12, 0x0a, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x73, 0x6e, 0x52, 0x65, 0x71, 0x1a, 0x0b,
	0x2e, 0x47, 0x65, 0x74, 0x44, 0x73, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x42, 0x06, 0x5a, 0x04, 0x2e,
	0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_qmysql_proto_rawDescOnce sync.Once
	file_qmysql_proto_rawDescData = file_qmysql_proto_rawDesc
)

func file_qmysql_proto_rawDescGZIP() []byte {
	file_qmysql_proto_rawDescOnce.Do(func() {
		file_qmysql_proto_rawDescData = protoimpl.X.CompressGZIP(file_qmysql_proto_rawDescData)
	})
	return file_qmysql_proto_rawDescData
}

var file_qmysql_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_qmysql_proto_goTypes = []interface{}{
	(*GetDsnReq)(nil),  // 0: GetDsnReq
	(*GetDsnResp)(nil), // 1: GetDsnResp
}
var file_qmysql_proto_depIdxs = []int32{
	0, // 0: MySql.GetDsn:input_type -> GetDsnReq
	1, // 1: MySql.GetDsn:output_type -> GetDsnResp
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_qmysql_proto_init() }
func file_qmysql_proto_init() {
	if File_qmysql_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_qmysql_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDsnReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_qmysql_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDsnResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_qmysql_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_qmysql_proto_goTypes,
		DependencyIndexes: file_qmysql_proto_depIdxs,
		MessageInfos:      file_qmysql_proto_msgTypes,
	}.Build()
	File_qmysql_proto = out.File
	file_qmysql_proto_rawDesc = nil
	file_qmysql_proto_goTypes = nil
	file_qmysql_proto_depIdxs = nil
}
