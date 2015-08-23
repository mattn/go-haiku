package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-haiku"
	"golang.org/x/net/html"
)

func walk(node *html.Node, buff *bytes.Buffer) {
	if node.Type == html.TextNode {
		data := strings.Trim(node.Data, "\r\n ")
		if data != "" {
			buff.WriteString("\n")
			buff.WriteString(data)
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		walk(c, buff)
	}
}

func text(reader io.Reader) (string, error) {
	var buffer bytes.Buffer
	doc, err := html.Parse(reader)
	if err != nil {
		return "", err
	}
	walk(doc, &buffer)
	return buffer.String(), nil
}

func rules(s string) ([]int, error) {
	r := []int{}
	for _, t := range strings.Split(s, ",") {
		i, err := strconv.Atoi(t)
		if err != nil {
			return nil, err
		}
		r = append(r, i)
	}
	return r, nil
}

func main() {
	u := flag.Bool("u", false, "handle as URL")
	rs := flag.String("r", "5,7,5", "rule of match (default: 5,7,5)")
	flag.Parse()

	r, err := rules(*rs)
	if err != nil {
		flag.Usage()
	}
	args := flag.Args()
	if len(args) == 0 {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		args = []string{string(b)}
	}
	for _, arg := range args {
		if *u {
			res, err := http.Get(arg)
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()
			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}
			arg = string(b)
		}
		for _, h := range haiku.Find(arg, r) {
			fmt.Println(h)
		}
	}
}
