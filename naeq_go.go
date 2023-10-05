package naeq_go

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	prose "github.com/jdkato/prose/v2"
	slices "golang.org/x/exp/slices"
	//markov "github.com/mb-14/gomarkov"
)

// a processor that calculates based on regular EQ can only keep calculating on that basis

type ALWCalculator interface {
	Calculate(ss []string) int
}

type EQBaseCalculator struct {
}

// sums all the word values together
func (eqbc *EQBaseCalculator) Calculate(ss []string) int {
	var value = 0
	for _, s := range ss {
		for _, c := range s {
			i := int(unicode.ToLower(c))
			if i >= int('a') {
				value += (i-int('a'))*19%26 + 1
			}
		}
	}
	return value
}

type EQFirstCalculator struct {
}

func (eqfc *EQFirstCalculator) Calculate(ss []string) int {
	var sb strings.Builder
	for _, s := range ss {
		sb.WriteByte(s[0])
	}
	var eqbc = new(EQBaseCalculator)
	var xs = make([]string, 1)
	xs[0] = sb.String()
	fmt.Println(xs[0])
	return eqbc.Calculate(xs)
}

type EQLastCalculator struct {
}

func (eqlc *EQLastCalculator) Calculate(ss []string) int {
	var sb strings.Builder
	for _, s := range ss {
		sb.WriteByte(s[len(s)-1])
	}
	var eqbc = new(EQBaseCalculator)
	var xs = make([]string, 1)
	xs[0] = sb.String()
	return eqbc.Calculate(xs)
}

type NAEQ_Processor struct {
	eqs        map[int][]string
	calculator ALWCalculator
}

func MakeNewNQP(calcmode string) *NAEQ_Processor {
	var pN = new(NAEQ_Processor)
	pN.eqs = make(map[int][]string)
	if calcmode == "alw" {
		pN.calculator = new(EQBaseCalculator)
	} else if calcmode == "first" {
		pN.calculator = new(EQFirstCalculator)
	} else if calcmode == "last" {
		pN.calculator = new(EQLastCalculator)
	} else {
		panic("what kind of mode is that?")
	}

	return pN
}

// core struct: NAEQ_Processor
// given string and mode, process the string (word and sentence mode)
// given bufio.scanner, process lines (only "line" mode)

func (n *NAEQ_Processor) CorrespondingWords(s string) []string {
	var xs = make([]string, 1)
	xs[0] = s
	return (n.eqs[n.calculator.Calculate(xs)])
}

func (n *NAEQ_Processor) ProcessString(s, mode string) {
	// modes can be "word" and "sent"
	if mode == "word" {
		doc, err := prose.NewDocument(s, prose.WithExtraction(false), prose.WithTagging(true))
		if err != nil {
			panic(err)
		}
		wordyTokens := []string{"SYM", ".", ",", "-", ":", ";", "\""}
		var val = 0
		var xs = make([]string, 1)
		for _, token := range doc.Tokens() {
			if !slices.Contains(wordyTokens, token.Tag) {
				xs[0] = token.Text
				val = n.calculator.Calculate(xs)
				(n.eqs)[val] = removeDuplicate(append((n.eqs)[val], strings.ToUpper(token.Text)))
			}
		}
	} else if mode == "sent" {
		doc, err := prose.NewDocument(s, prose.WithExtraction(false), prose.WithTagging(false))
		if err != nil {
			panic(err)
		}
		for _, sentence := range doc.Sentences() {
			n.processLine(sentence.Text)
		}
	} else {
		panic("not a valid mode, " + mode)
	}
}

func (n *NAEQ_Processor) ProcessBuf(s *bufio.Scanner) {
	s.Split(bufio.ScanLines)
	for s.Scan() {
		n.processLine(s.Text())
	}
}

func (n *NAEQ_Processor) processLine(line string) {
	doc, _ := prose.NewDocument(line, prose.WithExtraction(false), prose.WithTagging(true))
	val := 0
	wordyTokens := []string{"SYM", ".", ",", "-", ":", ";", "\""}
	ns := []string{}
	for _, token := range doc.Tokens() {
		// filter any not-word tokens from the string
		if !slices.Contains(wordyTokens, token.Tag) {
			ns = append(ns, token.Text)
		}
	}
	val = n.calculator.Calculate(ns)
	(n.eqs)[val] = removeDuplicate(append((n.eqs)[val], strings.ToUpper(strings.Join(ns, " "))))
}

func (n *NAEQ_Processor) Cleanup() {
	for k := range n.eqs {
		if k <= 0 {
			fmt.Println("deleting", k, n.eqs[k])
			delete(n.eqs, k)
		}
	}
}

func (n *NAEQ_Processor) Output(buf *bufio.Writer) {
	for val, words := range n.eqs {
		_, err := buf.WriteString(fmt.Sprintf("%d: %s\n", val, strings.Join(words, " ")))
		if err != nil {
			fmt.Println("Error writing to buffer")
			return
		}
	}
	buf.Flush()
}

// helper functions
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

func MergeWithMDFiles(directory string, n *NAEQ_Processor, autofix bool) error {
	// directory must exist. for all keys in the map, create/open NAEQ_key.md. Read the file in line-mode only, no tokenizing.
	// it's assumed these files have already been written by this utility!
	// reads the contents, adds to the eq-map, writes it out again
	dirContents, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	var wrongEQs *NAEQ_Processor = MakeNewNQP("alw")
	wrongEQs.calculator = n.calculator // just in case another calculator is used. Seems unlikely though.
	// go through the values
	// for each value, check if a file is there. create/open it and read values into a new eqs (for speed, rather than in the entire map again)
	// then write the values back into that file sorted alphabetically
	for file_eq_val := range n.eqs {
		var str = fmt.Sprintf("NAEQ_%d.md", file_eq_val)
		filename := filepath.Join(directory, str)
		// look for files in the directory. If it's there copy all its contents into the existing map
		for i := range dirContents {
			if dirContents[i].Name() == str {
				f, err := os.Open(filename)
				if err != nil {
					return err
				}
				defer f.Close()
				scanner := bufio.NewScanner(bufio.NewReader(f))
				scanner.Split(bufio.ScanLines)
				ns := []string{}
				// calculate line's value with no fancy tokenizing; assume each line is just ready to go
				for scanner.Scan() {
					ns[0] = scanner.Text()
					val := n.calculator.Calculate(ns)
					if val != file_eq_val {
						// there's a string in here that doesn't equal the file's value
						fmt.Println("found an error in file", str, "with string", scanner.Text())
						if autofix {
							(wrongEQs.eqs)[val] = removeDuplicate(append((n.eqs)[val], strings.ToUpper(scanner.Text())))
						}
					} else {
						(n.eqs)[val] = removeDuplicate(append((n.eqs)[val], strings.ToUpper(scanner.Text())))
					}
				}
			}
		}
		// now eqs has all the strings from the file as well as from the input. os.Create will truncate a file!
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		_, err = w.WriteString(strings.Join((n.eqs)[file_eq_val], "\n") + "\n")
		if err != nil {
			return err
		}
		w.Flush()
	}
	if len(wrongEQs.eqs) > 0 {
		fmt.Println("processing the wrong eqs recursively!")
		MergeWithMDFiles(directory, wrongEQs, false)
	}
	return nil
}
