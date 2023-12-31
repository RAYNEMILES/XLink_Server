// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.5
// source: message_cms/message_cms.proto

package message_cms

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

// MessageCMSClient is the client API for MessageCMS service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MessageCMSClient interface {
	BoradcastMessage(ctx context.Context, in *BoradcastMessageReq, opts ...grpc.CallOption) (*BoradcastMessageResp, error)
	MassSendMessage(ctx context.Context, in *MassSendMessageReq, opts ...grpc.CallOption) (*MassSendMessageResp, error)
	GetChatLogs(ctx context.Context, in *GetChatLogsReq, opts ...grpc.CallOption) (*GetChatLogsResp, error)
	GetChatLogsV1(ctx context.Context, in *GetChatLogsV1Req, opts ...grpc.CallOption) (*GetChatLogsResp, error)
	WithdrawMessage(ctx context.Context, in *WithdrawMessageReq, opts ...grpc.CallOption) (*WithdrawMessageResp, error)
}

type messageCMSClient struct {
	cc grpc.ClientConnInterface
}

func NewMessageCMSClient(cc grpc.ClientConnInterface) MessageCMSClient {
	return &messageCMSClient{cc}
}

func (c *messageCMSClient) BoradcastMessage(ctx context.Context, in *BoradcastMessageReq, opts ...grpc.CallOption) (*BoradcastMessageResp, error) {
	out := new(BoradcastMessageResp)
	err := c.cc.Invoke(ctx, "/message_cms.messageCMS/BoradcastMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageCMSClient) MassSendMessage(ctx context.Context, in *MassSendMessageReq, opts ...grpc.CallOption) (*MassSendMessageResp, error) {
	out := new(MassSendMessageResp)
	err := c.cc.Invoke(ctx, "/message_cms.messageCMS/MassSendMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageCMSClient) GetChatLogs(ctx context.Context, in *GetChatLogsReq, opts ...grpc.CallOption) (*GetChatLogsResp, error) {
	out := new(GetChatLogsResp)
	err := c.cc.Invoke(ctx, "/message_cms.messageCMS/GetChatLogs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageCMSClient) GetChatLogsV1(ctx context.Context, in *GetChatLogsV1Req, opts ...grpc.CallOption) (*GetChatLogsResp, error) {
	out := new(GetChatLogsResp)
	err := c.cc.Invoke(ctx, "/message_cms.messageCMS/GetChatLogsV1", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageCMSClient) WithdrawMessage(ctx context.Context, in *WithdrawMessageReq, opts ...grpc.CallOption) (*WithdrawMessageResp, error) {
	out := new(WithdrawMessageResp)
	err := c.cc.Invoke(ctx, "/message_cms.messageCMS/WithdrawMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MessageCMSServer is the server API for MessageCMS service.
// All implementations should embed UnimplementedMessageCMSServer
// for forward compatibility
type MessageCMSServer interface {
	BoradcastMessage(context.Context, *BoradcastMessageReq) (*BoradcastMessageResp, error)
	MassSendMessage(context.Context, *MassSendMessageReq) (*MassSendMessageResp, error)
	GetChatLogs(context.Context, *GetChatLogsReq) (*GetChatLogsResp, error)
	GetChatLogsV1(context.Context, *GetChatLogsV1Req) (*GetChatLogsResp, error)
	WithdrawMessage(context.Context, *WithdrawMessageReq) (*WithdrawMessageResp, error)
}

// UnimplementedMessageCMSServer should be embedded to have forward compatible implementations.
type UnimplementedMessageCMSServer struct {
}

func (UnimplementedMessageCMSServer) BoradcastMessage(context.Context, *BoradcastMessageReq) (*BoradcastMessageResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BoradcastMessage not implemented")
}
func (UnimplementedMessageCMSServer) MassSendMessage(context.Context, *MassSendMessageReq) (*MassSendMessageResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MassSendMessage not implemented")
}
func (UnimplementedMessageCMSServer) GetChatLogs(context.Context, *GetChatLogsReq) (*GetChatLogsResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetChatLogs not implemented")
}
func (UnimplementedMessageCMSServer) GetChatLogsV1(context.Context, *GetChatLogsV1Req) (*GetChatLogsResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetChatLogsV1 not implemented")
}
func (UnimplementedMessageCMSServer) WithdrawMessage(context.Context, *WithdrawMessageReq) (*WithdrawMessageResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WithdrawMessage not implemented")
}

// UnsafeMessageCMSServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MessageCMSServer will
// result in compilation errors.
type UnsafeMessageCMSServer interface {
	mustEmbedUnimplementedMessageCMSServer()
}

func RegisterMessageCMSServer(s grpc.ServiceRegistrar, srv MessageCMSServer) {
	s.RegisterService(&MessageCMS_ServiceDesc, srv)
}

func _MessageCMS_BoradcastMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BoradcastMessageReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageCMSServer).BoradcastMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/message_cms.messageCMS/BoradcastMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageCMSServer).BoradcastMessage(ctx, req.(*BoradcastMessageReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageCMS_MassSendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MassSendMessageReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageCMSServer).MassSendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/message_cms.messageCMS/MassSendMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageCMSServer).MassSendMessage(ctx, req.(*MassSendMessageReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageCMS_GetChatLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetChatLogsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageCMSServer).GetChatLogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/message_cms.messageCMS/GetChatLogs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageCMSServer).GetChatLogs(ctx, req.(*GetChatLogsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageCMS_GetChatLogsV1_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetChatLogsV1Req)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageCMSServer).GetChatLogsV1(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/message_cms.messageCMS/GetChatLogsV1",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageCMSServer).GetChatLogsV1(ctx, req.(*GetChatLogsV1Req))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageCMS_WithdrawMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WithdrawMessageReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageCMSServer).WithdrawMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/message_cms.messageCMS/WithdrawMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageCMSServer).WithdrawMessage(ctx, req.(*WithdrawMessageReq))
	}
	return interceptor(ctx, in, info, handler)
}

// MessageCMS_ServiceDesc is the grpc.ServiceDesc for MessageCMS service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MessageCMS_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "message_cms.messageCMS",
	HandlerType: (*MessageCMSServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BoradcastMessage",
			Handler:    _MessageCMS_BoradcastMessage_Handler,
		},
		{
			MethodName: "MassSendMessage",
			Handler:    _MessageCMS_MassSendMessage_Handler,
		},
		{
			MethodName: "GetChatLogs",
			Handler:    _MessageCMS_GetChatLogs_Handler,
		},
		{
			MethodName: "GetChatLogsV1",
			Handler:    _MessageCMS_GetChatLogsV1_Handler,
		},
		{
			MethodName: "WithdrawMessage",
			Handler:    _MessageCMS_WithdrawMessage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "message_cms/message_cms.proto",
}
