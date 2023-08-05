package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int
	keys     []int
	hashMap  map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(inputKeys ...string) {
	for _, realkey := range inputKeys {
		for i := 0; i < m.replicas; i++ {
			hashVal := int(m.hash([]byte(strconv.Itoa(i) + realkey)))
			m.keys = append(m.keys, hashVal)
			m.hashMap[hashVal] = realkey
		}
	}
}

func (m *Map) Get(key string) string {
	if len(key) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	//如果hash过大，索引==len(m.keys)，由于是哈希环idx此时应该取模为0
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
