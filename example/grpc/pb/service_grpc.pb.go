// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.12.4
// source: service.proto

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
	Test_Sum_FullMethodName      = "/Test/Sum"
	Test_PingPong_FullMethodName = "/Test/PingPong"
)

// TestClient is the client API for Test service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TestClient interface {
	Sum(ctx context.Context, in *SumReq, opts ...grpc.CallOption) (*SumResp, error)
	PingPong(ctx context.Context, opts ...grpc.CallOption) (Test_PingPongClient, error)
}

type testClient struct {
	cc grpc.ClientConnInterface
}

func NewTestClient(cc grpc.ClientConnInterface) TestClient {
	return &testClient{cc}
}

func (c *testClient) Sum(ctx context.Context, in *SumReq, opts ...grpc.CallOption) (*SumResp, error) {
	out := new(SumResp)
	err := c.cc.Invoke(ctx, Test_Sum_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *testClient) PingPong(ctx context.Context, opts ...grpc.CallOption) (Test_PingPongClient, error) {
	stream, err := c.cc.NewStream(ctx, &Test_ServiceDesc.Streams[0], Test_PingPong_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &testPingPongClient{stream}
	return x, nil
}

type Test_PingPongClient interface {
	Send(*Ping) error
	Recv() (*Pong, error)
	grpc.ClientStream
}

type testPingPongClient struct {
	grpc.ClientStream
}

func (x *testPingPongClient) Send(m *Ping) error {
	return x.ClientStream.SendMsg(m)
}

func (x *testPingPongClient) Recv() (*Pong, error) {
	m := new(Pong)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TestServer is the server API for Test service.
// All implementations must embed UnimplementedTestServer
// for forward compatibility
type TestServer interface {
	Sum(context.Context, *SumReq) (*SumResp, error)
	PingPong(Test_PingPongServer) error
	mustEmbedUnimplementedTestServer()
}

// UnimplementedTestServer must be embedded to have forward compatible implementations.
type UnimplementedTestServer struct {
}

func (UnimplementedTestServer) Sum(context.Context, *SumReq) (*SumResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Sum not implemented")
}
func (UnimplementedTestServer) PingPong(Test_PingPongServer) error {
	return status.Errorf(codes.Unimplemented, "method PingPong not implemented")
}
func (UnimplementedTestServer) mustEmbedUnimplementedTestServer() {}

// UnsafeTestServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TestServer will
// result in compilation errors.
type UnsafeTestServer interface {
	mustEmbedUnimplementedTestServer()
}

func RegisterTestServer(s grpc.ServiceRegistrar, srv TestServer) {
	s.RegisterService(&Test_ServiceDesc, srv)
}

func _Test_Sum_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SumReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestServer).Sum(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Test_Sum_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServer).Sum(ctx, req.(*SumReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Test_PingPong_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TestServer).PingPong(&testPingPongServer{stream})
}

type Test_PingPongServer interface {
	Send(*Pong) error
	Recv() (*Ping, error)
	grpc.ServerStream
}

type testPingPongServer struct {
	grpc.ServerStream
}

func (x *testPingPongServer) Send(m *Pong) error {
	return x.ServerStream.SendMsg(m)
}

func (x *testPingPongServer) Recv() (*Ping, error) {
	m := new(Ping)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Test_ServiceDesc is the grpc.ServiceDesc for Test service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Test_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Test",
	HandlerType: (*TestServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Sum",
			Handler:    _Test_Sum_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "PingPong",
			Handler:       _Test_PingPong_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "service.proto",
}
