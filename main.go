package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

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
		isCompressed := data[i]&128 == 128
		fmt.Printf("%x, %b, isCompressed: %t, len: %d\n", data[i], data[i], isCompressed, l)
		i += 1
		if isCompressed {
			for range l {
				res = append(res, data[i])
			}
			i += 1
		} else {
			res = append(res, data[i:i+int(l)]...)
			i += int(l)
		}
	}

	return res
}

func (sa *SimpleArchiver) CompressFile(inputPath, outputPath string) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла %s: %q", inputPath, err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла %s: %q", outputPath, err)
	}
	defer outputFile.Close()

	reader := bufio.NewReader(inputFile)
	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()

	filename := filepath.Base(inputPath)
	err = writer.WriteByte(byte(len(filename)))
	if err != nil {
		return fmt.Errorf("ошибка записи длины файла: %q", err)
	}

	_, err = writer.WriteString(filename)
	if err != nil {
		return fmt.Errorf("ошибка записи имени файла: %q", err)
	}

	for {
		n, err := reader.Read(sa.buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		compressed := sa.compress(sa.buffer[:n])
		var size [2]byte
		binary.BigEndian.PutUint16(size[:], uint16(len(compressed)))
		_, err = writer.Write(size[:])
		if err != nil {
			return err
		}
		_, err = writer.Write(compressed)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sa *SimpleArchiver) DecompressFile(inputPath, outputDir string) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла %s: %q", inputPath, err)
	}
	defer inputFile.Close()

	reader := bufio.NewReader(inputFile)
	filenameLength, err := reader.ReadByte()
	if err != nil {
		return err
	}
	filenameBytes := make([]byte, 0, filenameLength)
	_, err = reader.Read(filenameBytes)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(outputDir, string(filenameBytes))
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла %s: %q", outputPath, err)
	}
	defer outputFile.Close()
}

func main() {
	fmt.Println("Простой архиватор запущен")

	sa := NewArchiver("test")
	fmt.Println(string(sa.decompress(sa.compress([]byte("ABCDDDDDDDDDDDDDEF")))))
}
