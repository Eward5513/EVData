package proto_tools

import (
	"EVdata/common"
	"EVdata/proto_struct"
	"bufio"
	"encoding/binary"
	"google.golang.org/protobuf/proto"
	"os"
)

var (
	WriterBuffer = 1024 * 4
)

type ProtoBufWriter struct {
	fileHandle *os.File
	filename   string
	buf        *bufio.Writer
}

func NewProtoBufWriter(f string) *ProtoBufWriter {
	file, err := os.Create(f)
	if err != nil {
		common.ErrorLog("error when create file", err)
	}
	b := bufio.NewWriterSize(file, WriterBuffer)

	return &ProtoBufWriter{filename: f, fileHandle: file, buf: b}
}

func (w *ProtoBufWriter) Write(p *proto_struct.Track) {
	data, err := proto.Marshal(p)
	if err != nil {
		common.FatalLog("序列化错误: %v", err)
	}
	if err := binary.Write(w.buf, binary.LittleEndian, uint32(len(data))); err != nil {
		common.ErrorLog("error when writing length", err)
	}
	if _, err = w.buf.Write(data); err != nil {
		common.ErrorLog("写入文件错误", err)
	}
}

func (w *ProtoBufWriter) Close() {
	defer w.fileHandle.Close()

	_ = w.buf.Flush()
	//common.InfoLog("Finish writing file", w.filename)
}
