package main

import (
	"bytes"
	"compress/flate"
	"errors"
	"io/ioutil"
	"log"
)

var (
	decode6bitTable = make([]byte, 128)
)

func init() {
	encode6bitTable := make([]byte, 64)
	var b byte
	for ; b < 64; b++ {
		encode6bitTable[b] = encode6bit(b)
		decode6bitTable[encode6bitTable[b]] = b
	}
}

func encode6bit(b byte) byte {
	if b < 10 {
		return 48 + b
	}
	b -= 10
	if b < 26 {
		return 65 + b
	}
	b -= 26
	if b < 26 {
		return 97 + b
	}
	b -= 26
	if b == 0 {
		return '-'
	}
	if b == 1 {
		return '_'
	}
	return '?'
}

// Encode translate raw source code to plantuml compatible magic string
func Encode(data string) string {
	zStr := compressDeflate([]byte(data))
	res := encode64(zStr)
	if res == "" {
		return data
	}
	return res
}

// Decode parse plantuml compatible magic string to source code
func Decode(data string) string {
	zStr := decode64(data)
	res := uncompressInflate(zStr)
	if res == "" {
		return data
	}
	return res
}

func compressDeflate(str []byte) string {
	var b bytes.Buffer
	w, err := flate.NewWriter(&b, flate.BestSpeed)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(str)
	w.Close()
	return b.String()
}

func uncompressInflate(str []byte) string {
	b := bytes.NewReader(str)
	r := flate.NewReader(b)
	if r == nil {
		log.Fatal(errors.New("Creating new flate reader failed"))
	}
	defer r.Close()
	s, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	return string(s)
}

func decode64(s string) []byte {
	length := len(s)
	r := length % 4
	if r != 0 {
		length += 4 - r
	}
	r = (length*3 + 3) / 4
	var data []byte
	for i := 0; i < len(s); i += 4 {
		d := decode3bytes(s[i], s[i+1], s[i+2], s[i+3])
		data = append(data, d...)
	}
	return data
}

func decode3bytes(cc1, cc2, cc3, cc4 byte) []byte {
	c1 := decode6bitTable[cc1]
	c2 := decode6bitTable[cc2]
	c3 := decode6bitTable[cc3]
	c4 := decode6bitTable[cc4]
	r1 := c1<<2 | c2>>4
	r2 := (c2&0xf)<<4 | c3>>2
	r3 := (c3&0x3)<<6 | c4
	return []byte{r1, r2, r3}
}

func encode64(data string) string {
	var r string
	for i := 0; i < len(data); i += 3 {
		if i+2 == len(data) {
			r += append3bytes(data[i], data[i+1], 0)
		} else if i+1 == len(data) {
			r += append3bytes(data[i], 0, 0)
		} else {
			r += append3bytes(data[i], data[i+1], data[i+2])
		}
	}
	return r
}

func append3bytes(b1, b2, b3 byte) string {
	c1 := b1 >> 2
	c2 := ((b1 & 0x3) << 4) | (b2 >> 4)
	c3 := ((b2 & 0xF) << 2) | (b3 >> 6)
	c4 := b3 & 0x3F
	var r string
	r += string(encode6bit(c1 & 0x3F))
	r += string(encode6bit(c2 & 0x3F))
	r += string(encode6bit(c3 & 0x3F))
	r += string(encode6bit(c4 & 0x3F))
	return r
}
