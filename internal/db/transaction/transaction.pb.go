// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.17.3
// source: transaction.proto

package transaction

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Status int32

const (
	Status_UNKNOWN        Status = 0
	Status_RECEIVED       Status = 1
	Status_PENDING        Status = 2
	Status_ACCEPTED_ON_L2 Status = 3
	Status_ACCEPTED_ON_L1 Status = 4
	Status_REJECTED       Status = 5
)

// Enum value maps for Status.
var (
	Status_name = map[int32]string{
		0: "UNKNOWN",
		1: "RECEIVED",
		2: "PENDING",
		3: "ACCEPTED_ON_L2",
		4: "ACCEPTED_ON_L1",
		5: "REJECTED",
	}
	Status_value = map[string]int32{
		"UNKNOWN":        0,
		"RECEIVED":       1,
		"PENDING":        2,
		"ACCEPTED_ON_L2": 3,
		"ACCEPTED_ON_L1": 4,
		"REJECTED":       5,
	}
)

func (x Status) Enum() *Status {
	p := new(Status)
	*p = x
	return p
}

func (x Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status) Descriptor() protoreflect.EnumDescriptor {
	return file_transaction_proto_enumTypes[0].Descriptor()
}

func (Status) Type() protoreflect.EnumType {
	return &file_transaction_proto_enumTypes[0]
}

func (x Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status.Descriptor instead.
func (Status) EnumDescriptor() ([]byte, []int) {
	return file_transaction_proto_rawDescGZIP(), []int{0}
}

type Transaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hash []byte `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	// Types that are assignable to Tx:
	//	*Transaction_Deploy
	//	*Transaction_Invoke
	Tx isTransaction_Tx `protobuf_oneof:"tx"`
}

func (x *Transaction) Reset() {
	*x = Transaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transaction_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Transaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction) ProtoMessage() {}

func (x *Transaction) ProtoReflect() protoreflect.Message {
	mi := &file_transaction_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction.ProtoReflect.Descriptor instead.
func (*Transaction) Descriptor() ([]byte, []int) {
	return file_transaction_proto_rawDescGZIP(), []int{0}
}

func (x *Transaction) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

func (m *Transaction) GetTx() isTransaction_Tx {
	if m != nil {
		return m.Tx
	}
	return nil
}

func (x *Transaction) GetDeploy() *Deploy {
	if x, ok := x.GetTx().(*Transaction_Deploy); ok {
		return x.Deploy
	}
	return nil
}

func (x *Transaction) GetInvoke() *InvokeFunction {
	if x, ok := x.GetTx().(*Transaction_Invoke); ok {
		return x.Invoke
	}
	return nil
}

type isTransaction_Tx interface {
	isTransaction_Tx()
}

type Transaction_Deploy struct {
	Deploy *Deploy `protobuf:"bytes,2,opt,name=deploy,proto3,oneof"`
}

type Transaction_Invoke struct {
	Invoke *InvokeFunction `protobuf:"bytes,3,opt,name=invoke,proto3,oneof"`
}

func (*Transaction_Deploy) isTransaction_Tx() {}

func (*Transaction_Invoke) isTransaction_Tx() {}

type Deploy struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ContractAddressSalt []byte   `protobuf:"bytes,1,opt,name=contractAddressSalt,proto3" json:"contractAddressSalt,omitempty"`
	ConstructorCallData [][]byte `protobuf:"bytes,2,rep,name=constructorCallData,proto3" json:"constructorCallData,omitempty"`
}

func (x *Deploy) Reset() {
	*x = Deploy{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transaction_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Deploy) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Deploy) ProtoMessage() {}

func (x *Deploy) ProtoReflect() protoreflect.Message {
	mi := &file_transaction_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Deploy.ProtoReflect.Descriptor instead.
func (*Deploy) Descriptor() ([]byte, []int) {
	return file_transaction_proto_rawDescGZIP(), []int{1}
}

func (x *Deploy) GetContractAddressSalt() []byte {
	if x != nil {
		return x.ContractAddressSalt
	}
	return nil
}

func (x *Deploy) GetConstructorCallData() [][]byte {
	if x != nil {
		return x.ConstructorCallData
	}
	return nil
}

type InvokeFunction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ContractAddress    []byte   `protobuf:"bytes,2,opt,name=contractAddress,proto3" json:"contractAddress,omitempty"`
	EntryPointSelector []byte   `protobuf:"bytes,3,opt,name=entryPointSelector,proto3" json:"entryPointSelector,omitempty"`
	CallData           [][]byte `protobuf:"bytes,4,rep,name=callData,proto3" json:"callData,omitempty"`
	Signature          [][]byte `protobuf:"bytes,5,rep,name=signature,proto3" json:"signature,omitempty"`
	MaxFee             []byte   `protobuf:"bytes,6,opt,name=maxFee,proto3" json:"maxFee,omitempty"`
}

func (x *InvokeFunction) Reset() {
	*x = InvokeFunction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transaction_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InvokeFunction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InvokeFunction) ProtoMessage() {}

func (x *InvokeFunction) ProtoReflect() protoreflect.Message {
	mi := &file_transaction_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InvokeFunction.ProtoReflect.Descriptor instead.
func (*InvokeFunction) Descriptor() ([]byte, []int) {
	return file_transaction_proto_rawDescGZIP(), []int{2}
}

func (x *InvokeFunction) GetContractAddress() []byte {
	if x != nil {
		return x.ContractAddress
	}
	return nil
}

func (x *InvokeFunction) GetEntryPointSelector() []byte {
	if x != nil {
		return x.EntryPointSelector
	}
	return nil
}

func (x *InvokeFunction) GetCallData() [][]byte {
	if x != nil {
		return x.CallData
	}
	return nil
}

func (x *InvokeFunction) GetSignature() [][]byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *InvokeFunction) GetMaxFee() []byte {
	if x != nil {
		return x.MaxFee
	}
	return nil
}

type TransactionReceipt struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TxHash          []byte         `protobuf:"bytes,1,opt,name=txHash,proto3" json:"txHash,omitempty"`
	ActualFee       []byte         `protobuf:"bytes,2,opt,name=actualFee,proto3" json:"actualFee,omitempty"`
	Status          Status         `protobuf:"varint,3,opt,name=status,proto3,enum=Status" json:"status,omitempty"`
	StatusData      string         `protobuf:"bytes,4,opt,name=statusData,proto3" json:"statusData,omitempty"`
	MessagesSent    []*MessageToL1 `protobuf:"bytes,5,rep,name=messagesSent,proto3" json:"messagesSent,omitempty"`
	L1OriginMessage *MessageToL2   `protobuf:"bytes,6,opt,name=l1OriginMessage,proto3" json:"l1OriginMessage,omitempty"`
	Events          []*Event       `protobuf:"bytes,7,rep,name=events,proto3" json:"events,omitempty"`
}

func (x *TransactionReceipt) Reset() {
	*x = TransactionReceipt{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transaction_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TransactionReceipt) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransactionReceipt) ProtoMessage() {}

func (x *TransactionReceipt) ProtoReflect() protoreflect.Message {
	mi := &file_transaction_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransactionReceipt.ProtoReflect.Descriptor instead.
func (*TransactionReceipt) Descriptor() ([]byte, []int) {
	return file_transaction_proto_rawDescGZIP(), []int{3}
}

func (x *TransactionReceipt) GetTxHash() []byte {
	if x != nil {
		return x.TxHash
	}
	return nil
}

func (x *TransactionReceipt) GetActualFee() []byte {
	if x != nil {
		return x.ActualFee
	}
	return nil
}

func (x *TransactionReceipt) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_UNKNOWN
}

func (x *TransactionReceipt) GetStatusData() string {
	if x != nil {
		return x.StatusData
	}
	return ""
}

func (x *TransactionReceipt) GetMessagesSent() []*MessageToL1 {
	if x != nil {
		return x.MessagesSent
	}
	return nil
}

func (x *TransactionReceipt) GetL1OriginMessage() *MessageToL2 {
	if x != nil {
		return x.L1OriginMessage
	}
	return nil
}

func (x *TransactionReceipt) GetEvents() []*Event {
	if x != nil {
		return x.Events
	}
	return nil
}

type MessageToL1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ToAddress []byte   `protobuf:"bytes,1,opt,name=toAddress,proto3" json:"toAddress,omitempty"`
	Payload   [][]byte `protobuf:"bytes,2,rep,name=payload,proto3" json:"payload,omitempty"`
}

func (x *MessageToL1) Reset() {
	*x = MessageToL1{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transaction_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageToL1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageToL1) ProtoMessage() {}

func (x *MessageToL1) ProtoReflect() protoreflect.Message {
	mi := &file_transaction_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageToL1.ProtoReflect.Descriptor instead.
func (*MessageToL1) Descriptor() ([]byte, []int) {
	return file_transaction_proto_rawDescGZIP(), []int{4}
}

func (x *MessageToL1) GetToAddress() []byte {
	if x != nil {
		return x.ToAddress
	}
	return nil
}

func (x *MessageToL1) GetPayload() [][]byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

type MessageToL2 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FromAddress []byte   `protobuf:"bytes,1,opt,name=fromAddress,proto3" json:"fromAddress,omitempty"`
	Payload     [][]byte `protobuf:"bytes,2,rep,name=payload,proto3" json:"payload,omitempty"`
}

func (x *MessageToL2) Reset() {
	*x = MessageToL2{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transaction_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageToL2) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageToL2) ProtoMessage() {}

func (x *MessageToL2) ProtoReflect() protoreflect.Message {
	mi := &file_transaction_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageToL2.ProtoReflect.Descriptor instead.
func (*MessageToL2) Descriptor() ([]byte, []int) {
	return file_transaction_proto_rawDescGZIP(), []int{5}
}

func (x *MessageToL2) GetFromAddress() []byte {
	if x != nil {
		return x.FromAddress
	}
	return nil
}

func (x *MessageToL2) GetPayload() [][]byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FromAddress []byte   `protobuf:"bytes,1,opt,name=fromAddress,proto3" json:"fromAddress,omitempty"`
	Keys        [][]byte `protobuf:"bytes,2,rep,name=keys,proto3" json:"keys,omitempty"`
	Data        [][]byte `protobuf:"bytes,3,rep,name=data,proto3" json:"data,omitempty"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transaction_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_transaction_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_transaction_proto_rawDescGZIP(), []int{6}
}

func (x *Event) GetFromAddress() []byte {
	if x != nil {
		return x.FromAddress
	}
	return nil
}

func (x *Event) GetKeys() [][]byte {
	if x != nil {
		return x.Keys
	}
	return nil
}

func (x *Event) GetData() [][]byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_transaction_proto protoreflect.FileDescriptor

var file_transaction_proto_rawDesc = []byte{
	0x0a, 0x11, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x75, 0x0a, 0x0b, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x21, 0x0a, 0x06, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x48,
	0x00, 0x52, 0x06, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x12, 0x29, 0x0a, 0x06, 0x69, 0x6e, 0x76,
	0x6f, 0x6b, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x49, 0x6e, 0x76, 0x6f,
	0x6b, 0x65, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x06, 0x69, 0x6e,
	0x76, 0x6f, 0x6b, 0x65, 0x42, 0x04, 0x0a, 0x02, 0x74, 0x78, 0x22, 0x6c, 0x0a, 0x06, 0x44, 0x65,
	0x70, 0x6c, 0x6f, 0x79, 0x12, 0x30, 0x0a, 0x13, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74,
	0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x53, 0x61, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x13, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x53, 0x61, 0x6c, 0x74, 0x12, 0x30, 0x0a, 0x13, 0x63, 0x6f, 0x6e, 0x73, 0x74, 0x72,
	0x75, 0x63, 0x74, 0x6f, 0x72, 0x43, 0x61, 0x6c, 0x6c, 0x44, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0c, 0x52, 0x13, 0x63, 0x6f, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x6f, 0x72,
	0x43, 0x61, 0x6c, 0x6c, 0x44, 0x61, 0x74, 0x61, 0x22, 0xbc, 0x01, 0x0a, 0x0e, 0x49, 0x6e, 0x76,
	0x6f, 0x6b, 0x65, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x28, 0x0a, 0x0f, 0x63,
	0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x2e, 0x0a, 0x12, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x50, 0x6f,
	0x69, 0x6e, 0x74, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x12, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x53, 0x65, 0x6c,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x61, 0x6c, 0x6c, 0x44, 0x61, 0x74,
	0x61, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x08, 0x63, 0x61, 0x6c, 0x6c, 0x44, 0x61, 0x74,
	0x61, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x05,
	0x20, 0x03, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x6d, 0x61, 0x78, 0x46, 0x65, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x06, 0x6d, 0x61, 0x78, 0x46, 0x65, 0x65, 0x22, 0x95, 0x02, 0x0a, 0x12, 0x54, 0x72, 0x61, 0x6e,
	0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x12, 0x16,
	0x0a, 0x06, 0x74, 0x78, 0x48, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06,
	0x74, 0x78, 0x48, 0x61, 0x73, 0x68, 0x12, 0x1c, 0x0a, 0x09, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c,
	0x46, 0x65, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x61, 0x63, 0x74, 0x75, 0x61,
	0x6c, 0x46, 0x65, 0x65, 0x12, 0x1f, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x07, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x44,
	0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x44, 0x61, 0x74, 0x61, 0x12, 0x30, 0x0a, 0x0c, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x73, 0x53, 0x65, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x4c, 0x31, 0x52, 0x0c, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x73, 0x53, 0x65, 0x6e, 0x74, 0x12, 0x36, 0x0a, 0x0f, 0x6c, 0x31, 0x4f, 0x72, 0x69,
	0x67, 0x69, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0c, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x4c, 0x32, 0x52, 0x0f,
	0x6c, 0x31, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x1e, 0x0a, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x06, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x22,
	0x45, 0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x4c, 0x31, 0x12, 0x1c,
	0x0a, 0x09, 0x74, 0x6f, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x09, 0x74, 0x6f, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x18, 0x0a, 0x07,
	0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x07, 0x70,
	0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0x49, 0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x54, 0x6f, 0x4c, 0x32, 0x12, 0x20, 0x0a, 0x0b, 0x66, 0x72, 0x6f, 0x6d, 0x41, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x66, 0x72, 0x6f, 0x6d,
	0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61,
	0x64, 0x22, 0x51, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x66, 0x72,
	0x6f, 0x6d, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x0b, 0x66, 0x72, 0x6f, 0x6d, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x12, 0x0a, 0x04,
	0x6b, 0x65, 0x79, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73,
	0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x2a, 0x66, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0b,
	0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x52,
	0x45, 0x43, 0x45, 0x49, 0x56, 0x45, 0x44, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x50, 0x45, 0x4e,
	0x44, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x12, 0x12, 0x0a, 0x0e, 0x41, 0x43, 0x43, 0x45, 0x50, 0x54,
	0x45, 0x44, 0x5f, 0x4f, 0x4e, 0x5f, 0x4c, 0x32, 0x10, 0x03, 0x12, 0x12, 0x0a, 0x0e, 0x41, 0x43,
	0x43, 0x45, 0x50, 0x54, 0x45, 0x44, 0x5f, 0x4f, 0x4e, 0x5f, 0x4c, 0x31, 0x10, 0x04, 0x12, 0x0c,
	0x0a, 0x08, 0x52, 0x45, 0x4a, 0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x05, 0x42, 0x34, 0x5a, 0x32,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4e, 0x65, 0x74, 0x68, 0x65,
	0x72, 0x6d, 0x69, 0x6e, 0x64, 0x45, 0x74, 0x68, 0x2f, 0x6a, 0x75, 0x6e, 0x6f, 0x2f, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_transaction_proto_rawDescOnce sync.Once
	file_transaction_proto_rawDescData = file_transaction_proto_rawDesc
)

func file_transaction_proto_rawDescGZIP() []byte {
	file_transaction_proto_rawDescOnce.Do(func() {
		file_transaction_proto_rawDescData = protoimpl.X.CompressGZIP(file_transaction_proto_rawDescData)
	})
	return file_transaction_proto_rawDescData
}

var file_transaction_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_transaction_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_transaction_proto_goTypes = []interface{}{
	(Status)(0),                // 0: Status
	(*Transaction)(nil),        // 1: Transaction
	(*Deploy)(nil),             // 2: Deploy
	(*InvokeFunction)(nil),     // 3: InvokeFunction
	(*TransactionReceipt)(nil), // 4: TransactionReceipt
	(*MessageToL1)(nil),        // 5: MessageToL1
	(*MessageToL2)(nil),        // 6: MessageToL2
	(*Event)(nil),              // 7: Event
}
var file_transaction_proto_depIdxs = []int32{
	2, // 0: Transaction.deploy:type_name -> Deploy
	3, // 1: Transaction.invoke:type_name -> InvokeFunction
	0, // 2: TransactionReceipt.status:type_name -> Status
	5, // 3: TransactionReceipt.messagesSent:type_name -> MessageToL1
	6, // 4: TransactionReceipt.l1OriginMessage:type_name -> MessageToL2
	7, // 5: TransactionReceipt.events:type_name -> Event
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_transaction_proto_init() }
func file_transaction_proto_init() {
	if File_transaction_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_transaction_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Transaction); i {
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
		file_transaction_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Deploy); i {
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
		file_transaction_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InvokeFunction); i {
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
		file_transaction_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TransactionReceipt); i {
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
		file_transaction_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageToL1); i {
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
		file_transaction_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageToL2); i {
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
		file_transaction_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
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
	file_transaction_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Transaction_Deploy)(nil),
		(*Transaction_Invoke)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_transaction_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transaction_proto_goTypes,
		DependencyIndexes: file_transaction_proto_depIdxs,
		EnumInfos:         file_transaction_proto_enumTypes,
		MessageInfos:      file_transaction_proto_msgTypes,
	}.Build()
	File_transaction_proto = out.File
	file_transaction_proto_rawDesc = nil
	file_transaction_proto_goTypes = nil
	file_transaction_proto_depIdxs = nil
}
