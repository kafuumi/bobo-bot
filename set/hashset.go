// Package set
//哈希表实现的集合
package set

// HashSet 基于哈希表的集合
type HashSet[T comparable] struct {
	items map[T]struct{}
}

//New 创建一个空的 HashSet
func New[T comparable]() *HashSet[T] {
	return &HashSet[T]{
		items: make(map[T]struct{}),
	}
}

//NewSlice 根据参数 s 创建 HashSet
func NewSlice[T comparable](s []T) *HashSet[T] {
	set := New[T]()
	for _, item := range s {
		set.Add(item)
	}
	return set
}

// Add 向集合中添加元素
func (h *HashSet[T]) Add(item ...T) {
	for _, e := range item {
		h.items[e] = struct{}{}
	}
}

//Remove 移除集合中的某个元素
func (h *HashSet[T]) Remove(item ...T) {
	for _, e := range item {
		delete(h.items, e)
	}
}

//Clear 清空集合
func (h *HashSet[T]) Clear() {
	h.items = make(map[T]struct{})
}

//Contains 判断集合中是否包含某个元素
func (h *HashSet[T]) Contains(item ...T) bool {
	for _, e := range item {
		if _, ok := h.items[e]; !ok {
			return false
		}
	}
	return true
}

// Len 获取集合中的元素个数
func (h *HashSet[T]) Len() int {
	return len(h.items)
}
