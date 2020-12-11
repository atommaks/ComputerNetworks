package main

import (
	"fmt"
	"github.com/mgutz/logxi/v1"
	"golang.org/x/net/html"
	"net/http"
)

var topics = [4]string{"/finances", "/politics", "/society", "/technology_and_media"}

const (
	introduction_msg = "Разделы:\n1)Финансы\n2)Политика\n3)Общество\n4)Технологии и медиа\nВведите число, " +
		"интересующего вас раздела: "
	wrongNumber_msg = "Введите число, которое входит в диапозон\n"
)

type Item struct {
	Title, ImgRef, ArticleRef, ArticleText string
}

func main() {
	for {
		n := 0
		fmt.Print(introduction_msg)

		fmt.Scanf("%d", &n)
		if n > 0 && n < 5 {
			log.Info("Downloader started")
			items := downloadNews(topics[n - 1])
			for k, v := range items {
				fmt.Printf("%d)\n", k + 1)
				fmt.Println(v.Title)
				fmt.Println(v.ArticleText)
				fmt.Println(v.ArticleRef)
				fmt.Println(v.ImgRef)
			}
			log.Info("Download finished")
			break
		}
		fmt.Print(wrongNumber_msg)
	}
}

func downloadNews(topic string) []*Item {
	log.Info("sending request to https://www.rbc.ru" + topic)
	if response, err := http.Get("https://www.rbc.ru" + topic); err != nil {
		fmt.Println("request to https://www.rbc.ru" + topic + " failed")
		log.Error("request to https://www.rbc.ru" + topic + " failed", "error", err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		log.Info("got response from https://www.rbc.ru" + topic, "status", status)
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Error("invalid HTML from https://www.rbc.ru" + topic, "error", err)
			} else {
				log.Info("HTML from https://www.rbc.ru" + topic + " parsed successfully")
				return search(doc)
			}
		}
	}
	return nil
}

func search(node *html.Node) []*Item {
	if isDiv(node, "l-row js-load-container") {
		var items []*Item
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			//для статей с картинками
			if isDiv(c, "item js-category-item") {
				for d := c.FirstChild; d != nil; d = d.NextSibling {
					if isDiv(d, "item__wrap l-col-center") {
						if item := readItemWithImage(d); item != nil {
							items = append(items, item)
						}
					}
				}
			//для статей без картинок
			} else if isDiv(c, "item item_no-image js-category-item") {
				for d := c.FirstChild; d != nil; d = d.NextSibling {
					if isDiv(d, "item__wrap l-col-center") {
						if item := readItemWithNoImage(d); item != nil {
							items = append(items, item)
						}
					}
				}
			}
		}
		return items
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if items := search(c); items != nil {
			return items
		}
	}
	return nil
}

func readItemWithNoImage(item *html.Node) *Item {
	for _, v := range getChildren(item) {
		if isElem(v, "a") {
			cs := getChildren(v)
			if len(cs) == 3 && isSpan(cs[1], "item__title-wrap") {
				var title, articleRef string
				title = cs[1].FirstChild.NextSibling.FirstChild.Data
				title = title[21 : lastIndex(title) - 1]
				articleRef = getAttr(v, "href")

				return &Item{
					Title: 		 	title,
					ImgRef:  	 	"no-image",
					ArticleRef:	 	articleRef,
					ArticleText: 	downloadArticle(articleRef),
				}
			}
		}
	}
	return nil
}

func readItemWithImage(item *html.Node) *Item {
	for _, v := range getChildren(item) {
		if isElem(v, "a") {
			cs := getChildren(v)
			if len(cs) == 5 && isSpan(cs[1], "item__title-wrap") && isSpan(cs[3], "item__image-block") {
				var title, imgRef, articleRef string
				title = cs[1].FirstChild.NextSibling.FirstChild.Data
				title = title[21 : lastIndex(title) - 1]
				articleRef = getAttr(v, "href")
				for imgBlock := cs[3].FirstChild; imgBlock != nil; imgBlock = imgBlock.NextSibling {
					if len(imgBlock.Attr) == 3 {
						imgRef = getAttr(imgBlock, "src")
						break
					}
				}

				return &Item{
					Title: 			title,
					ImgRef:  		imgRef,
					ArticleRef: 	articleRef,
					ArticleText: 	downloadArticle(articleRef),
				}
			}
		}
	}
	return nil
}

func downloadArticle(url string) string {
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
				return getArticle(doc)
			}
		}
	}
	return ""
}

func getArticle (node *html.Node) string {
	//заходим в нужный div. Далее, если в нем натыкаемся на тэг p, получаем все дочерние элементы p и выводим их
	var text = ""
	if isDiv(node, "article__text article__text_free") {
		for c := node.FirstChild; c != nil; c = c.NextSibling{
			if isElem(c, "p") {
				for _, v := range getChildren(c) {
					if isText(v) {
						text += v.Data
					}
					if isElem(v, "a") {
						text += v.FirstChild.Data
					}
				}
			}
		}
		return text
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if t := getArticle(c); t != "" {
			return t
		}
	}
	return text
}

func getChildren(node *html.Node) []*html.Node {
	var children []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}
	return children
}

func getAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func isSpan(node *html.Node, class string) bool {
	return isElem(node, "span") && getAttr(node, "class") == class
}

func isDiv(node *html.Node, class string) bool {
	return isElem(node, "div") && getAttr(node, "class") == class
}

func isText(node *html.Node) bool {
	return node != nil && node.Type == html.TextNode
}

func isElem(node *html.Node, tag string) bool {
	return node != nil && node.Type == html.ElementNode && node.Data == tag
}

//функция получения индекса последнего символа, отличающегося от пробела
func lastIndex(str string) (l int) {
	l = 0
	space := false
	for i := 21; i < len(str); i++ {
		if str[i] == ' ' && space {
			return
		} else if str[i] == ' ' {
			l = i
			space = true
		} else {
			space = false
		}
	}

	return
}