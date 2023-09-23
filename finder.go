package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/exp/maps"
	"strings"
	"sync"
	"time"
)

type ThreeName struct {
	document   *goquery.Document
	list       map[string]*Status
	lastUpdate time.Time
	interval   time.Duration
	mu         sync.Mutex
}

func NewThreeName(every time.Duration) *ThreeName {
	return &ThreeName{list: make(map[string]*Status), interval: every}
}

func (t *ThreeName) Get(username string) *Status {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.refresh() {
		t.parse()
	}

	if username == "" {
		for i := range t.list {
			return t.list[i]
		}
	}
	return t.list[username]
}

func (t *ThreeName) GetByFilter(filter func(index int, name string) bool) *Status {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.refresh() {
		t.parse()
	}

	if filter == nil {
		for i := range t.list {
			return t.list[i]
		}
	}

	var index int
	for name := range t.list {
		if filter(index, name) {
			return t.list[name]
		}
		index++
	}
	return nil
}

func (t *ThreeName) refresh() bool {
	if t.lastUpdate.After(time.Now().Add(-t.interval)) {
		return false
	}
	doc, err := GetDocumentFromURL("https://3name.xyz/list")
	if err != nil {
		panic(err)
	}
	t.document = doc
	t.lastUpdate = time.Now()
	return true
}

func (t *ThreeName) parse() {
	t.document.Find("#body-container > div > div > div:nth-child(1)").Find("a").Each(func(i int, selection *goquery.Selection) {
		if strings.HasPrefix(selection.AttrOr("href", ""), "/name") {
			name := selection.Text()
			t.list[name] = &Status{Username: name, Available: true, drop: [2]time.Time{time.Now(), t.getDroppingEnd(name)}}
		}
	})
	t.document.Find("#timer-section").Find("a").Each(func(i int, selection *goquery.Selection) {
		name := selection.Text()
		t.list[name] = &Status{Username: name, Available: false, drop: [2]time.Time{}}
	})
	names := maps.Keys(t.list)
	t.document.Find("#timer-section").Find("span").Each(func(i int, selection *goquery.Selection) {
		t.list[names[i]].drop[0] = ParseTime(selection.AttrOr("data-lower-bound", ""))
	})

}

func (t *ThreeName) getDroppingEnd(username string) time.Time {
	doc, err := GetDocumentFromURL("https://3name.xyz/name/" + username)
	if err != nil {
		panic(err)
	}

	return ParseTime(doc.Find("#upper-bound-update").AttrOr("data-upper-bound", ""))
}

// NameMC Currently not working
type NameMC struct {
	document *goquery.Document
}

func NewNameMC() *NameMC {
	return &NameMC{}
}

func (n *NameMC) Get(username string) *Status {
	if username == "" {
		return nil
	}
	doc, err := GetDocumentFromURL("https://namemc.com/minecraft-names?sort=asc&length_op=le&length=3&lang=&searches=0")
	if err != nil {
		panic(err)
	}
	n.document = doc
	return &Status{Username: username, Available: n.parse()}
}

func (n *NameMC) parse() bool {
	switch n.document.Find("#status-bar > div > div:nth-child(1) > div:nth-child(2)").Text() {
	case "Available":
		return true
	case "Possibly Available":
		n.dropWindow()
		return true
	case "Locked":
		n.dropWindow()
		return false
	default:
		return false
	}
}

func (n *NameMC) dropWindow() [2]int64 {
	var drop [2]int64
	lower := n.document.Find("#status-bar > div > div:nth-child(3)").Next()
	upper := lower.Next()

	fmt.Println(lower.Text(), upper.Text())
	return drop
}
