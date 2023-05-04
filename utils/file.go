package utils

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// DirSize 获取一个目录的大小
func DirSize(dirPath string) (int64, error) {
	var size int64
	err := filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// AvailableDishSize 获取磁盘剩余可用空间大小
func AvailableDishSize() (uint64, error) {
	wd, err := syscall.Getwd()
	if err != nil {
		return 0, err
	}
	var stat syscall.Statfs_t
	if err = syscall.Statfs(wd, &stat); err != nil {
		return 0, err
	}
	return stat.Bavail * uint64(stat.Bsize), nil
}

// CopyDir 拷贝数据目录
func CopyDir(src, dest string, exclude []string) error {
	// 目标目录不存在则创建
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		if err := os.MkdirAll(dest, os.ModePerm); err != nil {
			return err
		}
	}
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		// "/tmp/a/11.data"获得11.data
		filename := strings.Replace(path, src, "", 1)
		if filename == "" {
			return nil
		}

		for _, e := range exclude {
			matched, err := filepath.Match(e, info.Name())
			if err != nil {
				return err
			}
			if matched {
				return nil
			}
		}
		// 如果是目录，则在目标目录创建新目录
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(dest, filename), info.Mode())
		}
		// 常规文件
		// 读取
		data, err := os.ReadFile(filepath.Join(src, filename))
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dest, filename), data, info.Mode())
	})
}
