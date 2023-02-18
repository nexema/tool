package utils

type OMap[K comparable, V any] struct {
	m    map[K]V
	keys []K
	len  int
}

func NewOMap[K comparable, V any]() *OMap[K, V] {
	return &OMap[K, V]{make(map[K]V), make([]K, 0), 0}
}

func (self *OMap[K, V]) Upsert(key K, update func(value V), def func() V) {
	val, ok := self.m[key]
	if !ok {
		self.insert(key, def())
		self.Upsert(key, update, def)
		return
	}

	update(val)
}

func (self *OMap[K, V]) Reverse(f func(k K, v V)) {
	for i := self.len - 1; i >= 0; i-- {
		key := self.keys[i]
		f(key, self.m[key])
	}
}

func (self *OMap[K, V]) insert(k K, v V) {
	self.m[k] = v
	self.keys = append(self.keys, k)
	self.len++
}
