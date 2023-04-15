package lfu

import (
	"container/heap"
	"github.com/zhaoziliang2019/cache"
)

//lfu是一个LFU cache，它不是并发安全的
type lfu struct {
	//缓存最大的容易，单位字节
	//groupcache 使用的是最大存放entry个数
	maxBytes int
	//当一个entry从缓存中移除时调用该回调函数，默认为nil
	//groupcache 中的key是任意的可比较类型，value是interface{}
	onEvicted func(key string, value interface{})
	//已使用的字节数，只包括值，key不算
	usedBytes int
	queue     *queue
	cache     map[string]*entry
}

//用New方法创建一个新的cache，如果maxBytes是0，则表示没有容量限制
func New(maxBytes int, onExvicted func(key string, value interface{})) cache.Cache {
	q := make(queue, 0, 1024)
	return &lfu{maxBytes: maxBytes, onEvicted: onExvicted, queue: &q, cache: make(map[string]*entry)}
}

//用set方法cache中增加一个元素（如果已经存在，则更新值，并增加权重，重新构建堆）
func (l *lfu) Set(key string, value interface{}) {
	if e, ok := l.cache[key]; ok {
		l.usedBytes = l.usedBytes - cache.CalcLen(e.value) + cache.CalcLen(value)
		l.queue.update(e, value, e.weight)
		return
	}
	en := &entry{key: key, value: value}
	heap.Push(l.queue, en)
	l.cache[key] = en
	l.usedBytes += en.Len()
	if l.maxBytes > 0 && l.usedBytes > l.maxBytes {
		l.removeElement(heap.Pop(l.queue))
	}
}
func (q *queue) update(en *entry, value interface{}, weight int) {
	en.value = value
	en.weight = weight
	heap.Fix(q, en.index)
}

//Get 方法会从cache中获取key对应的值，nil表示key不存在
func (l *lfu) Get(key string) interface{} {
	if e, ok := l.cache[key]; ok {
		l.queue.update(e, e.value, e.weight)
		return e.value
	}
	return nil
}

//Del 方法会从cache中删除key对应的元素
func (l *lfu) Del(key string) {
	if e, ok := l.cache[key]; ok {
		heap.Remove(l.queue, e.index)
		l.removeElement(e)
	}
}

//DelOldest 方法会从cache中删除最最旧的记录
func (l *lfu) DelOldest() {
	if l.queue.Len() == 0 {
		return
	}
	l.removeElement(heap.Pop(l.queue))
}
func (l *lfu) removeElement(x interface{}) {
	if x == nil {
		return
	}
	en := x.(*entry)
	delete(l.cache, en.key)
	l.usedBytes -= en.Len()
	if l.onEvicted != nil {
		l.onEvicted(en.key, en.value)
	}
}

//Len 会返回当前cache中的记录数
func (l *lfu) Len() int {
	return l.queue.Len()
}
