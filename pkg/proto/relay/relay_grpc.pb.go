// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.5
// source: relay/relay.proto

package pbRelay

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

// OnlineMessageRelayServiceClient is the client API for OnlineMessageRelayService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OnlineMessageRelayServiceClient interface {
	OnlinePushMsg(ctx context.Context, in *OnlinePushMsgReq, opts ...grpc.CallOption) (*OnlinePushMsgResp, error)
	GetUsersOnlineStatus(ctx context.Context, in *GetUsersOnlineStatusReq, opts ...grpc.CallOption) (*GetUsersOnlineStatusResp, error)
	OnlineBatchPushOneMsg(ctx context.Context, in *OnlineBatchPushOneMsgReq, opts ...grpc.CallOption) (*OnlineBatchPushOneMsgResp, error)
	SuperGroupOnlineBatchPushOneMsg(ctx context.Context, in *OnlineBatchPushOneMsgReq, opts ...grpc.CallOption) (*OnlineBatchPushOneMsgResp, error)
	KickUserOffline(ctx context.Context, in *KickUserOfflineReq, opts ...grpc.CallOption) (*KickUserOfflineResp, error)
}

type onlineMessageRelayServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOnlineMessageRelayServiceClient(cc grpc.ClientConnInterface) OnlineMessageRelayServiceClient {
	return &onlineMessageRelayServiceClient{cc}
}

func (c *onlineMessageRelayServiceClient) OnlinePushMsg(ctx context.Context, in *OnlinePushMsgReq, opts ...grpc.CallOption) (*OnlinePushMsgResp, error) {
	out := new(OnlinePushMsgResp)
	err := c.cc.Invoke(ctx, "/relay.OnlineMessageRelayService/OnlinePushMsg", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *onlineMessageRelayServiceClient) GetUsersOnlineStatus(ctx context.Context, in *GetUsersOnlineStatusReq, opts ...grpc.CallOption) (*GetUsersOnlineStatusResp, error) {
	out := new(GetUsersOnlineStatusResp)
	err := c.cc.Invoke(ctx, "/relay.OnlineMessageRelayService/GetUsersOnlineStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *onlineMessageRelayServiceClient) OnlineBatchPushOneMsg(ctx context.Context, in *OnlineBatchPushOneMsgReq, opts ...grpc.CallOption) (*OnlineBatchPushOneMsgResp, error) {
	out := new(OnlineBatchPushOneMsgResp)
	err := c.cc.Invoke(ctx, "/relay.OnlineMessageRelayService/OnlineBatchPushOneMsg", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *onlineMessageRelayServiceClient) SuperGroupOnlineBatchPushOneMsg(ctx context.Context, in *OnlineBatchPushOneMsgReq, opts ...grpc.CallOption) (*OnlineBatchPushOneMsgResp, error) {
	out := new(OnlineBatchPushOneMsgResp)
	err := c.cc.Invoke(ctx, "/relay.OnlineMessageRelayService/SuperGroupOnlineBatchPushOneMsg", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *onlineMessageRelayServiceClient) KickUserOffline(ctx context.Context, in *KickUserOfflineReq, opts ...grpc.CallOption) (*KickUserOfflineResp, error) {
	out := new(KickUserOfflineResp)
	err := c.cc.Invoke(ctx, "/relay.OnlineMessageRelayService/KickUserOffline", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OnlineMessageRelayServiceServer is the server API for OnlineMessageRelayService service.
// All implementations should embed UnimplementedOnlineMessageRelayServiceServer
// for forward compatibility
type OnlineMessageRelayServiceServer interface {
	OnlinePushMsg(context.Context, *OnlinePushMsgReq) (*OnlinePushMsgResp, error)
	GetUsersOnlineStatus(context.Context, *GetUsersOnlineStatusReq) (*GetUsersOnlineStatusResp, error)
	OnlineBatchPushOneMsg(context.Context, *OnlineBatchPushOneMsgReq) (*OnlineBatchPushOneMsgResp, error)
	SuperGroupOnlineBatchPushOneMsg(context.Context, *OnlineBatchPushOneMsgReq) (*OnlineBatchPushOneMsgResp, error)
	KickUserOffline(context.Context, *KickUserOfflineReq) (*KickUserOfflineResp, error)
}

// UnimplementedOnlineMessageRelayServiceServer should be embedded to have forward compatible implementations.
type UnimplementedOnlineMessageRelayServiceServer struct {
}

func (UnimplementedOnlineMessageRelayServiceServer) OnlinePushMsg(context.Context, *OnlinePushMsgReq) (*OnlinePushMsgResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnlinePushMsg not implemented")
}
func (UnimplementedOnlineMessageRelayServiceServer) GetUsersOnlineStatus(context.Context, *GetUsersOnlineStatusReq) (*GetUsersOnlineStatusResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUsersOnlineStatus not implemented")
}
func (UnimplementedOnlineMessageRelayServiceServer) OnlineBatchPushOneMsg(context.Context, *OnlineBatchPushOneMsgReq) (*OnlineBatchPushOneMsgResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnlineBatchPushOneMsg not implemented")
}
func (UnimplementedOnlineMessageRelayServiceServer) SuperGroupOnlineBatchPushOneMsg(context.Context, *OnlineBatchPushOneMsgReq) (*OnlineBatchPushOneMsgResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SuperGroupOnlineBatchPushOneMsg not implemented")
}
func (UnimplementedOnlineMessageRelayServiceServer) KickUserOffline(context.Context, *KickUserOfflineReq) (*KickUserOfflineResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method KickUserOffline not implemented")
}

// UnsafeOnlineMessageRelayServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OnlineMessageRelayServiceServer will
// result in compilation errors.
type UnsafeOnlineMessageRelayServiceServer interface {
	mustEmbedUnimplementedOnlineMessageRelayServiceServer()
}

func RegisterOnlineMessageRelayServiceServer(s grpc.ServiceRegistrar, srv OnlineMessageRelayServiceServer) {
	s.RegisterService(&OnlineMessageRelayService_ServiceDesc, srv)
}

func _OnlineMessageRelayService_OnlinePushMsg_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnlinePushMsgReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnlineMessageRelayServiceServer).OnlinePushMsg(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/relay.OnlineMessageRelayService/OnlinePushMsg",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnlineMessageRelayServiceServer).OnlinePushMsg(ctx, req.(*OnlinePushMsgReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _OnlineMessageRelayService_GetUsersOnlineStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUsersOnlineStatusReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnlineMessageRelayServiceServer).GetUsersOnlineStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/relay.OnlineMessageRelayService/GetUsersOnlineStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnlineMessageRelayServiceServer).GetUsersOnlineStatus(ctx, req.(*GetUsersOnlineStatusReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _OnlineMessageRelayService_OnlineBatchPushOneMsg_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnlineBatchPushOneMsgReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnlineMessageRelayServiceServer).OnlineBatchPushOneMsg(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/relay.OnlineMessageRelayService/OnlineBatchPushOneMsg",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnlineMessageRelayServiceServer).OnlineBatchPushOneMsg(ctx, req.(*OnlineBatchPushOneMsgReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _OnlineMessageRelayService_SuperGroupOnlineBatchPushOneMsg_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnlineBatchPushOneMsgReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnlineMessageRelayServiceServer).SuperGroupOnlineBatchPushOneMsg(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/relay.OnlineMessageRelayService/SuperGroupOnlineBatchPushOneMsg",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnlineMessageRelayServiceServer).SuperGroupOnlineBatchPushOneMsg(ctx, req.(*OnlineBatchPushOneMsgReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _OnlineMessageRelayService_KickUserOffline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KickUserOfflineReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnlineMessageRelayServiceServer).KickUserOffline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/relay.OnlineMessageRelayService/KickUserOffline",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnlineMessageRelayServiceServer).KickUserOffline(ctx, req.(*KickUserOfflineReq))
	}
	return interceptor(ctx, in, info, handler)
}

// OnlineMessageRelayService_ServiceDesc is the grpc.ServiceDesc for OnlineMessageRelayService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OnlineMessageRelayService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "relay.OnlineMessageRelayService",
	HandlerType: (*OnlineMessageRelayServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "OnlinePushMsg",
			Handler:    _OnlineMessageRelayService_OnlinePushMsg_Handler,
		},
		{
			MethodName: "GetUsersOnlineStatus",
			Handler:    _OnlineMessageRelayService_GetUsersOnlineStatus_Handler,
		},
		{
			MethodName: "OnlineBatchPushOneMsg",
			Handler:    _OnlineMessageRelayService_OnlineBatchPushOneMsg_Handler,
		},
		{
			MethodName: "SuperGroupOnlineBatchPushOneMsg",
			Handler:    _OnlineMessageRelayService_SuperGroupOnlineBatchPushOneMsg_Handler,
		},
		{
			MethodName: "KickUserOffline",
			Handler:    _OnlineMessageRelayService_KickUserOffline_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "relay/relay.proto",
}
