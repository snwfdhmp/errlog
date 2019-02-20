package errlog

func isIntInSlice(i int, s []int) bool {
	for vi := range s {
		if s[vi] == i {
			return true
		}
	}
	return false
}
