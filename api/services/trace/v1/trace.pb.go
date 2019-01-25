// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/containerd/containerd/api/services/trace/v1/trace.proto

/*
	Package trace is a generated protocol buffer package.

	It is generated from these files:
		github.com/containerd/containerd/api/services/trace/v1/trace.proto

	It has these top-level messages:
		KProbeConfig
		UProbeConfig
		ProbeRequest
		ProbeResponse
		UnloadRequest
*/
package trace

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/gogo/protobuf/types"

// skipping weak import gogoproto "github.com/gogo/protobuf/gogoproto"

import context "golang.org/x/net/context"
import grpc "google.golang.org/grpc"

import strings "strings"
import reflect "reflect"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type KProbeConfig struct {
	FunctionName string `protobuf:"bytes,1,opt,name=function_name,json=functionName,proto3" json:"function_name,omitempty"`
}

func (m *KProbeConfig) Reset()                    { *m = KProbeConfig{} }
func (*KProbeConfig) ProtoMessage()               {}
func (*KProbeConfig) Descriptor() ([]byte, []int) { return fileDescriptorTrace, []int{0} }

type UProbeConfig struct {
	Name   string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Symbol string `protobuf:"bytes,2,opt,name=symbol,proto3" json:"symbol,omitempty"`
}

func (m *UProbeConfig) Reset()                    { *m = UProbeConfig{} }
func (*UProbeConfig) ProtoMessage()               {}
func (*UProbeConfig) Descriptor() ([]byte, []int) { return fileDescriptorTrace, []int{1} }

type ProbeRequest struct {
	ID     string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Source string `protobuf:"bytes,2,opt,name=source,proto3" json:"source,omitempty"`
	// Types that are valid to be assigned to ProbeConfig:
	//	*ProbeRequest_KprobeConfig
	//	*ProbeRequest_UprobeConfig
	ProbeConfig        isProbeRequest_ProbeConfig `protobuf_oneof:"probe_config"`
	ProbeName          string                     `protobuf:"bytes,5,opt,name=probe_name,json=probeName,proto3" json:"probe_name,omitempty"`
	ReturnFunctionName string                     `protobuf:"bytes,7,opt,name=return_function_name,json=returnFunctionName,proto3" json:"return_function_name,omitempty"`
	TableID            string                     `protobuf:"bytes,8,opt,name=table_id,json=tableId,proto3" json:"table_id,omitempty"`
	ContainerID        string                     `protobuf:"bytes,9,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
}

func (m *ProbeRequest) Reset()                    { *m = ProbeRequest{} }
func (*ProbeRequest) ProtoMessage()               {}
func (*ProbeRequest) Descriptor() ([]byte, []int) { return fileDescriptorTrace, []int{2} }

type isProbeRequest_ProbeConfig interface {
	isProbeRequest_ProbeConfig()
	MarshalTo([]byte) (int, error)
	Size() int
}

type ProbeRequest_KprobeConfig struct {
	KprobeConfig *KProbeConfig `protobuf:"bytes,3,opt,name=kprobe_config,json=kprobeConfig,oneof"`
}
type ProbeRequest_UprobeConfig struct {
	UprobeConfig *UProbeConfig `protobuf:"bytes,4,opt,name=uprobe_config,json=uprobeConfig,oneof"`
}

func (*ProbeRequest_KprobeConfig) isProbeRequest_ProbeConfig() {}
func (*ProbeRequest_UprobeConfig) isProbeRequest_ProbeConfig() {}

func (m *ProbeRequest) GetProbeConfig() isProbeRequest_ProbeConfig {
	if m != nil {
		return m.ProbeConfig
	}
	return nil
}

func (m *ProbeRequest) GetKprobeConfig() *KProbeConfig {
	if x, ok := m.GetProbeConfig().(*ProbeRequest_KprobeConfig); ok {
		return x.KprobeConfig
	}
	return nil
}

func (m *ProbeRequest) GetUprobeConfig() *UProbeConfig {
	if x, ok := m.GetProbeConfig().(*ProbeRequest_UprobeConfig); ok {
		return x.UprobeConfig
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*ProbeRequest) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _ProbeRequest_OneofMarshaler, _ProbeRequest_OneofUnmarshaler, _ProbeRequest_OneofSizer, []interface{}{
		(*ProbeRequest_KprobeConfig)(nil),
		(*ProbeRequest_UprobeConfig)(nil),
	}
}

func _ProbeRequest_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*ProbeRequest)
	// probe_config
	switch x := m.ProbeConfig.(type) {
	case *ProbeRequest_KprobeConfig:
		_ = b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.KprobeConfig); err != nil {
			return err
		}
	case *ProbeRequest_UprobeConfig:
		_ = b.EncodeVarint(4<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.UprobeConfig); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("ProbeRequest.ProbeConfig has unexpected type %T", x)
	}
	return nil
}

func _ProbeRequest_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*ProbeRequest)
	switch tag {
	case 3: // probe_config.kprobe_config
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(KProbeConfig)
		err := b.DecodeMessage(msg)
		m.ProbeConfig = &ProbeRequest_KprobeConfig{msg}
		return true, err
	case 4: // probe_config.uprobe_config
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(UProbeConfig)
		err := b.DecodeMessage(msg)
		m.ProbeConfig = &ProbeRequest_UprobeConfig{msg}
		return true, err
	default:
		return false, nil
	}
}

func _ProbeRequest_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*ProbeRequest)
	// probe_config
	switch x := m.ProbeConfig.(type) {
	case *ProbeRequest_KprobeConfig:
		s := proto.Size(x.KprobeConfig)
		n += proto.SizeVarint(3<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *ProbeRequest_UprobeConfig:
		s := proto.Size(x.UprobeConfig)
		n += proto.SizeVarint(4<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type ProbeResponse struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *ProbeResponse) Reset()                    { *m = ProbeResponse{} }
func (*ProbeResponse) ProtoMessage()               {}
func (*ProbeResponse) Descriptor() ([]byte, []int) { return fileDescriptorTrace, []int{3} }

type UnloadRequest struct {
	ID string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (m *UnloadRequest) Reset()                    { *m = UnloadRequest{} }
func (*UnloadRequest) ProtoMessage()               {}
func (*UnloadRequest) Descriptor() ([]byte, []int) { return fileDescriptorTrace, []int{4} }

func init() {
	proto.RegisterType((*KProbeConfig)(nil), "containerd.services.trace.v1.KProbeConfig")
	proto.RegisterType((*UProbeConfig)(nil), "containerd.services.trace.v1.UProbeConfig")
	proto.RegisterType((*ProbeRequest)(nil), "containerd.services.trace.v1.ProbeRequest")
	proto.RegisterType((*ProbeResponse)(nil), "containerd.services.trace.v1.ProbeResponse")
	proto.RegisterType((*UnloadRequest)(nil), "containerd.services.trace.v1.UnloadRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Trace service

type TraceClient interface {
	Probe(ctx context.Context, in *ProbeRequest, opts ...grpc.CallOption) (Trace_ProbeClient, error)
	Unload(ctx context.Context, in *UnloadRequest, opts ...grpc.CallOption) (*google_protobuf.Empty, error)
}

type traceClient struct {
	cc *grpc.ClientConn
}

func NewTraceClient(cc *grpc.ClientConn) TraceClient {
	return &traceClient{cc}
}

func (c *traceClient) Probe(ctx context.Context, in *ProbeRequest, opts ...grpc.CallOption) (Trace_ProbeClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Trace_serviceDesc.Streams[0], c.cc, "/containerd.services.trace.v1.Trace/Probe", opts...)
	if err != nil {
		return nil, err
	}
	x := &traceProbeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Trace_ProbeClient interface {
	Recv() (*ProbeResponse, error)
	grpc.ClientStream
}

type traceProbeClient struct {
	grpc.ClientStream
}

func (x *traceProbeClient) Recv() (*ProbeResponse, error) {
	m := new(ProbeResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *traceClient) Unload(ctx context.Context, in *UnloadRequest, opts ...grpc.CallOption) (*google_protobuf.Empty, error) {
	out := new(google_protobuf.Empty)
	err := grpc.Invoke(ctx, "/containerd.services.trace.v1.Trace/Unload", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Trace service

type TraceServer interface {
	Probe(*ProbeRequest, Trace_ProbeServer) error
	Unload(context.Context, *UnloadRequest) (*google_protobuf.Empty, error)
}

func RegisterTraceServer(s *grpc.Server, srv TraceServer) {
	s.RegisterService(&_Trace_serviceDesc, srv)
}

func _Trace_Probe_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ProbeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TraceServer).Probe(m, &traceProbeServer{stream})
}

type Trace_ProbeServer interface {
	Send(*ProbeResponse) error
	grpc.ServerStream
}

type traceProbeServer struct {
	grpc.ServerStream
}

func (x *traceProbeServer) Send(m *ProbeResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Trace_Unload_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnloadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceServer).Unload(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/containerd.services.trace.v1.Trace/Unload",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceServer).Unload(ctx, req.(*UnloadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Trace_serviceDesc = grpc.ServiceDesc{
	ServiceName: "containerd.services.trace.v1.Trace",
	HandlerType: (*TraceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Unload",
			Handler:    _Trace_Unload_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Probe",
			Handler:       _Trace_Probe_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "github.com/containerd/containerd/api/services/trace/v1/trace.proto",
}

func (m *KProbeConfig) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *KProbeConfig) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.FunctionName) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintTrace(dAtA, i, uint64(len(m.FunctionName)))
		i += copy(dAtA[i:], m.FunctionName)
	}
	return i, nil
}

func (m *UProbeConfig) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *UProbeConfig) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintTrace(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if len(m.Symbol) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintTrace(dAtA, i, uint64(len(m.Symbol)))
		i += copy(dAtA[i:], m.Symbol)
	}
	return i, nil
}

func (m *ProbeRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ProbeRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.ID) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintTrace(dAtA, i, uint64(len(m.ID)))
		i += copy(dAtA[i:], m.ID)
	}
	if len(m.Source) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintTrace(dAtA, i, uint64(len(m.Source)))
		i += copy(dAtA[i:], m.Source)
	}
	if m.ProbeConfig != nil {
		nn1, err := m.ProbeConfig.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += nn1
	}
	if len(m.ProbeName) > 0 {
		dAtA[i] = 0x2a
		i++
		i = encodeVarintTrace(dAtA, i, uint64(len(m.ProbeName)))
		i += copy(dAtA[i:], m.ProbeName)
	}
	if len(m.ReturnFunctionName) > 0 {
		dAtA[i] = 0x3a
		i++
		i = encodeVarintTrace(dAtA, i, uint64(len(m.ReturnFunctionName)))
		i += copy(dAtA[i:], m.ReturnFunctionName)
	}
	if len(m.TableID) > 0 {
		dAtA[i] = 0x42
		i++
		i = encodeVarintTrace(dAtA, i, uint64(len(m.TableID)))
		i += copy(dAtA[i:], m.TableID)
	}
	if len(m.ContainerID) > 0 {
		dAtA[i] = 0x4a
		i++
		i = encodeVarintTrace(dAtA, i, uint64(len(m.ContainerID)))
		i += copy(dAtA[i:], m.ContainerID)
	}
	return i, nil
}

func (m *ProbeRequest_KprobeConfig) MarshalTo(dAtA []byte) (int, error) {
	i := 0
	if m.KprobeConfig != nil {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintTrace(dAtA, i, uint64(m.KprobeConfig.Size()))
		n2, err := m.KprobeConfig.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	return i, nil
}
func (m *ProbeRequest_UprobeConfig) MarshalTo(dAtA []byte) (int, error) {
	i := 0
	if m.UprobeConfig != nil {
		dAtA[i] = 0x22
		i++
		i = encodeVarintTrace(dAtA, i, uint64(m.UprobeConfig.Size()))
		n3, err := m.UprobeConfig.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n3
	}
	return i, nil
}
func (m *ProbeResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ProbeResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintTrace(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func (m *UnloadRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *UnloadRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.ID) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintTrace(dAtA, i, uint64(len(m.ID)))
		i += copy(dAtA[i:], m.ID)
	}
	return i, nil
}

func encodeVarintTrace(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *KProbeConfig) Size() (n int) {
	var l int
	_ = l
	l = len(m.FunctionName)
	if l > 0 {
		n += 1 + l + sovTrace(uint64(l))
	}
	return n
}

func (m *UProbeConfig) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovTrace(uint64(l))
	}
	l = len(m.Symbol)
	if l > 0 {
		n += 1 + l + sovTrace(uint64(l))
	}
	return n
}

func (m *ProbeRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.ID)
	if l > 0 {
		n += 1 + l + sovTrace(uint64(l))
	}
	l = len(m.Source)
	if l > 0 {
		n += 1 + l + sovTrace(uint64(l))
	}
	if m.ProbeConfig != nil {
		n += m.ProbeConfig.Size()
	}
	l = len(m.ProbeName)
	if l > 0 {
		n += 1 + l + sovTrace(uint64(l))
	}
	l = len(m.ReturnFunctionName)
	if l > 0 {
		n += 1 + l + sovTrace(uint64(l))
	}
	l = len(m.TableID)
	if l > 0 {
		n += 1 + l + sovTrace(uint64(l))
	}
	l = len(m.ContainerID)
	if l > 0 {
		n += 1 + l + sovTrace(uint64(l))
	}
	return n
}

func (m *ProbeRequest_KprobeConfig) Size() (n int) {
	var l int
	_ = l
	if m.KprobeConfig != nil {
		l = m.KprobeConfig.Size()
		n += 1 + l + sovTrace(uint64(l))
	}
	return n
}
func (m *ProbeRequest_UprobeConfig) Size() (n int) {
	var l int
	_ = l
	if m.UprobeConfig != nil {
		l = m.UprobeConfig.Size()
		n += 1 + l + sovTrace(uint64(l))
	}
	return n
}
func (m *ProbeResponse) Size() (n int) {
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovTrace(uint64(l))
	}
	return n
}

func (m *UnloadRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.ID)
	if l > 0 {
		n += 1 + l + sovTrace(uint64(l))
	}
	return n
}

func sovTrace(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozTrace(x uint64) (n int) {
	return sovTrace(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *KProbeConfig) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&KProbeConfig{`,
		`FunctionName:` + fmt.Sprintf("%v", this.FunctionName) + `,`,
		`}`,
	}, "")
	return s
}
func (this *UProbeConfig) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&UProbeConfig{`,
		`Name:` + fmt.Sprintf("%v", this.Name) + `,`,
		`Symbol:` + fmt.Sprintf("%v", this.Symbol) + `,`,
		`}`,
	}, "")
	return s
}
func (this *ProbeRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&ProbeRequest{`,
		`ID:` + fmt.Sprintf("%v", this.ID) + `,`,
		`Source:` + fmt.Sprintf("%v", this.Source) + `,`,
		`ProbeConfig:` + fmt.Sprintf("%v", this.ProbeConfig) + `,`,
		`ProbeName:` + fmt.Sprintf("%v", this.ProbeName) + `,`,
		`ReturnFunctionName:` + fmt.Sprintf("%v", this.ReturnFunctionName) + `,`,
		`TableID:` + fmt.Sprintf("%v", this.TableID) + `,`,
		`ContainerID:` + fmt.Sprintf("%v", this.ContainerID) + `,`,
		`}`,
	}, "")
	return s
}
func (this *ProbeRequest_KprobeConfig) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&ProbeRequest_KprobeConfig{`,
		`KprobeConfig:` + strings.Replace(fmt.Sprintf("%v", this.KprobeConfig), "KProbeConfig", "KProbeConfig", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *ProbeRequest_UprobeConfig) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&ProbeRequest_UprobeConfig{`,
		`UprobeConfig:` + strings.Replace(fmt.Sprintf("%v", this.UprobeConfig), "UProbeConfig", "UProbeConfig", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *ProbeResponse) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&ProbeResponse{`,
		`Data:` + fmt.Sprintf("%v", this.Data) + `,`,
		`}`,
	}, "")
	return s
}
func (this *UnloadRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&UnloadRequest{`,
		`ID:` + fmt.Sprintf("%v", this.ID) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringTrace(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *KProbeConfig) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTrace
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: KProbeConfig: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: KProbeConfig: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FunctionName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FunctionName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTrace(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTrace
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *UProbeConfig) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTrace
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: UProbeConfig: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: UProbeConfig: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Symbol", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Symbol = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTrace(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTrace
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ProbeRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTrace
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ProbeRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ProbeRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Source", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Source = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field KprobeConfig", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &KProbeConfig{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.ProbeConfig = &ProbeRequest_KprobeConfig{v}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UprobeConfig", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &UProbeConfig{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.ProbeConfig = &ProbeRequest_UprobeConfig{v}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProbeName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ProbeName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReturnFunctionName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ReturnFunctionName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TableID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TableID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContainerID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContainerID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTrace(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTrace
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ProbeResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTrace
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ProbeResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ProbeResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTrace(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTrace
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *UnloadRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTrace
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: UnloadRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: UnloadRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTrace
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTrace(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTrace
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTrace(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTrace
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTrace
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthTrace
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowTrace
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipTrace(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthTrace = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTrace   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/containerd/containerd/api/services/trace/v1/trace.proto", fileDescriptorTrace)
}

var fileDescriptorTrace = []byte{
	// 513 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x53, 0x4d, 0x6f, 0xd3, 0x40,
	0x10, 0xed, 0xa6, 0x4d, 0xd2, 0x4c, 0x6c, 0x90, 0x56, 0x55, 0x14, 0x05, 0x70, 0x2a, 0x57, 0x82,
	0x0a, 0x24, 0xbb, 0x4d, 0x6f, 0xc0, 0x29, 0x0d, 0x08, 0x0b, 0x81, 0xa8, 0xd5, 0x20, 0xc4, 0x25,
	0xf2, 0xc7, 0xc6, 0x58, 0xc4, 0x5e, 0x63, 0xaf, 0x23, 0xf5, 0xc6, 0x5f, 0xe1, 0x6f, 0xf0, 0x0b,
	0x7a, 0xe4, 0xc8, 0x29, 0xa2, 0xfe, 0x25, 0xc8, 0xb3, 0x49, 0x70, 0x38, 0x90, 0x8a, 0xdb, 0xf3,
	0xdb, 0x37, 0x6f, 0x67, 0xdf, 0x78, 0x60, 0x18, 0x84, 0xe2, 0x53, 0xee, 0x1a, 0x1e, 0x8f, 0x4c,
	0x8f, 0xc7, 0xc2, 0x09, 0x63, 0x96, 0xfa, 0x55, 0xe8, 0x24, 0xa1, 0x99, 0xb1, 0x74, 0x1e, 0x7a,
	0x2c, 0x33, 0x45, 0xea, 0x78, 0xcc, 0x9c, 0x9f, 0x4a, 0x60, 0x24, 0x29, 0x17, 0x9c, 0xde, 0xff,
	0xa3, 0x36, 0x56, 0x4a, 0x43, 0x0a, 0xe6, 0xa7, 0xbd, 0x7b, 0x01, 0xe7, 0xc1, 0x8c, 0x99, 0xa8,
	0x75, 0xf3, 0xa9, 0xc9, 0xa2, 0x44, 0x5c, 0xc9, 0xd2, 0xde, 0x41, 0xc0, 0x03, 0x8e, 0xd0, 0x2c,
	0x91, 0x64, 0xf5, 0x33, 0x50, 0x5e, 0xbf, 0x4b, 0xb9, 0xcb, 0xce, 0x79, 0x3c, 0x0d, 0x03, 0x7a,
	0x04, 0xea, 0x34, 0x8f, 0x3d, 0x11, 0xf2, 0x78, 0x12, 0x3b, 0x11, 0xeb, 0x92, 0x43, 0x72, 0xdc,
	0xb2, 0x95, 0x15, 0xf9, 0xd6, 0x89, 0x98, 0xfe, 0x14, 0x94, 0x71, 0xb5, 0x88, 0xc2, 0x5e, 0x45,
	0x8b, 0x98, 0x76, 0xa0, 0x91, 0x5d, 0x45, 0x2e, 0x9f, 0x75, 0x6b, 0xc8, 0x2e, 0xbf, 0xf4, 0x6f,
	0xbb, 0xa0, 0x60, 0xad, 0xcd, 0xbe, 0xe4, 0x2c, 0x13, 0xb4, 0x03, 0xb5, 0xd0, 0x97, 0xa5, 0xc3,
	0x46, 0xb1, 0xe8, 0xd7, 0xac, 0x91, 0x5d, 0x0b, 0x7d, 0x34, 0xe0, 0x79, 0xea, 0xb1, 0xb5, 0x01,
	0x7e, 0xd1, 0x0b, 0x50, 0x3f, 0x27, 0xa5, 0xc1, 0xc4, 0xc3, 0xdb, 0xbb, 0xbb, 0x87, 0xe4, 0xb8,
	0x3d, 0x78, 0x6c, 0xfc, 0x2b, 0x1a, 0xa3, 0xfa, 0xc8, 0x57, 0x3b, 0xb6, 0x22, 0x2d, 0x96, 0xfd,
	0x5f, 0x80, 0x9a, 0x6f, 0x58, 0xee, 0xdd, 0xc6, 0x72, 0xfc, 0x97, 0x65, 0x5e, 0xb5, 0x7c, 0x00,
	0x20, 0x1d, 0x31, 0x98, 0x3a, 0xbe, 0xa0, 0x85, 0x4c, 0x99, 0x20, 0x3d, 0x81, 0x83, 0x94, 0x89,
	0x3c, 0x8d, 0x27, 0x9b, 0x69, 0x37, 0x51, 0x48, 0xe5, 0xd9, 0xcb, 0x4a, 0xe6, 0xf4, 0x21, 0xec,
	0x0b, 0xc7, 0x9d, 0xb1, 0x49, 0xe8, 0x77, 0xf7, 0x31, 0xac, 0x76, 0xb1, 0xe8, 0x37, 0x2f, 0x4b,
	0xce, 0x1a, 0xd9, 0x4d, 0x3c, 0xb4, 0x7c, 0x3a, 0x00, 0x65, 0xdd, 0x75, 0xa9, 0x6d, 0xa1, 0xf6,
	0x6e, 0xb1, 0xe8, 0xb7, 0xcf, 0x57, 0xbc, 0x35, 0xb2, 0xdb, 0x6b, 0x91, 0xe5, 0x0f, 0xef, 0x80,
	0x52, 0x7d, 0xbe, 0x7e, 0x04, 0xea, 0x72, 0x44, 0x59, 0xc2, 0xe3, 0x8c, 0x95, 0x03, 0xf6, 0x1d,
	0xe1, 0xe0, 0x94, 0x14, 0x1b, 0xb1, 0xfe, 0x08, 0xd4, 0x71, 0x3c, 0xe3, 0x8e, 0xbf, 0x65, 0x90,
	0x83, 0xef, 0x04, 0xea, 0x97, 0x65, 0x68, 0xd4, 0x85, 0x3a, 0xfa, 0xd2, 0x2d, 0xc9, 0x56, 0xff,
	0x8f, 0xde, 0x93, 0x5b, 0x69, 0x65, 0xa3, 0x27, 0x84, 0xbe, 0x81, 0x86, 0x6c, 0x8b, 0x6e, 0x29,
	0xdc, 0x68, 0xbe, 0xd7, 0x31, 0xe4, 0xee, 0x18, 0xab, 0xdd, 0x31, 0x5e, 0x94, 0xbb, 0x33, 0x7c,
	0x7f, 0x7d, 0xa3, 0xed, 0xfc, 0xbc, 0xd1, 0x76, 0xbe, 0x16, 0x1a, 0xb9, 0x2e, 0x34, 0xf2, 0xa3,
	0xd0, 0xc8, 0xaf, 0x42, 0x23, 0x1f, 0x9f, 0xff, 0xdf, 0x3a, 0x3f, 0x43, 0xf0, 0x81, 0xb8, 0x0d,
	0xbc, 0xe9, 0xec, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0xbb, 0x3e, 0xec, 0x16, 0x17, 0x04, 0x00,
	0x00,
}
