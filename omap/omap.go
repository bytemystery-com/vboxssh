package omap

import "slices"

type OMap[K comparable, V any] struct {
	m map[K]V
	l []K
}

func NewOMap[K comparable, V any](number int) OMap[K, V] {
	om := OMap[K, V]{
		m: make(map[K]V, number),
		l: make([]K, 0, number),
	}
	return om
}

func (om *OMap[K, V]) Add(key K, value V) {
	om.m[key] = value
	om.l = append(om.l, key)
}

func (om *OMap[K, V]) Insert(index int, key K, value V) {
	om.m[key] = value
	om.l = slices.Insert(om.l, index, key)
}

func (om *OMap[K, V]) RemoveByIndex(index int) {
	k := om.l[index]
	om.l = slices.Delete(om.l, index, index+1)
	if slices.Index(om.l, k) < 0 {
		delete(om.m, k)
	}
}

func (om *OMap[K, V]) RemoveByKey(key K) {
	delete(om.m, key)
	l := make([]K, 0, len(om.l))
	for i := range len(om.l) {
		if om.l[i] != key {
			l = append(l, om.l[i])
		}
	}
	om.l = l
}

func (om *OMap[K, V]) ReplaceByKey(key K, value V) {
	om.m[key] = value
}

func (om *OMap[K, V]) ReplaceByIndex(index int, value V) {
	om.m[om.l[index]] = value
}

func (om *OMap[K, V]) Clear() {
	clear(om.m)
	om.l = om.l[:0]
}

func (om *OMap[K, V]) GetByKey(key K) V {
	return om.m[key]
}

func (om *OMap[K, V]) GetByIndex(index int) V {
	return om.m[om.l[index]]
}

func (om *OMap[K, V]) GetValues() []V {
	list := make([]V, 0, len(om.l))
	for i := range len(om.l) {
		list = append(list, om.m[om.l[i]])
	}
	return list
}

func (om *OMap[K, V]) Sort(cmp func(a, b V) int) {
	slices.SortFunc(om.l, func(a, b K) int {
		return cmp(om.m[a], om.m[b])
	})
}

func (om *OMap[K, V]) Len() int {
	return len(om.l)
}

/*
func (om *OMap[K, V]) Find(value V) []int {
	list := make([]int, 0, len(om.l))
	for i := range len(om.l) {
		if m.m[om.l[i]] == value {
			list = append(list, i)
		}
	}
	return list
}
*/
