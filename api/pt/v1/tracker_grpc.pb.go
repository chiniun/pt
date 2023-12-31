// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.4
// source: pt/v1/tracker.proto

package v1

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
	Tracker_Announce_FullMethodName = "/pt.v1.Tracker/Announce"
	Tracker_Scrape_FullMethodName   = "/pt.v1.Tracker/Scrape"
)

// TrackerClient is the client API for Tracker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TrackerClient interface {
	// Sends a greeting
	Announce(ctx context.Context, in *AnnounceRequest, opts ...grpc.CallOption) (*AnnounceReply, error)
	Scrape(ctx context.Context, in *ScrapeRequest, opts ...grpc.CallOption) (*ScrapeReply, error)
}

type trackerClient struct {
	cc grpc.ClientConnInterface
}

func NewTrackerClient(cc grpc.ClientConnInterface) TrackerClient {
	return &trackerClient{cc}
}

func (c *trackerClient) Announce(ctx context.Context, in *AnnounceRequest, opts ...grpc.CallOption) (*AnnounceReply, error) {
	out := new(AnnounceReply)
	err := c.cc.Invoke(ctx, Tracker_Announce_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trackerClient) Scrape(ctx context.Context, in *ScrapeRequest, opts ...grpc.CallOption) (*ScrapeReply, error) {
	out := new(ScrapeReply)
	err := c.cc.Invoke(ctx, Tracker_Scrape_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TrackerServer is the server API for Tracker service.
// All implementations must embed UnimplementedTrackerServer
// for forward compatibility
type TrackerServer interface {
	// Sends a greeting
	Announce(context.Context, *AnnounceRequest) (*AnnounceReply, error)
	Scrape(context.Context, *ScrapeRequest) (*ScrapeReply, error)
	mustEmbedUnimplementedTrackerServer()
}

// UnimplementedTrackerServer must be embedded to have forward compatible implementations.
type UnimplementedTrackerServer struct {
}

func (UnimplementedTrackerServer) Announce(context.Context, *AnnounceRequest) (*AnnounceReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Announce not implemented")
}
func (UnimplementedTrackerServer) Scrape(context.Context, *ScrapeRequest) (*ScrapeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Scrape not implemented")
}
func (UnimplementedTrackerServer) mustEmbedUnimplementedTrackerServer() {}

// UnsafeTrackerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TrackerServer will
// result in compilation errors.
type UnsafeTrackerServer interface {
	mustEmbedUnimplementedTrackerServer()
}

func RegisterTrackerServer(s grpc.ServiceRegistrar, srv TrackerServer) {
	s.RegisterService(&Tracker_ServiceDesc, srv)
}

func _Tracker_Announce_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AnnounceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrackerServer).Announce(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Tracker_Announce_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrackerServer).Announce(ctx, req.(*AnnounceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tracker_Scrape_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ScrapeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrackerServer).Scrape(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Tracker_Scrape_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrackerServer).Scrape(ctx, req.(*ScrapeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Tracker_ServiceDesc is the grpc.ServiceDesc for Tracker service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Tracker_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pt.v1.Tracker",
	HandlerType: (*TrackerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Announce",
			Handler:    _Tracker_Announce_Handler,
		},
		{
			MethodName: "Scrape",
			Handler:    _Tracker_Scrape_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pt/v1/tracker.proto",
}
