package dataTool

// Slice的拓展方法
type Slice[K comparable, V any] []V

// ToMap slice转map
/**
示例：
	var userList dataTools.Slice[uint, User]
	userList = []User{
		{ID: 1, Name: "test1", Password: "test2"},
		{ID: 2, Name: "test2", Password: "test3"},
		{ID: 3, Name: "test3", Password: "test3"},
	} // 或者 global.GvaDb.Limit(5).Find(&userList)
	userMap := userList.ToMap(func(user User) uint {
		return user.ID
	})
	fmt.Println(userMap) //将得到 map[1:{1 test1 test2} 2:{2 test2 test3}]
*/
func (s Slice[K, V]) ToMap(getKey func(V) K) (result map[K]V) {
	result = make(map[K]V)
	for _, v := range s {
		result[getKey(v)] = v
	}
	return
}

// ExtractValues 提取slice中的值
/**
示例：
	var userList dataTools.Slice[uint, User]
	userList = []User{
		{ID: 1, Name: "test1", Password: "test2"},
		{ID: 2, Name: "test2", Password: "test3"},
	} // 或者 global.GvaDb.Limit(2).Find(&userList)
	userIds := userList.ExtractValues(func(user User) uint {
		return user.ID
	})
	fmt.Println(userIds) //将得到 [1 2]
*/
func (s Slice[K, V]) ExtractValues(getVal func(V) K) (result []K) {
	result = make([]K, len(s), len(s))
	for i, v := range s {
		result[i] = getVal(v)
	}
	return
}

func (s *Slice[K, V]) Reverse() {
	n := len(*s)
	for i := 0; i < n/2; i++ {
		(*s)[i], (*s)[n-1-i] = (*s)[n-1-i], (*s)[i]
	}
}

func (s Slice[K, V]) CopyReverse() Slice[K, V] {
	list := make(Slice[K, V], len(s), len(s))
	copy(list, s)
	list.Reverse()
	return list
}

func Reverse[V any](s []V) {
	for i := 0; i < len(s)/2; i++ {
		s[i], s[len(s)-1-i] = s[len(s)-1-i], s[i]
	}
}

func CopyReverse[V any](s []V) []V {
	list := make([]V, len(s), len(s))
	copy(list, s)
	Reverse(list)
	return list
}

func ToMap[Slice ~[]V, V any, K comparable](s Slice, getKey func(V) K) (result map[K]V) {
	result = make(map[K]V)
	for _, v := range s {
		result[getKey(v)] = v
	}
	return
}
