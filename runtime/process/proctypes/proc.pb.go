// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/containerd/containerd/runtime/process/proctypes/proc.proto

/*
	Package proctypes is a generated protocol buffer package.

	It is generated from these files:
		github.com/containerd/containerd/runtime/process/proctypes/proc.proto

	It has these top-level messages:
		CreateOptions
		ProcessDetails
*/
package proctypes

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

// skipping weak import gogoproto "github.com/gogo/protobuf/gogoproto"

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

type CreateOptions struct {
	Runtime     string `protobuf:"bytes,1,opt,name=runtime,proto3" json:"runtime,omitempty"`
	RuntimeRoot string `protobuf:"bytes,2,opt,name=runtime_root,json=runtimeRoot,proto3" json:"runtime_root,omitempty"`
}

func (m *CreateOptions) Reset()                    { *m = CreateOptions{} }
func (*CreateOptions) ProtoMessage()               {}
func (*CreateOptions) Descriptor() ([]byte, []int) { return fileDescriptorProc, []int{0} }

type ProcessDetails struct {
	ProcessID uint32 `protobuf:"varint,1,opt,name=process_id,json=processId,proto3" json:"process_id,omitempty"`
	ExecID    string `protobuf:"bytes,2,opt,name=exec_id,json=execId,proto3" json:"exec_id,omitempty"`
}

func (m *ProcessDetails) Reset()                    { *m = ProcessDetails{} }
func (*ProcessDetails) ProtoMessage()               {}
func (*ProcessDetails) Descriptor() ([]byte, []int) { return fileDescriptorProc, []int{1} }

func init() {
	proto.RegisterType((*CreateOptions)(nil), "containerd.process.proc.CreateOptions")
	proto.RegisterType((*ProcessDetails)(nil), "containerd.process.proc.ProcessDetails")
}
func (m *CreateOptions) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CreateOptions) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Runtime) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintProc(dAtA, i, uint64(len(m.Runtime)))
		i += copy(dAtA[i:], m.Runtime)
	}
	if len(m.RuntimeRoot) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintProc(dAtA, i, uint64(len(m.RuntimeRoot)))
		i += copy(dAtA[i:], m.RuntimeRoot)
	}
	return i, nil
}

func (m *ProcessDetails) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ProcessDetails) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.ProcessID != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintProc(dAtA, i, uint64(m.ProcessID))
	}
	if len(m.ExecID) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintProc(dAtA, i, uint64(len(m.ExecID)))
		i += copy(dAtA[i:], m.ExecID)
	}
	return i, nil
}

func encodeVarintProc(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *CreateOptions) Size() (n int) {
	var l int
	_ = l
	l = len(m.Runtime)
	if l > 0 {
		n += 1 + l + sovProc(uint64(l))
	}
	l = len(m.RuntimeRoot)
	if l > 0 {
		n += 1 + l + sovProc(uint64(l))
	}
	return n
}

func (m *ProcessDetails) Size() (n int) {
	var l int
	_ = l
	if m.ProcessID != 0 {
		n += 1 + sovProc(uint64(m.ProcessID))
	}
	l = len(m.ExecID)
	if l > 0 {
		n += 1 + l + sovProc(uint64(l))
	}
	return n
}

func sovProc(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozProc(x uint64) (n int) {
	return sovProc(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *CreateOptions) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&CreateOptions{`,
		`Runtime:` + fmt.Sprintf("%v", this.Runtime) + `,`,
		`RuntimeRoot:` + fmt.Sprintf("%v", this.RuntimeRoot) + `,`,
		`}`,
	}, "")
	return s
}
func (this *ProcessDetails) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&ProcessDetails{`,
		`ProcessID:` + fmt.Sprintf("%v", this.ProcessID) + `,`,
		`ExecID:` + fmt.Sprintf("%v", this.ExecID) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringProc(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *CreateOptions) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProc
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
			return fmt.Errorf("proto: CreateOptions: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CreateOptions: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Runtime", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProc
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
				return ErrInvalidLengthProc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Runtime = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RuntimeRoot", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProc
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
				return ErrInvalidLengthProc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RuntimeRoot = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProc(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProc
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
func (m *ProcessDetails) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProc
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
			return fmt.Errorf("proto: ProcessDetails: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ProcessDetails: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProcessID", wireType)
			}
			m.ProcessID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProcessID |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExecID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProc
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
				return ErrInvalidLengthProc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ExecID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProc(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProc
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
func skipProc(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProc
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
					return 0, ErrIntOverflowProc
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
					return 0, ErrIntOverflowProc
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
				return 0, ErrInvalidLengthProc
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowProc
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
				next, err := skipProc(dAtA[start:])
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
	ErrInvalidLengthProc = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProc   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/containerd/containerd/runtime/process/proctypes/proc.proto", fileDescriptorProc)
}

var fileDescriptorProc = []byte{
	// 267 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x72, 0x4d, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xce, 0xcf, 0x2b, 0x49, 0xcc, 0xcc, 0x4b, 0x2d,
	0x4a, 0x41, 0x66, 0x16, 0x95, 0xe6, 0x95, 0x64, 0xe6, 0xa6, 0xea, 0x17, 0x14, 0xe5, 0x27, 0xa7,
	0x16, 0x17, 0x83, 0xe9, 0x92, 0xca, 0x82, 0x54, 0x08, 0x4b, 0xaf, 0xa0, 0x28, 0xbf, 0x24, 0x5f,
	0x48, 0x1c, 0xa1, 0x41, 0x0f, 0xaa, 0x10, 0x4c, 0x4b, 0x89, 0xa4, 0xe7, 0xa7, 0xe7, 0x83, 0xd5,
	0xe8, 0x83, 0x58, 0x10, 0xe5, 0x4a, 0x3e, 0x5c, 0xbc, 0xce, 0x45, 0xa9, 0x89, 0x25, 0xa9, 0xfe,
	0x05, 0x25, 0x99, 0xf9, 0x79, 0xc5, 0x42, 0x12, 0x5c, 0xec, 0x50, 0x7b, 0x24, 0x18, 0x15, 0x18,
	0x35, 0x38, 0x83, 0x60, 0x5c, 0x21, 0x45, 0x2e, 0x1e, 0x28, 0x33, 0xbe, 0x28, 0x3f, 0xbf, 0x44,
	0x82, 0x09, 0x2c, 0xcd, 0x0d, 0x15, 0x0b, 0xca, 0xcf, 0x2f, 0x51, 0x4a, 0xe6, 0xe2, 0x0b, 0x80,
	0xd8, 0xe9, 0x92, 0x5a, 0x92, 0x98, 0x99, 0x53, 0x2c, 0xa4, 0xc3, 0xc5, 0x05, 0x75, 0x45, 0x7c,
	0x66, 0x0a, 0xd8, 0x44, 0x5e, 0x27, 0xde, 0x47, 0xf7, 0xe4, 0x39, 0xa1, 0xea, 0x3c, 0x5d, 0x82,
	0x38, 0xa1, 0x0a, 0x3c, 0x53, 0x84, 0x94, 0xb9, 0xd8, 0x53, 0x2b, 0x52, 0x93, 0x41, 0x4a, 0xc1,
	0xa6, 0x3b, 0x71, 0x3d, 0xba, 0x27, 0xcf, 0xe6, 0x5a, 0x91, 0x9a, 0xec, 0xe9, 0x12, 0xc4, 0x06,
	0x92, 0xf2, 0x4c, 0x71, 0x8a, 0x3b, 0xf1, 0x50, 0x8e, 0xe1, 0xc6, 0x43, 0x39, 0x86, 0x86, 0x47,
	0x72, 0x8c, 0x27, 0x1e, 0xc9, 0x31, 0x5e, 0x78, 0x24, 0xc7, 0xf8, 0xe0, 0x91, 0x1c, 0x63, 0x94,
	0x0b, 0xf9, 0x41, 0x68, 0x0d, 0x67, 0x45, 0x30, 0x24, 0xb1, 0x81, 0xc3, 0xc6, 0x18, 0x10, 0x00,
	0x00, 0xff, 0xff, 0xd7, 0x67, 0x9c, 0x7d, 0x93, 0x01, 0x00, 0x00,
}
