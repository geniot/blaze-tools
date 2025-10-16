package main

import (
	"bytes"
	"embed"
	"hash/fnv"
)

func GetFileBytes(fileName string, fs *embed.FS) []byte {
	file, _ := fs.Open(fileName)
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(file); err != nil {
		println("Couldn't read from file: {}", err.Error())
		return nil
	}
	return buf.Bytes()
}

func First[T, U any](val T, _ U) T {
	return val
}

func Hash(s string) int {
	h := fnv.New32()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0
	}
	return int(h.Sum32()) % 2147483647 //max positive integer in PostgreSQL (4 bytes)
}

func IsDigitsOnly(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func If[T any](cond bool, vTrue, vFalse T) T {
	if cond {
		return vTrue
	}
	return vFalse
}
