//Accesses sqlite trigram databases, and provides useful query and manipulation functions
package trigrammr

import (
	"fmt"
	"log"
	"strings"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var Debug bool = false
var symbol_cache map[string]int

type DbDetails struct {
	db   *sql.DB
	name string
	url  string
}

type dbList []DbDetails

//Given a number (as a string type), return a string, as stored in the database
func FetchString_str(db DbDetails, sym string) string {
	stmt, err := db.db.Prepare("SELECT s FROM strings WHERE id = ?")
	if err != nil {
		log.Fatal("FetchString_str failed: ", err)
	}
	defer stmt.Close()
	var name string
	err = stmt.QueryRow(sym).Scan(&name)
	if err != nil {
		log.Fatal("FetchString_str failed: ", err)
	}
	return name
}

//Given a number, return a string, as stored in the database
func FetchString(db DbDetails, sym int) string {
	if Debug {
		log.Println("Looking up ", sym)
	}
	stmt, err := db.db.Prepare("SELECT s FROM strings WHERE id = ?")
	if err != nil {
		log.Fatal("FetchString failed: ", err)
	}
	defer stmt.Close()
	var name string
	err = stmt.QueryRow(sym).Scan(&name)
	if err != nil {
		log.Fatal("FetchString failed: ", err)
	}
	if Debug {
		log.Println("Found ", name)
	}
	return name
}

//Query the quadgram table, return the fourth gram, given the first 3.  This will be unique (or non-existent)
func FetchCForABC(db DbDetails, strs []string) []string {
	if Debug {
		log.Println("Looking up ", strs)
	}
	var out []string
	stmt, err := db.db.Prepare("SELECT DISTINCT c FROM trigram_symbols WHERE a = ? AND b = ? AND c=?")
	if err != nil {
		log.Println("FetchCforABC failed: ", err)
		return nil
	}
	defer stmt.Close()
	rows, err := stmt.Query(FetchSymbol(db, strs[0]), FetchSymbol(db, strs[1]), FetchSymbol(db, strs[2]))
	if err != nil {
		log.Println("FetchCforABC failed: ", err)
		return nil
	}
	for rows.Next() {
		var c string
		err = rows.Scan(&c)
		if err != nil {
			log.Println("FetchCforABC failed: ", err)
		}
		out = append(out, FetchString_str(db, c))
	}
	rows.Close()
	return out
}

//Query the trigrams table, return the grams, matching the first one
func FetchBForA(db DbDetails, strs []string) []string {
	if Debug {
		log.Println("Looking up ", strs)
	}
	out := []string{}
	stmt, err := db.db.Prepare("SELECT DISTINCT b FROM trigram_symbols WHERE a = ?")
	if err != nil {
		log.Println("FetchBForA failed:", err)
		return nil
	}
	defer stmt.Close()
	rows, err := stmt.Query(FetchSymbol(db, strs[0]))
	if err != nil {
		log.Println("FetchBForA failed:", err)
		return nil
	}

	tmp := []string{}
	for rows.Next() {
		var c string
		err = rows.Scan(&c)
		if err != nil {
			log.Fatal("FetchBForA failed:", err)
		}

		if Debug {
			log.Println("Found trigam: ", c)
		}
		tmp = append(tmp, c)
	}
	rows.Close()
	for _, c := range tmp {

		out = append(out, FetchString_str(db, c))
	}
	return out
}

//Query the quadgrams table, return the fourth gram, given the first three
func FetchDForABC(db DbDetails, strs []string) []string {
	if Debug {
		log.Println("Looking up ", strs)
	}
	out := []string{}
	stmt, err := db.db.Prepare("SELECT DISTINCT d FROM quadgram_symbols WHERE a = ? AND b = ? and c = ?")
	if err != nil {
		log.Println("FetchDForABC failed:", err)
		return nil
	}
	defer stmt.Close()
	rows, err := stmt.Query(FetchSymbol(db, strs[0]), FetchSymbol(db, strs[1]), FetchSymbol(db, strs[2]))
	if err != nil {
		log.Println("FetchDForABC failed:", err)
		return nil
	}

	tmp := []string{}
	for rows.Next() {
		var c string
		err = rows.Scan(&c)
		if err != nil {
			log.Fatal("FetchDForABC failed:", err)
		}

		tmp = append(tmp, c)
		if Debug {
			log.Println("Found quadgam: ", c)
		}
	}
	rows.Close()
	for _, c := range tmp {
		out = append(out, FetchString_str(db, c))
	}

	return out
}

//Query the trigrams table, return the third gram, given the first two
func FetchCForAB(db DbDetails, strs []string) []string {
	if Debug {
		log.Println("Looking up ", strs)
	}
	var out []string
	stmt, err := db.db.Prepare("SELECT c FROM trigram_symbols WHERE a = ? AND b = ?")
	if err != nil {
		log.Println("FetchCForAB failed:", err)
		return nil
	}
	defer stmt.Close()
	rows, err := stmt.Query(FetchSymbol(db, strs[0]), FetchSymbol(db, strs[1]))
	if err != nil {
		log.Println("FetchCForAB failed:", err)
		return nil
	}

	tmp := []string{}
	for rows.Next() {
		var c string
		err = rows.Scan(&c)
		if err != nil {
			log.Fatal("FetchCForAB failed:", err)
		}

		tmp = append(tmp, c)
		if Debug {
			log.Println("Found trigam: ", c)
		}
	}

	rows.Close()
	for _, c := range tmp {
		out = append(out, FetchString_str(db, c))
	}
	return out
}

func InsertStringCached(db DbDetails, str string) int {
	key := fmt.Sprintf("%v-%v", db.url, str)
	if symbol_cache == nil {
		symbol_cache = map[string]int{}
	}
	if val, ok := symbol_cache[key]; ok {
		return val
	}
	val := InsertString(db, str)
	symbol_cache[key] = val
	return val
}

//Insert a string into the symbol table, return the id of the new symbol
func InsertString(db DbDetails, str string) int {
	db.db.Exec("INSERT OR IGNORE INTO strings (s) VALUES (?)", str)
	sym := FetchSymbol(db, str)
	if sym < 0 {
		panic("Could not retrieve the symbol we just inserted!!!!")
	}
	return sym
}

//FIXME check errors
func InsertTrigramCached(db DbDetails, strs []string) {
	db.db.Exec("INSERT OR IGNORE INTO trigram_symbols (a,b,c) VALUES (?,?,?)", InsertStringCached(db, strs[0]), InsertStringCached(db, strs[1]), InsertStringCached(db, strs[2]))
}

func InsertQuadgramCached(db DbDetails, strs []string) {
	if Debug {
		log.Println("Inserting a = ", strs[0], " b = ", strs[1], " c = ", strs[2], " d = ", strs[3])
		log.Println("Inserting a = ", InsertStringCached(db, strs[0]), " b = ", InsertStringCached(db, strs[1]), " c = ", InsertStringCached(db, strs[2]), " d = ", InsertStringCached(db, strs[3]))
	}
	db.db.Exec("INSERT OR IGNORE INTO quadgram_symbols (a,b,c,d) VALUES (?,?,?,?)", InsertStringCached(db, strs[0]), InsertStringCached(db, strs[1]), InsertStringCached(db, strs[2]), InsertStringCached(db, strs[3]))
}

//FIXME check errors
func InsertTrigram(db DbDetails, strs []string) {
	db.db.Exec("INSERT OR IGNORE INTO trigram_symbols (a,b,c) VALUES (?,?,?)", InsertString(db, strs[0]), InsertString(db, strs[1]), InsertString(db, strs[2]))
}

//Given a number, find the matching string in the symbol table
func FetchSymbol(db DbDetails, str string) int {
	if Debug {
		log.Println("Looking up ", str)
	}
	stmt, err := db.db.Prepare("SELECT id FROM strings WHERE s = ?")
	if err != nil {
		log.Fatal("FetchSymbol prepare failed for string '", str, ": ", err)
	}
	defer stmt.Close()
	var id int
	err = stmt.QueryRow(str).Scan(&id)
	if err != nil {
		//log.Printf("FetchSymbol failed for string '%v': %v", str, err)
		return -1
		log.Fatal("FetchSymbol failed: ", err)
	}
	if Debug {
		log.Println("Found ", id)
	}
	return id
}

//Open an sqlite trigram database
func OpenDB(path string) (DbDetails, error) {
	log.Printf("Opening: %v\n", path)
	dbHandle, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Println(err)
		return DbDetails{}, err
	}

	_, err = dbHandle.Exec("PRAGMA synchronous = OFF;PRAGMA journal_mode = WAL;")
	if err != nil {
		log.Println(err)
		return DbDetails{}, err
	}
	_, err = dbHandle.Exec("CREATE TABLE IF NOT EXISTS trigrams ( a string, b string, c string, UNIQUE (a,b,c) ON CONFLICT IGNORE);")
	if err != nil {
		log.Println(err)
		return DbDetails{}, err
	}
	_, err = dbHandle.Exec("CREATE TABLE IF NOT EXISTS trigram_symbols( a INT, b INT, c INT, UNIQUE (a,b,c) ON CONFLICT IGNORE);")
	if err != nil {
		log.Println(err)
		return DbDetails{}, err
	}
	_, err = dbHandle.Exec("CREATE TABLE IF NOT EXISTS quadgram_symbols( a INT, b INT, c INT, d INT, UNIQUE (a,b,c,d) ON CONFLICT IGNORE);")
	if err != nil {
		log.Println(err)
		return DbDetails{}, err
	}
	_, err = dbHandle.Exec("CREATE TABLE IF NOT EXISTS strings ( id INTEGER PRIMARY KEY AUTOINCREMENT, s string, UNIQUE (s) ON CONFLICT IGNORE);")
	if err != nil {
		log.Println(err)
		return DbDetails{}, err
	}

	return DbDetails{dbHandle, path, path}, nil
}

func TopTenX(db DbDetails, colName string) map[string]int {
	out := map[string]int{}
	//FIXME sanitise XXXX
	stmt, err := db.db.Prepare(fmt.Sprintf("SELECT strings.s, count(%v) FROM trigram_symbols JOIN strings ON %v=id GROUP BY %v ORDER BY count(%v) DESC limit 10;", colName, colName, colName, colName))
	if err != nil {
		log.Fatal("Prepare failed in TopTenA: ", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Println("Query for toptenA failed:", err)
		return nil
	}
	for rows.Next() {
		var score int
		var key string
		err = rows.Scan(&key, &score)

		if err != nil {
			log.Fatal("FetchRow failed in TopTenA: ", err)
		}
		out[key] = score
	}
	rows.Close()
	return out

}

func formatResultList(l []string) []string {
	out := []string{}
	for _, v := range l {
		out = append(out, fmt.Sprintf("[%v]", v))
	}
	return out
}

func makeArgs(line string) []string {
	args := strings.Split(line, " ")
	return args[1:]
}

//Query all open trigram databases, for top ten lists and merge
//
//Counts the frequency of first words in the trigram databases, sorts
//them by frequency and returns the top 10
func TopTenA(dbs dbList, str string) map[string]int {
	out := map[string]int{}
	for _, db := range dbs {
		for k, v := range TopTenX(db, str) {
			out[k] = out[k] + v
		}
	}
	return out
}

//Query all open trigram databases, merge the results and return them (FIXME eliminate dupes)
//
//Given the first word in a trigram, returns all the know second words
func QueryAGetB(dbs dbList, strs []string) []string {
	out := []string{}
	for _, db := range dbs {
		results := FetchBForA(db, strs)
		for _, str := range results {
			out = append(out, str)
		}
	}
	return out
}

//Query all open trigram databases, merge the results and return them (FIXME eliminate dupes)
//
//Given the first word in a quadgram, returns all the known fourth words
func QueryABCGetD(dbs dbList, strs []string) []string {
	out := []string{}
	for _, db := range dbs {
		results := FetchDForABC(db, strs)
		for _, str := range results {
			out = append(out, str)
		}
	}
	return out
}

//Query all open trigram databases, merge the results and return them (FIXME eliminate dupes)
//
//Given the first two words in a trigram, returns all known third words
func QueryAB(dbs dbList, strs []string) []string {
	var out []string
	for _, db := range dbs {
		for _, str := range FetchCForAB(db, strs) {
			out = append(out, str)
		}
	}
	return out
}

//Check that a row exists in the trigrams table
func QueryABC(dbList dbList, strs []string) []string {
	var out []string
	for _, db := range dbList {
		out = append(out, FetchCForABC(db, strs)...)
	}
	return out
}

//Trims the whitespaces from every element in an array of strings.
func TrimWords(words []string) []string {
	out := []string{}
	for _, v := range words {
		v := strings.Trim(v, "\"'.,!? \r\n\t")
		/*
		   v = strings.Replace(v, "\"", "", -1)
		   v = strings.Replace(v, "'", "", -1)
		   v = strings.Replace(v, ".", "", -1)
		   v = strings.Replace(v, ",", "", -1)
		   v = strings.Replace(v, "!", "", -1)
		   v = strings.Replace(v, "?", "", -1)
		   v = strings.Replace(v, " ", "", -1)
		*/
		out = append(out, v)
	}
	return out
}

//Score a sentence against the trigram database
//
//Score() breaks the sentence up into trigrams, then searches the database for each trigram
//If the trigram is found, we add one point to the words from the trigram.
func Score(databases []DbDetails, args []string) []int {
	score := make([]int, len(args)+3)
	words := TrimWords(args)
	for i, _ := range words {
		if i < len(words)-2 {
			results := QueryABC(databases, []string{words[i], words[i+1], words[i+2]})
			if len(results) > 0 {
				score[i]++
				score[i+1]++
				score[i+2]++
			}
		}
	}
	return score
}
