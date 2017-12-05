package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-encoding"
	"github.com/mattn/go-haiku"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
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
		switch strings.ToLower(node.Data) {
		case "script":
			continue
		}
		walk(c, buff)
	}
}

func text(resp *http.Response) (string, error) {
	br := bufio.NewReader(resp.Body)
	var r io.Reader = br
	if data, err := br.Peek(1024); err == nil {
		enc, name, _ := charset.DetermineEncoding(data, resp.Header.Get("content-type"))
		if enc != nil {
			r = enc.NewDecoder().Reader(br)
		} else if name != "" {
			if enc := encoding.GetEncoding(name); enc != nil {
				r = enc.NewDecoder().Reader(br)
			}
		}
	}

	var buffer bytes.Buffer
	doc, err := html.Parse(r)
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
	u := flag.String("u", "", "handle as URL")
	d := flag.String("d", "", "user dic")
	rs := flag.String("r", "5,7,5", "rule of match (default: 5,7,5)")
	flag.Parse()

	r, err := rules(*rs)
	if err != nil {
		flag.Usage()
	}
	args := flag.Args()
	if len(args) == 0 && *u == "" {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		args = []string{string(b)}
	}
	http.DefaultTransport.(*http.Transport).TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)

	if *u != "" {
		resp, err := http.Get(*u)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		s, err := text(resp)
		if err != nil {
			log.Fatal(err)
		}
		args = []string{s}
	}
	for _, arg := range args {
		res, err := haiku.FindWithOpt(arg, r, &haiku.Opt{Udic: *d})
		if err != nil {
			log.Println(err)
			continue
		}
		for _, h := range res {
			fmt.Println(h)
		}
	}
}
