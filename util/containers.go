package util

func SetToSlice[T comparable](set map[T]struct{}) []T {
	res := []T{}
	for key := range set {
		res = append(res, key)
	}
	return res
}

func AddSliceToSet[T comparable](set map[T]struct{}, slice []T) {
	for _, val := range slice {
		set[val] = struct{}{}
	}
}
