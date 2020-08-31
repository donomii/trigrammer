package main

import (
	"fmt"
	//	"io/ioutil"

	"github.com/donomii/trigrammr"

	//"io"
	//"log"
	"bufio"
	"os"
	"strings"
	"text/scanner"
)

func insertTrigrams(db trigrammr.DbDetails, record []string) {
	if trigrammr.Debug {
		fmt.Println("Trigram: ", record)
	}
	trigrammr.InsertTrigramCached(db, record)
}

func makeWords(line string) []string {
	args := strings.Split(line, " ")
	return args
}

func main() {
	transitions := make(map[string]string)
	transCount := make(map[string]int)
	//	r := bufio.NewReader(os.Stdin)
	//	str, _ := ioutil.ReadAll(r)
	//	words := trigrammr.TrimWords(makeWords(string(str)))

	var s scanner.Scanner
	s.Init(bufio.NewReader(os.Stdin))
	s.Filename = "example"
	oldTok := ""
	for token := s.Scan(); token != scanner.EOF; token = s.Scan() {
		tok := s.TokenText()
		transitions[oldTok] = tok
		transCount[fmt.Sprintf("%v-%v", oldTok, tok)]++
		oldTok = tok
		//fmt.Printf("%s: %s\n", s.Position, s.TokenText())
	}

	/*

			for i:=1; i<len(words); i++ {
				record := []string{words[i-1], words[i]}
				transitions[record[0]]= record[1]
				transCount[fmt.Sprintf("%v-%v", record[0], record[1])]++
		        if trigrammr.Debug {
		            fmt.Println(record)
		        }
			}

	*/
	fmt.Printf("%+V", transCount)
}
