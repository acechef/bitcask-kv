package fio

import (
	"golang.org/x/exp/mmap"
	"os"
)

// MMap IO 内存文件映射
type MMap struct {
	readerAt *mmap.ReaderAt
}

// NewMMapIOManager 初始化MMap
func NewMMapIOManager(fileName string) (*MMap, error) {
	_, err := os.OpenFile(fileName, os.O_CREATE, DataFilePerm)
	if err != nil {
		return nil, err
	}
	readerAt, err := mmap.Open(fileName)
	if err != nil {
		return nil, err
	}
	return &MMap{readerAt: readerAt}, nil
}

func (M *MMap) Read(bytes []byte, i int64) (int, error) {
	return M.readerAt.ReadAt(bytes, i)
}

func (M *MMap) Write(bytes []byte) (int, error) {
	panic("implement me")
}

func (M *MMap) Sync() error {
	panic("implement me")
}

func (M *MMap) Close() error {
	return M.readerAt.Close()
}

func (M *MMap) Size() (int64, error) {
	return int64(M.readerAt.Len()), nil
}
