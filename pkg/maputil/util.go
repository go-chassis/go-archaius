package maputil

import "sort"

//Map2String convert map to a sorted string
func Map2String(m map[string]string) string {
	result := ""
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	l := len(keys)
	for i, k := range keys {
		if i != l-1 {
			result = result + k + "=" + m[k] + "|"
		} else {
			result = result + k + "=" + m[k]
		}

	}
	return result
}
