// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: cache/cache.proto

package cache

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

// CacheClient is the client API for Cache service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CacheClient interface {
	// userInfo
	GetUserInfoFromCache(ctx context.Context, in *GetUserInfoFromCacheReq, opts ...grpc.CallOption) (*GetUserInfoFromCacheResp, error)
	UpdateUserInfoToCache(ctx context.Context, in *UpdateUserInfoToCacheReq, opts ...grpc.CallOption) (*UpdateUserInfoToCacheResp, error)
	// friendInfo
	GetFriendIDListFromCache(ctx context.Context, in *GetFriendIDListFromCacheReq, opts ...grpc.CallOption) (*GetFriendIDListFromCacheResp, error)
	AddFriendToCache(ctx context.Context, in *AddFriendToCacheReq, opts ...grpc.CallOption) (*AddFriendToCacheResp, error)
	ReduceFriendFromCache(ctx context.Context, in *ReduceFriendFromCacheReq, opts ...grpc.CallOption) (*ReduceFriendFromCacheResp, error)
	// blackList
	GetBlackIDListFromCache(ctx context.Context, in *GetBlackIDListFromCacheReq, opts ...grpc.CallOption) (*GetBlackIDListFromCacheResp, error)
	AddBlackUserToCache(ctx context.Context, in *AddBlackUserToCacheReq, opts ...grpc.CallOption) (*AddBlackUserToCacheResp, error)
	ReduceBlackUserFromCache(ctx context.Context, in *ReduceBlackUserFromCacheReq, opts ...grpc.CallOption) (*ReduceBlackUserFromCacheResp, error)
	AddBlackFriendsToCache(ctx context.Context, in *AddBlackFriendsToCacheReq, opts ...grpc.CallOption) (*AddBlackFriendsToCacheResp, error)
	ReduceBlackFriendsFromCache(ctx context.Context, in *ReduceBlackFriendsFromCacheReq, opts ...grpc.CallOption) (*ReduceBlackFriendsFromCacheResp, error)
	// group
	GetGroupMemberIDListFromCache(ctx context.Context, in *GetGroupMemberIDListFromCacheReq, opts ...grpc.CallOption) (*GetGroupMemberIDListFromCacheResp, error)
	AddGroupMemberToCache(ctx context.Context, in *AddGroupMemberToCacheReq, opts ...grpc.CallOption) (*AddGroupMemberToCacheResp, error)
	ReduceGroupMemberFromCache(ctx context.Context, in *ReduceGroupMemberFromCacheReq, opts ...grpc.CallOption) (*ReduceGroupMemberFromCacheResp, error)
}

type cacheClient struct {
	cc grpc.ClientConnInterface
}

func NewCacheClient(cc grpc.ClientConnInterface) CacheClient {
	return &cacheClient{cc}
}

func (c *cacheClient) GetUserInfoFromCache(ctx context.Context, in *GetUserInfoFromCacheReq, opts ...grpc.CallOption) (*GetUserInfoFromCacheResp, error) {
	out := new(GetUserInfoFromCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/GetUserInfoFromCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) UpdateUserInfoToCache(ctx context.Context, in *UpdateUserInfoToCacheReq, opts ...grpc.CallOption) (*UpdateUserInfoToCacheResp, error) {
	out := new(UpdateUserInfoToCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/UpdateUserInfoToCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) GetFriendIDListFromCache(ctx context.Context, in *GetFriendIDListFromCacheReq, opts ...grpc.CallOption) (*GetFriendIDListFromCacheResp, error) {
	out := new(GetFriendIDListFromCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/GetFriendIDListFromCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) AddFriendToCache(ctx context.Context, in *AddFriendToCacheReq, opts ...grpc.CallOption) (*AddFriendToCacheResp, error) {
	out := new(AddFriendToCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/AddFriendToCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) ReduceFriendFromCache(ctx context.Context, in *ReduceFriendFromCacheReq, opts ...grpc.CallOption) (*ReduceFriendFromCacheResp, error) {
	out := new(ReduceFriendFromCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/ReduceFriendFromCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) GetBlackIDListFromCache(ctx context.Context, in *GetBlackIDListFromCacheReq, opts ...grpc.CallOption) (*GetBlackIDListFromCacheResp, error) {
	out := new(GetBlackIDListFromCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/GetBlackIDListFromCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) AddBlackUserToCache(ctx context.Context, in *AddBlackUserToCacheReq, opts ...grpc.CallOption) (*AddBlackUserToCacheResp, error) {
	out := new(AddBlackUserToCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/AddBlackUserToCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) ReduceBlackUserFromCache(ctx context.Context, in *ReduceBlackUserFromCacheReq, opts ...grpc.CallOption) (*ReduceBlackUserFromCacheResp, error) {
	out := new(ReduceBlackUserFromCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/ReduceBlackUserFromCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) AddBlackFriendsToCache(ctx context.Context, in *AddBlackFriendsToCacheReq, opts ...grpc.CallOption) (*AddBlackFriendsToCacheResp, error) {
	out := new(AddBlackFriendsToCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/AddBlackFriendsToCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) ReduceBlackFriendsFromCache(ctx context.Context, in *ReduceBlackFriendsFromCacheReq, opts ...grpc.CallOption) (*ReduceBlackFriendsFromCacheResp, error) {
	out := new(ReduceBlackFriendsFromCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/ReduceBlackFriendsFromCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) GetGroupMemberIDListFromCache(ctx context.Context, in *GetGroupMemberIDListFromCacheReq, opts ...grpc.CallOption) (*GetGroupMemberIDListFromCacheResp, error) {
	out := new(GetGroupMemberIDListFromCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/GetGroupMemberIDListFromCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) AddGroupMemberToCache(ctx context.Context, in *AddGroupMemberToCacheReq, opts ...grpc.CallOption) (*AddGroupMemberToCacheResp, error) {
	out := new(AddGroupMemberToCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/AddGroupMemberToCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) ReduceGroupMemberFromCache(ctx context.Context, in *ReduceGroupMemberFromCacheReq, opts ...grpc.CallOption) (*ReduceGroupMemberFromCacheResp, error) {
	out := new(ReduceGroupMemberFromCacheResp)
	err := c.cc.Invoke(ctx, "/cache.cache/ReduceGroupMemberFromCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CacheServer is the server API for Cache service.
// All implementations should embed UnimplementedCacheServer
// for forward compatibility
type CacheServer interface {
	// userInfo
	GetUserInfoFromCache(context.Context, *GetUserInfoFromCacheReq) (*GetUserInfoFromCacheResp, error)
	UpdateUserInfoToCache(context.Context, *UpdateUserInfoToCacheReq) (*UpdateUserInfoToCacheResp, error)
	// friendInfo
	GetFriendIDListFromCache(context.Context, *GetFriendIDListFromCacheReq) (*GetFriendIDListFromCacheResp, error)
	AddFriendToCache(context.Context, *AddFriendToCacheReq) (*AddFriendToCacheResp, error)
	ReduceFriendFromCache(context.Context, *ReduceFriendFromCacheReq) (*ReduceFriendFromCacheResp, error)
	// blackList
	GetBlackIDListFromCache(context.Context, *GetBlackIDListFromCacheReq) (*GetBlackIDListFromCacheResp, error)
	AddBlackUserToCache(context.Context, *AddBlackUserToCacheReq) (*AddBlackUserToCacheResp, error)
	ReduceBlackUserFromCache(context.Context, *ReduceBlackUserFromCacheReq) (*ReduceBlackUserFromCacheResp, error)
	AddBlackFriendsToCache(context.Context, *AddBlackFriendsToCacheReq) (*AddBlackFriendsToCacheResp, error)
	ReduceBlackFriendsFromCache(context.Context, *ReduceBlackFriendsFromCacheReq) (*ReduceBlackFriendsFromCacheResp, error)
	// group
	GetGroupMemberIDListFromCache(context.Context, *GetGroupMemberIDListFromCacheReq) (*GetGroupMemberIDListFromCacheResp, error)
	AddGroupMemberToCache(context.Context, *AddGroupMemberToCacheReq) (*AddGroupMemberToCacheResp, error)
	ReduceGroupMemberFromCache(context.Context, *ReduceGroupMemberFromCacheReq) (*ReduceGroupMemberFromCacheResp, error)
}

// UnimplementedCacheServer should be embedded to have forward compatible implementations.
type UnimplementedCacheServer struct {
}

func (UnimplementedCacheServer) GetUserInfoFromCache(context.Context, *GetUserInfoFromCacheReq) (*GetUserInfoFromCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserInfoFromCache not implemented")
}
func (UnimplementedCacheServer) UpdateUserInfoToCache(context.Context, *UpdateUserInfoToCacheReq) (*UpdateUserInfoToCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUserInfoToCache not implemented")
}
func (UnimplementedCacheServer) GetFriendIDListFromCache(context.Context, *GetFriendIDListFromCacheReq) (*GetFriendIDListFromCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFriendIDListFromCache not implemented")
}
func (UnimplementedCacheServer) AddFriendToCache(context.Context, *AddFriendToCacheReq) (*AddFriendToCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddFriendToCache not implemented")
}
func (UnimplementedCacheServer) ReduceFriendFromCache(context.Context, *ReduceFriendFromCacheReq) (*ReduceFriendFromCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReduceFriendFromCache not implemented")
}
func (UnimplementedCacheServer) GetBlackIDListFromCache(context.Context, *GetBlackIDListFromCacheReq) (*GetBlackIDListFromCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlackIDListFromCache not implemented")
}
func (UnimplementedCacheServer) AddBlackUserToCache(context.Context, *AddBlackUserToCacheReq) (*AddBlackUserToCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddBlackUserToCache not implemented")
}
func (UnimplementedCacheServer) ReduceBlackUserFromCache(context.Context, *ReduceBlackUserFromCacheReq) (*ReduceBlackUserFromCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReduceBlackUserFromCache not implemented")
}
func (UnimplementedCacheServer) AddBlackFriendsToCache(context.Context, *AddBlackFriendsToCacheReq) (*AddBlackFriendsToCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddBlackFriendsToCache not implemented")
}
func (UnimplementedCacheServer) ReduceBlackFriendsFromCache(context.Context, *ReduceBlackFriendsFromCacheReq) (*ReduceBlackFriendsFromCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReduceBlackFriendsFromCache not implemented")
}
func (UnimplementedCacheServer) GetGroupMemberIDListFromCache(context.Context, *GetGroupMemberIDListFromCacheReq) (*GetGroupMemberIDListFromCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGroupMemberIDListFromCache not implemented")
}
func (UnimplementedCacheServer) AddGroupMemberToCache(context.Context, *AddGroupMemberToCacheReq) (*AddGroupMemberToCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddGroupMemberToCache not implemented")
}
func (UnimplementedCacheServer) ReduceGroupMemberFromCache(context.Context, *ReduceGroupMemberFromCacheReq) (*ReduceGroupMemberFromCacheResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReduceGroupMemberFromCache not implemented")
}

// UnsafeCacheServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CacheServer will
// result in compilation errors.
type UnsafeCacheServer interface {
	mustEmbedUnimplementedCacheServer()
}

func RegisterCacheServer(s grpc.ServiceRegistrar, srv CacheServer) {
	s.RegisterService(&Cache_ServiceDesc, srv)
}

func _Cache_GetUserInfoFromCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserInfoFromCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).GetUserInfoFromCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/GetUserInfoFromCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).GetUserInfoFromCache(ctx, req.(*GetUserInfoFromCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_UpdateUserInfoToCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateUserInfoToCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).UpdateUserInfoToCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/UpdateUserInfoToCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).UpdateUserInfoToCache(ctx, req.(*UpdateUserInfoToCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_GetFriendIDListFromCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFriendIDListFromCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).GetFriendIDListFromCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/GetFriendIDListFromCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).GetFriendIDListFromCache(ctx, req.(*GetFriendIDListFromCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_AddFriendToCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddFriendToCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).AddFriendToCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/AddFriendToCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).AddFriendToCache(ctx, req.(*AddFriendToCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_ReduceFriendFromCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReduceFriendFromCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).ReduceFriendFromCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/ReduceFriendFromCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).ReduceFriendFromCache(ctx, req.(*ReduceFriendFromCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_GetBlackIDListFromCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlackIDListFromCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).GetBlackIDListFromCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/GetBlackIDListFromCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).GetBlackIDListFromCache(ctx, req.(*GetBlackIDListFromCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_AddBlackUserToCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddBlackUserToCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).AddBlackUserToCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/AddBlackUserToCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).AddBlackUserToCache(ctx, req.(*AddBlackUserToCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_ReduceBlackUserFromCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReduceBlackUserFromCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).ReduceBlackUserFromCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/ReduceBlackUserFromCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).ReduceBlackUserFromCache(ctx, req.(*ReduceBlackUserFromCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_AddBlackFriendsToCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddBlackFriendsToCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).AddBlackFriendsToCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/AddBlackFriendsToCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).AddBlackFriendsToCache(ctx, req.(*AddBlackFriendsToCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_ReduceBlackFriendsFromCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReduceBlackFriendsFromCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).ReduceBlackFriendsFromCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/ReduceBlackFriendsFromCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).ReduceBlackFriendsFromCache(ctx, req.(*ReduceBlackFriendsFromCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_GetGroupMemberIDListFromCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGroupMemberIDListFromCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).GetGroupMemberIDListFromCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/GetGroupMemberIDListFromCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).GetGroupMemberIDListFromCache(ctx, req.(*GetGroupMemberIDListFromCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_AddGroupMemberToCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddGroupMemberToCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).AddGroupMemberToCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/AddGroupMemberToCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).AddGroupMemberToCache(ctx, req.(*AddGroupMemberToCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_ReduceGroupMemberFromCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReduceGroupMemberFromCacheReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).ReduceGroupMemberFromCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cache.cache/ReduceGroupMemberFromCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).ReduceGroupMemberFromCache(ctx, req.(*ReduceGroupMemberFromCacheReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Cache_ServiceDesc is the grpc.ServiceDesc for Cache service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Cache_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "cache.cache",
	HandlerType: (*CacheServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetUserInfoFromCache",
			Handler:    _Cache_GetUserInfoFromCache_Handler,
		},
		{
			MethodName: "UpdateUserInfoToCache",
			Handler:    _Cache_UpdateUserInfoToCache_Handler,
		},
		{
			MethodName: "GetFriendIDListFromCache",
			Handler:    _Cache_GetFriendIDListFromCache_Handler,
		},
		{
			MethodName: "AddFriendToCache",
			Handler:    _Cache_AddFriendToCache_Handler,
		},
		{
			MethodName: "ReduceFriendFromCache",
			Handler:    _Cache_ReduceFriendFromCache_Handler,
		},
		{
			MethodName: "GetBlackIDListFromCache",
			Handler:    _Cache_GetBlackIDListFromCache_Handler,
		},
		{
			MethodName: "AddBlackUserToCache",
			Handler:    _Cache_AddBlackUserToCache_Handler,
		},
		{
			MethodName: "ReduceBlackUserFromCache",
			Handler:    _Cache_ReduceBlackUserFromCache_Handler,
		},
		{
			MethodName: "AddBlackFriendsToCache",
			Handler:    _Cache_AddBlackFriendsToCache_Handler,
		},
		{
			MethodName: "ReduceBlackFriendsFromCache",
			Handler:    _Cache_ReduceBlackFriendsFromCache_Handler,
		},
		{
			MethodName: "GetGroupMemberIDListFromCache",
			Handler:    _Cache_GetGroupMemberIDListFromCache_Handler,
		},
		{
			MethodName: "AddGroupMemberToCache",
			Handler:    _Cache_AddGroupMemberToCache_Handler,
		},
		{
			MethodName: "ReduceGroupMemberFromCache",
			Handler:    _Cache_ReduceGroupMemberFromCache_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cache/cache.proto",
}