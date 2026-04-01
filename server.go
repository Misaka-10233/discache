package discache

import (
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	cache  *Cache
	server *http.Server
}

func NewServer(port int, maxByte int64, onEvicted func(key string, value []byte)) *Server {
	cache := NewCache(maxByte, onEvicted)
	s := http.NewServeMux()
	s.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain")
			_, err := w.Write([]byte("key is empty"))
			if err != nil {
				log.Println(err)
				return
			}
			return
		}
		value, ok := cache.Get(key)
		if ok {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "binary/octet-stream")
			w.Header().Set("Content-Length", strconv.FormatInt(int64(len(value)), 10))
			_, err := w.Write(value)
			if err != nil {
				log.Println(err)
				return
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "text/plain")
			_, err := w.Write([]byte("key not found"))
			if err != nil {
				log.Println(err)
				return
			}
		}
	})
	return &Server{
		cache: cache,
		server: &http.Server{
			Addr:    ":" + strconv.Itoa(port),
			Handler: s,
		},
	}
}
