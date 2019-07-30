// Code generated by protoc-gen-go. DO NOT EDIT.
// source: opencensus/proto/agent/common/v1/common.proto

package v1

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type LibraryInfo_Language int32

const (
	LibraryInfo_LANGUAGE_UNSPECIFIED LibraryInfo_Language = 0
	LibraryInfo_CPP                  LibraryInfo_Language = 1
	LibraryInfo_C_SHARP              LibraryInfo_Language = 2
	LibraryInfo_ERLANG               LibraryInfo_Language = 3
	LibraryInfo_GO_LANG              LibraryInfo_Language = 4
	LibraryInfo_JAVA                 LibraryInfo_Language = 5
	LibraryInfo_NODE_JS              LibraryInfo_Language = 6
	LibraryInfo_PHP                  LibraryInfo_Language = 7
	LibraryInfo_PYTHON               LibraryInfo_Language = 8
	LibraryInfo_RUBY                 LibraryInfo_Language = 9
)

var LibraryInfo_Language_name = map[int32]string{
	0: "LANGUAGE_UNSPECIFIED",
	1: "CPP",
	2: "C_SHARP",
	3: "ERLANG",
	4: "GO_LANG",
	5: "JAVA",
	6: "NODE_JS",
	7: "PHP",
	8: "PYTHON",
	9: "RUBY",
}

var LibraryInfo_Language_value = map[string]int32{
	"LANGUAGE_UNSPECIFIED": 0,
	"CPP":     1,
	"C_SHARP": 2,
	"ERLANG":  3,
	"GO_LANG": 4,
	"JAVA":    5,
	"NODE_JS": 6,
	"PHP":     7,
	"PYTHON":  8,
	"RUBY":    9,
}

func (x LibraryInfo_Language) String() string {
	return proto.EnumName(LibraryInfo_Language_name, int32(x))
}

func (LibraryInfo_Language) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_126c72ed8a252c84, []int{2, 0}
}

// Identifier metadata of the Node that produces the span or tracing data.
// Note, this is not the metadata about the Node or service that is described by associated spans.
// In the future we plan to extend the identifier proto definition to support
// additional information (e.g cloud id, etc.)
type Node struct {
	// Identifier that uniquely identifies a process within a VM/container.
	Identifier *ProcessIdentifier `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	// Information on the OpenCensus Library that initiates the stream.
	LibraryInfo *LibraryInfo `protobuf:"bytes,2,opt,name=library_info,json=libraryInfo,proto3" json:"library_info,omitempty"`
	// Additional information on service.
	ServiceInfo *ServiceInfo `protobuf:"bytes,3,opt,name=service_info,json=serviceInfo,proto3" json:"service_info,omitempty"`
	// Additional attributes.
	Attributes           map[string]string `protobuf:"bytes,4,rep,name=attributes,proto3" json:"attributes,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Node) Reset()         { *m = Node{} }
func (m *Node) String() string { return proto.CompactTextString(m) }
func (*Node) ProtoMessage()    {}
func (*Node) Descriptor() ([]byte, []int) {
	return fileDescriptor_126c72ed8a252c84, []int{0}
}

func (m *Node) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Node.Unmarshal(m, b)
}
func (m *Node) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Node.Marshal(b, m, deterministic)
}
func (m *Node) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Node.Merge(m, src)
}
func (m *Node) XXX_Size() int {
	return xxx_messageInfo_Node.Size(m)
}
func (m *Node) XXX_DiscardUnknown() {
	xxx_messageInfo_Node.DiscardUnknown(m)
}

var xxx_messageInfo_Node proto.InternalMessageInfo

func (m *Node) GetIdentifier() *ProcessIdentifier {
	if m != nil {
		return m.Identifier
	}
	return nil
}

func (m *Node) GetLibraryInfo() *LibraryInfo {
	if m != nil {
		return m.LibraryInfo
	}
	return nil
}

func (m *Node) GetServiceInfo() *ServiceInfo {
	if m != nil {
		return m.ServiceInfo
	}
	return nil
}

func (m *Node) GetAttributes() map[string]string {
	if m != nil {
		return m.Attributes
	}
	return nil
}

// Identifier that uniquely identifies a process within a VM/container.
type ProcessIdentifier struct {
	// The host name. Usually refers to the machine/container name.
	// For example: os.Hostname() in Go, socket.gethostname() in Python.
	HostName string `protobuf:"bytes,1,opt,name=host_name,json=hostName,proto3" json:"host_name,omitempty"`
	// Process id.
	Pid uint32 `protobuf:"varint,2,opt,name=pid,proto3" json:"pid,omitempty"`
	// Start time of this ProcessIdentifier. Represented in epoch time.
	StartTimestamp       *timestamp.Timestamp `protobuf:"bytes,3,opt,name=start_timestamp,json=startTimestamp,proto3" json:"start_timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *ProcessIdentifier) Reset()         { *m = ProcessIdentifier{} }
func (m *ProcessIdentifier) String() string { return proto.CompactTextString(m) }
func (*ProcessIdentifier) ProtoMessage()    {}
func (*ProcessIdentifier) Descriptor() ([]byte, []int) {
	return fileDescriptor_126c72ed8a252c84, []int{1}
}

func (m *ProcessIdentifier) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProcessIdentifier.Unmarshal(m, b)
}
func (m *ProcessIdentifier) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProcessIdentifier.Marshal(b, m, deterministic)
}
func (m *ProcessIdentifier) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProcessIdentifier.Merge(m, src)
}
func (m *ProcessIdentifier) XXX_Size() int {
	return xxx_messageInfo_ProcessIdentifier.Size(m)
}
func (m *ProcessIdentifier) XXX_DiscardUnknown() {
	xxx_messageInfo_ProcessIdentifier.DiscardUnknown(m)
}

var xxx_messageInfo_ProcessIdentifier proto.InternalMessageInfo

func (m *ProcessIdentifier) GetHostName() string {
	if m != nil {
		return m.HostName
	}
	return ""
}

func (m *ProcessIdentifier) GetPid() uint32 {
	if m != nil {
		return m.Pid
	}
	return 0
}

func (m *ProcessIdentifier) GetStartTimestamp() *timestamp.Timestamp {
	if m != nil {
		return m.StartTimestamp
	}
	return nil
}

// Information on OpenCensus Library.
type LibraryInfo struct {
	// Language of OpenCensus Library.
	Language LibraryInfo_Language `protobuf:"varint,1,opt,name=language,proto3,enum=opencensus.proto.agent.common.v1.LibraryInfo_Language" json:"language,omitempty"`
	// Version of Agent exporter of Library.
	ExporterVersion string `protobuf:"bytes,2,opt,name=exporter_version,json=exporterVersion,proto3" json:"exporter_version,omitempty"`
	// Version of OpenCensus Library.
	CoreLibraryVersion   string   `protobuf:"bytes,3,opt,name=core_library_version,json=coreLibraryVersion,proto3" json:"core_library_version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LibraryInfo) Reset()         { *m = LibraryInfo{} }
func (m *LibraryInfo) String() string { return proto.CompactTextString(m) }
func (*LibraryInfo) ProtoMessage()    {}
func (*LibraryInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_126c72ed8a252c84, []int{2}
}

func (m *LibraryInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LibraryInfo.Unmarshal(m, b)
}
func (m *LibraryInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LibraryInfo.Marshal(b, m, deterministic)
}
func (m *LibraryInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LibraryInfo.Merge(m, src)
}
func (m *LibraryInfo) XXX_Size() int {
	return xxx_messageInfo_LibraryInfo.Size(m)
}
func (m *LibraryInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_LibraryInfo.DiscardUnknown(m)
}

var xxx_messageInfo_LibraryInfo proto.InternalMessageInfo

func (m *LibraryInfo) GetLanguage() LibraryInfo_Language {
	if m != nil {
		return m.Language
	}
	return LibraryInfo_LANGUAGE_UNSPECIFIED
}

func (m *LibraryInfo) GetExporterVersion() string {
	if m != nil {
		return m.ExporterVersion
	}
	return ""
}

func (m *LibraryInfo) GetCoreLibraryVersion() string {
	if m != nil {
		return m.CoreLibraryVersion
	}
	return ""
}

// Additional service information.
type ServiceInfo struct {
	// Name of the service.
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ServiceInfo) Reset()         { *m = ServiceInfo{} }
func (m *ServiceInfo) String() string { return proto.CompactTextString(m) }
func (*ServiceInfo) ProtoMessage()    {}
func (*ServiceInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_126c72ed8a252c84, []int{3}
}

func (m *ServiceInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceInfo.Unmarshal(m, b)
}
func (m *ServiceInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceInfo.Marshal(b, m, deterministic)
}
func (m *ServiceInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceInfo.Merge(m, src)
}
func (m *ServiceInfo) XXX_Size() int {
	return xxx_messageInfo_ServiceInfo.Size(m)
}
func (m *ServiceInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceInfo proto.InternalMessageInfo

func (m *ServiceInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func init() {
	proto.RegisterEnum("opencensus.proto.agent.common.v1.LibraryInfo_Language", LibraryInfo_Language_name, LibraryInfo_Language_value)
	proto.RegisterType((*Node)(nil), "opencensus.proto.agent.common.v1.Node")
	proto.RegisterMapType((map[string]string)(nil), "opencensus.proto.agent.common.v1.Node.AttributesEntry")
	proto.RegisterType((*ProcessIdentifier)(nil), "opencensus.proto.agent.common.v1.ProcessIdentifier")
	proto.RegisterType((*LibraryInfo)(nil), "opencensus.proto.agent.common.v1.LibraryInfo")
	proto.RegisterType((*ServiceInfo)(nil), "opencensus.proto.agent.common.v1.ServiceInfo")
}

func init() {
	proto.RegisterFile("opencensus/proto/agent/common/v1/common.proto", fileDescriptor_126c72ed8a252c84)
}

var fileDescriptor_126c72ed8a252c84 = []byte{
	// 590 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x94, 0x4f, 0x4f, 0xdb, 0x3e,
	0x1c, 0xc6, 0x7f, 0x69, 0x0a, 0xb4, 0xdf, 0xfc, 0x06, 0x99, 0xc5, 0xa1, 0x62, 0x87, 0xb1, 0xee,
	0xc2, 0x0e, 0x4d, 0x06, 0x48, 0xd3, 0x34, 0x69, 0x87, 0x52, 0x3a, 0x28, 0x42, 0x25, 0x72, 0x01,
	0x89, 0x5d, 0xa2, 0xb4, 0xb8, 0xc1, 0x5a, 0x63, 0x57, 0xb6, 0x53, 0x8d, 0xd3, 0x8e, 0xd3, 0xde,
	0xc0, 0x5e, 0xd4, 0x5e, 0xd5, 0x64, 0x3b, 0x69, 0xa3, 0x71, 0x28, 0xb7, 0xef, 0x9f, 0xe7, 0xf9,
	0x38, 0x7a, 0x6c, 0x05, 0x3a, 0x7c, 0x4e, 0xd8, 0x84, 0x30, 0x99, 0xcb, 0x70, 0x2e, 0xb8, 0xe2,
	0x61, 0x92, 0x12, 0xa6, 0xc2, 0x09, 0xcf, 0x32, 0xce, 0xc2, 0xc5, 0x61, 0x51, 0x05, 0x66, 0x89,
	0xf6, 0x57, 0x72, 0x3b, 0x09, 0x8c, 0x3c, 0x28, 0x44, 0x8b, 0xc3, 0xbd, 0xd7, 0x29, 0xe7, 0xe9,
	0x8c, 0x58, 0xd8, 0x38, 0x9f, 0x86, 0x8a, 0x66, 0x44, 0xaa, 0x24, 0x9b, 0x5b, 0x43, 0xfb, 0xb7,
	0x0b, 0xf5, 0x21, 0xbf, 0x27, 0x68, 0x04, 0x40, 0xef, 0x09, 0x53, 0x74, 0x4a, 0x89, 0x68, 0x39,
	0xfb, 0xce, 0x81, 0x77, 0x74, 0x1c, 0xac, 0x3b, 0x20, 0x88, 0x04, 0x9f, 0x10, 0x29, 0x07, 0x4b,
	0x2b, 0xae, 0x60, 0x50, 0x04, 0xff, 0xcf, 0xe8, 0x58, 0x24, 0xe2, 0x31, 0xa6, 0x6c, 0xca, 0x5b,
	0x35, 0x83, 0xed, 0xac, 0xc7, 0x5e, 0x5a, 0xd7, 0x80, 0x4d, 0x39, 0xf6, 0x66, 0xab, 0x46, 0x13,
	0x25, 0x11, 0x0b, 0x3a, 0x21, 0x96, 0xe8, 0x3e, 0x97, 0x38, 0xb2, 0x2e, 0x4b, 0x94, 0xab, 0x06,
	0xdd, 0x02, 0x24, 0x4a, 0x09, 0x3a, 0xce, 0x15, 0x91, 0xad, 0xfa, 0xbe, 0x7b, 0xe0, 0x1d, 0x7d,
	0x58, 0xcf, 0xd3, 0xa1, 0x05, 0xdd, 0xa5, 0xb1, 0xcf, 0x94, 0x78, 0xc4, 0x15, 0xd2, 0xde, 0x67,
	0xd8, 0xf9, 0x67, 0x8d, 0x7c, 0x70, 0xbf, 0x91, 0x47, 0x13, 0x6e, 0x13, 0xeb, 0x12, 0xed, 0xc2,
	0xc6, 0x22, 0x99, 0xe5, 0xc4, 0x24, 0xd3, 0xc4, 0xb6, 0xf9, 0x54, 0xfb, 0xe8, 0xb4, 0x7f, 0x3a,
	0xf0, 0xf2, 0x49, 0xb8, 0xe8, 0x15, 0x34, 0x1f, 0xb8, 0x54, 0x31, 0x4b, 0x32, 0x52, 0x70, 0x1a,
	0x7a, 0x30, 0x4c, 0x32, 0xa2, 0xf1, 0x73, 0x7a, 0x6f, 0x50, 0x2f, 0xb0, 0x2e, 0x51, 0x0f, 0x76,
	0xa4, 0x4a, 0x84, 0x8a, 0x97, 0xd7, 0x5e, 0x04, 0xb6, 0x17, 0xd8, 0x87, 0x11, 0x94, 0x0f, 0x23,
	0xb8, 0x2e, 0x15, 0x78, 0xdb, 0x58, 0x96, 0x7d, 0xfb, 0x4f, 0x0d, 0xbc, 0xca, 0x7d, 0x20, 0x0c,
	0x8d, 0x59, 0xc2, 0xd2, 0x3c, 0x49, 0xed, 0x27, 0x6c, 0x3f, 0x27, 0xae, 0x0a, 0x20, 0xb8, 0x2c,
	0xdc, 0x78, 0xc9, 0x41, 0xef, 0xc0, 0x27, 0xdf, 0xe7, 0x5c, 0x28, 0x22, 0xe2, 0x05, 0x11, 0x92,
	0x72, 0x56, 0x44, 0xb2, 0x53, 0xce, 0x6f, 0xed, 0x18, 0xbd, 0x87, 0xdd, 0x09, 0x17, 0x24, 0x2e,
	0x1f, 0x56, 0x29, 0x77, 0x8d, 0x1c, 0xe9, 0x5d, 0x71, 0x58, 0xe1, 0x68, 0xff, 0x72, 0xa0, 0x51,
	0x9e, 0x89, 0x5a, 0xb0, 0x7b, 0xd9, 0x1d, 0x9e, 0xdd, 0x74, 0xcf, 0xfa, 0xf1, 0xcd, 0x70, 0x14,
	0xf5, 0x7b, 0x83, 0x2f, 0x83, 0xfe, 0xa9, 0xff, 0x1f, 0xda, 0x02, 0xb7, 0x17, 0x45, 0xbe, 0x83,
	0x3c, 0xd8, 0xea, 0xc5, 0xa3, 0xf3, 0x2e, 0x8e, 0xfc, 0x1a, 0x02, 0xd8, 0xec, 0x63, 0xed, 0xf0,
	0x5d, 0xbd, 0x38, 0xbb, 0x8a, 0x4d, 0x53, 0x47, 0x0d, 0xa8, 0x5f, 0x74, 0x6f, 0xbb, 0xfe, 0x86,
	0x1e, 0x0f, 0xaf, 0x4e, 0xfb, 0xf1, 0xc5, 0xc8, 0xdf, 0xd4, 0x94, 0xe8, 0x3c, 0xf2, 0xb7, 0xb4,
	0x31, 0xba, 0xbb, 0x3e, 0xbf, 0x1a, 0xfa, 0x0d, 0xad, 0xc5, 0x37, 0x27, 0x77, 0x7e, 0xb3, 0xfd,
	0x06, 0xbc, 0xca, 0x4b, 0x44, 0x08, 0xea, 0x95, 0xab, 0x34, 0xf5, 0xc9, 0x0f, 0x78, 0x4b, 0xf9,
	0xda, 0x44, 0x4f, 0xbc, 0x9e, 0x29, 0x23, 0xbd, 0x8c, 0x9c, 0xaf, 0x83, 0x94, 0xaa, 0x87, 0x7c,
	0xac, 0x05, 0xa1, 0xf5, 0x75, 0x28, 0x93, 0x4a, 0xe4, 0x19, 0x61, 0x2a, 0x51, 0x94, 0xb3, 0x70,
	0x85, 0xec, 0xd8, 0x9f, 0x4b, 0x4a, 0x58, 0x27, 0x7d, 0xf2, 0x8f, 0x19, 0x6f, 0x9a, 0xed, 0xf1,
	0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x94, 0xe5, 0x77, 0x76, 0x8e, 0x04, 0x00, 0x00,
}
