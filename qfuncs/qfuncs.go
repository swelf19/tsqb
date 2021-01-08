package qfuncs

func RemoveEmpty(src []string) []string {
	dst := []string{}
	for _, s := range src {
		if s != "" {
			dst = append(dst, s)
		}
	}
	return dst
}
