package data

import (
	"bitcask-go/fio"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"path/filepath"
)

var (
	ErrInvalidCRC = errors.New("invalid crc value")
)

const DataFileNameSuffix = ".data"
const HintFileName = "hint-index"
const HintFinishedFileName = "merge-finished"

type DataFile struct {
	FileId    uint32        // 文件id
	WriteOff  int64         // 文件写到了哪个位置
	IOManager fio.IOManager // io读写管理
}

// OpenDataFile 打开新的数据文件
func OpenDataFile(dirPath string, fileId uint32) (*DataFile, error) {
	fileName := GetDataFileName(dirPath, fileId)
	return newDataFile(fileName, fileId)
}

// OpenHintFile 打开Hint索引文件
func OpenHintFile(dirPath string) (*DataFile, error) {
	fileName := filepath.Join(dirPath, HintFileName)
	return newDataFile(fileName, 0)
}

// OpenMergeFinishedFile 打开标识merge完成的文件
func OpenMergeFinishedFile(dirPath string) (*DataFile, error) {
	fileName := filepath.Join(dirPath, HintFinishedFileName)
	return newDataFile(fileName, 0)
}

func GetDataFileName(dirPath string, fileId uint32) string {
	return filepath.Join(dirPath, fmt.Sprintf("%09d", fileId)+DataFileNameSuffix)
}

func newDataFile(fileName string, fileId uint32) (*DataFile, error) {
	ioManager, err := fio.NewIOManager(fileName)
	if err != nil {
		return nil, err
	}
	return &DataFile{
		FileId:    fileId,
		WriteOff:  0,
		IOManager: ioManager,
	}, nil
}

func (df *DataFile) ReadLogRecord(offset int64) (*LogRecord, int64, error) {
	fileSize, err := df.IOManager.Size()
	if err != nil {
		return nil, 0, err
	}

	// 注意：如果读取的最大header长度已经超过了文件的长度，则直接读取到文件末尾即可
	var headerBytes int64 = maxLOgRecordHeaderSize
	if offset+maxLOgRecordHeaderSize > fileSize {
		headerBytes = fileSize - offset
	}

	// 读取header信息
	headerBuf, err := df.readNBytes(headerBytes, offset)
	if err != nil {
		return nil, 0, err
	}

	// 下面两个条件表示读取到了文件的末尾，直接返回EOF错误
	header, headerSize := decodeLogRecordHeader(headerBuf)
	if header == nil {
		return nil, 0, io.EOF
	}
	if header.crc == 0 && header.keySize == 0 && header.valueSize == 0 {
		return nil, 0, io.EOF
	}

	// 取出对应的key 和value的长度
	keySize, valueSize := int64(header.keySize), int64(header.valueSize)
	var recordSize = headerSize + keySize + valueSize

	logRecord := &LogRecord{Type: header.recordType}
	// 开始读取用户实际存储的key/value数据
	if keySize > 0 || valueSize > 0 {
		// kvBuf为key+value的结果
		kvBuf, err := df.readNBytes(keySize+valueSize, offset+headerSize)
		if err != nil {
			return nil, 0, err
		}
		// 解出 key 和 value
		logRecord.Key = kvBuf[:keySize]
		logRecord.Value = kvBuf[keySize:]
	}
	// 校验数据的crc是否正确
	crc := getLogRecordCRC(logRecord, headerBuf[crc32.Size:headerSize])
	if crc != header.crc {
		return nil, 0, ErrInvalidCRC
	}

	return logRecord, recordSize, nil
}

func (df *DataFile) Write(buf []byte) error {
	n, err := df.IOManager.Write(buf)
	if err != nil {
		return err
	}
	// 更新 WriteOff
	df.WriteOff += int64(n)
	return nil
}

// WriteHintRecord 写入索引信息到hint文件中
func (df *DataFile) WriteHintRecord(key []byte, pos *LogRecordPos) error {
	record := &LogRecord{
		Key:   key,
		Value: EncodeLogRecordPos(pos),
	}
	encodeLogRecord, _ := EncodeLogRecord(record)
	return df.Write(encodeLogRecord)
}

func (df *DataFile) Sync() error {
	return df.IOManager.Sync()
}

func (df *DataFile) Close() error {
	return df.IOManager.Close()
}

func (df *DataFile) readNBytes(n int64, offset int64) (b []byte, err error) {
	b = make([]byte, n)
	_, err = df.IOManager.Read(b, offset)
	return
}
