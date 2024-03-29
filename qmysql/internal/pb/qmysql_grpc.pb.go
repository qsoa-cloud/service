// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.12.4
// source: qmysql.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	MySql_GetDsn_FullMethodName = "/MySql/GetDsn"
)

// MySqlClient is the client API for MySql service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MySqlClient interface {
	GetDsn(ctx context.Context, in *GetDsnReq, opts ...grpc.CallOption) (*GetDsnResp, error)
}

type mySqlClient struct {
	cc grpc.ClientConnInterface
}

func NewMySqlClient(cc grpc.ClientConnInterface) MySqlClient {
	return &mySqlClient{cc}
}

func (c *mySqlClient) GetDsn(ctx context.Context, in *GetDsnReq, opts ...grpc.CallOption) (*GetDsnResp, error) {
	out := new(GetDsnResp)
	err := c.cc.Invoke(ctx, MySql_GetDsn_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MySqlServer is the server API for MySql service.
// All implementations must embed UnimplementedMySqlServer
// for forward compatibility
type MySqlServer interface {
	GetDsn(context.Context, *GetDsnReq) (*GetDsnResp, error)
	mustEmbedUnimplementedMySqlServer()
}

// UnimplementedMySqlServer must be embedded to have forward compatible implementations.
type UnimplementedMySqlServer struct {
}

func (UnimplementedMySqlServer) GetDsn(context.Context, *GetDsnReq) (*GetDsnResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDsn not implemented")
}
func (UnimplementedMySqlServer) mustEmbedUnimplementedMySqlServer() {}

// UnsafeMySqlServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MySqlServer will
// result in compilation errors.
type UnsafeMySqlServer interface {
	mustEmbedUnimplementedMySqlServer()
}

func RegisterMySqlServer(s grpc.ServiceRegistrar, srv MySqlServer) {
	s.RegisterService(&MySql_ServiceDesc, srv)
}

func _MySql_GetDsn_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDsnReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MySqlServer).GetDsn(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MySql_GetDsn_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MySqlServer).GetDsn(ctx, req.(*GetDsnReq))
	}
	return interceptor(ctx, in, info, handler)
}

// MySql_ServiceDesc is the grpc.ServiceDesc for MySql service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MySql_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "MySql",
	HandlerType: (*MySqlServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetDsn",
			Handler:    _MySql_GetDsn_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "qmysql.proto",
}
