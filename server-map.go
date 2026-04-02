package discache

import (
	"log"
	"slices"
	"sort"
	"strconv"
)

type ServerMap struct {
	nodeReplica int
	dataReplica int
	hashMap     map[uint64]string
	keys        []uint64
}

func NewServerMap(nodeReplica int, dataReplica int, url ...string) *ServerMap {
	if len(url) == 0 || nodeReplica <= 0 {
		log.Printf("ServerMap NewServerMap url is empty")
		return nil
	}
	s := &ServerMap{
		nodeReplica: nodeReplica,
		dataReplica: dataReplica,
		hashMap:     make(map[uint64]string),
		keys:        make([]uint64, 0),
	}
	for _, u := range url {
		s.RegisterServer(u)
	}
	return s
}

func (s *ServerMap) RegisterServer(url string) {
	for i := range s.nodeReplica {
		hash := HashString(url + strconv.Itoa(i))
		s.hashMap[hash] = url
		index := sort.Search(len(s.keys), func(i int) bool {
			return s.keys[i] >= hash
		})
		s.keys = slices.Insert(s.keys, index, hash)
	}
}

func (c *ServerMap) GetServerUrls(key string) []string {
	hash := HashString(key)
	index := sort.Search(len(c.keys), func(i int) bool {
		return c.keys[i] >= hash
	})
	urls := make([]string, 0, c.dataReplica)
	seen := make(map[string]bool)
	for i := 0; i < len(c.keys) && len(urls) < c.dataReplica; i++ {
		idx := (index + i) % len(c.keys)
		url := c.hashMap[c.keys[idx]]
		if !seen[url] {
			seen[url] = true
			urls = append(urls, url)
		}
	}
	return urls
}
