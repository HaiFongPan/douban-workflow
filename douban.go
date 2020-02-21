package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"gopkg.in/resty.v1"
	"net/url"
	"os"
	"strings"
)

type UrlItem struct {
	Url string
}

type SearchResultItem struct {
	Title     string
	Url       string
	Subtitles string
}

type AlfredItem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Arg      string `json:"arg"`
	Icon     struct {
		Path string `json:"path"`
	} `json:"icon"`
}

var urlMapping = map[string]UrlItem{
	"movie": {
		Url: "https://www.douban.com/search?cat=1002&q=%s",
	},
	"music": {
		Url: "https://www.douban.com/search?cat=1003&q=%s",
	},
	"book": {
		Url: "https://www.douban.com/search?cat=1001&q=%s",
	},
}

func getNodeAttr(node *html.Node, attrName string) string {
	for _, a := range node.Attr {
		if a.Key == attrName {
			return a.Val
		}
	}
	return ""
}

func getItems(searchType string, searchString string) *[]SearchResultItem {
	if v, ok := urlMapping[searchType]; ok {
		resp, _ := resty.R().Get(fmt.Sprintf(v.Url, searchString))
		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
		return getItemsFromDoc(doc)
	}
	return nil
}

func getItemsFromDoc(doc *goquery.Document) *[]SearchResultItem {
	r := make([]SearchResultItem, 0)
	s := doc.Find("div.search-result > div.result-list > div.result")

	var href, title, score, desc string

	for _, n := range s.Nodes {
		node := goquery.NewDocumentFromNode(n)
		titleNode := node.Find("div.content > div.title a")
		href, _ = url.QueryUnescape(getNodeAttr(titleNode.Nodes[0], "href"))
		start := strings.Index(href, "url=")
		end := strings.Index(href, "\u0026")

		title = titleNode.Text()
		score = node.Find("span.rating_nums").Text()
		desc = node.Find("span.subject-cast").Text()

		r = append(r, SearchResultItem{
			Title:     title,
			Url:       href[start+4 : end],
			Subtitles: "star: " + score + " " + desc,
		})
	}
	return &r
}

func generateResponse(items *[]SearchResultItem, searchType string) {
	r := make([]AlfredItem, 0)
	for _, i := range *items {
		r = append(r, AlfredItem{
			Type:     "file",
			Title:    i.Title,
			Subtitle: i.Subtitles,
			Arg:      i.Url,
			Icon: struct {
				Path string `json:"path"`
			}{
				Path: fmt.Sprintf("imgs/%s.png", searchType),
			},
		})
	}
	finalRes, _ := json.Marshal(struct {
		Items []AlfredItem `json:"items"`
	}{
		Items: r,
	})
	fmt.Println(string(finalRes))
}

func main() {
	searchType := os.Args[1]
	query := strings.Join(os.Args[2:], " ")
	generateResponse(getItems(searchType, query), searchType)
}
