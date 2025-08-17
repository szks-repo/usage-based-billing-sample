package db

import "strings"

func MakeValues(numColumn, numValues int) string {
	var values string
	var buf strings.Builder
	buf.WriteString("(")
	for i := range numColumn {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString("?")
	}
	buf.WriteString(")")

	values = buf.String()
	buf.Reset()

	for i := range numValues {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(values)
	}

	return "VALUES " + buf.String()
}
