package data

import (
	"encoding/binary"
	"hash/crc32"
)

type LogRecordType = byte

const (
	LogRecordNormal LogRecordType = iota
	LogRecordDeleted
	LogRecordTxnFinished
)

// crc type keySize valueSize
// 4 + 1 + 5（变长） + 5（变长） = 15
const maxLOgRecordHeaderSize = binary.MaxVarintLen32*2 + 5

// LogRecord 写入到数据文件的记录
// 之所以叫日志，因为数据文件中的记录是追加写入的，类似于日志
type LogRecord struct {
	Key   []byte
	Value []byte
	Type  LogRecordType
}

// logRecord 的头部信息
type logRecordHeader struct {
	crc        uint32        // crc校验值
	recordType LogRecordType // 标识LogRecord的类型
	keySize    uint32        // key的长度
	valueSize  uint32        // value的长度
}

// LogRecordPos 数据内存索引，主要是描述数据在磁盘上的位置
type LogRecordPos struct {
	// 文件id，表示将数据存储到了哪个文件当中
	Fid uint32
	// 偏移量
	Offset int64
	// 标识数据在磁盘上的大小
	Size uint32
}

// TransactionRecord 暂存事务相关数据
type TransactionRecord struct {
	Record *LogRecord
	Pos    *LogRecordPos
}

// EncodeLogRecord LogRecord进行编码，返回字节数组及长度
// +------------+------------+------------+------------+------------+------------+
// |  crc校验值  |   type类型  |   key size | value size |      key   |   value    |
// +------------+------------+------------+------------+------------+------------+
//
//	4字节        1字节      变长（最大5）  变长（最大5）     变长          变长
func EncodeLogRecord(record *LogRecord) ([]byte, int64) {
	// 初始化一个header部分的字节数组
	header := make([]byte, maxLOgRecordHeaderSize)

	// 第五个字节存储 Type
	header[4] = record.Type
	var index = 5
	// 5 字节之后，存储的是key 和 value的长度信息
	// 使用变长类型，节省空间
	index += binary.PutVarint(header[index:], int64(len(record.Key)))
	index += binary.PutVarint(header[index:], int64(len(record.Value)))

	var size = index + len(record.Key) + len(record.Value)
	encBytes := make([]byte, size)

	// 将header部分的内容拷贝过来
	copy(encBytes[:index], header[:index])
	// 将 key 和 value 数据拷贝到字节数组中
	copy(encBytes[index:], record.Key)
	copy(encBytes[index+len(record.Key):], record.Value)

	// 对整个 LogRecord的数据进行校验
	crc := crc32.ChecksumIEEE(encBytes[4:])
	// 小断续
	binary.LittleEndian.PutUint32(encBytes[:4], crc)

	return encBytes, int64(size)
}

// EncodeLogRecordPos 对位置信息进行编码
func EncodeLogRecordPos(pos *LogRecordPos) []byte {
	buf := make([]byte, binary.MaxVarintLen32+binary.MaxVarintLen64)
	var index = 0
	index += binary.PutVarint(buf[index:], int64(pos.Fid))
	index += binary.PutVarint(buf[index:], pos.Offset)
	index += binary.PutVarint(buf[index:], int64(pos.Size))
	return buf[:index]
}

// DecodeLogRecordPos 解码LogRecordPos
func DecodeLogRecordPos(buf []byte) *LogRecordPos {
	var index = 0
	fileId, n := binary.Varint(buf[index:])
	index += n
	offset, n := binary.Varint(buf[index:])
	index += n
	size, _ := binary.Varint(buf[index:])

	return &LogRecordPos{
		Fid:    uint32(fileId),
		Offset: offset,
		Size:   uint32(size),
	}
}

// 对字节数组中的header信息进行解码
func decodeLogRecordHeader(buf []byte) (*logRecordHeader, int64) {
	// 连crc的长度都不够
	if len(buf) <= 4 {
		return nil, 0
	}

	header := &logRecordHeader{
		crc:        binary.LittleEndian.Uint32(buf[:4]),
		recordType: buf[4],
	}
	var index = 5
	// 取出实际的 key size
	keySize, n := binary.Varint(buf[index:])
	header.keySize = uint32(keySize)
	index += n

	// 取出实际的value size
	valueSize, n := binary.Varint(buf[index:])
	header.valueSize = uint32(valueSize)
	index += n

	return header, int64(index)
}

func getLogRecordCRC(record *LogRecord, header []byte) uint32 {
	if record == nil {
		return 0
	}
	crc := crc32.ChecksumIEEE(header[:])
	crc = crc32.Update(crc, crc32.IEEETable, record.Key)
	crc = crc32.Update(crc, crc32.IEEETable, record.Value)
	return crc
}
