package main

import (
	"encoding/json"
	"fmt"
	"sync"
	// cmap "github.com/orcaman/concurrent-map"
)

/*
github.com/orcaman/concurrent-map/concurrent_map.go
concurrent-map  通过分片的方式(shard)减少 锁竞争。
分片使用 fnv32 哈希算法
*/

var SHARD_COUNT = 32

type ConcurrentMap []*ConcurrentMapShared

type ConcurrentMapShared struct {
	items map[string]interface{}
	sync.RWMutex
}

func New() ConcurrentMap {
	m := make(ConcurrentMap, SHARD_COUNT)
	for i := 0; i < SHARD_COUNT; i++ {
		m[i] = &ConcurrentMapShared{items: make(map[string]interface{})}
	}
	return m
}

// 使用到的 hash 算法
func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

func (m ConcurrentMap) GetShard(key string) *ConcurrentMapShared {
	return m[uint(fnv32(key))%uint(SHARD_COUNT)]
}

// set 多个
func (m ConcurrentMap) Mset(data map[string]interface{}) {
	for k, v := range data {
		shard := m.GetShard(k)
		shard.Lock()
		shard.items[k] = v
		shard.Unlock()
	}

}

func (m ConcurrentMap) Set(key string, value interface{}) {
	shard := m.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

type UpsertCb func(exist bool, valueInMap interface{}, newValue interface{}) interface{}

// insert or update
func (m ConcurrentMap) Upsert(key string, value interface{}, cb UpsertCb) (res interface{}) {
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	res = cb(ok, v, value)
	shard.items[key] = res
	shard.Unlock()
	return res
}

// 没有 key 就设置
func (m ConcurrentMap) SetIfAbsent(key string, value interface{}) bool {
	shard := m.GetShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return !ok
}

func (m ConcurrentMap) Get(key string) (interface{}, bool) {
	shard := m.GetShard(key)
	shard.RLock()
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

// 返回元素个数
func (m ConcurrentMap) Count() int {
	cnt := 0
	for i := 0; i < SHARD_COUNT; i++ {
		shard := m[i]
		shard.RLock()
		cnt += len(shard.items)
		shard.RUnlock()
	}
	return cnt
}

func (m ConcurrentMap) Has(key string) bool {
	shard := m.GetShard(key)
	shard.RLock()
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

func (m ConcurrentMap) Remove(key string) {
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

type RemoveCb func(key string, v interface{}, exists bool) bool

func (m ConcurrentMap) RemoveCb(key string, cb RemoveCb) bool {
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	remove := cb(key, v, ok)
	if remove && ok {
		delete(shard.items, key)
	}
	shard.Unlock()
	return remove
}

func (m ConcurrentMap) Pop(key string) (v interface{}, exists bool) {
	shard := m.GetShard(key)
	shard.Lock()
	val, ok := shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return val, ok
}

func (m ConcurrentMap) IsEmpty() bool {
	return m.Count() == 0
}

type Tuple struct {
	Key string
	Val interface{}
}

func (m ConcurrentMap) IterBuffered() <-chan Tuple {
	chans := snapshot(m)
	total := 0
	for _, c := range chans {
		total += cap(c)
	}
	ch := make(chan Tuple, total)
	go fanIn(chans, ch)
	return ch
}

func snapshot(m ConcurrentMap) (chans []chan Tuple) {
	chans = make([]chan Tuple, SHARD_COUNT)
	wg := sync.WaitGroup{}
	wg.Add(SHARD_COUNT)
	for idx, shard := range m {
		go func(idx int, shard *ConcurrentMapShared) {
			shard.RLock()
			chans[idx] = make(chan Tuple, len(shard.items))
			wg.Done()
			for k, v := range shard.items {
				chans[idx] <- Tuple{k, v}
			}
			shard.RUnlock()
			close(chans[idx])
		}(idx, shard)
	}
	wg.Wait()
	return chans
}

func fanIn(chans []chan Tuple, out chan Tuple) {
	wg := sync.WaitGroup{}
	wg.Add(len(chans))
	for _, ch := range chans {
		go func(ch chan Tuple) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(out)
}

func (m ConcurrentMap) Items() map[string]interface{} {
	tmp := make(map[string]interface{})

	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}
	return tmp
}

type IterCb func(key string, v interface{})

func (m ConcurrentMap) IterCb(fn IterCb) {
	for idx := range m {
		// shard := (m)[idx] // TODO 注意这个? 为啥不直接用下标？
		shard := m[idx] // TODO 注意这个?
		shard.RLock()
		for k, v := range shard.items {
			fn(k, v)
		}
		shard.RUnlock()
	}
}

func (m ConcurrentMap) Keys() []string {
	count := m.Count()
	ch := make(chan string, count)

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(SHARD_COUNT)
		for _, shard := range m {
			go func(shard *ConcurrentMapShared) {
				shard.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.RUnlock()
				wg.Done()
			}(shard)
		}
		wg.Wait()
		close(ch)
	}()

	keys := make([]string, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}

func (m ConcurrentMap) MarshaJSON() ([]byte, error) {
	tmp := make(map[string]interface{})

	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}

func main() {
	// 创建一个新的 map.
	// m := cmap.New()
	m := New()

	// 设置变量m一个键为“foo”值为“bar”键值对
	m.Set("foo", "======")

	// 从m中获取指定键值.
	if tmp, ok := m.Get("foo"); ok {
		bar := tmp.(string)
		fmt.Println(bar)
	}

	// 删除键为“foo”的项
	m.Remove("foo")
}
