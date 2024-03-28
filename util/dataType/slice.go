package dataType

type Slice[K comparable, V any] []V

func (s *Slice[K, V]) ToMap(getKey func(V) K) (result map[K]V) {
	result = make(map[K]V)
	for _, v := range *s {
		result[getKey(v)] = v
	}
	return
}
func (s *Slice[K, V]) ExtractValues(getVal func(V) K) (result []K) {
	result = make([]K, len(*s), len(*s))
	for i, v := range *s {
		result[i] = getVal(v)
	}
	return
}
