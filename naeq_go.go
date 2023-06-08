package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"unicode"
)

// "strings"
// add options to read a text file, process it (with optional processing modes) and output it to naeq_X.md
// command line options

// processing modes: -p=
//  word -- processes each word as an individual calculatable thing (this is the default)
//  line -- processes each line as the calculatable thing
//  markov -- uses markov-chain chunking and processes each chung as the calculatable thing

// file options
// -f=filename -- reads the named file and processes all words in it
// -d=directory -- reads all files in the directory and processes them (sequentially)
// -o=directory -- writes/appends to NAEQ_X.md files in named directory

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// some helper
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	processPtr := flag.String("p", "word", "processing mode: word, line or markov")
	inputDirPtr := flag.String("d", ".", "input directory; incompatible with -f")
	outputDirPtr := flag.String("o", "", "output directory; if not specified, everything goes to stdout. if specified, output will be to individual files for each EQ value")
	filePtr := flag.String("f", "", "input file, incompatible with -d and command-line words to process")

	flag.Parse()

	// check for incompatibilities
	if isFlagPassed("d") && isFlagPassed("f") {
		//fmt.Println("-d and -f are incompatible, please see help")
		flag.PrintDefaults()
		return
	} else if (isFlagPassed("f") || isFlagPassed("d")) && len(flag.Args()) > 0 {
		flag.PrintDefaults()
		return
	}
	// check that processing mode parameter is within scope
	if !(*processPtr == "word" || *processPtr == "line" || *processPtr == "markov") {
		flag.PrintDefaults()
		return
	}

	// let's process a file and output it to stdout
	if isFlagPassed("f") {
		fmt.Println("gonna process a file!", *filePtr)
		f, err := os.Open(*filePtr)
		check(err)

		fio := bufio.NewReader(f)
		scanner := bufio.NewScanner(fio)

		if *processPtr == "word" {
			fmt.Println("using word processing mode")
			scanner.Split(bufio.ScanWords)

		} else if *processPtr == "line" {
			fmt.Println("using line processing mode")
			scanner.Split(bufio.ScanLines)
		} else {
			fmt.Fprintln(os.Stderr, "sorry can't do markov yet")
			return
		}

		for scanner.Scan() {
			fmt.Printf("%s: %d\n", scanner.Text(), EQalculateMod(scanner.Text()))
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading input:", err)
		}

	}

	fmt.Println("process:", *processPtr)
	fmt.Println("inputDir", *inputDirPtr)
	fmt.Println("outputDir", *outputDirPtr)
	fmt.Println("file", *filePtr)
	fmt.Println("tail:", flag.Args())
	for _, w := range flag.Args() {
		fmt.Println(w, EQalculateMod(w))
	}
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z]+`)

func clearString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

// calculate EQ of word using % operator
func EQalculateMod(word string) int {
	value := 0
	for _, c := range clearString(word) {
		value += int(unicode.ToLower(c)-'a')*19%26 + 1
	}
	return value
}
