// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package Replica

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

// ReplicaServiceClient is the client API for ReplicaService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ReplicaServiceClient interface {
	CheckLeaderStatus(ctx context.Context, in *GetStatusRequest, opts ...grpc.CallOption) (*LeaderStatusResponse, error)
	ChooseNewLeader(ctx context.Context, in *WantToLeadRequest, opts ...grpc.CallOption) (*StatusResponse, error)
	CheckStatus(ctx context.Context, in *GetStatusRequest, opts ...grpc.CallOption) (*StatusResponse, error)
}

type replicaServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewReplicaServiceClient(cc grpc.ClientConnInterface) ReplicaServiceClient {
	return &replicaServiceClient{cc}
}

func (c *replicaServiceClient) CheckLeaderStatus(ctx context.Context, in *GetStatusRequest, opts ...grpc.CallOption) (*LeaderStatusResponse, error) {
	out := new(LeaderStatusResponse)
	err := c.cc.Invoke(ctx, "/Replica.ReplicaService/CheckLeaderStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replicaServiceClient) ChooseNewLeader(ctx context.Context, in *WantToLeadRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := c.cc.Invoke(ctx, "/Replica.ReplicaService/ChooseNewLeader", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replicaServiceClient) CheckStatus(ctx context.Context, in *GetStatusRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := c.cc.Invoke(ctx, "/Replica.ReplicaService/CheckStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReplicaServiceServer is the server API for ReplicaService service.
// All implementations must embed UnimplementedReplicaServiceServer
// for forward compatibility
type ReplicaServiceServer interface {
	CheckLeaderStatus(context.Context, *GetStatusRequest) (*LeaderStatusResponse, error)
	ChooseNewLeader(context.Context, *WantToLeadRequest) (*StatusResponse, error)
	CheckStatus(context.Context, *GetStatusRequest) (*StatusResponse, error)
	mustEmbedUnimplementedReplicaServiceServer()
}

// UnimplementedReplicaServiceServer must be embedded to have forward compatible implementations.
type UnimplementedReplicaServiceServer struct {
}

func (UnimplementedReplicaServiceServer) CheckLeaderStatus(context.Context, *GetStatusRequest) (*LeaderStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckLeaderStatus not implemented")
}
func (UnimplementedReplicaServiceServer) ChooseNewLeader(context.Context, *WantToLeadRequest) (*StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChooseNewLeader not implemented")
}
func (UnimplementedReplicaServiceServer) CheckStatus(context.Context, *GetStatusRequest) (*StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckStatus not implemented")
}
func (UnimplementedReplicaServiceServer) mustEmbedUnimplementedReplicaServiceServer() {}

// UnsafeReplicaServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReplicaServiceServer will
// result in compilation errors.
type UnsafeReplicaServiceServer interface {
	mustEmbedUnimplementedReplicaServiceServer()
}

func RegisterReplicaServiceServer(s grpc.ServiceRegistrar, srv ReplicaServiceServer) {
	s.RegisterService(&ReplicaService_ServiceDesc, srv)
}

func _ReplicaService_CheckLeaderStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicaServiceServer).CheckLeaderStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Replica.ReplicaService/CheckLeaderStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicaServiceServer).CheckLeaderStatus(ctx, req.(*GetStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplicaService_ChooseNewLeader_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WantToLeadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicaServiceServer).ChooseNewLeader(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Replica.ReplicaService/ChooseNewLeader",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicaServiceServer).ChooseNewLeader(ctx, req.(*WantToLeadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplicaService_CheckStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicaServiceServer).CheckStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Replica.ReplicaService/CheckStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicaServiceServer).CheckStatus(ctx, req.(*GetStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ReplicaService_ServiceDesc is the grpc.ServiceDesc for ReplicaService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ReplicaService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Replica.ReplicaService",
	HandlerType: (*ReplicaServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckLeaderStatus",
			Handler:    _ReplicaService_CheckLeaderStatus_Handler,
		},
		{
			MethodName: "ChooseNewLeader",
			Handler:    _ReplicaService_ChooseNewLeader_Handler,
		},
		{
			MethodName: "CheckStatus",
			Handler:    _ReplicaService_CheckStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "Replica/Replica.proto",
}
