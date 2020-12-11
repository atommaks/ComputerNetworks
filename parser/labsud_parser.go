package main

import (
	"fmt"
	"github.com/mgutz/logxi/v1"
	"golang.org/x/net/html"
	"net/http"
	"strings"
)

const (
	address = "http://lab-sud.ru"
	xml_url = "http://lab-sud.ru/sitemap.xml"
)

var (
	data []*Item
	xmlData []string
	isEverything = true
)

type Item struct {
	Title, Text, Ref string
}

func main () {
	data = make([]*Item, 0)
	log.Info("Download started")
	downloadServices("http://lab-sud.ru/uslugi/", 0)
	log.Info("Download finished")
	for _, v := range data {
		fmt.Println("Title: " + v.Title)
		fmt.Println("Article: " + v.Text)
		fmt.Println("Ref: " + v.Ref)
		fmt.Println("-------------------------------------------------------------------")
	}
	fmt.Println("")
	fmt.Println("-----     Сравнение url с xml-файлом     -----")
	fmt.Println()
	xmlData = make([]string, 0)
	downloadXML(xml_url)
	for _, v := range data {
		if !contains(xmlData, v.Ref) {
			fmt.Println("Отсутствует - " + v.Ref)
			isEverything = false
		}
	}
	fmt.Println("")
	if isEverything {
		fmt.Println("Все ссылки присутствуют в xml-файле, программсит - красавчик!")
	} else {
		fmt.Println("К сожалению, не все ссылки есть в xml-файле. Программист - дурачок. Пусть учит CEO")
	}
	fmt.Printf("Размер моего массива - %d\nРазмер массива xml - %d\n", len(data), len(xmlData))
}

func contains (arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

func downloadServices (url string, num uint) {
	log.Info("sending request to " + url)
	if response, err := http.Get(url); err != nil {
		fmt.Println("request to " + url + " failed")
		log.Error("request to " + url + " failed", "error", err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		log.Info("got response from " + url, "status", status)
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Error("invalid HTML from " + url, "error", err)
			} else {
				log.Info("HTML from " + url + " parsed successfully")
				if num == 0 {
					getService(doc)
				} else if num == 1 {
					getSubService(doc)
				} else if num == 2 {
					getSubSubService(doc)
				}
			}
		}
	}
}

func readItem (item *html.Node) *Item {
	if a := item.FirstChild; isElem(a, "a") {
		if cs := getChildren(a); len(cs) == 1 && isText(cs[0]) {
			title := translateText(cs[0].Data)
			ref := address + getAttr(a, "href")
			text := translateText(getText(getTextDoc(ref)))
			return &Item{
				Title: title,
				Ref: ref,
				Text: text,
			}
		}
	}
	return  nil
}

func getService (node *html.Node) {
	if isDiv(node, "col-sm-4") {
		for _, v := range getChildren(node.FirstChild.NextSibling) {
			if v != nil {
				if item := readItem(v); item != nil {
					data = append(data, item)
					downloadServices(item.Ref, 1)
				}
			}
		}
		return
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		getService(c)
	}
}

func getSubService (node *html.Node) {
	if isDiv(node, "col-sm-4") {
		for _, v := range getChildren(node.FirstChild.NextSibling) {
			if v != nil {
				if isLi(v, "next") {
					for _, u := range getChildren(v.FirstChild) {
						if u != nil {
							if item := readItem(u); item != nil {
								data = append(data, item)
								downloadServices(item.Ref, 2)
							}
						}
					}
					return
				}
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		getSubService(c)
	}
}

func getSubSubService (node *html.Node) {
	if isDiv(node, "col-sm-4") {
		for _, v := range getChildren(node.FirstChild.NextSibling) {
			if v != nil {
				if isLi(v, "next") {
					for _, u := range getChildren(v.FirstChild) {
						if u != nil {
							if isLi(u, "next") {
								for _, w := range getChildren(u.FirstChild) {
									if w != nil {
										if item := readItem(w); item != nil {
											data = append(data, item)
										}
									}
								}
								return
							}
						}
					}
				}
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		getSubSubService(c)
	}
}

func getTextDoc (url string) *html.Node {
	log.Info("sending request to " + url)
	if response, err := http.Get(url); err != nil {
		fmt.Println("request to " + url + " failed")
		log.Error("request to " + url + " failed", "error", err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		log.Info("got response from " + url, "status", status)
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Error("invalid HTML from " + url, "error", err)
			} else {
				log.Info("HTML from " + url + " parsed successfully")
				return doc
			}
		}
	}
	return nil
}

func getBottomText (node *html.Node) string {
	text := ""
	for c := node; c != nil; c = c.NextSibling {
		if isElem(c, "h1") {
			for s := c; s != nil; s = s.NextSibling {
				if isElem(s, "p") {
					for _, v := range getChildren(s) {
						if isText(v) {
							text += v.Data
						}
					}
				}
			}
			return text
		}
	}
	return text
}

func getText (node *html.Node) string {
	if isDiv(node, "main_display") {
		text := ""
		for _, v := range getChildren(node) {
			if v != nil {
				if isElem(v, "p") {
					for _, u := range getChildren(v) {
						if isText(u) {
							text += u.Data
						}
					}
				}
			}
		}
		return text + getBottomText(node)
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if text := getText(c); text != "" {
			return  text
		}
	}

	return ""
}

func translateText (text string) string {
	str := ""
	for i := 0; i < len(text); i++ {
		if int(text[i]) >= 0 && int(text[i]) <= 127 {
			str += string(int(text[i]))
		} else {
			str += string(int(text[i]) + 848)
		}
	}

	return str
}

func downloadXML (url string) {
	log.Info("sending request to " + url)
	if response, err := http.Get(url); err != nil {
		fmt.Println("request to " + url + " failed")
		log.Error("request to " + url + " failed", "error", err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		log.Info("got response from " + url, "status", status)
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Error("invalid HTML from " + url, "error", err)
			} else {
				log.Info("HTML from " + url + " parsed successfully")
				getXML(doc)
			}
		}
	}
}

func getXML(node *html.Node) {
	if isElem(node, "url") {
		for _, v := range getChildren(node) {
			if v != nil {
				if isElem(v, "loc") {
					str := strings.Split(v.FirstChild.Data, "/")
					if str[3] == "uslugi" || str[3] == "main" {
						xmlData = append(xmlData, v.FirstChild.Data)
					}
				}
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		getXML(c)
	}
}

func getChildren (node *html.Node) []*html.Node {
	var children []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}
	return children
}

func getAttr (node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func isLi (node *html.Node, class string) bool {
	return isElem(node, "li") && getAttr(node, "class") == class
}

func isDiv (node *html.Node, class string) bool {
	return isElem(node, "div") && getAttr(node, "class") == class
}

func isText (node *html.Node) bool {
	return node != nil && node.Type == html.TextNode
}

func isElem (node *html.Node, tag string) bool {
	return node != nil && node.Type == html.ElementNode && node.Data == tag
}