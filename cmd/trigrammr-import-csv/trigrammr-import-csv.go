package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/donomii/trigrammr"
)

func lc(s string) string {
	return strings.ToLower(s)
}

func insertTrigrams(db trigrammr.DbDetails, names []string, record []string) {
	if noColumns {
		for i := range record {
			if i < len(record)-2 {
				trigrammr.InsertTrigramCached(db, []string{lc(record[i]), lc(record[i+1]), lc(record[i+2])})
				//trigrammr.InsertTrigramCached(db, []string{lc(record[i] + names[i]), lc(record[i+1] + names[i+1]), lc(record[i+2] + names[i+2])})
				trigrammr.InsertQuadgramCached(db, []string{lc(record[i]), lc(record[i+1]), lc(record[i+2]), lc(fmt.Sprintf("%v", record))})
			}
		}
	} else {
		for i, _ := range names {
			for j := 0; j < len(names); j++ {
				if len(record[i]) > 0 && len(record[j]) > 0 && i != j {
					if record[i] != "0" && record[j] != "0" {
						ri := lc(record[i])
						rj := lc(record[j])
						ni := lc(names[i])
						nj := lc(names[j])
						if trigrammr.Debug {
							//fmt.Println("Trigram: ", ri, " - ", nj, " - ", rj)
							//fmt.Println("Trigram: ", ni, " - ", ri, " - ", nj)
						}
						trigrammr.InsertTrigramCached(db, []string{ri, nj, rj})
						trigrammr.InsertTrigramCached(db, []string{ni, ri, nj})

						trigrammr.InsertQuadgramCached(db, []string{ri, nj, rj, fmt.Sprintf("%v", record)})
						trigrammr.InsertQuadgramCached(db, []string{ni, ri, nj, fmt.Sprintf("%v", record)})
					}
				}
			}
		}
	}
}

func reverse(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

var noColumns bool
var verbose bool
var addreverse bool
var tsv bool

func main() {
	flag.BoolVar(&noColumns, "no_columns", false, "Do not link column names to data, treat each row as a list of trigrams")
	flag.BoolVar(&tsv, "tsv", false, "Input is tab-separated-values, not comma-separated-values")
	flag.BoolVar(&verbose, "verbose", false, "Print records as they are inserted")
	flag.BoolVar(&addreverse, "add-reverse", false, "Also add reverse trigrams")
	flag.Parse()
	dbNames := flag.Args()
	if len(dbNames) == 0 {
		fmt.Println(`
Use: cat data.csv | trigrammr-import-csv database-name.sqlite

`)
		os.Exit(1)
	}
	db, _ := trigrammr.OpenDB(dbNames[0])
	r := csv.NewReader(os.Stdin)
	r.Comment = '#'
	r.LazyQuotes = true
	r.FieldsPerRecord = 4
	if tsv {
		text := []rune("\t")
		r.Comma = text[0]
	}

	columnNames, err := r.Read()
	if err != nil {
		log.Fatal(err)
	}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
		} else {
			insertTrigrams(db, columnNames, record)
			if addreverse {
				reverse(record)
				insertTrigrams(db, columnNames, record)
			}
			if verbose {
				fmt.Println(strings.Join(record, "<|>"))
			}
		}
	}
}
