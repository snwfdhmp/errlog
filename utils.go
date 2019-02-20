package errlog

import "strings"

func isIntInSlice(i int, s []int) bool {
	for vi := range s {
		if s[vi] == i {
			return true
		}
	}
	return false
}

func deleteBankLinesFromRange(lines []string, start, end *int) {
	//clean leading blank lines
	for (*start) <= (*end) {
		if strings.Trim(lines[(*start)], " \n\t") != "" {
			break
		}
		(*start)++
	}

	//clean trailing blank lines
	for (*end) >= (*start) {
		if strings.Trim(lines[(*end)], " \n\t") != "" {
			break
		}
		(*end)--
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
