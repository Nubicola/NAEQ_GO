package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	prose "github.com/jdkato/prose/v2"
	slices "golang.org/x/exp/slices"
)

type NAEQ_Processor struct {
	eqs map[int][]string
	/*
	   input_mode, output_mode           string
	   input_directory, output_directory string
	*/
}

// from stack overflow: https://stackoverflow.com/questions/66643946/how-to-remove-duplicates-strings-or-int-from-slice-in-go
func removeDuplicate[T string | int](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func (n *NAEQ_Processor) ProcessString(s string) {
	doc, err := prose.NewDocument(s, prose.WithExtraction(false), prose.WithTagging(true))
	if err != nil {
		panic(err)
	}
	wordyTokens := []string{"SYM", ".", ",", "-", ":", ";", "\""}
	for _, token := range doc.Tokens() {
		val := EQalculateMod(token.Text)
		if !slices.Contains(wordyTokens, token.Tag) {
			(n.eqs)[val] = removeDuplicate(append((n.eqs)[val], strings.ToUpper(token.Text)))
		}
	}
}

func (n *NAEQ_Processor) ProcessTokens(filename string) {
	f, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	n.ProcessString(string(f))
}

func (n *NAEQ_Processor) processLine(line string) {
	doc, _ := prose.NewDocument(line, prose.WithExtraction(false), prose.WithTagging(true))
	val := 0
	wordyTokens := []string{"SYM", ".", ",", "-", ":", ";", "\""}
	ns := []string{}
	for _, token := range doc.Tokens() {
		if !slices.Contains(wordyTokens, token.Tag) {
			val += EQalculateMod(token.Text)
			ns = append(ns, token.Text)
		}
	}
	(n.eqs)[val] = removeDuplicate(append((n.eqs)[val], strings.ToUpper(strings.Join(ns, " "))))
}

// calculates EQ value on a line-by-line basis
func (n *NAEQ_Processor) ProcessLines(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(bufio.NewReader(f))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		n.processLine(scanner.Text())
	}
}

// calculates EQ value on a sentence basis
func (n *NAEQ_Processor) ProcessSentences(filename string) {
	f, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	doc, err := prose.NewDocument(string(f), prose.WithExtraction(false), prose.WithTagging(false))
	if err != nil {
		panic(err)
	}
	for _, sentence := range doc.Sentences() {
		n.processLine(sentence.Text)
	}
}

func (n *NAEQ_Processor) ProcessFile(filename string, pmode string) error {
	if pmode == "word" {
		n.ProcessTokens(filename)
	} else if pmode == "sent" {
		n.ProcessSentences(filename)
	} else if pmode == "line" {
		n.ProcessLines(filename)
	} else {
		return errors.New("can't do markov yet")
	}
	return nil
}

func (n *NAEQ_Processor) Verify() {
	for k := range n.eqs {
		if k <= 0 {
			fmt.Println("deleting", k, n.eqs[k])
			delete(n.eqs, k)
		}
	}

}
func (n *NAEQ_Processor) WriteToFiles(directory string) error {
	// directory must exist. for all keys in the map, create/open NAEQ_key.md. Parse the whole file (just like the scanner above, actually)
	// and append to the eqs value.
	dirContents, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	// go through the values
	// for each value, check if a file is there. create/open it and read values into a new eqs (for speed, rather than in the entire map again)
	// then write the values back into that file sorted alphabetically
	for val := range n.eqs {
		var str = fmt.Sprintf("NAEQ_%d.md", val)
		filename := filepath.Join(directory, str)
		// look for files in the directory. If it's there copy all its contents into the existing map
		for i := range dirContents {
			if dirContents[i].Name() == str {
				n.ProcessFile(filename, "line")
			}
		}
		// now eqs has all the strings from the file as well as from the input. os.Create will truncate a file!
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		_, err = w.WriteString(strings.Join((n.eqs)[val], "\n") + "\n")
		if err != nil {
			return err
		}
		w.Flush()
	}
	return nil
}

func MakeNewNQP() *NAEQ_Processor {
	var pN = new(NAEQ_Processor)
	pN.eqs = make(map[int][]string)
	/*	pN.input_mode =
		pN.input_directory = */
	return pN
}

// function is called when traversing directories for each element
// doing a little closure so the processing-mode flag can be passed in
func visit(pmode string, pN *NAEQ_Processor) filepath.WalkFunc {
	return func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			err = pN.ProcessFile(p, pmode)
		}
		return err
	}
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

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
func main() {
	processPtr := flag.String("p", "word", "processing mode: word, line, sent or markov")
	inputDirPtr := flag.String("d", ".", "input directory; incompatible with -f")
	outputDirPtr := flag.String("o", "", "output directory; if not specified, everything goes to stdout. if specified, output will be to individual files for each EQ value")
	filePtr := flag.String("f", "", "input file, incompatible with -d and command-line words to process")

	flag.Parse()

	// check for incompatibile flags
	if isFlagPassed("d") && isFlagPassed("f") {
		flag.PrintDefaults()
		return
	} else if (isFlagPassed("f") || isFlagPassed("d")) && len(flag.Args()) > 0 {
		flag.PrintDefaults()
		return
	} else if (!isFlagPassed("f") && !isFlagPassed("d")) && len(flag.Args()) == 0 {
		flag.PrintDefaults()
		return
	}
	// check that processing mode parameter is within scope
	if !(*processPtr == "word" || *processPtr == "line" || *processPtr == "markov" || *processPtr == "sent") {
		flag.PrintDefaults()
		return
	}

	// handle the input source - a file, a directory of files or command line args. Build the map of eq values (value -> list of strings)
	var pN *NAEQ_Processor = MakeNewNQP()

	if isFlagPassed("f") {
		pN.ProcessFile(*filePtr, *processPtr)
	} else if isFlagPassed("d") {
		// read and process all files in a directory
		filepath.Walk(*inputDirPtr, visit(*processPtr, pN))
	} else { // not f or d, must be somthing in the args. Later: also handle stdin
		pN.ProcessString(strings.Join(flag.Args(), " "))
	}

	pN.Verify()
	// now handle output; write it to files in the output directory, or write to stdout
	if isFlagPassed("o") {
		//		check(writeToFiles(*outputDirPtr, &(pN.eqs)))
		_ = pN.WriteToFiles(*outputDirPtr)
	} else {
		for val, words := range pN.eqs {
			fmt.Println(val, words)
		}
	}
}

// calculate EQ of word
func EQalculateMod(word string) int {
	value := 0
	for _, c := range word {
		i := int(unicode.ToLower(c))
		if i >= int('a') {
			value += (i-int('a'))*19%26 + 1
		}
	}
	return value
}
