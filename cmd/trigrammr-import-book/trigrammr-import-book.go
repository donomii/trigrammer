package main

import (
	"fmt"
	"io/ioutil"

	"github.com/donomii/trigrammr"

	//"io"
	//"log"
	"bufio"
	"os"
	"strings"
)

func insertTrigrams(db trigrammr.DbDetails, record []string) {
	if trigrammr.Debug {
		fmt.Println("Trigram: ", record)
	}
	trigrammr.InsertTrigramCached(db, record)
}

func insertQuadgrams(db trigrammr.DbDetails, record []string) {
	if trigrammr.Debug {
		fmt.Println("Quadgram: ", record)
	}
	trigrammr.InsertQuadgramCached(db, record)
}

func makeWords(line string) []string {
	args := strings.Split(line, " ")
	return args
}

func main() {
	db, _ := trigrammr.OpenDB(os.Args[1])
	r := bufio.NewReader(os.Stdin)
	str, _ := ioutil.ReadAll(r)
	words := trigrammr.TrimWords(makeWords(string(str)))

	for i := 2; i < len(words); i++ {
		record := []string{words[i-2], words[i-1], words[i]}
		//fmt.Println(record)
		insertTrigrams(db, record)
		if trigrammr.Debug {
			fmt.Println(record)
		}
	}

	for i := 3; i < len(words); i++ {
		record := []string{words[i-3], words[i-2], words[i-1], words[i]}
		//fmt.Println(record)
		insertQuadgrams(db, record)
		if trigrammr.Debug {
			fmt.Println(record)
		}
	}
}
