package utils

func SliceToMap(slice []string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, v := range slice {
		result[v] = struct{}{}
	}
	return result
}

func Contains(mapSet map[string]struct{}, item string) bool {
	_, exists := mapSet[item]
	return exists
}
