package proto_tools

import (
	"EVdata/common"
	"EVdata/proto_struct"
	"bytes"
	"encoding/binary"
	"google.golang.org/protobuf/proto"
	"io"
	"os"
)

func ReadTrackFromProto(filename string) []*proto_struct.Track {
	file, _ := os.Open(filename)
	defer file.Close()
	bs, err := io.ReadAll(file)
	if err != nil {
		common.ErrorLog("ReadTrackFromProto err:", err)
	}
	reader := bytes.NewReader(bs)
	res := make([]*proto_struct.Track, 0)
	var length uint32
	for {
		if err = binary.Read(reader, binary.LittleEndian, &length); err == io.EOF {
			break
		}
		//common.InfoLog("ReadTrackFromProto length:", length)
		msgbf := make([]byte, length)
		rlen, err := reader.Read(msgbf)
		if err != nil || rlen != int(length) {
			common.ErrorLog("ReadTrackFromProto err:", err)
		}
		//common.InfoLog("ReadTrackFromProto length:", length)
		var track proto_struct.Track
		if err = proto.Unmarshal(msgbf, &track); err != nil {
			common.ErrorLog("ReadTrackFromProto err:", err)
		}
		//common.InfoLog("read one", track.Vin, track.Date, track.Tid, len(track.TrackSegs))
		res = append(res, &track)
	}

	return res
}
