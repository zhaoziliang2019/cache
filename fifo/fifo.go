package fifo

import (
	"container/list"
	"github.com/zhaoziliang2019/cache"
)

//fifo是一个FIFO cache，它不是并发安全的
type fifo struct {
	//缓存最大的容量，单位字节
	//groupcache 使用的是最大存放entry个数
	maxBytes int
	//当一个entry从缓存中移除时调用该回调函数，默认为nil
	//groupcache中的key时任意的可比较类型：value是interface{}
	onEvicted func(key string, value interface{})
	//已使用的字节数，只包括值，key不算
	usedBytes int
	li        *list.List
	cache     map[string]*list.Element
}

type entry struct {
	key   string
	value interface{}
}

func (e *entry) Len() int {
	return cache.CalcLen(e.value)
}
func New(maxBytes int, onEvicted func(key string, value interface{})) cache.Cache {
	return &fifo{
		maxBytes:  maxBytes,
		onEvicted: onEvicted,
		li:        list.New(),
		cache:     make(map[string]*list.Element),
	}
}

//通过Set方法往Cache尾部增加一个元素（如果已经存在，则移到尾部，并修改值）
func (f *fifo) Set(key string, value interface{}) {
	if e, ok := f.cache[key]; ok {
		f.li.MoveToBack(e)
		en := e.Value.(*entry)
		f.usedBytes = f.usedBytes - cache.CalcLen(en.value) + cache.CalcLen(value)
		e.Value = value
		return
	}
	en := &entry{key: key, value: value}
	e := f.li.PushBack(en)
	f.cache[key] = e
	f.usedBytes += en.Len()
	if f.maxBytes > 0 && f.usedBytes > f.maxBytes {
		f.DelOldest()
	}
}

//Get方法会从cache中获取key对应的值，nil表示key不存在
func (f *fifo) Get(key string) interface{} {
	if e, ok := f.cache[key]; ok {
		return e.Value.(*entry).value
	}
	return nil
}

//Del 方法会从cache中删除key对应的记录
func (f *fifo) Del(key string) {
	if e, ok := f.cache[key]; ok {
		f.removeElement(e)
	}
}

//DelOldest方法会从cache中删除最旧的记录
func (f *fifo) DelOldest() {
	f.removeElement(f.li.Front())
}
func (f *fifo) removeElement(e *list.Element) {
	if e == nil {
		return
	}
	f.li.Remove(e)
	en := e.Value.(*entry)
	f.usedBytes -= en.Len()
	delete(f.cache, en.key)
	if f.onEvicted != nil {
		f.onEvicted(en.key, en.value)
	}
}

//Len返回当前cache中的记录数
func (f *fifo) Len() int {
	return f.li.Len()
}
