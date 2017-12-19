package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func main() {

	foundUrls := make(map[string]bool)
	seedUrls := os.Args[1:]
	chUrls := make(chan string)
	chFinished := make(chan bool)

	webname := ("http://" + strings.Join(os.Args[1:], " ") + "/")
	//webname := ("http://localhost:8000/")

	fmt.Println("Check reposnse from: " + webname)

	_, err := http.Get(webname)
	if err != nil {
		fmt.Println("\n" + webname + " - Shit, site doesn't working")
	} else {

		go deepSearch(webname, chUrls, chFinished)

		for c := 0; c < len(seedUrls); {
			select {
			case url := <-chUrls:
				foundUrls[url] = true
			case <-chFinished:
				c++
			}
		}

		fmt.Println("\nFound", len(foundUrls), "unique urls:")

		for url := range foundUrls {
			fmt.Println(" - " + url)
		}

		close(chUrls)
	}

}

func getHref(t html.Token) (ok bool, href string) {

	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}

	return
}

func deepSearch(url string, ch chan string, chFinished chan bool) {

	resp, _ := http.Get(url)

	defer func() {
		// Notify that we're done after this function
		chFinished <- true
	}()

	b := resp.Body
	defer b.Close() // close Body when the function returns

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			// Check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// Extract the href value, if there is one
			ok, url := getHref(t)
			if !ok {
				continue
			}

			// Make sure the url begines in http**
			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				ch <- url
			}
		}
	}
}
