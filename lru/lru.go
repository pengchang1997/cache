package lru

import "container/list"

// Value接口使用len函数计算其占用的字节
type Value interface {
	Len() int
}

// 链表节点
type entry struct {
	key   string
	value Value
}

// LRU cache
type Cache struct {
	// 缓存的最大容量（单位为字节）
	capacity int64

	// 已使用的缓存空间（单位为字节）
	size int64

	// 双向链表
	doubleLinkedList *list.List

	// 存储key与链表节点映射关系的哈希表
	cache map[string]*list.Element

	// 可选的回调函数，在发生缓存条目清除时被执行
	OnEvicted func(key string, value Value)
}

// 实例化LRU cache
func New(capacity int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		capacity:         capacity,
		doubleLinkedList: list.New(),
		cache:            make(map[string]*list.Element),
		OnEvicted:        onEvicted,
	}
}

// 实现查找功能
func (c *Cache) Get(key string) (value Value, ok bool) {
	// 如果在哈希表中查找到了key
	if element, ok := c.cache[key]; ok {
		// 将对应的链表节点移动到链表最前面
		c.doubleLinkedList.MoveToFront(element)

		// 获取链表节点存储的键值对
		keyValue := element.Value.(*entry)

		// 返回value
		return keyValue.value, true
	}

	return
}

// 实现缓存淘汰功能
func (c *Cache) RemoveOldest() {
	// 获取尾节点
	oldest := c.doubleLinkedList.Back()

	if oldest != nil {
		// 从链表中删除节点
		c.doubleLinkedList.Remove(oldest)

		// 获取链表节点存储的键值对
		keyValue := oldest.Value.(*entry)

		// 获取key
		key := keyValue.key

		// 从哈希表中删除key对应的记录
		delete(c.cache, key)

		// 更新缓存大小
		c.size -= int64(len(keyValue.key)) + int64(keyValue.value.Len())

		// 调用回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(key, keyValue.value)
		}
	}
}

// 实现新增与修改功能
func (c *Cache) Add(key string, value Value) {
	// 如果在哈希表中查找到了key
	if element, ok := c.cache[key]; ok {
		// 将链表节点移动到链表最前面
		c.doubleLinkedList.MoveToFront(element)

		// 获取链表节点对应的键值对
		keyValue := element.Value.(*entry)

		// 更新缓存大小
		c.size = c.size - int64(keyValue.value.Len()) + int64(value.Len())

		// 更新键值对
		keyValue.value = value
	} else {
		// 如果没有在哈希表中查找到key，则先新建一个节点并插入到链表最前面
		element := c.doubleLinkedList.PushFront(&entry{key: key, value: value})

		// 在哈希表中建立映射关系
		c.cache[key] = element

		// 更新缓存大小
		c.size += int64(len(key)) + int64(value.Len())
	}

	// 如果缓存大小大于缓存容量，则持续移除最近最少访问的节点
	for c.capacity != 0 && c.capacity < c.size {
		c.RemoveOldest()
	}
}

// 获取缓存的条目数量
func (c *Cache) Len() int {
	return c.doubleLinkedList.Len()
}
