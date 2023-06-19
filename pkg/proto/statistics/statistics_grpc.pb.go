// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.5
// source: statistics/statistics.proto

package statistics

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

// UserClient is the client API for User service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserClient interface {
	GetActiveUser(ctx context.Context, in *GetActiveUserReq, opts ...grpc.CallOption) (*GetActiveUserResp, error)
	GetActiveGroup(ctx context.Context, in *GetActiveGroupReq, opts ...grpc.CallOption) (*GetActiveGroupResp, error)
	GetMessageStatistics(ctx context.Context, in *GetMessageStatisticsReq, opts ...grpc.CallOption) (*GetMessageStatisticsResp, error)
	GetGroupStatistics(ctx context.Context, in *GetGroupStatisticsReq, opts ...grpc.CallOption) (*GetGroupStatisticsResp, error)
	GetUserStatistics(ctx context.Context, in *GetUserStatisticsReq, opts ...grpc.CallOption) (*GetUserStatisticsResp, error)
	GetGameStatistics(ctx context.Context, in *GetGameStatisticsReq, opts ...grpc.CallOption) (*GetGameStatisticsResp, error)
}

type userClient struct {
	cc grpc.ClientConnInterface
}

func NewUserClient(cc grpc.ClientConnInterface) UserClient {
	return &userClient{cc}
}

func (c *userClient) GetActiveUser(ctx context.Context, in *GetActiveUserReq, opts ...grpc.CallOption) (*GetActiveUserResp, error) {
	out := new(GetActiveUserResp)
	err := c.cc.Invoke(ctx, "/statistics.user/GetActiveUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) GetActiveGroup(ctx context.Context, in *GetActiveGroupReq, opts ...grpc.CallOption) (*GetActiveGroupResp, error) {
	out := new(GetActiveGroupResp)
	err := c.cc.Invoke(ctx, "/statistics.user/GetActiveGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) GetMessageStatistics(ctx context.Context, in *GetMessageStatisticsReq, opts ...grpc.CallOption) (*GetMessageStatisticsResp, error) {
	out := new(GetMessageStatisticsResp)
	err := c.cc.Invoke(ctx, "/statistics.user/GetMessageStatistics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) GetGroupStatistics(ctx context.Context, in *GetGroupStatisticsReq, opts ...grpc.CallOption) (*GetGroupStatisticsResp, error) {
	out := new(GetGroupStatisticsResp)
	err := c.cc.Invoke(ctx, "/statistics.user/GetGroupStatistics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) GetUserStatistics(ctx context.Context, in *GetUserStatisticsReq, opts ...grpc.CallOption) (*GetUserStatisticsResp, error) {
	out := new(GetUserStatisticsResp)
	err := c.cc.Invoke(ctx, "/statistics.user/GetUserStatistics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) GetGameStatistics(ctx context.Context, in *GetGameStatisticsReq, opts ...grpc.CallOption) (*GetGameStatisticsResp, error) {
	out := new(GetGameStatisticsResp)
	err := c.cc.Invoke(ctx, "/statistics.user/GetGameStatistics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserServer is the server API for User service.
// All implementations should embed UnimplementedUserServer
// for forward compatibility
type UserServer interface {
	GetActiveUser(context.Context, *GetActiveUserReq) (*GetActiveUserResp, error)
	GetActiveGroup(context.Context, *GetActiveGroupReq) (*GetActiveGroupResp, error)
	GetMessageStatistics(context.Context, *GetMessageStatisticsReq) (*GetMessageStatisticsResp, error)
	GetGroupStatistics(context.Context, *GetGroupStatisticsReq) (*GetGroupStatisticsResp, error)
	GetUserStatistics(context.Context, *GetUserStatisticsReq) (*GetUserStatisticsResp, error)
	GetGameStatistics(context.Context, *GetGameStatisticsReq) (*GetGameStatisticsResp, error)
}

// UnimplementedUserServer should be embedded to have forward compatible implementations.
type UnimplementedUserServer struct {
}

func (UnimplementedUserServer) GetActiveUser(context.Context, *GetActiveUserReq) (*GetActiveUserResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetActiveUser not implemented")
}
func (UnimplementedUserServer) GetActiveGroup(context.Context, *GetActiveGroupReq) (*GetActiveGroupResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetActiveGroup not implemented")
}
func (UnimplementedUserServer) GetMessageStatistics(context.Context, *GetMessageStatisticsReq) (*GetMessageStatisticsResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMessageStatistics not implemented")
}
func (UnimplementedUserServer) GetGroupStatistics(context.Context, *GetGroupStatisticsReq) (*GetGroupStatisticsResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGroupStatistics not implemented")
}
func (UnimplementedUserServer) GetUserStatistics(context.Context, *GetUserStatisticsReq) (*GetUserStatisticsResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserStatistics not implemented")
}
func (UnimplementedUserServer) GetGameStatistics(context.Context, *GetGameStatisticsReq) (*GetGameStatisticsResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGameStatistics not implemented")
}

// UnsafeUserServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserServer will
// result in compilation errors.
type UnsafeUserServer interface {
	mustEmbedUnimplementedUserServer()
}

func RegisterUserServer(s grpc.ServiceRegistrar, srv UserServer) {
	s.RegisterService(&User_ServiceDesc, srv)
}

func _User_GetActiveUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetActiveUserReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).GetActiveUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/statistics.user/GetActiveUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).GetActiveUser(ctx, req.(*GetActiveUserReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_GetActiveGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetActiveGroupReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).GetActiveGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/statistics.user/GetActiveGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).GetActiveGroup(ctx, req.(*GetActiveGroupReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_GetMessageStatistics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMessageStatisticsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).GetMessageStatistics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/statistics.user/GetMessageStatistics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).GetMessageStatistics(ctx, req.(*GetMessageStatisticsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_GetGroupStatistics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGroupStatisticsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).GetGroupStatistics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/statistics.user/GetGroupStatistics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).GetGroupStatistics(ctx, req.(*GetGroupStatisticsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_GetUserStatistics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserStatisticsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).GetUserStatistics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/statistics.user/GetUserStatistics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).GetUserStatistics(ctx, req.(*GetUserStatisticsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_GetGameStatistics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGameStatisticsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).GetGameStatistics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/statistics.user/GetGameStatistics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).GetGameStatistics(ctx, req.(*GetGameStatisticsReq))
	}
	return interceptor(ctx, in, info, handler)
}

// User_ServiceDesc is the grpc.ServiceDesc for User service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var User_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "statistics.user",
	HandlerType: (*UserServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetActiveUser",
			Handler:    _User_GetActiveUser_Handler,
		},
		{
			MethodName: "GetActiveGroup",
			Handler:    _User_GetActiveGroup_Handler,
		},
		{
			MethodName: "GetMessageStatistics",
			Handler:    _User_GetMessageStatistics_Handler,
		},
		{
			MethodName: "GetGroupStatistics",
			Handler:    _User_GetGroupStatistics_Handler,
		},
		{
			MethodName: "GetUserStatistics",
			Handler:    _User_GetUserStatistics_Handler,
		},
		{
			MethodName: "GetGameStatistics",
			Handler:    _User_GetGameStatistics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "statistics/statistics.proto",
}
