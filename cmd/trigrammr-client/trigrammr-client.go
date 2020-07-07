package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	_ "database/sql"

	"github.com/chzyer/readline"
	"github.com/donomii/trigrammr"
	_ "github.com/mattn/go-sqlite3"
)

var debug bool
var shortDisplay bool
var databases []trigrammr.DbDetails
var categories = []string{"Category A", "Category B", "Category C"}

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
}

// Function constructor - constructs new function for listing given directory
func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("mode",
		readline.PcItem("vi"),
		readline.PcItem("emacs"),
	),
	readline.PcItem("login"),
	readline.PcItem("say",
		readline.PcItemDynamic(listFiles("./"),
			readline.PcItem("with",
				readline.PcItem("following"),
				readline.PcItem("items"),
			),
		),
		readline.PcItem("hello"),
		readline.PcItem("bye"),
	),
	readline.PcItem("setprompt"),
	readline.PcItem("setpassword"),
	readline.PcItem("bye"),
	readline.PcItem("help"),
	readline.PcItem("go",
		readline.PcItem("build", readline.PcItem("-o"), readline.PcItem("-v")),
		readline.PcItem("install",
			readline.PcItem("-v"),
			readline.PcItem("-vv"),
			readline.PcItem("-vvv"),
		),
		readline.PcItem("test"),
	),
	readline.PcItem("sleep"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func formatResultList(l []string) []string {
	out := []string{}
	for i, v := range l {
		if !shortDisplay || i < 30 {
			out = append(out, fmt.Sprintf("[%v]", v))
		} else if i == 30 {
			out = append(out, "...")
		}
	}
	return out
}

func makeArgs(line string) []string {
	args := strings.Split(line, " ")
	return args[1:]
}

var searchGram []string

func main() {
	shortDisplay = true
	var dbFiles = os.Args[1:]
	for _, file := range dbFiles {
		db, err := trigrammr.OpenDB(file)
		if err == nil {
			databases = append(databases, db)
		}
	}
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	l.SetVimMode(true)
	l = updateCompleter(l, categories)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	setPasswordCfg := l.GenPasswordConfig()
	setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		l.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
		l.Refresh()
		return nil, 0, false
	})

	log.SetOutput(l.Stderr())
	for {
		//fmt.Println("Current search: ", searchGram)
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		args := makeArgs(line)
		switch {
		case strings.HasPrefix(line, "mode "):
			switch line[5:] {
			case "vi":
				l.SetVimMode(true)
			case "emacs":
				l.SetVimMode(false)
			default:
				println("invalid mode:", line[5:])
			}
		case strings.HasPrefix(line, "summary"):
			fmt.Println(trigrammr.TopTenA(databases, "A"))
		case strings.HasPrefix(line, "summarise"):
			fmt.Println(trigrammr.TopTenA(databases, "A"))
		case strings.HasPrefix(line, "summarize"):
			fmt.Println(trigrammr.TopTenA(databases, "A"))
		case strings.HasPrefix(line, "score "):
			score := trigrammr.Score(databases, args)
			fmt.Println("Score:")
			for k, v := range args {
				fmt.Printf("%v(%v) ", v, score[k])
			}
			fmt.Println()
		case strings.HasPrefix(line, "f "):
			fmt.Println(formatResultList(trigrammr.QueryAB(databases, makeArgs(line))))
		case strings.HasPrefix(line, "db "):
			db, err := trigrammr.OpenDB(fmt.Sprintf("%v.sqlite", args[0]))
			if err == nil {
				databases = append(databases, db)
			}
		case line == "reset":
			searchGram = []string{}
		case line == ".":
			searchGram = searchGram[0:1]
			doSearch(l, searchGram)
		case line == "..":
			searchGram = searchGram[0 : len(searchGram)-1]
			doSearch(l, searchGram)
		case line == "mode":
			if l.IsVimMode() {
				println("current mode: vim")
			} else {
				println("current mode: emacs")
			}
		case line == "login":
			/*pswd, err := l.ReadPassword("please enter your password: ")
			  if err != nil {
			      break
			  }*/
			//println("you enter:", strconv.Quote(string(pswd)))
		case line == "help":
			usage(l.Stderr())
		case line == "setpassword":
			/*pswd, err := l.ReadPasswordWithConfig(setPasswordCfg)
			  if err == nil {
			      //println("you set:", strconv.Quote(string(pswd)))
			  }*/
		case strings.HasPrefix(line, "setprompt"):
			if len(line) <= 10 {
				log.Println("setprompt <prompt>")
				break
			}
			l.SetPrompt(line[10:])
		case strings.HasPrefix(line, "say"):
			line := strings.TrimSpace(line[3:])
			if len(line) == 0 {
				log.Println("say what?")
				break
			}
			go func() {
				for range time.Tick(time.Second) {
					log.Println(line)
				}
			}()
		case strings.HasPrefix(line, "dump"):
			if len(searchGram) < 2 {
				log.Println("Don't do that yet, you need more search terms")
			} else {
				args := makeArgs(line)
				if len(args) > 0 {
					term := args[0]
					results := trigrammr.QueryABCGetD(databases, []string{searchGram[0], searchGram[1], term})
					fmt.Println("Found ", len(results), " results")
					for _, v := range results {
						fmt.Printf("%v %v %v\n", searchGram, term, v)
					}
				} else {
					res1 := trigrammr.QueryAB(databases, searchGram[len(searchGram)-2:])
					for _, term := range res1 {
						results := trigrammr.QueryABCGetD(databases, []string{searchGram[0], searchGram[1], term})
						//fmt.Println("Found ", len(results), " results")
						for _, v := range results {
							fmt.Printf("%v %v %v\n", searchGram, term, v)
						}
					}
				}
			}
		case line == "long":
			shortDisplay = false
			doSearch(l, searchGram)
		case line == "short":
			shortDisplay = true
			doSearch(l, searchGram)
		case line == "":
		default:
			searchGram = append(searchGram, line)
			if len(searchGram) > 2 {
				searchGram = searchGram[1:]
			}
			doSearch(l, searchGram)
		}
		l.SetPrompt(fmt.Sprintf("%v \033[31m»\033[0m ", searchGram))
		fmt.Println()
	}
}

func doSearch(l *readline.Instance, searchGram []string) {
	fmt.Println()
	if len(searchGram) > 1 {
		//log.Println("Searching for pair ", searchGram)
		res1 := trigrammr.QueryAB(databases, searchGram[len(searchGram)-2:])
		fmt.Println(formatResultList(res1))
		l = updateCompleter(l, trigrammr.QueryAB(databases, searchGram[len(searchGram)-2:]))
		term := searchGram[len(searchGram)-1:]
		fmt.Println("\nAdditional results for only ", term, "")
		res2 := trigrammr.QueryAGetB(databases, term)
		fmt.Println(formatResultList(diff(res1, res2)))
	} else if len(searchGram) > 0 {
		//log.Println("Searching for single ", searchGram)
		fmt.Println(formatResultList(trigrammr.QueryAGetB(databases, searchGram[0:1])))
		l = updateCompleter(l, trigrammr.QueryAGetB(databases, searchGram[0:1]))
	}
}

func diff(a, b []string) []string {
	if len(b) > len(a) {
		c := b
		b = a
		a = c
	}

	a_h := map[string]bool{}
	b_h := map[string]bool{}
	res := []string{}
	for _, v := range a {
		a_h[v] = true
	}
	for _, v := range b {
		b_h[v] = true
	}
	for ka, _ := range a_h {
		_, ok := b_h[ka]
		if !ok {
			res = append(res, ka)
		}
	}
	return res
}

func updateCompleter(l *readline.Instance, categories []string) *readline.Instance {
	var items []readline.PrefixCompleterInterface

	for _, category := range categories {
		items = append(items, readline.PcItem(category))
	}
	items = append(items, readline.PcItem("dump", items...))
	items = append(items, readline.PcItem("reset"))
	items = append(items, readline.PcItem("summary"))
	items = append(items, readline.PcItem("summarise"))
	items = append(items, readline.PcItem("summarize"))
	items = append(items, readline.PcItem("score"))
	items = append(items, readline.PcItem("db"))
	items = append(items, readline.PcItem("help"))

	completer = readline.NewPrefixCompleter(items...)

	l.Config.AutoComplete = completer
	return l
}
