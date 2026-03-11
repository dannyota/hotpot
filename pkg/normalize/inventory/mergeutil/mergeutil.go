package mergeutil

// NamespacedKeys flattens a MergeKeys map into "type:value" strings for indexing.
func NamespacedKeys(mergeKeys map[string][]string) []string {
	var keys []string
	for col, vals := range mergeKeys {
		for _, v := range vals {
			if v != "" {
				keys = append(keys, col+":"+v)
			}
		}
	}
	return keys
}

// SetIfEmpty sets *dst to val if *dst is empty.
func SetIfEmpty(dst *string, val string) {
	if *dst == "" {
		*dst = val
	}
}
