// Code generated by protoc-gen-go. DO NOT EDIT.
// source: dapr/proto/internals/v1/service_invocation.proto

package internals

import (
	context "context"
	fmt "fmt"
	v1 "github.com/dapr/dapr/pkg/proto/common/v1"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Actor represents actor using actor_type and actor_id
type Actor struct {
	// Required. The type of actor.
	ActorType string `protobuf:"bytes,1,opt,name=actor_type,json=actorType,proto3" json:"actor_type,omitempty"`
	// Required. The ID of actor type (actor_type)
	ActorId              string   `protobuf:"bytes,2,opt,name=actor_id,json=actorId,proto3" json:"actor_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Actor) Reset()         { *m = Actor{} }
func (m *Actor) String() string { return proto.CompactTextString(m) }
func (*Actor) ProtoMessage()    {}
func (*Actor) Descriptor() ([]byte, []int) {
	return fileDescriptor_6a1e51ba6ea480e4, []int{0}
}

func (m *Actor) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Actor.Unmarshal(m, b)
}
func (m *Actor) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Actor.Marshal(b, m, deterministic)
}
func (m *Actor) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Actor.Merge(m, src)
}
func (m *Actor) XXX_Size() int {
	return xxx_messageInfo_Actor.Size(m)
}
func (m *Actor) XXX_DiscardUnknown() {
	xxx_messageInfo_Actor.DiscardUnknown(m)
}

var xxx_messageInfo_Actor proto.InternalMessageInfo

func (m *Actor) GetActorType() string {
	if m != nil {
		return m.ActorType
	}
	return ""
}

func (m *Actor) GetActorId() string {
	if m != nil {
		return m.ActorId
	}
	return ""
}

// InternalInvokeRequest is the message to transfer caller's data to callee
// for service invocaton. This includes callee's app id and caller's request data.
type InternalInvokeRequest struct {
	// Required. The version of Dapr runtime API.
	Ver APIVersion `protobuf:"varint,1,opt,name=ver,proto3,enum=dapr.proto.internals.v1.APIVersion" json:"ver,omitempty"`
	// Required. metadata holds caller's HTTP headers or gRPC metadata.
	Metadata map[string]*ListStringValue `protobuf:"bytes,2,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Required. message including caller's invocation request.
	Message *v1.InvokeRequest `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	// Actor type and id. This field is used only for
	// actor service invocation.
	Actor                *Actor   `protobuf:"bytes,4,opt,name=actor,proto3" json:"actor,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *InternalInvokeRequest) Reset()         { *m = InternalInvokeRequest{} }
func (m *InternalInvokeRequest) String() string { return proto.CompactTextString(m) }
func (*InternalInvokeRequest) ProtoMessage()    {}
func (*InternalInvokeRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_6a1e51ba6ea480e4, []int{1}
}

func (m *InternalInvokeRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InternalInvokeRequest.Unmarshal(m, b)
}
func (m *InternalInvokeRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InternalInvokeRequest.Marshal(b, m, deterministic)
}
func (m *InternalInvokeRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InternalInvokeRequest.Merge(m, src)
}
func (m *InternalInvokeRequest) XXX_Size() int {
	return xxx_messageInfo_InternalInvokeRequest.Size(m)
}
func (m *InternalInvokeRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_InternalInvokeRequest.DiscardUnknown(m)
}

var xxx_messageInfo_InternalInvokeRequest proto.InternalMessageInfo

func (m *InternalInvokeRequest) GetVer() APIVersion {
	if m != nil {
		return m.Ver
	}
	return APIVersion_APIVERSION_UNSPECIFIED
}

func (m *InternalInvokeRequest) GetMetadata() map[string]*ListStringValue {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *InternalInvokeRequest) GetMessage() *v1.InvokeRequest {
	if m != nil {
		return m.Message
	}
	return nil
}

func (m *InternalInvokeRequest) GetActor() *Actor {
	if m != nil {
		return m.Actor
	}
	return nil
}

// InternalInvokeResponse is the message to transfer callee's response to caller
// for service invocaton.
type InternalInvokeResponse struct {
	// Required. HTTP/gRPC status.
	Status *Status `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	// Required. The app callback response headers.
	Headers map[string]*ListStringValue `protobuf:"bytes,2,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// App callback response trailers.
	// This will be used only for gRPC app callback
	Trailers map[string]*ListStringValue `protobuf:"bytes,3,rep,name=trailers,proto3" json:"trailers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Callee's invocation response message.
	Message              *v1.InvokeResponse `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *InternalInvokeResponse) Reset()         { *m = InternalInvokeResponse{} }
func (m *InternalInvokeResponse) String() string { return proto.CompactTextString(m) }
func (*InternalInvokeResponse) ProtoMessage()    {}
func (*InternalInvokeResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_6a1e51ba6ea480e4, []int{2}
}

func (m *InternalInvokeResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InternalInvokeResponse.Unmarshal(m, b)
}
func (m *InternalInvokeResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InternalInvokeResponse.Marshal(b, m, deterministic)
}
func (m *InternalInvokeResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InternalInvokeResponse.Merge(m, src)
}
func (m *InternalInvokeResponse) XXX_Size() int {
	return xxx_messageInfo_InternalInvokeResponse.Size(m)
}
func (m *InternalInvokeResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_InternalInvokeResponse.DiscardUnknown(m)
}

var xxx_messageInfo_InternalInvokeResponse proto.InternalMessageInfo

func (m *InternalInvokeResponse) GetStatus() *Status {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *InternalInvokeResponse) GetHeaders() map[string]*ListStringValue {
	if m != nil {
		return m.Headers
	}
	return nil
}

func (m *InternalInvokeResponse) GetTrailers() map[string]*ListStringValue {
	if m != nil {
		return m.Trailers
	}
	return nil
}

func (m *InternalInvokeResponse) GetMessage() *v1.InvokeResponse {
	if m != nil {
		return m.Message
	}
	return nil
}

// ListStringValue represents string value array
type ListStringValue struct {
	// The array of string.
	Values               []string `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListStringValue) Reset()         { *m = ListStringValue{} }
func (m *ListStringValue) String() string { return proto.CompactTextString(m) }
func (*ListStringValue) ProtoMessage()    {}
func (*ListStringValue) Descriptor() ([]byte, []int) {
	return fileDescriptor_6a1e51ba6ea480e4, []int{3}
}

func (m *ListStringValue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListStringValue.Unmarshal(m, b)
}
func (m *ListStringValue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListStringValue.Marshal(b, m, deterministic)
}
func (m *ListStringValue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListStringValue.Merge(m, src)
}
func (m *ListStringValue) XXX_Size() int {
	return xxx_messageInfo_ListStringValue.Size(m)
}
func (m *ListStringValue) XXX_DiscardUnknown() {
	xxx_messageInfo_ListStringValue.DiscardUnknown(m)
}

var xxx_messageInfo_ListStringValue proto.InternalMessageInfo

func (m *ListStringValue) GetValues() []string {
	if m != nil {
		return m.Values
	}
	return nil
}

func init() {
	proto.RegisterType((*Actor)(nil), "dapr.proto.internals.v1.Actor")
	proto.RegisterType((*InternalInvokeRequest)(nil), "dapr.proto.internals.v1.InternalInvokeRequest")
	proto.RegisterMapType((map[string]*ListStringValue)(nil), "dapr.proto.internals.v1.InternalInvokeRequest.MetadataEntry")
	proto.RegisterType((*InternalInvokeResponse)(nil), "dapr.proto.internals.v1.InternalInvokeResponse")
	proto.RegisterMapType((map[string]*ListStringValue)(nil), "dapr.proto.internals.v1.InternalInvokeResponse.HeadersEntry")
	proto.RegisterMapType((map[string]*ListStringValue)(nil), "dapr.proto.internals.v1.InternalInvokeResponse.TrailersEntry")
	proto.RegisterType((*ListStringValue)(nil), "dapr.proto.internals.v1.ListStringValue")
}

func init() {
	proto.RegisterFile("dapr/proto/internals/v1/service_invocation.proto", fileDescriptor_6a1e51ba6ea480e4)
}

var fileDescriptor_6a1e51ba6ea480e4 = []byte{
	// 535 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x54, 0xdd, 0x6e, 0xd3, 0x30,
	0x14, 0xa6, 0xcd, 0xfa, 0x77, 0xca, 0xaf, 0x25, 0x46, 0x89, 0x04, 0x94, 0xb2, 0x8b, 0x70, 0x93,
	0xb0, 0xc2, 0x34, 0x04, 0x6c, 0x52, 0x41, 0x48, 0x44, 0x1a, 0x12, 0x4a, 0xa7, 0x0a, 0xb8, 0x99,
	0xbc, 0xc4, 0xea, 0xac, 0xa6, 0x76, 0xb0, 0xdd, 0x48, 0xbd, 0xe7, 0x05, 0x78, 0x56, 0x5e, 0x00,
	0xd9, 0x4e, 0x43, 0x3b, 0x2d, 0x48, 0xbd, 0x18, 0x77, 0x76, 0xce, 0xf7, 0x73, 0xec, 0xcf, 0x39,
	0xf0, 0x22, 0xc1, 0x99, 0x08, 0x32, 0xc1, 0x15, 0x0f, 0x28, 0x53, 0x44, 0x30, 0x9c, 0xca, 0x20,
	0xdf, 0x0f, 0x24, 0x11, 0x39, 0x8d, 0xc9, 0x19, 0x65, 0x39, 0x8f, 0xb1, 0xa2, 0x9c, 0xf9, 0x06,
	0x85, 0x1e, 0x68, 0x86, 0x5d, 0xfb, 0x25, 0xc3, 0xcf, 0xf7, 0xdd, 0xa7, 0x6b, 0x52, 0x31, 0x9f,
	0xcf, 0x39, 0xd3, 0x3a, 0x76, 0x65, 0xf1, 0xae, 0x57, 0xe5, 0x86, 0x33, 0x9a, 0x13, 0x21, 0x4b,
	0x17, 0x77, 0xaf, 0xb2, 0x2f, 0x85, 0xd5, 0x42, 0x5a, 0xd4, 0x60, 0x04, 0x8d, 0x51, 0xac, 0xb8,
	0x40, 0x8f, 0x00, 0xb0, 0x5e, 0x9c, 0xa9, 0x65, 0x46, 0x7a, 0xb5, 0x7e, 0xcd, 0xeb, 0x44, 0x1d,
	0xf3, 0xe5, 0x74, 0x99, 0x11, 0xf4, 0x10, 0xda, 0xb6, 0x4c, 0x93, 0x5e, 0xdd, 0x14, 0x5b, 0x66,
	0x1f, 0x26, 0x83, 0x9f, 0x0e, 0xdc, 0x0f, 0x0b, 0x83, 0x90, 0xe5, 0x7c, 0x46, 0x22, 0xf2, 0x63,
	0x41, 0xa4, 0x42, 0x07, 0xe0, 0xe4, 0x44, 0x18, 0xb1, 0xdb, 0xc3, 0x67, 0x7e, 0xc5, 0xb1, 0xfd,
	0xd1, 0x97, 0x70, 0x62, 0x5b, 0x8f, 0x34, 0x1e, 0x7d, 0x85, 0xf6, 0x9c, 0x28, 0x9c, 0x60, 0x85,
	0x7b, 0xf5, 0xbe, 0xe3, 0x75, 0x87, 0xef, 0x2a, 0xb9, 0x57, 0x1a, 0xfb, 0x9f, 0x0b, 0xfa, 0x47,
	0xa6, 0xc4, 0x32, 0x2a, 0xd5, 0xd0, 0x11, 0xb4, 0xe6, 0x44, 0x4a, 0x3c, 0x25, 0x3d, 0xa7, 0x5f,
	0xf3, 0xba, 0x9b, 0x4d, 0x15, 0x17, 0x6d, 0x54, 0xd7, 0xd4, 0xa2, 0x15, 0x07, 0xbd, 0x82, 0x86,
	0x39, 0x74, 0x6f, 0xc7, 0x90, 0x1f, 0x57, 0x9f, 0x48, 0xa3, 0x22, 0x0b, 0x76, 0x09, 0xdc, 0xda,
	0xe8, 0x07, 0xdd, 0x05, 0x67, 0x46, 0x96, 0xc5, 0x1d, 0xeb, 0x25, 0x3a, 0x86, 0x46, 0x8e, 0xd3,
	0x05, 0x31, 0x57, 0xdb, 0x1d, 0x7a, 0x95, 0xc2, 0x27, 0x54, 0xaa, 0xb1, 0x12, 0x94, 0x4d, 0x27,
	0x1a, 0x1f, 0x59, 0xda, 0x9b, 0xfa, 0xeb, 0xda, 0xe0, 0xd7, 0x0e, 0xec, 0x5e, 0xbe, 0x0d, 0x99,
	0x71, 0x26, 0x09, 0x3a, 0x84, 0xa6, 0x0d, 0xdd, 0x78, 0x76, 0x87, 0x4f, 0x2a, 0xf5, 0xc7, 0x06,
	0x16, 0x15, 0x70, 0x34, 0x81, 0xd6, 0x05, 0xc1, 0x09, 0x11, 0x72, 0xeb, 0x20, 0xac, 0xb5, 0xff,
	0xc9, 0xd2, 0x6d, 0x10, 0x2b, 0x31, 0xf4, 0x0d, 0xda, 0x4a, 0x60, 0x9a, 0x6a, 0x61, 0xc7, 0x08,
	0x1f, 0x6d, 0x2b, 0x7c, 0x5a, 0xf0, 0x8b, 0x88, 0x57, 0x72, 0xe8, 0xf8, 0x6f, 0xc4, 0x36, 0xa5,
	0xbd, 0x7f, 0x47, 0x6c, 0xe5, 0xca, 0x8c, 0xdd, 0x04, 0x6e, 0xae, 0xf7, 0x7c, 0x3d, 0x61, 0xe9,
	0x37, 0xb1, 0x71, 0x80, 0x6b, 0x7a, 0x13, 0xcf, 0xe1, 0xce, 0xa5, 0x2a, 0xda, 0x85, 0xa6, 0xa9,
	0xeb, 0xb7, 0xe0, 0x78, 0x9d, 0xa8, 0xd8, 0x0d, 0x7f, 0xd7, 0xe0, 0xde, 0xd8, 0x4e, 0xac, 0xb0,
	0x1c, 0x58, 0x88, 0x41, 0xe7, 0x03, 0x4e, 0x53, 0x3b, 0x22, 0xfc, 0xed, 0xfe, 0x42, 0x37, 0xd8,
	0x32, 0xd3, 0xc1, 0x8d, 0x95, 0xdf, 0x09, 0x8f, 0x71, 0xfa, 0x1f, 0xfc, 0xde, 0x1f, 0x7e, 0x3f,
	0x98, 0x52, 0x75, 0xb1, 0x38, 0xd7, 0x2f, 0x23, 0x30, 0x13, 0xd3, 0x8e, 0xcd, 0xd9, 0xf4, 0x8a,
	0xd1, 0xf9, 0xb6, 0xdc, 0x9c, 0x37, 0x4d, 0xf5, 0xe5, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xcc,
	0x35, 0xd8, 0x19, 0xfe, 0x05, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ServiceInvocationClient is the client API for ServiceInvocation service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ServiceInvocationClient interface {
	// Invokes a method of the specific actor.
	CallActor(ctx context.Context, in *InternalInvokeRequest, opts ...grpc.CallOption) (*InternalInvokeResponse, error)
	// Invokes a method of the specific service.
	CallLocal(ctx context.Context, in *InternalInvokeRequest, opts ...grpc.CallOption) (*InternalInvokeResponse, error)
}

type serviceInvocationClient struct {
	cc *grpc.ClientConn
}

func NewServiceInvocationClient(cc *grpc.ClientConn) ServiceInvocationClient {
	return &serviceInvocationClient{cc}
}

func (c *serviceInvocationClient) CallActor(ctx context.Context, in *InternalInvokeRequest, opts ...grpc.CallOption) (*InternalInvokeResponse, error) {
	out := new(InternalInvokeResponse)
	err := c.cc.Invoke(ctx, "/dapr.proto.internals.v1.ServiceInvocation/CallActor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceInvocationClient) CallLocal(ctx context.Context, in *InternalInvokeRequest, opts ...grpc.CallOption) (*InternalInvokeResponse, error) {
	out := new(InternalInvokeResponse)
	err := c.cc.Invoke(ctx, "/dapr.proto.internals.v1.ServiceInvocation/CallLocal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServiceInvocationServer is the server API for ServiceInvocation service.
type ServiceInvocationServer interface {
	// Invokes a method of the specific actor.
	CallActor(context.Context, *InternalInvokeRequest) (*InternalInvokeResponse, error)
	// Invokes a method of the specific service.
	CallLocal(context.Context, *InternalInvokeRequest) (*InternalInvokeResponse, error)
}

// UnimplementedServiceInvocationServer can be embedded to have forward compatible implementations.
type UnimplementedServiceInvocationServer struct {
}

func (*UnimplementedServiceInvocationServer) CallActor(ctx context.Context, req *InternalInvokeRequest) (*InternalInvokeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CallActor not implemented")
}
func (*UnimplementedServiceInvocationServer) CallLocal(ctx context.Context, req *InternalInvokeRequest) (*InternalInvokeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CallLocal not implemented")
}

func RegisterServiceInvocationServer(s *grpc.Server, srv ServiceInvocationServer) {
	s.RegisterService(&_ServiceInvocation_serviceDesc, srv)
}

func _ServiceInvocation_CallActor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InternalInvokeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceInvocationServer).CallActor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dapr.proto.internals.v1.ServiceInvocation/CallActor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceInvocationServer).CallActor(ctx, req.(*InternalInvokeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ServiceInvocation_CallLocal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InternalInvokeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceInvocationServer).CallLocal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dapr.proto.internals.v1.ServiceInvocation/CallLocal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceInvocationServer).CallLocal(ctx, req.(*InternalInvokeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ServiceInvocation_serviceDesc = grpc.ServiceDesc{
	ServiceName: "dapr.proto.internals.v1.ServiceInvocation",
	HandlerType: (*ServiceInvocationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CallActor",
			Handler:    _ServiceInvocation_CallActor_Handler,
		},
		{
			MethodName: "CallLocal",
			Handler:    _ServiceInvocation_CallLocal_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "dapr/proto/internals/v1/service_invocation.proto",
}
