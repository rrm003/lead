package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/net/html"
)

const (
	siteURL = "https://www.chessgames.com/chessecohelp.html"
)

type Wiki struct {
	Key   string
	Name  string
	Value string
}

type Response struct {
	Data []Wiki
}

var result = map[string][]string{}

func FetchData() (map[string][]string, error) {
	cli := http.Client{}

	url := siteURL
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		log.Fatal(err)
	}

	var f func(*html.Node)
	count := 0
	key := ""
	flag := false
	f = func(n *html.Node) {
		if n.Data == "tr" {
			flag = true
		}
		if n.Type == html.TextNode && flag {
			if count == 0 {
				key = n.Data
			} else if count < 2 && n.Data != "" {
				result[key] = append(result[key], n.Data)
			} else if count == 3 {
				result[key] = append(result[key], n.Data)
				key = ""
				count = -1
				flag = false
			}
			count++
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	return result, err
}

func main() {
	data, _ := FetchData()
	var wiki = []Wiki{}
	t, err := template.ParseFiles("resp.html")
	if err != nil {
		fmt.Println(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for key, val := range data {
			temp := Wiki{
				Key:   key,
				Name:  val[0],
				Value: val[1],
			}
			wiki = append(wiki, temp)
		}
		result := Response{
			Data: wiki,
		}
		t.Execute(w, result)
	})

	r.HandleFunc("/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["key"]
		if len(data[key]) < 2 {
			http.NotFound(w, r)
			return
		}
		temp := Wiki{
			Key:   key,
			Name:  data[key][0],
			Value: data[key][1],
		}
		wiki = append(wiki, temp)
		result := Response{
			Data: wiki,
		}
		t.Execute(w, result)
	})

	log.Fatal(http.ListenAndServe(":8081", r))

}
