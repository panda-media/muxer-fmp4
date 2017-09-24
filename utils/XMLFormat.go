package utils

import "bytes"

func FormatXML(data []byte) []byte {
	buf := new(bytes.Buffer)
	tabCount := 0
	for i, v := range data {
		buf.WriteByte(v)
		if v == '<' {
			tabCount++
		}
		if v == '/' && i+1 < len(data) && data[i+1] == '>' {
			tabCount--
		}
		if v == '/' && data[i-1] == '<' {
			tabCount -= 2
		}
		if v == '>' {
			buf.WriteByte('\n')
			if tabCount < 0 {
				tabCount = 0
			}
			for c := 0; c < tabCount; c++ {
				buf.WriteByte('\t')
			}

		}
	}
	return buf.Bytes()
}
