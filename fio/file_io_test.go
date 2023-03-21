package fio

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func destroyFile(name string) {
	err := os.RemoveAll(name)
	if err != nil {
		panic(err)
	}
}

func TestNewFileIOManager(t *testing.T) {
	path := filepath.Join("/tmp", "a.data")
	file, err := NewFileIOManager(path)
	defer destroyFile(path)

	assert.Nil(t, err)
	assert.NotNil(t, file)
}

func TestFileIO_Write(t *testing.T) {
	path := filepath.Join("/tmp", "a.data")
	file, err := NewFileIOManager(path)
	defer destroyFile(path)

	assert.Nil(t, err)
	assert.NotNil(t, file)

	n, err := file.Write([]byte(""))
	assert.Nil(t, err)
	assert.Equal(t, 0, n)

	n, err = file.Write([]byte("bitcask kv"))
	assert.Nil(t, err)
	assert.Equal(t, 10, n)
	n, err = file.Write([]byte("storage"))
	assert.Nil(t, err)
	assert.Equal(t, 7, n)
}

func TestFileIO_Read(t *testing.T) {
	path := filepath.Join("/tmp", "a.data")
	file, err := NewFileIOManager(path)
	defer destroyFile(path)

	assert.Nil(t, err)
	assert.NotNil(t, file)

	_, err = file.Write([]byte("key-a"))
	assert.Nil(t, err)

	_, err = file.Write([]byte("key-b"))
	assert.Nil(t, err)

	b1 := make([]byte, 5)
	n, err := file.Read(b1, 0)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte("key-a"), b1)

	b2 := make([]byte, 5)
	n, err = file.Read(b2, 5)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte("key-b"), b2)
}

func TestFileIO_Sync(t *testing.T) {
	path := filepath.Join("/tmp", "a.data")
	file, err := NewFileIOManager(path)
	defer destroyFile(path)
	
	assert.Nil(t, err)
	assert.NotNil(t, file)

	err = file.Sync()
	assert.Nil(t, err)
}

func TestFileIO_Close(t *testing.T) {
	file, err := NewFileIOManager(filepath.Join("/tmp", "0001.data"))
	assert.Nil(t, err)
	assert.NotNil(t, file)

	err = file.Close()
	assert.Nil(t, err)
}
