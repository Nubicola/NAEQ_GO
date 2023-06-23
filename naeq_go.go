package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

func copyEqsMap(dest_eqs map[int][]string, src_eqs map[int][]string) {
	for k, v := range src_eqs {
		// value is a slice; append to the slice
		dest_eqs[k] = removeDuplicate[string](append(dest_eqs[k], v...))
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

// error checking
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// function is called when traversing directories for each element
// doing a little closure so the processing-mode flag can be passed in
func visit(pmode string, eqs *map[int][]string) filepath.WalkFunc {
	return func(p string, info os.FileInfo, err error) error {
		check(err)
		if !info.IsDir() {
			err := processFile(p, pmode, eqs)
			return err
		}
		return nil
	}
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

func scanStrings(scanner *bufio.Scanner, eqs *map[int][]string) {
	leqs := *eqs
	for scanner.Scan() {
		val := EQalculateMod(scanner.Text())
		leqs[val] = removeDuplicate[string](append(leqs[val], strings.ToUpper(scanner.Text())))
	}
}

func processFile(filename string, pmode string, eqs *map[int][]string) error {
	// map kv: keys are the EQ value, values are a slice of string {
	//fmt.Println("gonna process a file!", filename)
	f, err := os.Open(filename)
	check(err)
	defer f.Close()

	scanner := bufio.NewScanner(bufio.NewReader(f))

	if pmode == "word" {
		//		fmt.Println("using word processing mode")
		scanner.Split(bufio.ScanWords)
	} else if pmode == "line" {
		//		fmt.Println("using line processing mode")
		scanner.Split(bufio.ScanLines)
	} else {
		return errors.New("can't do markov yet")
	}

	scanStrings(scanner, eqs)
	return nil
}

func writeToFiles(directory string, eqs *map[int][]string) error {
	// directory must exist. for all keys in the map, create/open NAEQ_key.md. Parse the whole file (just like the scanner above, actually)
	// and append to the eqs value.
	dirContents, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	// go through the values
	// for each value, check if a fiel is there. create/open it and read values into a new eqs (for speed, rather than in the entire map again)
	// then write the values back into that file sorted alphabetically
	//fmt.Println("eqs before", *eqs)
	for val := range *eqs {
		var str = fmt.Sprintf("NAEQ_%d.md", val)
		//fmt.Println("looking for file", str)
		filename := filepath.Join(directory, str)
		// look for files in the directory. If it's there copy all its contents into the existing map
		for i := range dirContents {
			if dirContents[i].Name() == str {
				//fmt.Println("found it!")
				file_eqs := make(map[int][]string)
				err = processFile(filename, "line", &file_eqs)
				if err != nil {
					return err
				}
				copyEqsMap(*eqs, file_eqs)
			}
		}
		// now eqs has all the strings from the file as well as from the input. os.Create will truncate a file!
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		_, err = w.WriteString(strings.Join((*eqs)[val], "\n") + "\n")
		if err != nil {
			return err
		}
		w.Flush()
	}
	//fmt.Println("eqs after", *eqs)
	return nil
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
	processPtr := flag.String("p", "word", "processing mode: word, line or markov")
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
	if !(*processPtr == "word" || *processPtr == "line" || *processPtr == "markov") {
		flag.PrintDefaults()
		return
	}

	// handle the input source - a file, a directory of files or command line args. Build the map of eq values (value -> list of strings)
	eqs := make(map[int][]string)
	if isFlagPassed("f") {
		check(processFile(*filePtr, *processPtr, &eqs))
	} else if isFlagPassed("d") {
		// read and process all files in a directory
		filepath.Walk(*inputDirPtr, visit(*processPtr, &eqs))
	} else { // not f or d, must be somthing in the args. Later: also handle stdin
		r := strings.NewReader(strings.Join(flag.Args(), "\n"))
		scanner := bufio.NewScanner(r)
		scanStrings(scanner, &eqs)
	}

	// now handle output; write it to files in the output directory, or write to stdout
	if isFlagPassed("o") {
		check(writeToFiles(*outputDirPtr, &eqs))
	} else {
		for val, words := range eqs {
			fmt.Println(val, words)
		}
	}
}

// var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z]+`)
var nonAlphanumericRegex = regexp.MustCompile(`(\b[A-Z0-9]['A-Z0-9]+\b|\b[A-Z]\b)\|?`)

func clearString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

// calculate EQ of word
func EQalculateMod(word string) int {
	value := 0
	for _, c := range clearString(word) {
		value += int(unicode.ToLower(c)-'a')*19%26 + 1
	}
	return value
}
