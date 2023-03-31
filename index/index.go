package index

import (
	"bitcask-go/data"
	"bytes"
	"github.com/google/btree"
)

// Indexer 抽象索引接口，后续如果想要接入其他的数据结构，则直接实现这个接口即可
type Indexer interface {
	// Put 向索引中存储key对应的数据位置信息
	Put(key []byte, pos *data.LogRecordPos) bool
	// Get 根据key取出对应的索引位置信息
	Get(key []byte) *data.LogRecordPos
	// Delete 根据key删除对应的索引位置信息
	Delete(key []byte) bool
	// Size 索引中的数据量
	Size() int
	// Iterator 索引迭代器
	Iterator(reverse bool) Iterator
}

type IndexType = int8

const (
	Btree IndexType = iota + 1
	ART
)

// NewIndexer 根据类型初始化索引
func NewIndexer(indexType IndexType) Indexer {
	switch indexType {
	case Btree:
		return NewBTree()
	case ART:
		return nil
	default:
		panic("unsupported index type")
	}
}

type Item struct {
	key []byte
	pos *data.LogRecordPos
}

func (ai *Item) Less(bi btree.Item) bool {
	return bytes.Compare(ai.key, bi.(*Item).key) == -1
}

// Iterator 通用索引迭代器
type Iterator interface {
	// Rewind 重新回到迭代器的起点，即第一个数据
	Rewind()
	// Seek 根据传入的key查找到第一个大于（或小于）等于的目标key， 根据从这个key开始遍历
	Seek(key []byte)
	// Next 跳转到下一个key
	Next()
	// Valid 是否有效，即是否已经遍历完了所有的key，用于退出遍历
	Valid() bool
	// Key 当前遍历位置的key数据
	Key() []byte
	// Value 当前遍历位置的Value数据
	Value() *data.LogRecordPos
	// Close 关闭迭代器，释放相应资源
	Close()
}
