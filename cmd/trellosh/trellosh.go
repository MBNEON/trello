package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hackerlist/trello"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type Creds struct {
	Key, Secret, Token, Member, Organization string
}

var (
	creds Creds
	c     *trello.Client
	load  sync.Once
)

func loadCreds() error {
	b, err := ioutil.ReadFile(os.Getenv("HOME") + "/.trelloshrc")

	if err != nil {
		fmt.Printf("No credentials found.\n")
		return err
	}

	err = json.Unmarshal(b, &creds)

	if err != nil {
		fmt.Printf("Invalid credentials.\n")
		return err
	}

	creds.Key = "09f16319e72a2488397b119be7560215"
	return nil
}

func main() {
	flag.Parse()
	loadCreds()

	c = trello.New(creds.Key, creds.Secret, creds.Token)

	sc := bufio.NewScanner(os.Stdin)

	me, err := c.Member(creds.Member)

	if err != nil {
		fmt.Printf("error loading me: %s", err)
		return
	}

	boardsf := func() {
		if boards, err := me.Boards(); err != nil {
			fmt.Printf("error: %s\n", err)
		} else {
			boardsprint(boards)
		}
	}

	fmt.Printf("trellosh> ")
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 1 {
			boardsf()
		} else {
			switch f[0] {
			case "boards":
				boardsf()
			case "board":
				if len(f) > 1 {
					boardrepl(f[1], sc)
				} else {
					fmt.Println("missing board argument\n")
				}
			default:
				fallthrough
			case "help":
				fmt.Printf("commands:\n")
				for _, cmd := range []string{"boards", "board id"} {
					fmt.Printf("  %s\n", cmd)
				}
			case "exit":
				return
			}
		}
		fmt.Printf("trellosh> ")
	}
}

func boardsprint(boards []trello.Board) {
	fmt.Printf("%-24.24s %-20.20s %-24.24s:\n", "id", "name", "shorturl")
	for _, b := range boards {
		fmt.Printf("%24.24s %-20.20s %-24.24s\n", b.Id(), b.Name(), b.ShortUrl())
	}
}

func boardrepl(id string, sc *bufio.Scanner) error {
	board, err := c.Board(id)
	if err != nil {
		fmt.Printf("load board error: %s\n", err)
		return err
	}

	var last func()

	lists := func() {
		if lists, err := board.Lists(); err != nil {
			fmt.Printf("lists error: %s\n", err)
		} else {
			listsprint(lists)
		}
	}

	last = lists

	last()

	fmt.Printf("board %s> ", board.Name())
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 1 {
			if last != nil {
				last()
			}
		} else {
			switch f[0] {
			case "lists":
				last = lists
				last()
			case "list":
				if len(f) > 1 {
					listrepl(f[1], sc)
				} else {
					fmt.Println("missing list argument\n")
				}
			default:
				fallthrough
			case "help":
				fmt.Printf("commands:\n")
				for _, cmd := range []string{"lists", "list id"} {
					fmt.Printf("  %s\n", cmd)
				}
			case "exit":
				return nil
			}
		}
		fmt.Printf("board %s> ", board.Name())
	}
	return sc.Err()
}

func listsprint(lists []trello.List) {
	fmt.Printf("%-24.24s %-20.20s\n", "id", "name")
	for _, l := range lists {
		fmt.Printf("%-24.24s %-20.20s\n", l.Id(), l.Name())
	}
}

func listrepl(id string, sc *bufio.Scanner) error {
	list, err := c.List(id)
	if err != nil {
		fmt.Printf("load list error: %s\n", err)
		return err
	}

	var last func()

	cards := func() {
		if cards, err := list.Cards(); err != nil {
			fmt.Printf("cards error: %s\n", err)
		} else {
			cardsprint(cards)
		}
	}

	last = cards

	last()

	fmt.Printf("list %s> ", list.Name())
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 1 {
			if last != nil {
				last()
			}
		} else {
			switch f[0] {
			case "cards":
				last = cards
				last()
			case "card":
				if len(f) > 1 {
					cardrepl(f[1], sc)
				} else {
					fmt.Println("missing card argument\n")
				}
			default:
				fallthrough
			case "help":
				fmt.Printf("commands:\n")
				for _, cmd := range []string{"cards", "card id"} {
					fmt.Printf("  %s\n", cmd)
				}
			case "exit":
				return nil
			}
		}
		fmt.Printf("list %s> ", list.Name())
	}
	return sc.Err()
}

func cardsprint(cards []trello.Card) {
	fmt.Printf("%-24.24s %-20.20s\n", "id", "name")
	for _, c := range cards {
		fmt.Printf("%-24.24s %-20.20s\n", c.Id(), c.Name())
	}
}

func cardrepl(id string, sc *bufio.Scanner) error {
	card, err := c.Card(id)
	if err != nil {
		fmt.Printf("load card error: %s\n", err)
		return err
	}

	var last func()

	actions := func() {
		if actions, err := card.Actions(); err != nil {
			fmt.Printf("cards error: %s\n", err)
		} else {
			actionsprint(actions)
		}
	}

	last = actions

	last()

	fmt.Printf("card %s> ", card.Name())
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 1 {
			if last != nil {
				last()
			}
		} else {
			switch f[0] {
			case "comment":
				if len(f) > 1 {
					card.AddComment(strings.Join(f[1:], " "))
				} else {
					fmt.Printf("usage: comment text...\n")
				}
			case "actions":
				last = actions
				last()
			default:
				fallthrough
			case "help":
				fmt.Printf("commands:\n")
				for _, cmd := range []string{"actions", "comment text"} {
					fmt.Printf("  %s\n", cmd)
				}
			case "exit":
				return nil
			}
		}
		fmt.Printf("card %s> ", card.Name())
	}
	return sc.Err()
}

func actionsprint(actions []trello.Action) {
	fmt.Printf("%-24.24s %-20.20s %-20.20s\n", "id", "type", "text")
	for _, a := range actions {
		fmt.Printf("%-24.24s %-20.20s %-20.20s\n", a.Id(), a.Type(), a.DataText())
	}
}
