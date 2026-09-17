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

func main() {
	fmt.Println("Простой архиватор запущен")
}
