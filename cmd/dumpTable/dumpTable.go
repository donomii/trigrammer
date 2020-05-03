package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func dumpTable(rows *sql.Rows, w *csv.Writer) error {
	colNames, err := rows.Columns()
	if err != nil {
		panic(err)
	}
	w.Write(colNames)
	readCols := make([]interface{}, len(colNames))
	writeCols := make([]sql.NullString, len(colNames))
	for i, _ := range writeCols {
		readCols[i] = &writeCols[i]
	}
	for rows.Next() {
		err := rows.Scan(readCols...)
		if err != nil {
			panic(err)
		}
		//out.Write(writeCols)
		//fmt.Println(colNames)
		//fmt.Println(writeCols)
		out := []string{}
		for i, _ := range colNames {
			out = append(out, fmt.Sprintf("%v", writeCols[i].String))
		}
		w.Write(out)
		//fmt.Println("")
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	//out.Flush()
	return nil
}

func perlDsnToMysql(username, password, dsn string) string {
	dsn = strings.Replace(dsn, "dbi:mysql:host=", "tcp(", 1)
	dsn = strings.Replace(dsn, ";", ":3306)/", 1)
	dsn = strings.Replace(dsn, "database=", "", 1)
	mysql_dsn := fmt.Sprintf("%v:%v@%v", username, password, dsn)
	return mysql_dsn
}

func checkErr(name string, err error) {
	if err != nil {
		log.Fatal("Error in ", name, ": ", err)
	}
}

func main() {
	if len(os.Args) < 4 {
		log.Fatalln("Use: dumpTable server username password database-name")
	}
	addr, user, pass, dbName := os.Args[1], os.Args[2], os.Args[3], os.Args[4]

	/*
	   tableName := os.Args[2]

	   stmt, err := db.Query(fmt.Sprintf("SELECT * FROM %v;", tableName))
	   checkErr("Querying database", err)
	*/

	w := csv.NewWriter(os.Stdout)

	//db, err := sql.Open("mysql", dbName)
	//db := mysql.New(proto, "", addr, user, pass, dbname)
	connect := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, addr, dbName)
	log.Printf("Connecting with: %s\n", connect)
	db, err := sql.Open("mysql", connect)
	checkErr("Querying database", err)
	dumpTables(db, w)
}

func dumpTables(db *sql.DB, w *csv.Writer) error {

	rows, err := db.Query(fmt.Sprintf("SELECT TABLE_SCHEMA, TABLE_NAME, COLUMN_NAME, COLUMN_TYPE FROM information_schema.columns;"))
	checkErr("Querying database", err)
	colNames, err := rows.Columns()
	if err != nil {
		panic(err)
	}
	w.Write(colNames)
	readCols := make([]interface{}, len(colNames))
	writeCols := make([]sql.NullString, len(colNames))
	for i, _ := range writeCols {
		readCols[i] = &writeCols[i]
	}
	for rows.Next() {
		err := rows.Scan(readCols...)
		if err != nil {
			panic(err)
		}
		//out.Write(writeCols)
		//fmt.Println(colNames)
		//fmt.Println(writeCols)
		out := []string{}
		for i, _ := range colNames {
			out = append(out, fmt.Sprintf("%v", writeCols[i].String))
		}
		w.Write(out)
		//fmt.Println("")
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	w.Flush()
	return nil
}
