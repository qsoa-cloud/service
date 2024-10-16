// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
// source: dfs.proto

package dfspb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Dfs_File_FullMethodName      = "/Dfs/File"
	Dfs_MkDir_FullMethodName     = "/Dfs/MkDir"
	Dfs_RemoveAll_FullMethodName = "/Dfs/RemoveAll"
	Dfs_Rename_FullMethodName    = "/Dfs/Rename"
	Dfs_Stat_FullMethodName      = "/Dfs/Stat"
)

// DfsClient is the client API for Dfs service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DfsClient interface {
	File(ctx context.Context, opts ...grpc.CallOption) (Dfs_FileClient, error)
	MkDir(ctx context.Context, in *MkDirReq, opts ...grpc.CallOption) (*emptypb.Empty, error)
	RemoveAll(ctx context.Context, in *RemoveAllReq, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Rename(ctx context.Context, in *RenameReq, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Stat(ctx context.Context, in *StatReq, opts ...grpc.CallOption) (*StatResp, error)
}

type dfsClient struct {
	cc grpc.ClientConnInterface
}

func NewDfsClient(cc grpc.ClientConnInterface) DfsClient {
	return &dfsClient{cc}
}

func (c *dfsClient) File(ctx context.Context, opts ...grpc.CallOption) (Dfs_FileClient, error) {
	stream, err := c.cc.NewStream(ctx, &Dfs_ServiceDesc.Streams[0], Dfs_File_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &dfsFileClient{stream}
	return x, nil
}

type Dfs_FileClient interface {
	Send(*FileReq) error
	Recv() (*FileResp, error)
	grpc.ClientStream
}

type dfsFileClient struct {
	grpc.ClientStream
}

func (x *dfsFileClient) Send(m *FileReq) error {
	return x.ClientStream.SendMsg(m)
}

func (x *dfsFileClient) Recv() (*FileResp, error) {
	m := new(FileResp)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *dfsClient) MkDir(ctx context.Context, in *MkDirReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Dfs_MkDir_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dfsClient) RemoveAll(ctx context.Context, in *RemoveAllReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Dfs_RemoveAll_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dfsClient) Rename(ctx context.Context, in *RenameReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Dfs_Rename_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dfsClient) Stat(ctx context.Context, in *StatReq, opts ...grpc.CallOption) (*StatResp, error) {
	out := new(StatResp)
	err := c.cc.Invoke(ctx, Dfs_Stat_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DfsServer is the server API for Dfs service.
// All implementations must embed UnimplementedDfsServer
// for forward compatibility
type DfsServer interface {
	File(Dfs_FileServer) error
	MkDir(context.Context, *MkDirReq) (*emptypb.Empty, error)
	RemoveAll(context.Context, *RemoveAllReq) (*emptypb.Empty, error)
	Rename(context.Context, *RenameReq) (*emptypb.Empty, error)
	Stat(context.Context, *StatReq) (*StatResp, error)
	mustEmbedUnimplementedDfsServer()
}

// UnimplementedDfsServer must be embedded to have forward compatible implementations.
type UnimplementedDfsServer struct {
}

func (UnimplementedDfsServer) File(Dfs_FileServer) error {
	return status.Errorf(codes.Unimplemented, "method File not implemented")
}
func (UnimplementedDfsServer) MkDir(context.Context, *MkDirReq) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MkDir not implemented")
}
func (UnimplementedDfsServer) RemoveAll(context.Context, *RemoveAllReq) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveAll not implemented")
}
func (UnimplementedDfsServer) Rename(context.Context, *RenameReq) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Rename not implemented")
}
func (UnimplementedDfsServer) Stat(context.Context, *StatReq) (*StatResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stat not implemented")
}
func (UnimplementedDfsServer) mustEmbedUnimplementedDfsServer() {}

// UnsafeDfsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DfsServer will
// result in compilation errors.
type UnsafeDfsServer interface {
	mustEmbedUnimplementedDfsServer()
}

func RegisterDfsServer(s grpc.ServiceRegistrar, srv DfsServer) {
	s.RegisterService(&Dfs_ServiceDesc, srv)
}

func _Dfs_File_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(DfsServer).File(&dfsFileServer{stream})
}

type Dfs_FileServer interface {
	Send(*FileResp) error
	Recv() (*FileReq, error)
	grpc.ServerStream
}

type dfsFileServer struct {
	grpc.ServerStream
}

func (x *dfsFileServer) Send(m *FileResp) error {
	return x.ServerStream.SendMsg(m)
}

func (x *dfsFileServer) Recv() (*FileReq, error) {
	m := new(FileReq)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Dfs_MkDir_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MkDirReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DfsServer).MkDir(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Dfs_MkDir_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DfsServer).MkDir(ctx, req.(*MkDirReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Dfs_RemoveAll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveAllReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DfsServer).RemoveAll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Dfs_RemoveAll_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DfsServer).RemoveAll(ctx, req.(*RemoveAllReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Dfs_Rename_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RenameReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DfsServer).Rename(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Dfs_Rename_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DfsServer).Rename(ctx, req.(*RenameReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Dfs_Stat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DfsServer).Stat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Dfs_Stat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DfsServer).Stat(ctx, req.(*StatReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Dfs_ServiceDesc is the grpc.ServiceDesc for Dfs service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Dfs_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Dfs",
	HandlerType: (*DfsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MkDir",
			Handler:    _Dfs_MkDir_Handler,
		},
		{
			MethodName: "RemoveAll",
			Handler:    _Dfs_RemoveAll_Handler,
		},
		{
			MethodName: "Rename",
			Handler:    _Dfs_Rename_Handler,
		},
		{
			MethodName: "Stat",
			Handler:    _Dfs_Stat_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "File",
			Handler:       _Dfs_File_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "dfs.proto",
}
