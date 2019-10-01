// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v1/services/keyword_plan_ad_group_service.proto

package services

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/wrappers"
	resources "google.golang.org/genproto/googleapis/ads/googleads/v1/resources"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	status "google.golang.org/genproto/googleapis/rpc/status"
	field_mask "google.golang.org/genproto/protobuf/field_mask"
	grpc "google.golang.org/grpc"
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

// Request message for
// [KeywordPlanAdGroupService.GetKeywordPlanAdGroup][google.ads.googleads.v1.services.KeywordPlanAdGroupService.GetKeywordPlanAdGroup].
type GetKeywordPlanAdGroupRequest struct {
	// The resource name of the Keyword Plan ad group to fetch.
	ResourceName         string   `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetKeywordPlanAdGroupRequest) Reset()         { *m = GetKeywordPlanAdGroupRequest{} }
func (m *GetKeywordPlanAdGroupRequest) String() string { return proto.CompactTextString(m) }
func (*GetKeywordPlanAdGroupRequest) ProtoMessage()    {}
func (*GetKeywordPlanAdGroupRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ccc3a69fa60910db, []int{0}
}

func (m *GetKeywordPlanAdGroupRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetKeywordPlanAdGroupRequest.Unmarshal(m, b)
}
func (m *GetKeywordPlanAdGroupRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetKeywordPlanAdGroupRequest.Marshal(b, m, deterministic)
}
func (m *GetKeywordPlanAdGroupRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetKeywordPlanAdGroupRequest.Merge(m, src)
}
func (m *GetKeywordPlanAdGroupRequest) XXX_Size() int {
	return xxx_messageInfo_GetKeywordPlanAdGroupRequest.Size(m)
}
func (m *GetKeywordPlanAdGroupRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetKeywordPlanAdGroupRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetKeywordPlanAdGroupRequest proto.InternalMessageInfo

func (m *GetKeywordPlanAdGroupRequest) GetResourceName() string {
	if m != nil {
		return m.ResourceName
	}
	return ""
}

// Request message for
// [KeywordPlanAdGroupService.MutateKeywordPlanAdGroups][google.ads.googleads.v1.services.KeywordPlanAdGroupService.MutateKeywordPlanAdGroups].
type MutateKeywordPlanAdGroupsRequest struct {
	// The ID of the customer whose Keyword Plan ad groups are being modified.
	CustomerId string `protobuf:"bytes,1,opt,name=customer_id,json=customerId,proto3" json:"customer_id,omitempty"`
	// The list of operations to perform on individual Keyword Plan ad groups.
	Operations []*KeywordPlanAdGroupOperation `protobuf:"bytes,2,rep,name=operations,proto3" json:"operations,omitempty"`
	// If true, successful operations will be carried out and invalid
	// operations will return errors. If false, all operations will be carried
	// out in one transaction if and only if they are all valid.
	// Default is false.
	PartialFailure bool `protobuf:"varint,3,opt,name=partial_failure,json=partialFailure,proto3" json:"partial_failure,omitempty"`
	// If true, the request is validated but not executed. Only errors are
	// returned, not results.
	ValidateOnly         bool     `protobuf:"varint,4,opt,name=validate_only,json=validateOnly,proto3" json:"validate_only,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MutateKeywordPlanAdGroupsRequest) Reset()         { *m = MutateKeywordPlanAdGroupsRequest{} }
func (m *MutateKeywordPlanAdGroupsRequest) String() string { return proto.CompactTextString(m) }
func (*MutateKeywordPlanAdGroupsRequest) ProtoMessage()    {}
func (*MutateKeywordPlanAdGroupsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ccc3a69fa60910db, []int{1}
}

func (m *MutateKeywordPlanAdGroupsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MutateKeywordPlanAdGroupsRequest.Unmarshal(m, b)
}
func (m *MutateKeywordPlanAdGroupsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MutateKeywordPlanAdGroupsRequest.Marshal(b, m, deterministic)
}
func (m *MutateKeywordPlanAdGroupsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MutateKeywordPlanAdGroupsRequest.Merge(m, src)
}
func (m *MutateKeywordPlanAdGroupsRequest) XXX_Size() int {
	return xxx_messageInfo_MutateKeywordPlanAdGroupsRequest.Size(m)
}
func (m *MutateKeywordPlanAdGroupsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MutateKeywordPlanAdGroupsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MutateKeywordPlanAdGroupsRequest proto.InternalMessageInfo

func (m *MutateKeywordPlanAdGroupsRequest) GetCustomerId() string {
	if m != nil {
		return m.CustomerId
	}
	return ""
}

func (m *MutateKeywordPlanAdGroupsRequest) GetOperations() []*KeywordPlanAdGroupOperation {
	if m != nil {
		return m.Operations
	}
	return nil
}

func (m *MutateKeywordPlanAdGroupsRequest) GetPartialFailure() bool {
	if m != nil {
		return m.PartialFailure
	}
	return false
}

func (m *MutateKeywordPlanAdGroupsRequest) GetValidateOnly() bool {
	if m != nil {
		return m.ValidateOnly
	}
	return false
}

// A single operation (create, update, remove) on a Keyword Plan ad group.
type KeywordPlanAdGroupOperation struct {
	// The FieldMask that determines which resource fields are modified in an
	// update.
	UpdateMask *field_mask.FieldMask `protobuf:"bytes,4,opt,name=update_mask,json=updateMask,proto3" json:"update_mask,omitempty"`
	// The mutate operation.
	//
	// Types that are valid to be assigned to Operation:
	//	*KeywordPlanAdGroupOperation_Create
	//	*KeywordPlanAdGroupOperation_Update
	//	*KeywordPlanAdGroupOperation_Remove
	Operation            isKeywordPlanAdGroupOperation_Operation `protobuf_oneof:"operation"`
	XXX_NoUnkeyedLiteral struct{}                                `json:"-"`
	XXX_unrecognized     []byte                                  `json:"-"`
	XXX_sizecache        int32                                   `json:"-"`
}

func (m *KeywordPlanAdGroupOperation) Reset()         { *m = KeywordPlanAdGroupOperation{} }
func (m *KeywordPlanAdGroupOperation) String() string { return proto.CompactTextString(m) }
func (*KeywordPlanAdGroupOperation) ProtoMessage()    {}
func (*KeywordPlanAdGroupOperation) Descriptor() ([]byte, []int) {
	return fileDescriptor_ccc3a69fa60910db, []int{2}
}

func (m *KeywordPlanAdGroupOperation) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeywordPlanAdGroupOperation.Unmarshal(m, b)
}
func (m *KeywordPlanAdGroupOperation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeywordPlanAdGroupOperation.Marshal(b, m, deterministic)
}
func (m *KeywordPlanAdGroupOperation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeywordPlanAdGroupOperation.Merge(m, src)
}
func (m *KeywordPlanAdGroupOperation) XXX_Size() int {
	return xxx_messageInfo_KeywordPlanAdGroupOperation.Size(m)
}
func (m *KeywordPlanAdGroupOperation) XXX_DiscardUnknown() {
	xxx_messageInfo_KeywordPlanAdGroupOperation.DiscardUnknown(m)
}

var xxx_messageInfo_KeywordPlanAdGroupOperation proto.InternalMessageInfo

func (m *KeywordPlanAdGroupOperation) GetUpdateMask() *field_mask.FieldMask {
	if m != nil {
		return m.UpdateMask
	}
	return nil
}

type isKeywordPlanAdGroupOperation_Operation interface {
	isKeywordPlanAdGroupOperation_Operation()
}

type KeywordPlanAdGroupOperation_Create struct {
	Create *resources.KeywordPlanAdGroup `protobuf:"bytes,1,opt,name=create,proto3,oneof"`
}

type KeywordPlanAdGroupOperation_Update struct {
	Update *resources.KeywordPlanAdGroup `protobuf:"bytes,2,opt,name=update,proto3,oneof"`
}

type KeywordPlanAdGroupOperation_Remove struct {
	Remove string `protobuf:"bytes,3,opt,name=remove,proto3,oneof"`
}

func (*KeywordPlanAdGroupOperation_Create) isKeywordPlanAdGroupOperation_Operation() {}

func (*KeywordPlanAdGroupOperation_Update) isKeywordPlanAdGroupOperation_Operation() {}

func (*KeywordPlanAdGroupOperation_Remove) isKeywordPlanAdGroupOperation_Operation() {}

func (m *KeywordPlanAdGroupOperation) GetOperation() isKeywordPlanAdGroupOperation_Operation {
	if m != nil {
		return m.Operation
	}
	return nil
}

func (m *KeywordPlanAdGroupOperation) GetCreate() *resources.KeywordPlanAdGroup {
	if x, ok := m.GetOperation().(*KeywordPlanAdGroupOperation_Create); ok {
		return x.Create
	}
	return nil
}

func (m *KeywordPlanAdGroupOperation) GetUpdate() *resources.KeywordPlanAdGroup {
	if x, ok := m.GetOperation().(*KeywordPlanAdGroupOperation_Update); ok {
		return x.Update
	}
	return nil
}

func (m *KeywordPlanAdGroupOperation) GetRemove() string {
	if x, ok := m.GetOperation().(*KeywordPlanAdGroupOperation_Remove); ok {
		return x.Remove
	}
	return ""
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*KeywordPlanAdGroupOperation) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*KeywordPlanAdGroupOperation_Create)(nil),
		(*KeywordPlanAdGroupOperation_Update)(nil),
		(*KeywordPlanAdGroupOperation_Remove)(nil),
	}
}

// Response message for a Keyword Plan ad group mutate.
type MutateKeywordPlanAdGroupsResponse struct {
	// Errors that pertain to operation failures in the partial failure mode.
	// Returned only when partial_failure = true and all errors occur inside the
	// operations. If any errors occur outside the operations (e.g. auth errors),
	// we return an RPC level error.
	PartialFailureError *status.Status `protobuf:"bytes,3,opt,name=partial_failure_error,json=partialFailureError,proto3" json:"partial_failure_error,omitempty"`
	// All results for the mutate.
	Results              []*MutateKeywordPlanAdGroupResult `protobuf:"bytes,2,rep,name=results,proto3" json:"results,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                          `json:"-"`
	XXX_unrecognized     []byte                            `json:"-"`
	XXX_sizecache        int32                             `json:"-"`
}

func (m *MutateKeywordPlanAdGroupsResponse) Reset()         { *m = MutateKeywordPlanAdGroupsResponse{} }
func (m *MutateKeywordPlanAdGroupsResponse) String() string { return proto.CompactTextString(m) }
func (*MutateKeywordPlanAdGroupsResponse) ProtoMessage()    {}
func (*MutateKeywordPlanAdGroupsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ccc3a69fa60910db, []int{3}
}

func (m *MutateKeywordPlanAdGroupsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MutateKeywordPlanAdGroupsResponse.Unmarshal(m, b)
}
func (m *MutateKeywordPlanAdGroupsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MutateKeywordPlanAdGroupsResponse.Marshal(b, m, deterministic)
}
func (m *MutateKeywordPlanAdGroupsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MutateKeywordPlanAdGroupsResponse.Merge(m, src)
}
func (m *MutateKeywordPlanAdGroupsResponse) XXX_Size() int {
	return xxx_messageInfo_MutateKeywordPlanAdGroupsResponse.Size(m)
}
func (m *MutateKeywordPlanAdGroupsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MutateKeywordPlanAdGroupsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MutateKeywordPlanAdGroupsResponse proto.InternalMessageInfo

func (m *MutateKeywordPlanAdGroupsResponse) GetPartialFailureError() *status.Status {
	if m != nil {
		return m.PartialFailureError
	}
	return nil
}

func (m *MutateKeywordPlanAdGroupsResponse) GetResults() []*MutateKeywordPlanAdGroupResult {
	if m != nil {
		return m.Results
	}
	return nil
}

// The result for the Keyword Plan ad group mutate.
type MutateKeywordPlanAdGroupResult struct {
	// Returned for successful operations.
	ResourceName         string   `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MutateKeywordPlanAdGroupResult) Reset()         { *m = MutateKeywordPlanAdGroupResult{} }
func (m *MutateKeywordPlanAdGroupResult) String() string { return proto.CompactTextString(m) }
func (*MutateKeywordPlanAdGroupResult) ProtoMessage()    {}
func (*MutateKeywordPlanAdGroupResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_ccc3a69fa60910db, []int{4}
}

func (m *MutateKeywordPlanAdGroupResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MutateKeywordPlanAdGroupResult.Unmarshal(m, b)
}
func (m *MutateKeywordPlanAdGroupResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MutateKeywordPlanAdGroupResult.Marshal(b, m, deterministic)
}
func (m *MutateKeywordPlanAdGroupResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MutateKeywordPlanAdGroupResult.Merge(m, src)
}
func (m *MutateKeywordPlanAdGroupResult) XXX_Size() int {
	return xxx_messageInfo_MutateKeywordPlanAdGroupResult.Size(m)
}
func (m *MutateKeywordPlanAdGroupResult) XXX_DiscardUnknown() {
	xxx_messageInfo_MutateKeywordPlanAdGroupResult.DiscardUnknown(m)
}

var xxx_messageInfo_MutateKeywordPlanAdGroupResult proto.InternalMessageInfo

func (m *MutateKeywordPlanAdGroupResult) GetResourceName() string {
	if m != nil {
		return m.ResourceName
	}
	return ""
}

func init() {
	proto.RegisterType((*GetKeywordPlanAdGroupRequest)(nil), "google.ads.googleads.v1.services.GetKeywordPlanAdGroupRequest")
	proto.RegisterType((*MutateKeywordPlanAdGroupsRequest)(nil), "google.ads.googleads.v1.services.MutateKeywordPlanAdGroupsRequest")
	proto.RegisterType((*KeywordPlanAdGroupOperation)(nil), "google.ads.googleads.v1.services.KeywordPlanAdGroupOperation")
	proto.RegisterType((*MutateKeywordPlanAdGroupsResponse)(nil), "google.ads.googleads.v1.services.MutateKeywordPlanAdGroupsResponse")
	proto.RegisterType((*MutateKeywordPlanAdGroupResult)(nil), "google.ads.googleads.v1.services.MutateKeywordPlanAdGroupResult")
}

func init() {
	proto.RegisterFile("google/ads/googleads/v1/services/keyword_plan_ad_group_service.proto", fileDescriptor_ccc3a69fa60910db)
}

var fileDescriptor_ccc3a69fa60910db = []byte{
	// 730 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x95, 0xdd, 0x6a, 0xd4, 0x4e,
	0x14, 0xc0, 0xff, 0xc9, 0xfe, 0xa9, 0x76, 0xb6, 0x2a, 0x8c, 0x14, 0xb7, 0x6b, 0xa9, 0x6b, 0x2c,
	0x58, 0xf6, 0x22, 0x61, 0x57, 0x8a, 0x92, 0xb2, 0xe2, 0x6e, 0x6d, 0xb7, 0x22, 0xb5, 0x25, 0x85,
	0x5e, 0x94, 0x95, 0x30, 0xdd, 0x4c, 0x43, 0x68, 0x92, 0x89, 0x33, 0x93, 0x2d, 0xa5, 0xf4, 0x46,
	0xf0, 0x09, 0x7c, 0x03, 0xbd, 0xf3, 0x45, 0x04, 0xc1, 0x2b, 0x2f, 0x7c, 0x01, 0x6f, 0xf4, 0xca,
	0x47, 0x90, 0xc9, 0x64, 0xd6, 0x7e, 0x65, 0x57, 0xda, 0xbb, 0x93, 0x33, 0x27, 0xbf, 0xf3, 0x39,
	0x67, 0xc0, 0x73, 0x9f, 0x10, 0x3f, 0xc4, 0x16, 0xf2, 0x98, 0x25, 0x45, 0x21, 0x0d, 0x1a, 0x16,
	0xc3, 0x74, 0x10, 0xf4, 0x31, 0xb3, 0xf6, 0xf1, 0xe1, 0x01, 0xa1, 0x9e, 0x9b, 0x84, 0x28, 0x76,
	0x91, 0xe7, 0xfa, 0x94, 0xa4, 0x89, 0x9b, 0x1f, 0x9b, 0x09, 0x25, 0x9c, 0xc0, 0x9a, 0xfc, 0xd5,
	0x44, 0x1e, 0x33, 0x87, 0x14, 0x73, 0xd0, 0x30, 0x15, 0xa5, 0xda, 0x2a, 0xf2, 0x43, 0x31, 0x23,
	0x29, 0x2d, 0x74, 0x24, 0x1d, 0x54, 0x67, 0xd5, 0xef, 0x49, 0x60, 0xa1, 0x38, 0x26, 0x1c, 0xf1,
	0x80, 0xc4, 0x2c, 0x3f, 0xcd, 0xdd, 0x5b, 0xd9, 0xd7, 0x6e, 0xba, 0x67, 0xed, 0x05, 0x38, 0xf4,
	0xdc, 0x08, 0xb1, 0xfd, 0xdc, 0x62, 0xee, 0xac, 0xc5, 0x01, 0x45, 0x49, 0x82, 0xa9, 0x22, 0xdc,
	0xc9, 0xcf, 0x69, 0xd2, 0xb7, 0x18, 0x47, 0x3c, 0xcd, 0x0f, 0x8c, 0x65, 0x30, 0xdb, 0xc5, 0xfc,
	0xa5, 0x0c, 0x6d, 0x33, 0x44, 0x71, 0xdb, 0xeb, 0x8a, 0xb8, 0x1c, 0xfc, 0x26, 0xc5, 0x8c, 0xc3,
	0x07, 0xe0, 0x86, 0xca, 0xc0, 0x8d, 0x51, 0x84, 0x2b, 0x5a, 0x4d, 0x5b, 0x98, 0x74, 0xa6, 0x94,
	0xf2, 0x15, 0x8a, 0xb0, 0xf1, 0x5b, 0x03, 0xb5, 0xf5, 0x94, 0x23, 0x8e, 0xcf, 0x83, 0x98, 0x22,
	0xdd, 0x03, 0xe5, 0x7e, 0xca, 0x38, 0x89, 0x30, 0x75, 0x03, 0x2f, 0xe7, 0x00, 0xa5, 0x7a, 0xe1,
	0xc1, 0xd7, 0x00, 0x90, 0x04, 0x53, 0x99, 0x79, 0x45, 0xaf, 0x95, 0x16, 0xca, 0xcd, 0x96, 0x39,
	0xae, 0xf2, 0xe6, 0x79, 0x97, 0x1b, 0x8a, 0xe2, 0x9c, 0x00, 0xc2, 0x87, 0xe0, 0x56, 0x82, 0x28,
	0x0f, 0x50, 0xe8, 0xee, 0xa1, 0x20, 0x4c, 0x29, 0xae, 0x94, 0x6a, 0xda, 0xc2, 0x75, 0xe7, 0x66,
	0xae, 0x5e, 0x95, 0x5a, 0x91, 0xf2, 0x00, 0x85, 0x81, 0x87, 0x38, 0x76, 0x49, 0x1c, 0x1e, 0x56,
	0xfe, 0xcf, 0xcc, 0xa6, 0x94, 0x72, 0x23, 0x0e, 0x0f, 0x8d, 0x8f, 0x3a, 0xb8, 0x3b, 0xc2, 0x33,
	0x5c, 0x02, 0xe5, 0x34, 0xc9, 0x10, 0xa2, 0x4b, 0x19, 0xa2, 0xdc, 0xac, 0xaa, 0x6c, 0x54, 0x9b,
	0xcc, 0x55, 0xd1, 0xc8, 0x75, 0xc4, 0xf6, 0x1d, 0x20, 0xcd, 0x85, 0x0c, 0x37, 0xc0, 0x44, 0x9f,
	0x62, 0xc4, 0x65, 0xb5, 0xcb, 0xcd, 0xc5, 0xc2, 0x2a, 0x0c, 0xa7, 0xeb, 0x82, 0x32, 0xac, 0xfd,
	0xe7, 0xe4, 0x18, 0x01, 0x94, 0xf8, 0x8a, 0x7e, 0x45, 0xa0, 0xc4, 0xc0, 0x0a, 0x98, 0xa0, 0x38,
	0x22, 0x03, 0x59, 0xc3, 0x49, 0x71, 0x22, 0xbf, 0x3b, 0x65, 0x30, 0x39, 0x2c, 0xba, 0xf1, 0x59,
	0x03, 0xf7, 0x47, 0x0c, 0x06, 0x4b, 0x48, 0xcc, 0x30, 0x5c, 0x05, 0xd3, 0x67, 0x3a, 0xe3, 0x62,
	0x4a, 0x09, 0xcd, 0xd8, 0xe5, 0x26, 0x54, 0xc1, 0xd2, 0xa4, 0x6f, 0x6e, 0x65, 0xc3, 0xeb, 0xdc,
	0x3e, 0xdd, 0xb3, 0x15, 0x61, 0x0e, 0x77, 0xc0, 0x35, 0x8a, 0x59, 0x1a, 0x72, 0x35, 0x3d, 0xcf,
	0xc6, 0x4f, 0x4f, 0x51, 0x74, 0x4e, 0x06, 0x72, 0x14, 0xd0, 0x58, 0x01, 0x73, 0xa3, 0x4d, 0xff,
	0xe9, 0xa6, 0x34, 0xbf, 0x97, 0xc0, 0xcc, 0x79, 0xc2, 0x96, 0x8c, 0x06, 0x7e, 0xd5, 0xc0, 0xf4,
	0x85, 0xb7, 0x11, 0x3e, 0x1d, 0x9f, 0xc9, 0xa8, 0x6b, 0x5c, 0xbd, 0x5c, 0xc3, 0x8d, 0xd6, 0xdb,
	0x6f, 0x3f, 0xde, 0xeb, 0x8f, 0xe1, 0xa2, 0xd8, 0x64, 0x47, 0xa7, 0xd2, 0x6b, 0xa9, 0x9b, 0xcb,
	0xac, 0xba, 0x5a, 0x6d, 0x27, 0xbb, 0x6b, 0xd5, 0x8f, 0xe1, 0x4f, 0x0d, 0xcc, 0x14, 0xb6, 0x1f,
	0x76, 0x2e, 0xdf, 0x1d, 0xb5, 0x54, 0xaa, 0xcb, 0x57, 0x62, 0xc8, 0xf9, 0x33, 0x96, 0xb3, 0x2c,
	0x5b, 0xc6, 0x13, 0x91, 0xe5, 0xdf, 0xb4, 0x8e, 0x4e, 0xac, 0xab, 0x56, 0xfd, 0xf8, 0xa2, 0x24,
	0xed, 0x28, 0x83, 0xdb, 0x5a, 0xbd, 0xf3, 0x4e, 0x07, 0xf3, 0x7d, 0x12, 0x8d, 0x8d, 0xa7, 0x33,
	0x57, 0xd8, 0xff, 0x4d, 0xb1, 0x15, 0x36, 0xb5, 0x9d, 0xb5, 0x9c, 0xe1, 0x93, 0x10, 0xc5, 0xbe,
	0x49, 0xa8, 0x6f, 0xf9, 0x38, 0xce, 0x76, 0x86, 0x7a, 0x5b, 0x92, 0x80, 0x15, 0x3f, 0x69, 0x4b,
	0x4a, 0xf8, 0xa0, 0x97, 0xba, 0xed, 0xf6, 0x27, 0xbd, 0xd6, 0x95, 0xc0, 0xb6, 0xc7, 0x4c, 0x29,
	0x0a, 0x69, 0xbb, 0x61, 0xe6, 0x8e, 0xd9, 0x17, 0x65, 0xd2, 0x6b, 0x7b, 0xac, 0x37, 0x34, 0xe9,
	0x6d, 0x37, 0x7a, 0xca, 0xe4, 0x97, 0x3e, 0x2f, 0xf5, 0xb6, 0xdd, 0xf6, 0x98, 0x6d, 0x0f, 0x8d,
	0x6c, 0x7b, 0xbb, 0x61, 0xdb, 0xca, 0x6c, 0x77, 0x22, 0x8b, 0xf3, 0xd1, 0x9f, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xc6, 0x4d, 0x17, 0xa6, 0x79, 0x07, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// KeywordPlanAdGroupServiceClient is the client API for KeywordPlanAdGroupService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type KeywordPlanAdGroupServiceClient interface {
	// Returns the requested Keyword Plan ad group in full detail.
	GetKeywordPlanAdGroup(ctx context.Context, in *GetKeywordPlanAdGroupRequest, opts ...grpc.CallOption) (*resources.KeywordPlanAdGroup, error)
	// Creates, updates, or removes Keyword Plan ad groups. Operation statuses are
	// returned.
	MutateKeywordPlanAdGroups(ctx context.Context, in *MutateKeywordPlanAdGroupsRequest, opts ...grpc.CallOption) (*MutateKeywordPlanAdGroupsResponse, error)
}

type keywordPlanAdGroupServiceClient struct {
	cc *grpc.ClientConn
}

func NewKeywordPlanAdGroupServiceClient(cc *grpc.ClientConn) KeywordPlanAdGroupServiceClient {
	return &keywordPlanAdGroupServiceClient{cc}
}

func (c *keywordPlanAdGroupServiceClient) GetKeywordPlanAdGroup(ctx context.Context, in *GetKeywordPlanAdGroupRequest, opts ...grpc.CallOption) (*resources.KeywordPlanAdGroup, error) {
	out := new(resources.KeywordPlanAdGroup)
	err := c.cc.Invoke(ctx, "/google.ads.googleads.v1.services.KeywordPlanAdGroupService/GetKeywordPlanAdGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keywordPlanAdGroupServiceClient) MutateKeywordPlanAdGroups(ctx context.Context, in *MutateKeywordPlanAdGroupsRequest, opts ...grpc.CallOption) (*MutateKeywordPlanAdGroupsResponse, error) {
	out := new(MutateKeywordPlanAdGroupsResponse)
	err := c.cc.Invoke(ctx, "/google.ads.googleads.v1.services.KeywordPlanAdGroupService/MutateKeywordPlanAdGroups", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KeywordPlanAdGroupServiceServer is the server API for KeywordPlanAdGroupService service.
type KeywordPlanAdGroupServiceServer interface {
	// Returns the requested Keyword Plan ad group in full detail.
	GetKeywordPlanAdGroup(context.Context, *GetKeywordPlanAdGroupRequest) (*resources.KeywordPlanAdGroup, error)
	// Creates, updates, or removes Keyword Plan ad groups. Operation statuses are
	// returned.
	MutateKeywordPlanAdGroups(context.Context, *MutateKeywordPlanAdGroupsRequest) (*MutateKeywordPlanAdGroupsResponse, error)
}

func RegisterKeywordPlanAdGroupServiceServer(s *grpc.Server, srv KeywordPlanAdGroupServiceServer) {
	s.RegisterService(&_KeywordPlanAdGroupService_serviceDesc, srv)
}

func _KeywordPlanAdGroupService_GetKeywordPlanAdGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetKeywordPlanAdGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeywordPlanAdGroupServiceServer).GetKeywordPlanAdGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.ads.googleads.v1.services.KeywordPlanAdGroupService/GetKeywordPlanAdGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeywordPlanAdGroupServiceServer).GetKeywordPlanAdGroup(ctx, req.(*GetKeywordPlanAdGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeywordPlanAdGroupService_MutateKeywordPlanAdGroups_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MutateKeywordPlanAdGroupsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeywordPlanAdGroupServiceServer).MutateKeywordPlanAdGroups(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.ads.googleads.v1.services.KeywordPlanAdGroupService/MutateKeywordPlanAdGroups",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeywordPlanAdGroupServiceServer).MutateKeywordPlanAdGroups(ctx, req.(*MutateKeywordPlanAdGroupsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _KeywordPlanAdGroupService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "google.ads.googleads.v1.services.KeywordPlanAdGroupService",
	HandlerType: (*KeywordPlanAdGroupServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetKeywordPlanAdGroup",
			Handler:    _KeywordPlanAdGroupService_GetKeywordPlanAdGroup_Handler,
		},
		{
			MethodName: "MutateKeywordPlanAdGroups",
			Handler:    _KeywordPlanAdGroupService_MutateKeywordPlanAdGroups_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "google/ads/googleads/v1/services/keyword_plan_ad_group_service.proto",
}
