package main

import (
	"fmt"

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

func makeWords(line string) []string {
	line = strings.Replace(line, ".", " ", -1)
	line = strings.Replace(line, ",", " ", -1)
	line = strings.Replace(line, "\"", " ", -1)
	args := strings.Split(line, " ")
	return args
}

func main() {
	db, _ := trigrammr.OpenDB(os.Args[1])
	r := bufio.NewReader(os.Stdin)
	for {
		str, err := r.ReadString([]byte("\n")[0])
		if err != nil {
			panic(err)
		}
		words := trigrammr.TrimWords(makeWords(string(str)))

		for i := 2; i < len(words); i++ {
			record := []string{words[i-2], words[i-1], words[i]}
			//fmt.Println(record)
			insertTrigrams(db, record)
			if trigrammr.Debug {
				fmt.Println(record)
			}
		}
	}
}
