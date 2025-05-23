// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.2
// source: track.proto

package proto_struct

import (
	"encoding/json"
	"fmt"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	"time"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// 轨迹点信息
type TrackPoint struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	Vin            int32                  `parquet:"vin" protobuf:"varint,1,opt,name=vin,proto3" json:"vin,omitempty"`                                             // 车辆识别号
	CollectionTime int64                  `parquet:"collectiontime" protobuf:"varint,2,opt,name=collection_time,json=collectionTime,proto3" json:"collection_time,omitempty"` // 采集时间戳
	Date           string                 `parquet:"date" protobuf:"bytes,3,opt,name=date,proto3" json:"date,omitempty"`                                            // 日期
	Timestamp      string                 `parquet:"timestamp" protobuf:"bytes,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`                                  // 时间戳字符串
	Hour           int32                  `parquet:"hour" protobuf:"varint,5,opt,name=hour,proto3" json:"hour,omitempty"`                                           // 小时
	Speed          float64                `parquet:"speed" protobuf:"fixed64,6,opt,name=speed,proto3" json:"speed,omitempty"`                                        // 速度
	Longitude      float64                `parquet:"longitude" protobuf:"fixed64,7,opt,name=longitude,proto3" json:"longitude,omitempty"`                                // 经度
	Latitude       float64                `parquet:"latitude" protobuf:"fixed64,8,opt,name=latitude,proto3" json:"latitude,omitempty"`
	StartTime      string                 `protobuf:"bytes,9,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	EndTime        string                 `protobuf:"bytes,10,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`// 纬度
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (data *TrackPoint) ToCsv() []string {
	t := time.UnixMilli(data.CollectionTime)
	csvData := []string{
		fmt.Sprint(data.Vin),
		fmt.Sprintf("%d", data.CollectionTime),
		t.Format("2006-01-02"),
		t.Format("15:04:05"),
		fmt.Sprintf("%d", t.Hour()),
		fmt.Sprintf("%.1f", data.Speed),
		fmt.Sprintf("%.6f", data.Longitude),
		fmt.Sprintf("%.6f", data.Latitude),
	}
	return csvData
}

func (x *TrackPoint) Reset() {
	*x = TrackPoint{}
	mi := &file_track_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TrackPoint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrackPoint) ProtoMessage() {}

func (x *TrackPoint) ProtoReflect() protoreflect.Message {
	mi := &file_track_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrackPoint.ProtoReflect.Descriptor instead.
func (*TrackPoint) Descriptor() ([]byte, []int) {
	return file_track_proto_rawDescGZIP(), []int{0}
}

func (x *TrackPoint) GetVin() int32 {
	if x != nil {
		return x.Vin
	}
	return 0
}

func (x *TrackPoint) GetCollectionTime() int64 {
	if x != nil {
		return x.CollectionTime
	}
	return 0
}

func (x *TrackPoint) GetDate() string {
	if x != nil {
		return x.Date
	}
	return ""
}

func (x *TrackPoint) GetTimestamp() string {
	if x != nil {
		return x.Timestamp
	}
	return ""
}

func (x *TrackPoint) GetHour() int32 {
	if x != nil {
		return x.Hour
	}
	return 0
}

func (x *TrackPoint) GetSpeed() float64 {
	if x != nil {
		return x.Speed
	}
	return 0
}

func (x *TrackPoint) GetLongitude() float64 {
	if x != nil {
		return x.Longitude
	}
	return 0
}

func (x *TrackPoint) GetLatitude() float64 {
	if x != nil {
		return x.Latitude
	}
	return 0
}

func (x *TrackPoint) GetStartTime() string {
	if x != nil {
		return x.StartTime
	}
	return ""
}

func (x *TrackPoint) GetEndTime() string {
	if x != nil {
		return x.EndTime
	}
	return ""
}

// 轨迹段信息
type TrackSegment struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	StartTime      string                 `protobuf:"bytes,1,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`                // 开始时间
	EndTime        string                 `protobuf:"bytes,2,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`                      // 结束时间
	RoadId         int64                  `protobuf:"varint,3,opt,name=road_id,json=roadId,proto3" json:"road_id,omitempty"`                        // 道路ID
	TrackPoints    []*TrackPoint          `protobuf:"bytes,4,rep,name=track_points,json=trackPoints,proto3" json:"track_points,omitempty"`          // 轨迹点列表
	OriginalPoints []*TrackPoint          `protobuf:"bytes,5,rep,name=original_points,json=originalPoints,proto3" json:"original_points,omitempty"` // 原始轨迹点列表
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *TrackSegment) Reset() {
	*x = TrackSegment{}
	mi := &file_track_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TrackSegment) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrackSegment) ProtoMessage() {}

func (x *TrackSegment) ProtoReflect() protoreflect.Message {
	mi := &file_track_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrackSegment.ProtoReflect.Descriptor instead.
func (*TrackSegment) Descriptor() ([]byte, []int) {
	return file_track_proto_rawDescGZIP(), []int{1}
}

func (x *TrackSegment) GetStartTime() string {
	if x != nil {
		return x.StartTime
	}
	return ""
}

func (x *TrackSegment) GetEndTime() string {
	if x != nil {
		return x.EndTime
	}
	return ""
}

func (x *TrackSegment) GetRoadId() int64 {
	if x != nil {
		return x.RoadId
	}
	return 0
}

func (x *TrackSegment) GetTrackPoints() []*TrackPoint {
	if x != nil {
		return x.TrackPoints
	}
	return nil
}

func (x *TrackSegment) GetOriginalPoints() []*TrackPoint {
	if x != nil {
		return x.OriginalPoints
	}
	return nil
}

// 完整轨迹信息
type Track struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Vin           int32                  `protobuf:"varint,1,opt,name=vin,proto3" json:"vin,omitempty"`                             // 车辆识别号
	Tid           int32                  `protobuf:"varint,2,opt,name=tid,proto3" json:"tid,omitempty"`                             // 轨迹ID
	StartTime     string                 `protobuf:"bytes,3,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"` // 开始时间
	EndTime       string                 `protobuf:"bytes,4,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`       // 结束时间
	Date          string                 `protobuf:"bytes,5,opt,name=date,proto3" json:"date,omitempty"`                            // 日期
	TrackSegs     []*TrackSegment        `protobuf:"bytes,6,rep,name=track_segs,json=trackSegs,proto3" json:"track_segs,omitempty"` // 轨迹段列表
	Probability   float64                `protobuf:"fixed64,7,opt,name=probability,proto3" json:"probability,omitempty"`            // 概率值
	IsBad         int32                  `protobuf:"varint,8,opt,name=isBad,proto3" json:"isBad,omitempty"`
	DisCount      int32                  `protobuf:"varint,9,opt,name=disCount,proto3" json:"disCount,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (t *Track) ToCsv() [][]string {
	result := make([][]string, 0, len(t.TrackSegs))

	for _, seg := range t.TrackSegs {
		// 将单个轨迹段的轨迹点转换为JSON字符串
		trackPointsBytes, _ := json.Marshal(seg.TrackPoints)
		originalPointsBytes, _ := json.Marshal(seg.OriginalPoints)

		row := []string{
			fmt.Sprintf("%d", t.Vin),
			t.Date,
			fmt.Sprintf("%d", t.Tid),
			seg.StartTime,
			seg.EndTime,
			fmt.Sprintf("%d", seg.RoadId),
			string(trackPointsBytes),
			string(originalPointsBytes),
		}

		result = append(result, row)
	}

	// 如果没有轨迹段，返回一个包含基本信息的行
	if len(result) == 0 {
		result = append(result, []string{
			fmt.Sprintf("%d", t.Vin),
			fmt.Sprintf("%d", t.Tid),
			t.StartTime,
			t.EndTime,
			t.Date,
			"",
			"[]",
			"[]",
		})
	}

	return result
}

func (x *Track) Reset() {
	*x = Track{}
	mi := &file_track_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Track) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Track) ProtoMessage() {}

func (x *Track) ProtoReflect() protoreflect.Message {
	mi := &file_track_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Track.ProtoReflect.Descriptor instead.
func (*Track) Descriptor() ([]byte, []int) {
	return file_track_proto_rawDescGZIP(), []int{2}
}

func (x *Track) GetVin() int32 {
	if x != nil {
		return x.Vin
	}
	return 0
}

func (x *Track) GetTid() int32 {
	if x != nil {
		return x.Tid
	}
	return 0
}

func (x *Track) GetStartTime() string {
	if x != nil {
		return x.StartTime
	}
	return ""
}

func (x *Track) GetEndTime() string {
	if x != nil {
		return x.EndTime
	}
	return ""
}

func (x *Track) GetDate() string {
	if x != nil {
		return x.Date
	}
	return ""
}

func (x *Track) GetTrackSegs() []*TrackSegment {
	if x != nil {
		return x.TrackSegs
	}
	return nil
}

func (x *Track) GetProbability() float64 {
	if x != nil {
		return x.Probability
	}
	return 0
}

func (x *Track) GetIsBad() int32 {
	if x != nil {
		return x.IsBad
	}
	return 0
}

func (x *Track) GetDisCount() int32 {
	if x != nil {
		return x.DisCount
	}
	return 0
}

var File_track_proto protoreflect.FileDescriptor

const file_track_proto_rawDesc = "" +
	"\n" +
	"\vtrack.proto\x12\x05track\"\x97\x02\n" +
	"\n" +
	"TrackPoint\x12\x10\n" +
	"\x03vin\x18\x01 \x01(\x05R\x03vin\x12'\n" +
	"\x0fcollection_time\x18\x02 \x01(\x03R\x0ecollectionTime\x12\x12\n" +
	"\x04date\x18\x03 \x01(\tR\x04date\x12\x1c\n" +
	"\ttimestamp\x18\x04 \x01(\tR\ttimestamp\x12\x12\n" +
	"\x04hour\x18\x05 \x01(\x05R\x04hour\x12\x14\n" +
	"\x05speed\x18\x06 \x01(\x01R\x05speed\x12\x1c\n" +
	"\tlongitude\x18\a \x01(\x01R\tlongitude\x12\x1a\n" +
	"\blatitude\x18\b \x01(\x01R\blatitude\x12\x1d\n" +
	"\n" +
	"start_time\x18\t \x01(\tR\tstartTime\x12\x19\n" +
	"\bend_time\x18\n" +
	" \x01(\tR\aendTime\"\xd3\x01\n" +
	"\fTrackSegment\x12\x1d\n" +
	"\n" +
	"start_time\x18\x01 \x01(\tR\tstartTime\x12\x19\n" +
	"\bend_time\x18\x02 \x01(\tR\aendTime\x12\x17\n" +
	"\aroad_id\x18\x03 \x01(\x03R\x06roadId\x124\n" +
	"\ftrack_points\x18\x04 \x03(\v2\x11.track.TrackPointR\vtrackPoints\x12:\n" +
	"\x0foriginal_points\x18\x05 \x03(\v2\x11.track.TrackPointR\x0eoriginalPoints\"\x81\x02\n" +
	"\x05Track\x12\x10\n" +
	"\x03vin\x18\x01 \x01(\x05R\x03vin\x12\x10\n" +
	"\x03tid\x18\x02 \x01(\x05R\x03tid\x12\x1d\n" +
	"\n" +
	"start_time\x18\x03 \x01(\tR\tstartTime\x12\x19\n" +
	"\bend_time\x18\x04 \x01(\tR\aendTime\x12\x12\n" +
	"\x04date\x18\x05 \x01(\tR\x04date\x122\n" +
	"\n" +
	"track_segs\x18\x06 \x03(\v2\x13.track.TrackSegmentR\ttrackSegs\x12 \n" +
	"\vprobability\x18\a \x01(\x01R\vprobability\x12\x14\n" +
	"\x05isBad\x18\b \x01(\x05R\x05isBad\x12\x1a\n" +
	"\bdisCount\x18\t \x01(\x05R\bdisCountB\x10Z\x0e.;proto_structb\x06proto3"

var (
	file_track_proto_rawDescOnce sync.Once
	file_track_proto_rawDescData []byte
)

func file_track_proto_rawDescGZIP() []byte {
	file_track_proto_rawDescOnce.Do(func() {
		file_track_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_track_proto_rawDesc), len(file_track_proto_rawDesc)))
	})
	return file_track_proto_rawDescData
}

var file_track_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_track_proto_goTypes = []any{
	(*TrackPoint)(nil),   // 0: track.TrackPoint
	(*TrackSegment)(nil), // 1: track.TrackSegment
	(*Track)(nil),        // 2: track.Track
}
var file_track_proto_depIdxs = []int32{
	0, // 0: track.TrackSegment.track_points:type_name -> track.TrackPoint
	0, // 1: track.TrackSegment.original_points:type_name -> track.TrackPoint
	1, // 2: track.Track.track_segs:type_name -> track.TrackSegment
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_track_proto_init() }
func file_track_proto_init() {
	if File_track_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_track_proto_rawDesc), len(file_track_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_track_proto_goTypes,
		DependencyIndexes: file_track_proto_depIdxs,
		MessageInfos:      file_track_proto_msgTypes,
	}.Build()
	File_track_proto = out.File
	file_track_proto_goTypes = nil
	file_track_proto_depIdxs = nil
}
