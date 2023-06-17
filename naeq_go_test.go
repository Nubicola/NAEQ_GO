package main

import (
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
)

func BenchmarkEQalculateMod(b *testing.B) {
	data := getCorpusStrings(b)
	words := regexp.MustCompile(`[\w'\-]+`).FindAll(data, -1)
	for _, wordSlice := range words {
		word := strings.ToLower(string(wordSlice))
		_ = EQalculateMod(word)
	}

}

func getCorpusStrings(b *testing.B) []byte {
	data, err := ioutil.ReadFile("corpus/lib10.txt")
	if err != nil {
		b.Error("Unable to open file")
	}
	data2, err := ioutil.ReadFile("corpus/lib65.txt")
	if err != nil {
		b.Error("unable to open next file")
	}
	data3, err := ioutil.ReadFile("corpus/lib7a.txt")
	if err != nil {
		b.Error("third file not available")
	}
	data = append(data, data2...)
	data = append(data, data3...)
	return (data)
	//return (append(data, data2...))

}
