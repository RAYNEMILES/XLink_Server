// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.5
// source: friend/friend.proto

package friend

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

// FriendClient is the client API for Friend service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FriendClient interface {
	// rpc getFriendsInfo(GetFriendsInfoReq) returns(GetFriendInfoResp);
	AddFriend(ctx context.Context, in *AddFriendReq, opts ...grpc.CallOption) (*AddFriendResp, error)
	GetFriendApplyList(ctx context.Context, in *GetFriendApplyListReq, opts ...grpc.CallOption) (*GetFriendApplyListResp, error)
	GetSelfApplyList(ctx context.Context, in *GetSelfApplyListReq, opts ...grpc.CallOption) (*GetSelfApplyListResp, error)
	GetFriendList(ctx context.Context, in *GetFriendListReq, opts ...grpc.CallOption) (*GetFriendListResp, error)
	GetFriendsInfo(ctx context.Context, in *GetFriendsInfoReq, opts ...grpc.CallOption) (*GetFriendsInfoResp, error)
	AddBlacklist(ctx context.Context, in *AddBlacklistReq, opts ...grpc.CallOption) (*AddBlacklistResp, error)
	RemoveBlacklist(ctx context.Context, in *RemoveBlacklistReq, opts ...grpc.CallOption) (*RemoveBlacklistResp, error)
	IsFriend(ctx context.Context, in *IsFriendReq, opts ...grpc.CallOption) (*IsFriendResp, error)
	IsInBlackList(ctx context.Context, in *IsInBlackListReq, opts ...grpc.CallOption) (*IsInBlackListResp, error)
	GetBlacklist(ctx context.Context, in *GetBlacklistReq, opts ...grpc.CallOption) (*GetBlacklistResp, error)
	DeleteFriend(ctx context.Context, in *DeleteFriendReq, opts ...grpc.CallOption) (*DeleteFriendResp, error)
	AddFriendResponse(ctx context.Context, in *AddFriendResponseReq, opts ...grpc.CallOption) (*AddFriendResponseResp, error)
	SetFriendRemark(ctx context.Context, in *SetFriendRemarkReq, opts ...grpc.CallOption) (*SetFriendRemarkResp, error)
	GetFriendRemarkOrNick(ctx context.Context, in *GetFriendRemarkOrNickReq, opts ...grpc.CallOption) (*GetFriendRemarkOrNickResp, error)
	ImportFriend(ctx context.Context, in *ImportFriendReq, opts ...grpc.CallOption) (*ImportFriendResp, error)
	AddBlackFriends(ctx context.Context, in *AddBlackFriendsReq, opts ...grpc.CallOption) (*AddBlackFriendsResp, error)
	GetBlackFriends(ctx context.Context, in *GetBlackFriendsReq, opts ...grpc.CallOption) (*GetBlackFriendsResp, error)
	RemoveBlackFriends(ctx context.Context, in *RemoveBlackFriendsReq, opts ...grpc.CallOption) (*RemoveBlackFriendsResp, error)
	CheckFriendFromCache(ctx context.Context, in *IsFriendReq, opts ...grpc.CallOption) (*IsFriendResp, error)
	CheckBlockFromCache(ctx context.Context, in *IsInBlackListReq, opts ...grpc.CallOption) (*IsFriendResp, error)
	AutoAddFriend(ctx context.Context, in *AutoAddFriendRequset, opts ...grpc.CallOption) (*AutoAddFriendResponse, error)
	ChannelAddFriend(ctx context.Context, in *ChannelAddFriendRequset, opts ...grpc.CallOption) (*ChannelAddFriendResponse, error)
	GetBlacks(ctx context.Context, in *GetBlacksReq, opts ...grpc.CallOption) (*GetBlacksResp, error)
	RemoveBlack(ctx context.Context, in *RemoveBlackReq, opts ...grpc.CallOption) (*RemoveBlackResp, error)
	AlterRemark(ctx context.Context, in *AlterRemarkReq, opts ...grpc.CallOption) (*AlterRemarkResp, error)
}

type friendClient struct {
	cc grpc.ClientConnInterface
}

func NewFriendClient(cc grpc.ClientConnInterface) FriendClient {
	return &friendClient{cc}
}

func (c *friendClient) AddFriend(ctx context.Context, in *AddFriendReq, opts ...grpc.CallOption) (*AddFriendResp, error) {
	out := new(AddFriendResp)
	err := c.cc.Invoke(ctx, "/friend.friend/addFriend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) GetFriendApplyList(ctx context.Context, in *GetFriendApplyListReq, opts ...grpc.CallOption) (*GetFriendApplyListResp, error) {
	out := new(GetFriendApplyListResp)
	err := c.cc.Invoke(ctx, "/friend.friend/getFriendApplyList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) GetSelfApplyList(ctx context.Context, in *GetSelfApplyListReq, opts ...grpc.CallOption) (*GetSelfApplyListResp, error) {
	out := new(GetSelfApplyListResp)
	err := c.cc.Invoke(ctx, "/friend.friend/getSelfApplyList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) GetFriendList(ctx context.Context, in *GetFriendListReq, opts ...grpc.CallOption) (*GetFriendListResp, error) {
	out := new(GetFriendListResp)
	err := c.cc.Invoke(ctx, "/friend.friend/getFriendList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) GetFriendsInfo(ctx context.Context, in *GetFriendsInfoReq, opts ...grpc.CallOption) (*GetFriendsInfoResp, error) {
	out := new(GetFriendsInfoResp)
	err := c.cc.Invoke(ctx, "/friend.friend/getFriendsInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) AddBlacklist(ctx context.Context, in *AddBlacklistReq, opts ...grpc.CallOption) (*AddBlacklistResp, error) {
	out := new(AddBlacklistResp)
	err := c.cc.Invoke(ctx, "/friend.friend/addBlacklist", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) RemoveBlacklist(ctx context.Context, in *RemoveBlacklistReq, opts ...grpc.CallOption) (*RemoveBlacklistResp, error) {
	out := new(RemoveBlacklistResp)
	err := c.cc.Invoke(ctx, "/friend.friend/removeBlacklist", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) IsFriend(ctx context.Context, in *IsFriendReq, opts ...grpc.CallOption) (*IsFriendResp, error) {
	out := new(IsFriendResp)
	err := c.cc.Invoke(ctx, "/friend.friend/isFriend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) IsInBlackList(ctx context.Context, in *IsInBlackListReq, opts ...grpc.CallOption) (*IsInBlackListResp, error) {
	out := new(IsInBlackListResp)
	err := c.cc.Invoke(ctx, "/friend.friend/isInBlackList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) GetBlacklist(ctx context.Context, in *GetBlacklistReq, opts ...grpc.CallOption) (*GetBlacklistResp, error) {
	out := new(GetBlacklistResp)
	err := c.cc.Invoke(ctx, "/friend.friend/getBlacklist", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) DeleteFriend(ctx context.Context, in *DeleteFriendReq, opts ...grpc.CallOption) (*DeleteFriendResp, error) {
	out := new(DeleteFriendResp)
	err := c.cc.Invoke(ctx, "/friend.friend/deleteFriend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) AddFriendResponse(ctx context.Context, in *AddFriendResponseReq, opts ...grpc.CallOption) (*AddFriendResponseResp, error) {
	out := new(AddFriendResponseResp)
	err := c.cc.Invoke(ctx, "/friend.friend/addFriendResponse", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) SetFriendRemark(ctx context.Context, in *SetFriendRemarkReq, opts ...grpc.CallOption) (*SetFriendRemarkResp, error) {
	out := new(SetFriendRemarkResp)
	err := c.cc.Invoke(ctx, "/friend.friend/setFriendRemark", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) GetFriendRemarkOrNick(ctx context.Context, in *GetFriendRemarkOrNickReq, opts ...grpc.CallOption) (*GetFriendRemarkOrNickResp, error) {
	out := new(GetFriendRemarkOrNickResp)
	err := c.cc.Invoke(ctx, "/friend.friend/getFriendRemarkOrNick", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) ImportFriend(ctx context.Context, in *ImportFriendReq, opts ...grpc.CallOption) (*ImportFriendResp, error) {
	out := new(ImportFriendResp)
	err := c.cc.Invoke(ctx, "/friend.friend/importFriend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) AddBlackFriends(ctx context.Context, in *AddBlackFriendsReq, opts ...grpc.CallOption) (*AddBlackFriendsResp, error) {
	out := new(AddBlackFriendsResp)
	err := c.cc.Invoke(ctx, "/friend.friend/addBlackFriends", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) GetBlackFriends(ctx context.Context, in *GetBlackFriendsReq, opts ...grpc.CallOption) (*GetBlackFriendsResp, error) {
	out := new(GetBlackFriendsResp)
	err := c.cc.Invoke(ctx, "/friend.friend/getBlackFriends", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) RemoveBlackFriends(ctx context.Context, in *RemoveBlackFriendsReq, opts ...grpc.CallOption) (*RemoveBlackFriendsResp, error) {
	out := new(RemoveBlackFriendsResp)
	err := c.cc.Invoke(ctx, "/friend.friend/removeBlackFriends", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) CheckFriendFromCache(ctx context.Context, in *IsFriendReq, opts ...grpc.CallOption) (*IsFriendResp, error) {
	out := new(IsFriendResp)
	err := c.cc.Invoke(ctx, "/friend.friend/CheckFriendFromCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) CheckBlockFromCache(ctx context.Context, in *IsInBlackListReq, opts ...grpc.CallOption) (*IsFriendResp, error) {
	out := new(IsFriendResp)
	err := c.cc.Invoke(ctx, "/friend.friend/CheckBlockFromCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) AutoAddFriend(ctx context.Context, in *AutoAddFriendRequset, opts ...grpc.CallOption) (*AutoAddFriendResponse, error) {
	out := new(AutoAddFriendResponse)
	err := c.cc.Invoke(ctx, "/friend.friend/AutoAddFriend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) ChannelAddFriend(ctx context.Context, in *ChannelAddFriendRequset, opts ...grpc.CallOption) (*ChannelAddFriendResponse, error) {
	out := new(ChannelAddFriendResponse)
	err := c.cc.Invoke(ctx, "/friend.friend/ChannelAddFriend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) GetBlacks(ctx context.Context, in *GetBlacksReq, opts ...grpc.CallOption) (*GetBlacksResp, error) {
	out := new(GetBlacksResp)
	err := c.cc.Invoke(ctx, "/friend.friend/GetBlacks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) RemoveBlack(ctx context.Context, in *RemoveBlackReq, opts ...grpc.CallOption) (*RemoveBlackResp, error) {
	out := new(RemoveBlackResp)
	err := c.cc.Invoke(ctx, "/friend.friend/RemoveBlack", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *friendClient) AlterRemark(ctx context.Context, in *AlterRemarkReq, opts ...grpc.CallOption) (*AlterRemarkResp, error) {
	out := new(AlterRemarkResp)
	err := c.cc.Invoke(ctx, "/friend.friend/AlterRemark", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FriendServer is the server API for Friend service.
// All implementations should embed UnimplementedFriendServer
// for forward compatibility
type FriendServer interface {
	// rpc getFriendsInfo(GetFriendsInfoReq) returns(GetFriendInfoResp);
	AddFriend(context.Context, *AddFriendReq) (*AddFriendResp, error)
	GetFriendApplyList(context.Context, *GetFriendApplyListReq) (*GetFriendApplyListResp, error)
	GetSelfApplyList(context.Context, *GetSelfApplyListReq) (*GetSelfApplyListResp, error)
	GetFriendList(context.Context, *GetFriendListReq) (*GetFriendListResp, error)
	GetFriendsInfo(context.Context, *GetFriendsInfoReq) (*GetFriendsInfoResp, error)
	AddBlacklist(context.Context, *AddBlacklistReq) (*AddBlacklistResp, error)
	RemoveBlacklist(context.Context, *RemoveBlacklistReq) (*RemoveBlacklistResp, error)
	IsFriend(context.Context, *IsFriendReq) (*IsFriendResp, error)
	IsInBlackList(context.Context, *IsInBlackListReq) (*IsInBlackListResp, error)
	GetBlacklist(context.Context, *GetBlacklistReq) (*GetBlacklistResp, error)
	DeleteFriend(context.Context, *DeleteFriendReq) (*DeleteFriendResp, error)
	AddFriendResponse(context.Context, *AddFriendResponseReq) (*AddFriendResponseResp, error)
	SetFriendRemark(context.Context, *SetFriendRemarkReq) (*SetFriendRemarkResp, error)
	GetFriendRemarkOrNick(context.Context, *GetFriendRemarkOrNickReq) (*GetFriendRemarkOrNickResp, error)
	ImportFriend(context.Context, *ImportFriendReq) (*ImportFriendResp, error)
	AddBlackFriends(context.Context, *AddBlackFriendsReq) (*AddBlackFriendsResp, error)
	GetBlackFriends(context.Context, *GetBlackFriendsReq) (*GetBlackFriendsResp, error)
	RemoveBlackFriends(context.Context, *RemoveBlackFriendsReq) (*RemoveBlackFriendsResp, error)
	CheckFriendFromCache(context.Context, *IsFriendReq) (*IsFriendResp, error)
	CheckBlockFromCache(context.Context, *IsInBlackListReq) (*IsFriendResp, error)
	AutoAddFriend(context.Context, *AutoAddFriendRequset) (*AutoAddFriendResponse, error)
	ChannelAddFriend(context.Context, *ChannelAddFriendRequset) (*ChannelAddFriendResponse, error)
	GetBlacks(context.Context, *GetBlacksReq) (*GetBlacksResp, error)
	RemoveBlack(context.Context, *RemoveBlackReq) (*RemoveBlackResp, error)
	AlterRemark(context.Context, *AlterRemarkReq) (*AlterRemarkResp, error)
}

// UnimplementedFriendServer should be embedded to have forward compatible implementations.
type UnimplementedFriendServer struct {
}

func (UnimplementedFriendServer) AddFriend(context.Context, *AddFriendReq) (*AddFriendResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddFriend not implemented")
}
func (UnimplementedFriendServer) GetFriendApplyList(context.Context, *GetFriendApplyListReq) (*GetFriendApplyListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFriendApplyList not implemented")
}
func (UnimplementedFriendServer) GetSelfApplyList(context.Context, *GetSelfApplyListReq) (*GetSelfApplyListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSelfApplyList not implemented")
}
func (UnimplementedFriendServer) GetFriendList(context.Context, *GetFriendListReq) (*GetFriendListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFriendList not implemented")
}
func (UnimplementedFriendServer) GetFriendsInfo(context.Context, *GetFriendsInfoReq) (*GetFriendsInfoResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFriendsInfo not implemented")
}
func (UnimplementedFriendServer) AddBlacklist(context.Context, *AddBlacklistReq) (*AddBlacklistResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddBlacklist not implemented")
}
func (UnimplementedFriendServer) RemoveBlacklist(context.Context, *RemoveBlacklistReq) (*RemoveBlacklistResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveBlacklist not implemented")
}
func (UnimplementedFriendServer) IsFriend(context.Context, *IsFriendReq) (*IsFriendResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsFriend not implemented")
}
func (UnimplementedFriendServer) IsInBlackList(context.Context, *IsInBlackListReq) (*IsInBlackListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsInBlackList not implemented")
}
func (UnimplementedFriendServer) GetBlacklist(context.Context, *GetBlacklistReq) (*GetBlacklistResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlacklist not implemented")
}
func (UnimplementedFriendServer) DeleteFriend(context.Context, *DeleteFriendReq) (*DeleteFriendResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteFriend not implemented")
}
func (UnimplementedFriendServer) AddFriendResponse(context.Context, *AddFriendResponseReq) (*AddFriendResponseResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddFriendResponse not implemented")
}
func (UnimplementedFriendServer) SetFriendRemark(context.Context, *SetFriendRemarkReq) (*SetFriendRemarkResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetFriendRemark not implemented")
}
func (UnimplementedFriendServer) GetFriendRemarkOrNick(context.Context, *GetFriendRemarkOrNickReq) (*GetFriendRemarkOrNickResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFriendRemarkOrNick not implemented")
}
func (UnimplementedFriendServer) ImportFriend(context.Context, *ImportFriendReq) (*ImportFriendResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ImportFriend not implemented")
}
func (UnimplementedFriendServer) AddBlackFriends(context.Context, *AddBlackFriendsReq) (*AddBlackFriendsResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddBlackFriends not implemented")
}
func (UnimplementedFriendServer) GetBlackFriends(context.Context, *GetBlackFriendsReq) (*GetBlackFriendsResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlackFriends not implemented")
}
func (UnimplementedFriendServer) RemoveBlackFriends(context.Context, *RemoveBlackFriendsReq) (*RemoveBlackFriendsResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveBlackFriends not implemented")
}
func (UnimplementedFriendServer) CheckFriendFromCache(context.Context, *IsFriendReq) (*IsFriendResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckFriendFromCache not implemented")
}
func (UnimplementedFriendServer) CheckBlockFromCache(context.Context, *IsInBlackListReq) (*IsFriendResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckBlockFromCache not implemented")
}
func (UnimplementedFriendServer) AutoAddFriend(context.Context, *AutoAddFriendRequset) (*AutoAddFriendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AutoAddFriend not implemented")
}
func (UnimplementedFriendServer) ChannelAddFriend(context.Context, *ChannelAddFriendRequset) (*ChannelAddFriendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChannelAddFriend not implemented")
}
func (UnimplementedFriendServer) GetBlacks(context.Context, *GetBlacksReq) (*GetBlacksResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlacks not implemented")
}
func (UnimplementedFriendServer) RemoveBlack(context.Context, *RemoveBlackReq) (*RemoveBlackResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveBlack not implemented")
}
func (UnimplementedFriendServer) AlterRemark(context.Context, *AlterRemarkReq) (*AlterRemarkResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AlterRemark not implemented")
}

// UnsafeFriendServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FriendServer will
// result in compilation errors.
type UnsafeFriendServer interface {
	mustEmbedUnimplementedFriendServer()
}

func RegisterFriendServer(s grpc.ServiceRegistrar, srv FriendServer) {
	s.RegisterService(&Friend_ServiceDesc, srv)
}

func _Friend_AddFriend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddFriendReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).AddFriend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/addFriend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).AddFriend(ctx, req.(*AddFriendReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_GetFriendApplyList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFriendApplyListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).GetFriendApplyList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/getFriendApplyList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).GetFriendApplyList(ctx, req.(*GetFriendApplyListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_GetSelfApplyList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSelfApplyListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).GetSelfApplyList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/getSelfApplyList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).GetSelfApplyList(ctx, req.(*GetSelfApplyListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_GetFriendList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFriendListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).GetFriendList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/getFriendList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).GetFriendList(ctx, req.(*GetFriendListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_GetFriendsInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFriendsInfoReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).GetFriendsInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/getFriendsInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).GetFriendsInfo(ctx, req.(*GetFriendsInfoReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_AddBlacklist_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddBlacklistReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).AddBlacklist(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/addBlacklist",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).AddBlacklist(ctx, req.(*AddBlacklistReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_RemoveBlacklist_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveBlacklistReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).RemoveBlacklist(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/removeBlacklist",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).RemoveBlacklist(ctx, req.(*RemoveBlacklistReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_IsFriend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsFriendReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).IsFriend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/isFriend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).IsFriend(ctx, req.(*IsFriendReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_IsInBlackList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsInBlackListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).IsInBlackList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/isInBlackList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).IsInBlackList(ctx, req.(*IsInBlackListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_GetBlacklist_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlacklistReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).GetBlacklist(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/getBlacklist",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).GetBlacklist(ctx, req.(*GetBlacklistReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_DeleteFriend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteFriendReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).DeleteFriend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/deleteFriend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).DeleteFriend(ctx, req.(*DeleteFriendReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_AddFriendResponse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddFriendResponseReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).AddFriendResponse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/addFriendResponse",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).AddFriendResponse(ctx, req.(*AddFriendResponseReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_SetFriendRemark_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetFriendRemarkReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).SetFriendRemark(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/setFriendRemark",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).SetFriendRemark(ctx, req.(*SetFriendRemarkReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_GetFriendRemarkOrNick_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFriendRemarkOrNickReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).GetFriendRemarkOrNick(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/getFriendRemarkOrNick",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).GetFriendRemarkOrNick(ctx, req.(*GetFriendRemarkOrNickReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_ImportFriend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ImportFriendReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).ImportFriend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/importFriend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).ImportFriend(ctx, req.(*ImportFriendReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_AddBlackFriends_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddBlackFriendsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).AddBlackFriends(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/addBlackFriends",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).AddBlackFriends(ctx, req.(*AddBlackFriendsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_GetBlackFriends_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlackFriendsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).GetBlackFriends(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/getBlackFriends",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).GetBlackFriends(ctx, req.(*GetBlackFriendsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_RemoveBlackFriends_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveBlackFriendsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).RemoveBlackFriends(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/removeBlackFriends",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).RemoveBlackFriends(ctx, req.(*RemoveBlackFriendsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_CheckFriendFromCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsFriendReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).CheckFriendFromCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/CheckFriendFromCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).CheckFriendFromCache(ctx, req.(*IsFriendReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_CheckBlockFromCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsInBlackListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).CheckBlockFromCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/CheckBlockFromCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).CheckBlockFromCache(ctx, req.(*IsInBlackListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_AutoAddFriend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AutoAddFriendRequset)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).AutoAddFriend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/AutoAddFriend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).AutoAddFriend(ctx, req.(*AutoAddFriendRequset))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_ChannelAddFriend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChannelAddFriendRequset)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).ChannelAddFriend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/ChannelAddFriend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).ChannelAddFriend(ctx, req.(*ChannelAddFriendRequset))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_GetBlacks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlacksReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).GetBlacks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/GetBlacks",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).GetBlacks(ctx, req.(*GetBlacksReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_RemoveBlack_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveBlackReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).RemoveBlack(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/RemoveBlack",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).RemoveBlack(ctx, req.(*RemoveBlackReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Friend_AlterRemark_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AlterRemarkReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FriendServer).AlterRemark(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/friend.friend/AlterRemark",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FriendServer).AlterRemark(ctx, req.(*AlterRemarkReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Friend_ServiceDesc is the grpc.ServiceDesc for Friend service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Friend_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "friend.friend",
	HandlerType: (*FriendServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "addFriend",
			Handler:    _Friend_AddFriend_Handler,
		},
		{
			MethodName: "getFriendApplyList",
			Handler:    _Friend_GetFriendApplyList_Handler,
		},
		{
			MethodName: "getSelfApplyList",
			Handler:    _Friend_GetSelfApplyList_Handler,
		},
		{
			MethodName: "getFriendList",
			Handler:    _Friend_GetFriendList_Handler,
		},
		{
			MethodName: "getFriendsInfo",
			Handler:    _Friend_GetFriendsInfo_Handler,
		},
		{
			MethodName: "addBlacklist",
			Handler:    _Friend_AddBlacklist_Handler,
		},
		{
			MethodName: "removeBlacklist",
			Handler:    _Friend_RemoveBlacklist_Handler,
		},
		{
			MethodName: "isFriend",
			Handler:    _Friend_IsFriend_Handler,
		},
		{
			MethodName: "isInBlackList",
			Handler:    _Friend_IsInBlackList_Handler,
		},
		{
			MethodName: "getBlacklist",
			Handler:    _Friend_GetBlacklist_Handler,
		},
		{
			MethodName: "deleteFriend",
			Handler:    _Friend_DeleteFriend_Handler,
		},
		{
			MethodName: "addFriendResponse",
			Handler:    _Friend_AddFriendResponse_Handler,
		},
		{
			MethodName: "setFriendRemark",
			Handler:    _Friend_SetFriendRemark_Handler,
		},
		{
			MethodName: "getFriendRemarkOrNick",
			Handler:    _Friend_GetFriendRemarkOrNick_Handler,
		},
		{
			MethodName: "importFriend",
			Handler:    _Friend_ImportFriend_Handler,
		},
		{
			MethodName: "addBlackFriends",
			Handler:    _Friend_AddBlackFriends_Handler,
		},
		{
			MethodName: "getBlackFriends",
			Handler:    _Friend_GetBlackFriends_Handler,
		},
		{
			MethodName: "removeBlackFriends",
			Handler:    _Friend_RemoveBlackFriends_Handler,
		},
		{
			MethodName: "CheckFriendFromCache",
			Handler:    _Friend_CheckFriendFromCache_Handler,
		},
		{
			MethodName: "CheckBlockFromCache",
			Handler:    _Friend_CheckBlockFromCache_Handler,
		},
		{
			MethodName: "AutoAddFriend",
			Handler:    _Friend_AutoAddFriend_Handler,
		},
		{
			MethodName: "ChannelAddFriend",
			Handler:    _Friend_ChannelAddFriend_Handler,
		},
		{
			MethodName: "GetBlacks",
			Handler:    _Friend_GetBlacks_Handler,
		},
		{
			MethodName: "RemoveBlack",
			Handler:    _Friend_RemoveBlack_Handler,
		},
		{
			MethodName: "AlterRemark",
			Handler:    _Friend_AlterRemark_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "friend/friend.proto",
}
