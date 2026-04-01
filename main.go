package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Misaka-10233/discache"
)

func main() {
	// 测试基本功能
	testBasicCache()
	
	// 测试缓存淘汰
	testCacheEviction()
	
	// 测试并发访问
	testConcurrentAccess()
}

func testBasicCache() {
	log.Println("=== 测试基本缓存功能 ===")
	
	// 创建缓存，最大容量100字节
	cache := discache.NewCache(100, func(key string, value []byte) {
		log.Printf("缓存淘汰: %s -> %s", key, value)
	})
	
	// 测试获取缓存（第一次从DB获取）
	value, found := cache.Get("a")
	fmt.Printf("获取键 'a': %s, 找到: %t\n", value, found)
	
	// 再次获取（应该从缓存获取）
	value, found = cache.Get("a")
	fmt.Printf("再次获取键 'a': %s, 找到: %t\n", value, found)
	
	// 测试不存在的键
	value, found = cache.Get("d")
	fmt.Printf("获取不存在的键 'd': %s, 找到: %t\n", value, found)
	
	fmt.Printf("缓存大小: %d, 缓存项数: %d\n", cache.Size(), cache.Len())
	log.Println()
}

func testCacheEviction() {
	log.Println("=== 测试缓存淘汰功能 ===")
	
	// 创建一个较小容量的缓存，确保会触发淘汰
	cache := discache.NewCache(50, func(key string, value []byte) {
		log.Printf("缓存淘汰: %s -> %s", key, value)
	})
	
	// 填充缓存
	cache.Get("a") // 从DB获取
	cache.Get("b") // 从DB获取
	cache.Get("c") // 从DB获取
	
	fmt.Printf("填充后 - 缓存大小: %d, 缓存项数: %d\n", cache.Size(), cache.Len())
	
	// 再次访问a，使其成为最近使用的
	cache.Get("a")
	
	// 访问b，使其成为最近使用的
	cache.Get("b")
	
	// 此时c应该是最久未使用的，会被淘汰
	fmt.Printf("访问后 - 缓存大小: %d, 缓存项数: %d\n", cache.Size(), cache.Len())
	log.Println()
}

func testConcurrentAccess() {
	log.Println("=== 测试并发访问 ===")
	
	cache := discache.NewCache(100, func(key string, value []byte) {
		log.Printf("缓存淘汰: %s -> %s", key, value)
	})
	
	var wg sync.WaitGroup
	const goroutines = 10
	const accesses = 5
	
	// 启动多个goroutine并发访问缓存
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < accesses; j++ {
				key := string('a' + (id+j)%3)
				value, found := cache.Get(key)
				fmt.Printf("Goroutine %d 访问键 '%s': %s, 找到: %t\n", id, key, value, found)
				time.Sleep(10 * time.Millisecond) // 模拟处理时间
			}
		}(i)
	}
	
	wg.Wait()
	fmt.Printf("并发测试后 - 缓存大小: %d, 缓存项数: %d\n", cache.Size(), cache.Len())
	log.Println()
}
