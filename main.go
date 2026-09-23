package main

import "fmt"

type SimpleArchiver struct {
	inputPath  string
	outputPath string
	buffer     []byte
}

func NewArchiver(inputPath string) *SimpleArchiver {
	return &SimpleArchiver{
		inputPath: inputPath,
		buffer:    make([]byte, 1024*8),
	}
}

func (sa *SimpleArchiver) compressEmpty(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	return data
}

func (sa *SimpleArchiver) countRepeating(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}

	res := make([]byte, 0)
	var count byte = 1
	for i := 1; i < len(data); i++ {
		if data[i] == data[i-1] {
			count++
		} else {
			res = append(res, count, data[i-1])
			count = 1
		}
	}
	res = append(res, count, data[len(data)-1])
	return res
}

func (sa *SimpleArchiver) createControlByte(count int, isCompressed bool) byte {
	if count > 127 {
		count = 127
	}

	if isCompressed {
		return byte(128 + count)
	}

	return byte(count)
}

func (sa *SimpleArchiver) compress(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}

	minRepeats := 3
	maxGroupLen := 127

	res := make([]byte, 0)
	left, repeats := 0, 1
	for right := 1; right < len(data); right++ {
		if data[right] == data[right-1] {
			repeats++

			if repeats == minRepeats {
				notCompressedLen := right - repeats - left + 1
				for notCompressedLen != 0 {
					l := min(notCompressedLen, maxGroupLen)
					res = append(res, sa.createControlByte(l, false))
					res = append(res, data[left:left+l]...)
					left += l
					notCompressedLen -= l
				}
			}
		} else {
			if repeats >= minRepeats {
				for repeats != 0 {
					l := min(repeats, maxGroupLen)
					res = append(res, sa.createControlByte(l, true))
					res = append(res, data[right-1])
					left += l
					repeats -= l
				}
			}
			repeats = 1
		}
	}

	if repeats >= minRepeats {
		for repeats != 0 {
			l := min(repeats, maxGroupLen)
			res = append(res, sa.createControlByte(l, true))
			res = append(res, data[len(data)-1])
			left += l
			repeats -= l
		}
	} else {
		notCompressedLen := len(data) - left
		for notCompressedLen != 0 {
			l := min(notCompressedLen, maxGroupLen)
			res = append(res, sa.createControlByte(l, false))
			res = append(res, data[left:left+l]...)
			left += l
			notCompressedLen -= l
		}
	}

	return res
}

func (sa *SimpleArchiver) decompress(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}

	res := make([]byte, 0)
	i := 0
	for i < len(data) {
		l := data[i] & 127
		isCompressed := data[i]&128 == 1
		fmt.Printf("%x, %b\n", data[i], data[i])
		if isCompressed {
			i += 2
		} else {
			i += int(l) + 1
		}
	}

	return res
}

func main() {
	fmt.Println("Простой архиватор запущен")
}
