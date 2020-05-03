package main

import (
    "github.com/donomii/trigrammr"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
)

func insertTrigrams(db trigrammr.DbDetails, record []string) {
    if trigrammr.Debug {
        fmt.Println("Trigram: ", record)
    }
    trigrammr.InsertTrigramCached(db, record)
}

func main() {
    db, _ := trigrammr.OpenDB(os.Args[1])
	r := csv.NewReader(os.Stdin)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		insertTrigrams(db, record)
        if trigrammr.Debug {
            fmt.Println(record)
        }
	}
}
