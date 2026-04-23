package utils

// UniqueSlice функция возвращает слайс уникальных строк, содержащихся в исходном слайсе.
func UniqueSlice(slice *[]string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, v := range *slice {
		if _, ok := seen[v]; !ok {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}
