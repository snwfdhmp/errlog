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

//deleteBlankLinesFromRange increments and decrements respectively start and end so they are not representing an empty line (in slice lines)
func deleteBlankLinesFromRange(lines []string, start, end *int) {
	//clean from out of range values
	(*start) = max(*start, 0)
	(*end) = min(*end, len(lines)-1)

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
