package main

import (
	"encoding/json"
	"fmt"
	"log"

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
	args := strings.Split(line, " ")
	return args
}

func main() {
	db, _ := trigrammr.OpenDB(os.Args[1])
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		//fmt.Println(scanner.Text())

		jsonStr := scanner.Text()
		jsonMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(jsonStr), &jsonMap)
		if err != nil {
			panic(err)
		}
		str := jsonMap["body"].(string)
		//fmt.Println(str)
		words := trigrammr.TrimWords(makeWords(string(str)))

		for i := 2; i < len(words); i++ {
			record := []string{words[i-2], words[i-1], words[i]}
			//fmt.Println(record)
			insertTrigrams(db, record)
			if trigrammr.Debug {
				fmt.Println(record)
			}
			fmt.Println(record)
		}

		if err := scanner.Err(); err != nil {
			log.Println(err)
		}
	}
}
