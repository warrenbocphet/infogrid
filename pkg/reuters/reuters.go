package reuters

import (
	"github.com/vitsensei/infogrid/pkg/extractor"
	"golang.org/x/net/html"
	"regexp"
	"strings"
	"time"
)

const (
	reuterBasedURL = "https://www.reuters.com/"
)

type Article struct {
	URL            string `json:"url"`
	Title          string `json:"title"`
	Section        string `json:"section"`
	DateCreated    string `json:"published_date"`
	Text           string
	SummarisedText string
	Tags           []string
}

type API struct {
	urls     map[string]string
	articles []Article
}

func NewAPI() *API {
	urls := map[string]string{
		"world":      "https://www.reuters.com/news/world",
		"technology": "https://www.reuters.com/news/technology"}

	return &API{urls: urls}
}

func (a *API) GenerateArticles() error {
	for section, url := range a.urls {
		articles, err := generateArticles(url)
		if err != nil {
			return err
		}

		for i := range articles {
			articles[i].Section = section
			articles[i].DateCreated = time.Now().String()

			text, err := ExtractText(articles[i].URL)
			if err == nil {
				articles[i].Text = text

				tags, err := extractor.ExtractTags(text, 3)
				if err == nil {
					articles[i].Tags = tags
				}
			}

		}

		a.articles = append(a.articles, articles...)
	}

	return nil
}

func generateArticles(url string) ([]Article, error) {
	var articles []Article

	bodyString, err := extractor.ExtractTextFromURL(url)

	doc, err := html.Parse(strings.NewReader(bodyString))

	if err != nil {
		return nil, err
	}

	var storyContentNodes []*html.Node

	var findStoryConent func(*html.Node)
	findStoryConent = func(node *html.Node) {
		if isStoryConent(node) {
			storyContentNodes = append(storyContentNodes, node)
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			findStoryConent(c)
		}
	}
	findStoryConent(doc)

	for _, node := range storyContentNodes {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			isContainedLink, link := isContainLink(c)
			if isContainedLink {
				newArticle := Article{URL: link}

				for t := c.FirstChild; t != nil; t = t.NextSibling {
					isContainedTitle, title := isContainTitle(t)
					if isContainedTitle {
						newArticle.Title = title
						break
					}
				}

				articles = append(articles, newArticle)

				break
			}
		}
	}

	return articles, nil
}

func isStoryConent(node *html.Node) bool {
	for _, a := range node.Attr {
		if a.Key == "class" && a.Val == "story-content" {
			return true
		}
	}

	return false
}

func isContainLink(node *html.Node) (bool, string) {
	if node.Data == "a" {
		for _, a := range node.Attr {
			if a.Key == "href" {
				return true, reuterBasedURL + a.Val
			}
		}
	}

	return false, ""
}

func isContainTitle(node *html.Node) (bool, string) {
	if node.Data == "h3" {
		for _, a := range node.Attr {
			if a.Key == "class" && a.Val == "story-title" {
				return true, strings.TrimSpace(node.FirstChild.Data)
			}
		}
	}

	return false, ""
}

func (a *API) GetArticles() []Article {
	return a.articles
}

func isArticleBody(n *html.Node) bool {
	for _, a := range n.Attr {
		if a.Key == "class" && a.Val == "ArticleBodyWrapper" {
			return true
		}
	}

	return false
}

func hasArticleText(n *html.Node) bool {
	for _, a := range n.Attr {
		result, _ := regexp.MatchString("Paragraph", a.Val)
		if a.Key == "class" && result {
			return true
		}
	}

	return false
}

func ExtractText(url string) (string, error) {
	var paragraph string

	bodyString, err := extractor.ExtractTextFromURL(url)

	doc, err := html.Parse(strings.NewReader(bodyString))

	if err != nil {
		return "", err
	}

	var articleBodyNode *html.Node

	// All the actual writing is in ArticleBodyWrapper node. Find this node
	// and extract text from it to avoid extracting rubbish
	var findArticleBodyNode func(*html.Node)
	findArticleBodyNode = func(n *html.Node) {
		if isArticleBody(n) {
			articleBodyNode = n
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findArticleBodyNode(c)
		}
	}
	findArticleBodyNode(doc)

	// Given the article body node, extract the text
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode && n.Parent.Data == "p" && hasArticleText(n.Parent) {
			paragraph = paragraph + n.Data + "\n"
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}

	}
	f(articleBodyNode)

	return paragraph, nil
}
