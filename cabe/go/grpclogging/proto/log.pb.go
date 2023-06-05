// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.12
// source: log.proto

package proto

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Entry is a structured log message for grpc events.
// It can represent either side of an rpc: client or server.
type Entry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start         *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=start,proto3" json:"start,omitempty"`
	Elapsed       *durationpb.Duration   `protobuf:"bytes,2,opt,name=elapsed,proto3" json:"elapsed,omitempty"`
	StatusCode    int32                  `protobuf:"varint,3,opt,name=status_code,json=statusCode,proto3" json:"status_code,omitempty"`
	StatusMessage string                 `protobuf:"bytes,4,opt,name=status_message,json=statusMessage,proto3" json:"status_message,omitempty"`
	StatusDetails []*anypb.Any           `protobuf:"bytes,5,rep,name=status_details,json=statusDetails,proto3" json:"status_details,omitempty"`
	Error         string                 `protobuf:"bytes,6,opt,name=error,proto3" json:"error,omitempty"`
	// When cabe receives an incoming gRPC request, that would
	// be logged in the server_unary field. Since cabe does
	// not currently implement any streaming server calls,
	// there is no corresponding server_stream field.
	ServerUnary *ServerUnary `protobuf:"bytes,7,opt,name=server_unary,json=serverUnary,proto3" json:"server_unary,omitempty"`
	// When cabe makes an outgoing unary gRPC call to some
	// other service, that gets logged in client_unary.
	ClientUnary *ClientUnary `protobuf:"bytes,8,opt,name=client_unary,json=clientUnary,proto3" json:"client_unary,omitempty"`
	// When cabe makes an outgoing streaming gRPC call to some
	// other service, that gets logged in client_stream.
	ClientStream *ClientStream `protobuf:"bytes,9,opt,name=client_stream,json=clientStream,proto3" json:"client_stream,omitempty"`
}

func (x *Entry) Reset() {
	*x = Entry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_log_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Entry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Entry) ProtoMessage() {}

func (x *Entry) ProtoReflect() protoreflect.Message {
	mi := &file_log_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Entry.ProtoReflect.Descriptor instead.
func (*Entry) Descriptor() ([]byte, []int) {
	return file_log_proto_rawDescGZIP(), []int{0}
}

func (x *Entry) GetStart() *timestamppb.Timestamp {
	if x != nil {
		return x.Start
	}
	return nil
}

func (x *Entry) GetElapsed() *durationpb.Duration {
	if x != nil {
		return x.Elapsed
	}
	return nil
}

func (x *Entry) GetStatusCode() int32 {
	if x != nil {
		return x.StatusCode
	}
	return 0
}

func (x *Entry) GetStatusMessage() string {
	if x != nil {
		return x.StatusMessage
	}
	return ""
}

func (x *Entry) GetStatusDetails() []*anypb.Any {
	if x != nil {
		return x.StatusDetails
	}
	return nil
}

func (x *Entry) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *Entry) GetServerUnary() *ServerUnary {
	if x != nil {
		return x.ServerUnary
	}
	return nil
}

func (x *Entry) GetClientUnary() *ClientUnary {
	if x != nil {
		return x.ClientUnary
	}
	return nil
}

func (x *Entry) GetClientStream() *ClientStream {
	if x != nil {
		return x.ClientStream
	}
	return nil
}

// ServerUnary logs the server side of a grpc unary request, as intercepted
// by https://pkg.go.dev/google.golang.org/grpc#UnaryServerInterceptor
type ServerUnary struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// request is the incoming grpc request proto sent by
	// a client calling cabe.
	Request *anypb.Any `protobuf:"bytes,1,opt,name=request,proto3" json:"request,omitempty"`
	// response is the outgoing grpc response proto returned
	// by cabe to the caller.
	Response *anypb.Any `protobuf:"bytes,2,opt,name=response,proto3" json:"response,omitempty"`
	// full_method is the fully qualified grpc method being
	// called, e.g. "cabe.proto.Analysis/GetAnalysis"
	FullMethod string `protobuf:"bytes,3,opt,name=full_method,json=fullMethod,proto3" json:"full_method,omitempty"`
	// user is the identity of the user making the incoming server request.
	User string `protobuf:"bytes,4,opt,name=user,proto3" json:"user,omitempty"`
}

func (x *ServerUnary) Reset() {
	*x = ServerUnary{}
	if protoimpl.UnsafeEnabled {
		mi := &file_log_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerUnary) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerUnary) ProtoMessage() {}

func (x *ServerUnary) ProtoReflect() protoreflect.Message {
	mi := &file_log_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerUnary.ProtoReflect.Descriptor instead.
func (*ServerUnary) Descriptor() ([]byte, []int) {
	return file_log_proto_rawDescGZIP(), []int{1}
}

func (x *ServerUnary) GetRequest() *anypb.Any {
	if x != nil {
		return x.Request
	}
	return nil
}

func (x *ServerUnary) GetResponse() *anypb.Any {
	if x != nil {
		return x.Response
	}
	return nil
}

func (x *ServerUnary) GetFullMethod() string {
	if x != nil {
		return x.FullMethod
	}
	return ""
}

func (x *ServerUnary) GetUser() string {
	if x != nil {
		return x.User
	}
	return ""
}

// ClientUnary logs the client side of a grpc unary request, as intercepted
// by https://pkg.go.dev/google.golang.org/grpc#UnaryClientInterceptor
type ClientUnary struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Method   string     `protobuf:"bytes,1,opt,name=method,proto3" json:"method,omitempty"`
	Request  *anypb.Any `protobuf:"bytes,2,opt,name=request,proto3" json:"request,omitempty"`
	Response *anypb.Any `protobuf:"bytes,3,opt,name=response,proto3" json:"response,omitempty"`
}

func (x *ClientUnary) Reset() {
	*x = ClientUnary{}
	if protoimpl.UnsafeEnabled {
		mi := &file_log_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientUnary) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientUnary) ProtoMessage() {}

func (x *ClientUnary) ProtoReflect() protoreflect.Message {
	mi := &file_log_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientUnary.ProtoReflect.Descriptor instead.
func (*ClientUnary) Descriptor() ([]byte, []int) {
	return file_log_proto_rawDescGZIP(), []int{2}
}

func (x *ClientUnary) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *ClientUnary) GetRequest() *anypb.Any {
	if x != nil {
		return x.Request
	}
	return nil
}

func (x *ClientUnary) GetResponse() *anypb.Any {
	if x != nil {
		return x.Response
	}
	return nil
}

// ClientStream logs the client side of a grpc stream request, as intercepted
// by https://pkg.go.dev/google.golang.org/grpc#StreamClientInterceptor
type ClientStream struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Method        string `protobuf:"bytes,1,opt,name=method,proto3" json:"method,omitempty"`
	StreamName    string `protobuf:"bytes,2,opt,name=stream_name,json=streamName,proto3" json:"stream_name,omitempty"`
	ServerStreams bool   `protobuf:"varint,3,opt,name=server_streams,json=serverStreams,proto3" json:"server_streams,omitempty"`
	ClientStreams bool   `protobuf:"varint,4,opt,name=client_streams,json=clientStreams,proto3" json:"client_streams,omitempty"`
}

func (x *ClientStream) Reset() {
	*x = ClientStream{}
	if protoimpl.UnsafeEnabled {
		mi := &file_log_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientStream) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientStream) ProtoMessage() {}

func (x *ClientStream) ProtoReflect() protoreflect.Message {
	mi := &file_log_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientStream.ProtoReflect.Descriptor instead.
func (*ClientStream) Descriptor() ([]byte, []int) {
	return file_log_proto_rawDescGZIP(), []int{3}
}

func (x *ClientStream) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *ClientStream) GetStreamName() string {
	if x != nil {
		return x.StreamName
	}
	return ""
}

func (x *ClientStream) GetServerStreams() bool {
	if x != nil {
		return x.ServerStreams
	}
	return false
}

func (x *ClientStream) GetClientStreams() bool {
	if x != nil {
		return x.ClientStreams
	}
	return false
}

var File_log_proto protoreflect.FileDescriptor

var file_log_proto_rawDesc = []byte{
	0x0a, 0x09, 0x6c, 0x6f, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x63, 0x61, 0x62,
	0x65, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6c, 0x6f, 0x67, 0x73, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc9, 0x03, 0x0a, 0x05, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x30, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x05, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x12, 0x33, 0x0a, 0x07, 0x65, 0x6c, 0x61, 0x70, 0x73, 0x65, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x07,
	0x65, 0x6c, 0x61, 0x70, 0x73, 0x65, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x25, 0x0a, 0x0e, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x3b, 0x0a, 0x0e, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c,
	0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x0d, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x12, 0x14, 0x0a, 0x05,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x12, 0x3d, 0x0a, 0x0c, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x75, 0x6e, 0x61,
	0x72, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x63, 0x61, 0x62, 0x65, 0x2e,
	0x67, 0x72, 0x70, 0x63, 0x6c, 0x6f, 0x67, 0x73, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x55,
	0x6e, 0x61, 0x72, 0x79, 0x52, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x55, 0x6e, 0x61, 0x72,
	0x79, 0x12, 0x3d, 0x0a, 0x0c, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x75, 0x6e, 0x61, 0x72,
	0x79, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x63, 0x61, 0x62, 0x65, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x6c, 0x6f, 0x67, 0x73, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x55, 0x6e,
	0x61, 0x72, 0x79, 0x52, 0x0b, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x55, 0x6e, 0x61, 0x72, 0x79,
	0x12, 0x40, 0x0a, 0x0d, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x73, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x63, 0x61, 0x62, 0x65, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x6c, 0x6f, 0x67, 0x73, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x52, 0x0c, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x22, 0xa4, 0x01, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x55, 0x6e, 0x61,
	0x72, 0x79, 0x12, 0x2e, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x30, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x08, 0x72, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x66, 0x75, 0x6c, 0x6c, 0x5f, 0x6d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x66, 0x75, 0x6c, 0x6c, 0x4d,
	0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x73, 0x65, 0x72, 0x22, 0x87, 0x01, 0x0a, 0x0b, 0x43, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x55, 0x6e, 0x61, 0x72, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f,
	0x64, 0x12, 0x2e, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x30, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x95, 0x01, 0x0a, 0x0c, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x1f, 0x0a, 0x0b,
	0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0a, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x25, 0x0a,
	0x0e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x53, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x73, 0x12, 0x25, 0x0a, 0x0e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x73,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x42, 0x2d, 0x5a, 0x2b, 0x67,
	0x6f, 0x2e, 0x73, 0x6b, 0x69, 0x61, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x69, 0x6e, 0x66, 0x72, 0x61,
	0x2f, 0x63, 0x61, 0x62, 0x65, 0x2f, 0x67, 0x6f, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x6c, 0x6f, 0x67,
	0x67, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_log_proto_rawDescOnce sync.Once
	file_log_proto_rawDescData = file_log_proto_rawDesc
)

func file_log_proto_rawDescGZIP() []byte {
	file_log_proto_rawDescOnce.Do(func() {
		file_log_proto_rawDescData = protoimpl.X.CompressGZIP(file_log_proto_rawDescData)
	})
	return file_log_proto_rawDescData
}

var file_log_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_log_proto_goTypes = []interface{}{
	(*Entry)(nil),                 // 0: cabe.grpclogs.Entry
	(*ServerUnary)(nil),           // 1: cabe.grpclogs.ServerUnary
	(*ClientUnary)(nil),           // 2: cabe.grpclogs.ClientUnary
	(*ClientStream)(nil),          // 3: cabe.grpclogs.ClientStream
	(*timestamppb.Timestamp)(nil), // 4: google.protobuf.Timestamp
	(*durationpb.Duration)(nil),   // 5: google.protobuf.Duration
	(*anypb.Any)(nil),             // 6: google.protobuf.Any
}
var file_log_proto_depIdxs = []int32{
	4,  // 0: cabe.grpclogs.Entry.start:type_name -> google.protobuf.Timestamp
	5,  // 1: cabe.grpclogs.Entry.elapsed:type_name -> google.protobuf.Duration
	6,  // 2: cabe.grpclogs.Entry.status_details:type_name -> google.protobuf.Any
	1,  // 3: cabe.grpclogs.Entry.server_unary:type_name -> cabe.grpclogs.ServerUnary
	2,  // 4: cabe.grpclogs.Entry.client_unary:type_name -> cabe.grpclogs.ClientUnary
	3,  // 5: cabe.grpclogs.Entry.client_stream:type_name -> cabe.grpclogs.ClientStream
	6,  // 6: cabe.grpclogs.ServerUnary.request:type_name -> google.protobuf.Any
	6,  // 7: cabe.grpclogs.ServerUnary.response:type_name -> google.protobuf.Any
	6,  // 8: cabe.grpclogs.ClientUnary.request:type_name -> google.protobuf.Any
	6,  // 9: cabe.grpclogs.ClientUnary.response:type_name -> google.protobuf.Any
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_log_proto_init() }
func file_log_proto_init() {
	if File_log_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_log_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Entry); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_log_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerUnary); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_log_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientUnary); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_log_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientStream); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_log_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_log_proto_goTypes,
		DependencyIndexes: file_log_proto_depIdxs,
		MessageInfos:      file_log_proto_msgTypes,
	}.Build()
	File_log_proto = out.File
	file_log_proto_rawDesc = nil
	file_log_proto_goTypes = nil
	file_log_proto_depIdxs = nil
}
