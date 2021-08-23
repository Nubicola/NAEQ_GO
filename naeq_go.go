package main

import (
	"fmt"
	"strings"
)

// this is the main map that translates characters to values

var eqMap map[rune]int

func main() {
	//this could be done in a much nicer way, but for now, this does the trick I guess
	vals := [26]int{1, 20, 13, 6, 25, 18, 11, 4, 23, 16, 9, 2, 21, 14, 7, 26, 19, 12, 5, 24, 17, 10, 3, 22, 15, 8}
	chs := [26]rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}
	eqMap = make(map[rune]int)
	for i := 0; i < len(vals); i++ {
		eqMap[chs[i]] = vals[i]
	}
	fmt.Println(eqMap)
	fmt.Println("Sum of 'hello', ", eqalculate("hello"))
}

// calculates EQ Value of (exactly) 1 word
func eqalculate(word string) int {
	splitStrings := strings.Split(word, " ")
	if len(splitStrings) > 1 {
		panic("this function calculates 1 word only!")
	}
	oneWord := strings.ToLower(splitStrings[0])
	eq := 0
	for r := range oneWord {
		eq += eqMap[rune(oneWord[r])]
	}
	return eq
}
