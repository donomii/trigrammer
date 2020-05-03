# trigrammr
Explore trigram databases

trigrammr is a highly experimental tool for exploring data.  Using trigrammr, you can navigate your data, explore links between data and reveal hidden connections.

## Get it

    go get -u github.com/donomii/trigrammr

## Build it

    go build cmd/trigrammr-import-csv/trigrammr-import-csv.go
    go build cmd/trigrammr-client/trigrammr-client.go
    go build cmd/trigrammr-import-book/trigrammr-import-book.go 

## Install it


    go install github.com/donomii/trigrammr/...


## Import data

    cat data.csv | trigrammr-import-csv mydb.sqlite

or

    cat example.txt | ./trigrammr-import-book cat.sqlite

trigrammr-import-csv reads the data from STDIN and saves it in cat.sqlite.

## Large example database

You can download a large example database by typing 

    bash download_wikipedia.bash

You will need to install BASH, PERL and CURL for this to work.  They come pre-installed on Linux and MacOSX.

## Explore your database

### Load a database

    > ./trigrammr-client cat.sqlite

or

    > ./trigrammr-client
    » db cat
    Opening: cat.sqlite

Before you can search for anything, you must load the data from a database.

### Get a summary of your database

    » summary
    map[the:1 cat:1 sat:1 on:1]

Print a summary of the top ten words in the database.

### Explore links

Note that all words are imported as lower-case.

Search for the word "the"

    » the
    
    [[cat]]

Here, the [[cat]] is the only known word that follows [[the]].  Trigrammr displays every bigram (2-gram) that starts with [[the]].

Typing "cat" displays every trigram (3-gram) that matches [[the]] [[cat]].

    » cat

    Searching for  [the cat]
    
    [[sat]]

### Scoring

Trigrammr can also assist with text analysis.  The "score" command will take a sentence and print out the number of trigrams that match each word.

    » score The cat sat on the mat
    Score:
    the(1) cat(2) sat(3) on(3) the(2) mat(1)


Scoring prints out the number of matches for a sentence.  The sentence is broken up into trigrams, and each trigram is looked up in the database.  If the trigram exists, we increase the score of each word from the trigram.

Note that the end words can only score 1, while "sat" scores 3, because it is part of three trigrams.


    » score The cat stood on the mat

    Score:
    the(0) cat(0) stood(0) on(1) the(1) mat(1)

This scores much worse, because the word "stood" is not in the database, and any trigram that contains it will fail.

## Commands

Trigrammr has several commands to help you navigate the database.

### db NAME

Loads database NAME.sqlite from disk

### summary

Prints out the top ten words in the database, to help you get started on a search

### short

Trims the output if it gets too long

### long

Prints the entire search output, no matter how many pages it might take up

### dump WORD

Trigrammer stores the original record when it is imported.  "dump" will print out all records that match WORD

### reset

Resets the search (to nothing)

### ..

(Two dots) moves back to the previous search

### .

(One dot) drops the first word in the current search trigram.


## TODO

  * Fix autocomplete (add builtin commands, finish on dump etc)
  * Count trigrams for markov models
  * Convert download script to golang
