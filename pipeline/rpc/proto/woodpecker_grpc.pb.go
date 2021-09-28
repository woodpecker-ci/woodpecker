// Copyright 2021 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proto

import (
	"context"

	"google.golang.org/grpc"
)

// Client API for Woodpecker service

type WoodpeckerClient interface {
	Next(ctx context.Context, in *NextRequest, opts ...grpc.CallOption) (*NextReply, error)
	Init(ctx context.Context, in *InitRequest, opts ...grpc.CallOption) (*Empty, error)
	Wait(ctx context.Context, in *WaitRequest, opts ...grpc.CallOption) (*Empty, error)
	Done(ctx context.Context, in *DoneRequest, opts ...grpc.CallOption) (*Empty, error)
	Extend(ctx context.Context, in *ExtendRequest, opts ...grpc.CallOption) (*Empty, error)
	Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*Empty, error)
	Upload(ctx context.Context, in *UploadRequest, opts ...grpc.CallOption) (*Empty, error)
	Log(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (*Empty, error)
}

type woodpeckerClient struct {
	cc *grpc.ClientConn
}

func NewWoodpeckerClient(cc *grpc.ClientConn) WoodpeckerClient {
	return &woodpeckerClient{cc}
}

func (c *woodpeckerClient) Next(ctx context.Context, in *NextRequest, opts ...grpc.CallOption) (*NextReply, error) {
	out := new(NextReply)
	err := c.cc.Invoke(ctx, "/proto.Woodpecker/Next", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *woodpeckerClient) Init(ctx context.Context, in *InitRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/proto.Woodpecker/Init", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *woodpeckerClient) Wait(ctx context.Context, in *WaitRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/proto.Woodpecker/Wait", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *woodpeckerClient) Done(ctx context.Context, in *DoneRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/proto.Woodpecker/Done", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *woodpeckerClient) Extend(ctx context.Context, in *ExtendRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/proto.Woodpecker/Extend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *woodpeckerClient) Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/proto.Woodpecker/Update", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *woodpeckerClient) Upload(ctx context.Context, in *UploadRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/proto.Woodpecker/Upload", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *woodpeckerClient) Log(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/proto.Woodpecker/Log", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Woodpecker service

type WoodpeckerServer interface {
	Next(context.Context, *NextRequest) (*NextReply, error)
	Init(context.Context, *InitRequest) (*Empty, error)
	Wait(context.Context, *WaitRequest) (*Empty, error)
	Done(context.Context, *DoneRequest) (*Empty, error)
	Extend(context.Context, *ExtendRequest) (*Empty, error)
	Update(context.Context, *UpdateRequest) (*Empty, error)
	Upload(context.Context, *UploadRequest) (*Empty, error)
	Log(context.Context, *LogRequest) (*Empty, error)
}

func RegisterWoodpeckerServer(s *grpc.Server, srv WoodpeckerServer) {
	s.RegisterService(&WoodpeckerServiceDesc, srv)
}

func NextHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NextRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WoodpeckerServer).Next(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Woodpecker/Next",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WoodpeckerServer).Next(ctx, req.(*NextRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func InitHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WoodpeckerServer).Init(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Woodpecker/Init",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WoodpeckerServer).Init(ctx, req.(*InitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func WaitHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WaitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WoodpeckerServer).Wait(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Woodpecker/Wait",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WoodpeckerServer).Wait(ctx, req.(*WaitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func DoneHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DoneRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WoodpeckerServer).Done(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Woodpecker/Done",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WoodpeckerServer).Done(ctx, req.(*DoneRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func ExtendHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExtendRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WoodpeckerServer).Extend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Woodpecker/Extend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WoodpeckerServer).Extend(ctx, req.(*ExtendRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func UpdateHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WoodpeckerServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Woodpecker/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WoodpeckerServer).Update(ctx, req.(*UpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func UploadHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WoodpeckerServer).Upload(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Woodpecker/Upload",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WoodpeckerServer).Upload(ctx, req.(*UploadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func LogHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WoodpeckerServer).Log(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Woodpecker/Log",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WoodpeckerServer).Log(ctx, req.(*LogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var WoodpeckerServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Woodpecker",
	HandlerType: (*WoodpeckerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Next",
			Handler:    NextHandler,
		},
		{
			MethodName: "Init",
			Handler:    InitHandler,
		},
		{
			MethodName: "Wait",
			Handler:    WaitHandler,
		},
		{
			MethodName: "Done",
			Handler:    DoneHandler,
		},
		{
			MethodName: "Extend",
			Handler:    ExtendHandler,
		},
		{
			MethodName: "Update",
			Handler:    UpdateHandler,
		},
		{
			MethodName: "Upload",
			Handler:    UploadHandler,
		},
		{
			MethodName: "Log",
			Handler:    LogHandler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "woodpecker.proto",
}

// Client API for Health service

type HealthClient interface {
	Check(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error)
}

type healthClient struct {
	cc *grpc.ClientConn
}

func NewHealthClient(cc *grpc.ClientConn) HealthClient {
	return &healthClient{cc}
}

func (c *healthClient) Check(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error) {
	out := new(HealthCheckResponse)
	err := c.cc.Invoke(ctx, "/proto.Health/Check", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HealthServer Server API for Health service
type HealthServer interface {
	Check(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error)
}

func RegisterHealthServer(s *grpc.Server, srv HealthServer) {
	s.RegisterService(&HealthServiceDesc, srv)
}

func HealthCheckHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthCheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HealthServer).Check(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Health/Check",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HealthServer).Check(ctx, req.(*HealthCheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var HealthServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Health",
	HandlerType: (*HealthServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Check",
			Handler:    HealthCheckHandler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "woodpecker.proto",
}
