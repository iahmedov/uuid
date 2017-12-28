package uuid

import (
	"fmt"
)

var maskTable [9]byte = [9]byte{0x00, 0x01, 0x03, 0x07, 0x0F, 0x1F, 0x3F, 0x7F, 0xFF}

func mask(size uint) byte {
	return maskTable[size]
}

func BitSet(res []byte, start, end int, data []byte) {
	if start >= end {
		return
	}
	firstByte := start / 8
	firstByteStartBit := uint(start % 8)
	lastByte := end / 8
	lastByteLastBit := uint(end % 8)

	bits := uint(end - start)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			fmt.Printf("values: fb(%d), fbsb(%d), lb(%d), lblb(%d), start(%d), end(%d)\n",
				firstByte, firstByteStartBit, lastByte, lastByteLastBit, start, end)
			panic("fail")
		}
	}()

	if bits <= 8 {
		if firstByte == lastByte {
			// copy bits inside one byte
			// example: 0000 0000 (2,5, "010")
			res[firstByte] = res[firstByte] | data[0]&mask(bits)<<firstByteStartBit
		} else {
			// copy bits in intersection of bytes
			// example: copy (5,12, "0010100")
			part1 := (data[0] & mask(8-firstByteStartBit)) << firstByteStartBit // 10101010 & 0b11 = 0b10 -> 0b1000 0000
			res[firstByte] = res[firstByte] | part1

			if lastByteLastBit != 0 {
				part2 := (data[0] & (mask(lastByteLastBit) << (8 - firstByteStartBit))) >> (8 - firstByteStartBit) // 10101010 & 0x11111100 >> 2
				res[lastByte] = res[lastByte] | part2
			}
		}
	} else {
		var localEnd int = start
		var localBegin int = start
		bt := [1]byte{}
		for _, b := range data {
			localEnd = localEnd + 8
			if end < localEnd {
				localEnd = end
			}
			bt[0] = b
			BitSet(res, localBegin, localEnd, bt[:])
			localBegin = localEnd
		}
	}
}

func BitGet(src []byte, start, end int, dst []byte) {
	firstByte := start / 8
	firstByteStartBit := uint(start % 8)
	lastByte := end / 8
	lastByteLastBit := uint(end % 8)

	bits := uint(end - start)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			fmt.Printf("values: fb(%d), fbsb(%d), lb(%d), lblb(%d), start(%d), end(%d)\n",
				firstByte, firstByteStartBit, lastByte, lastByteLastBit, start, end)
			panic("fail")
		}
	}()

	if bits <= 8 {
		if firstByte == lastByte {
			// copy bits inside one byte
			// example: 0001 0000 (2,5, "010")
			dst[0] = src[firstByte] & mask(bits) << firstByteStartBit
		} else {
			// copy bits in intersection of bytes
			// example: 0000 0(001 010)0 0000  (5,11, "010100")
			x1 := src[firstByte] >> firstByteStartBit
			var x2 byte = 0
			if lastByteLastBit != 0 {
				x2 = src[lastByte] & mask(lastByteLastBit) << (8 - firstByteStartBit)
			}
			dst[0] = x1 | x2
		}
	} else {
		localEnd := start
		counter := 0
		for i := start; i < end; i += 8 {
			if i+8 > end {
				localEnd = end
			} else {
				localEnd = i + 8
			}
			BitGet(src, i, localEnd, dst[counter:])
			counter += 1
		}
	}
}
