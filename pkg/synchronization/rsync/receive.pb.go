// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.4
// source: synchronization/rsync/receive.proto

package rsync

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// ReceivingStatus encodes that status of an rsync receiver.
type ReceiverStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Path is the path currently being received.
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// Received is the number of paths that have already been received.
	Received uint64 `protobuf:"varint,2,opt,name=received,proto3" json:"received,omitempty"`
	// Total is the total number of paths expected.
	Total uint64 `protobuf:"varint,3,opt,name=total,proto3" json:"total,omitempty"`
}

func (x *ReceiverStatus) Reset() {
	*x = ReceiverStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_synchronization_rsync_receive_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReceiverStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReceiverStatus) ProtoMessage() {}

func (x *ReceiverStatus) ProtoReflect() protoreflect.Message {
	mi := &file_synchronization_rsync_receive_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReceiverStatus.ProtoReflect.Descriptor instead.
func (*ReceiverStatus) Descriptor() ([]byte, []int) {
	return file_synchronization_rsync_receive_proto_rawDescGZIP(), []int{0}
}

func (x *ReceiverStatus) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *ReceiverStatus) GetReceived() uint64 {
	if x != nil {
		return x.Received
	}
	return 0
}

func (x *ReceiverStatus) GetTotal() uint64 {
	if x != nil {
		return x.Total
	}
	return 0
}

var File_synchronization_rsync_receive_proto protoreflect.FileDescriptor

var file_synchronization_rsync_receive_proto_rawDesc = []byte{
	0x0a, 0x23, 0x73, 0x79, 0x6e, 0x63, 0x68, 0x72, 0x6f, 0x6e, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x2f, 0x72, 0x73, 0x79, 0x6e, 0x63, 0x2f, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x72, 0x73, 0x79, 0x6e, 0x63, 0x22, 0x56, 0x0a, 0x0e,
	0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x12,
	0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61,
	0x74, 0x68, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x64, 0x12, 0x14,
	0x0a, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x74,
	0x6f, 0x74, 0x61, 0x6c, 0x42, 0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6d, 0x75, 0x74, 0x61, 0x67, 0x65, 0x6e, 0x2d, 0x69, 0x6f, 0x2f, 0x6d, 0x75,
	0x74, 0x61, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x73, 0x79, 0x6e, 0x63, 0x68, 0x72,
	0x6f, 0x6e, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x72, 0x73, 0x79, 0x6e, 0x63, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_synchronization_rsync_receive_proto_rawDescOnce sync.Once
	file_synchronization_rsync_receive_proto_rawDescData = file_synchronization_rsync_receive_proto_rawDesc
)

func file_synchronization_rsync_receive_proto_rawDescGZIP() []byte {
	file_synchronization_rsync_receive_proto_rawDescOnce.Do(func() {
		file_synchronization_rsync_receive_proto_rawDescData = protoimpl.X.CompressGZIP(file_synchronization_rsync_receive_proto_rawDescData)
	})
	return file_synchronization_rsync_receive_proto_rawDescData
}

var file_synchronization_rsync_receive_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_synchronization_rsync_receive_proto_goTypes = []interface{}{
	(*ReceiverStatus)(nil), // 0: rsync.ReceiverStatus
}
var file_synchronization_rsync_receive_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_synchronization_rsync_receive_proto_init() }
func file_synchronization_rsync_receive_proto_init() {
	if File_synchronization_rsync_receive_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_synchronization_rsync_receive_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReceiverStatus); i {
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
			RawDescriptor: file_synchronization_rsync_receive_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_synchronization_rsync_receive_proto_goTypes,
		DependencyIndexes: file_synchronization_rsync_receive_proto_depIdxs,
		MessageInfos:      file_synchronization_rsync_receive_proto_msgTypes,
	}.Build()
	File_synchronization_rsync_receive_proto = out.File
	file_synchronization_rsync_receive_proto_rawDesc = nil
	file_synchronization_rsync_receive_proto_goTypes = nil
	file_synchronization_rsync_receive_proto_depIdxs = nil
}
