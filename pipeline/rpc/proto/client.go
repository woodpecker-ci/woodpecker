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

// Client API for Drone service

type DroneClient interface {
	Next(ctx context.Context, in *NextRequest, opts ...grpc.CallOption) (*NextReply, error)
	Init(ctx context.Context, in *InitRequest, opts ...grpc.CallOption) (*Empty, error)
	Wait(ctx context.Context, in *WaitRequest, opts ...grpc.CallOption) (*Empty, error)
	Done(ctx context.Context, in *DoneRequest, opts ...grpc.CallOption) (*Empty, error)
	Extend(ctx context.Context, in *ExtendRequest, opts ...grpc.CallOption) (*Empty, error)
	Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*Empty, error)
	Upload(ctx context.Context, in *UploadRequest, opts ...grpc.CallOption) (*Empty, error)
	Log(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (*Empty, error)
}

type droneClient struct {
	cc *grpc.ClientConn
}

func NewDroneClient(cc *grpc.ClientConn) DroneClient {
	return &droneClient{cc}
}

func (c *droneClient) Next(ctx context.Context, in *NextRequest, opts ...grpc.CallOption) (*NextReply, error) {
	out := new(NextReply)
	// DEPRECATED: Use ClientConn.Invoke instead
	err := grpc.Invoke(ctx, "/proto.Drone/Next", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *droneClient) Init(ctx context.Context, in *InitRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	// DEPRECATED: Use ClientConn.Invoke instead
	err := grpc.Invoke(ctx, "/proto.Drone/Init", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *droneClient) Wait(ctx context.Context, in *WaitRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	// DEPRECATED: Use ClientConn.Invoke instead
	err := grpc.Invoke(ctx, "/proto.Drone/Wait", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *droneClient) Done(ctx context.Context, in *DoneRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	// DEPRECATED: Use ClientConn.Invoke instead
	err := grpc.Invoke(ctx, "/proto.Drone/Done", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *droneClient) Extend(ctx context.Context, in *ExtendRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	// DEPRECATED: Use ClientConn.Invoke instead
	err := grpc.Invoke(ctx, "/proto.Drone/Extend", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *droneClient) Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	// DEPRECATED: Use ClientConn.Invoke instead
	err := grpc.Invoke(ctx, "/proto.Drone/Update", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *droneClient) Upload(ctx context.Context, in *UploadRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	// DEPRECATED: Use ClientConn.Invoke instead
	err := grpc.Invoke(ctx, "/proto.Drone/Upload", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *droneClient) Log(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	// DEPRECATED: Use ClientConn.Invoke instead
	err := grpc.Invoke(ctx, "/proto.Drone/Log", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Drone service

type DroneServer interface {
	Next(context.Context, *NextRequest) (*NextReply, error)
	Init(context.Context, *InitRequest) (*Empty, error)
	Wait(context.Context, *WaitRequest) (*Empty, error)
	Done(context.Context, *DoneRequest) (*Empty, error)
	Extend(context.Context, *ExtendRequest) (*Empty, error)
	Update(context.Context, *UpdateRequest) (*Empty, error)
	Upload(context.Context, *UploadRequest) (*Empty, error)
	Log(context.Context, *LogRequest) (*Empty, error)
}

func RegisterDroneServer(s *grpc.Server, srv DroneServer) {
	s.RegisterService(&DroneServiceDesc, srv)
}

func NextHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NextRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DroneServer).Next(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Drone/Next",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DroneServer).Next(ctx, req.(*NextRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func InitHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DroneServer).Init(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Drone/Init",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DroneServer).Init(ctx, req.(*InitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func WaitHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WaitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DroneServer).Wait(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Drone/Wait",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DroneServer).Wait(ctx, req.(*WaitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func DoneHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DoneRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DroneServer).Done(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Drone/Done",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DroneServer).Done(ctx, req.(*DoneRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func ExtendHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExtendRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DroneServer).Extend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Drone/Extend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DroneServer).Extend(ctx, req.(*ExtendRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func UpdateHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DroneServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Drone/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DroneServer).Update(ctx, req.(*UpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func UploadHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DroneServer).Upload(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Drone/Upload",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DroneServer).Upload(ctx, req.(*UploadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func LogHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DroneServer).Log(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Drone/Log",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DroneServer).Log(ctx, req.(*LogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var DroneServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Drone",
	HandlerType: (*DroneServer)(nil),
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
	Metadata: "drone.proto",
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
	// DEPRECATED: Use ClientConn.Invoke instead
	err := grpc.Invoke(ctx, "/proto.Health/Check", in, out, c.cc, opts...)
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
	Metadata: "drone.proto",
}
