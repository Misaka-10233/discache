package discache

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"
)

// var (
// 	ERR_MARSHAL_FAILED = fmt.Errorf("marshal failed")
// )

type Server struct {
	nodeId string
	ctx    context.Context
	cache  *Cache
	server *http.Server
}

func NewServer(nodeId string, port int, maxByte int64, onEvicted func(key string, value []byte), onMissed func(key string) ([]byte, error)) *Server {
	server := &Server{}
	cache := NewCache(maxByte, onEvicted, onMissed)
	s := http.NewServeMux()
	s.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		// 获取key
		key := r.URL.Query().Get("key")
		// 检查key是否为空
		if key == "" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("key is empty"))
			log.Println("Server error: ", err)
			return
		}
		// 从缓存中获取值
		value, ok := cache.Get(key)
		if ok {
			w.Header().Set("Content-Type", "binary/octet-stream")
			w.Header().Set("Content-Length", strconv.FormatInt(int64(len(value)), 10))
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(value)
			if err != nil {
				log.Println("Server error: ", err)
			}
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("key not found"))
			if err != nil {
				log.Println("Server error: ", err)
			}
		}
	})

	// s.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
	// 	defer r.Body.Close()
	// 	data, err := io.ReadAll(r.Body)
	// 	if err != nil {
	// 		w.Header().Set("Content-Type", "text/plain")
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		_, err = w.Write([]byte("Add failed"))
	// 		if err != nil {
	// 			log.Println("Server error: ", err)
	// 		}
	// 		return
	// 	}
	// 	// 解析protobuf
	// 	var pair Pair
	// 	err = proto.Unmarshal(data, &pair)
	// 	if err != nil {
	// 		w.Header().Set("Content-Type", "text/plain")
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		_, err = w.Write([]byte("Add failed"))
	// 		if err != nil {
	// 			log.Println("Server error: ", err)
	// 		}
	// 		return
	// 	}
	// 	// 添加到缓存
	// 	cache.Add(pair.GetKey(), pair.GetValue())
	// 	w.Header().Set("Content-Type", "text/plain")
	// 	w.WriteHeader(http.StatusOK)
	// 	_, err = w.Write([]byte("Add success"))
	// 	if err != nil {
	// 		log.Println("Server error: ", err)
	// 	}
	// })

	s.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		stats := server.generateStats()
		data, err := proto.Marshal(stats)
		if err != nil {
			log.Println("Server error: ", err)
		}
		_, err = w.Write(data)
		if err != nil {
			log.Println("Server error: ", err)
		}
	})

	server.nodeId = nodeId
	server.ctx = context.Background()
	server.cache = cache
	server.server = &http.Server{
		Addr:    "127.0.0.1:" + strconv.Itoa(port),
		Handler: s,
	}

	return server
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Stop() {
	err := s.server.Shutdown(s.ctx)
	if err != nil {
		log.Printf("Server Stop err %v", err)
	}
}

func (s *Server) generateStats() *Stats {
	return &Stats{
		NodeId:       s.nodeId,
		Status:       "healthy",
		CacheSize:    s.cache.Size(),
		CacheCount:   int32(s.cache.Len()),
		RequestCount: s.cache.GetRequestCount(),
		HitCount:     s.cache.GetHitCount(),
		Timestamp:    time.Now().Unix(),
	}
}
