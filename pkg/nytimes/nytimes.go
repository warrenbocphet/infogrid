package nytimes

import (
	"encoding/json"
	"github.com/vitsensei/infogrid/pkg/extractor"
	"github.com/vitsensei/infogrid/pkg/models"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	apiKey             = os.Getenv("NYTIMES_KEY")
	partialTopStoryURL = "https://api.nytimes.com/svc/topstories/v2/home.json?api-key="
	wg                 sync.WaitGroup
)

// The json of the response from NYTimes API
type TopStories struct {
	Articles []models.Article `json:"results"`
}

// The API for other package to interact with
type API struct {
	url             string
	allowedSections []string
	TopStories      TopStories `json:"body"`
}

func NewAPI() *API {
	return &API{
		allowedSections: []string{"business", "politics", "technology", "us", "world"},
	}
}

// Used in controller/article to filter out the "non-news" sections
func (a *API) FilterBySections() {
	var filteredArticles []models.Article

	for _, article := range a.TopStories.Articles {
		for _, allowedSection := range a.allowedSections {
			if article.Section == allowedSection {
				filteredArticles = append(filteredArticles, article)
				break
			}
		}
	}

	a.TopStories.Articles = filteredArticles

}

func (a *API) generateURL() {
	a.url = partialTopStoryURL + apiKey
}

// Used in ExtractText to detect ArticleBody node
func isArticleBody(n html.Node) bool {
	for _, a := range n.Attr {
		if a.Key == "name" && a.Val == "articleBody" {
			return true
		}
	}

	return false
}

// Given a URL, the text will be extracted (if exist)
func ExtractText(url string) (string, error) {
	var paragraph string

	bodyString, err := extractor.ExtractTextFromURL(url)

	doc, err := html.Parse(strings.NewReader(bodyString))

	if err != nil {
		return "", err
	}

	var articleBodyNode *html.Node

	// All the actual writing is in Article Body node. Find this node
	// and extract text from it to avoid extracting rubbish
	var findArticleBodyNode func(*html.Node)
	findArticleBodyNode = func(n *html.Node) {
		if isArticleBody(*n) {
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
		if n.Type == html.TextNode && n.Parent.Data == "p" {
			paragraph = paragraph + n.Data + "\n"
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}

	}

	// NYTimes loves interactive articles (and they are amazing!). Unfortunately, it is not
	// the usual text format and therefore cannot be extract
	// (for example: https://www.nytimes.com/interactive/2020/obituaries/people-died-coronavirus-obituaries.html#lloyd-porter).
	// Most likely they don't have an article node, and we will skip those interactive ones.
	if articleBodyNode != nil {
		f(articleBodyNode)
	}

	return paragraph, nil
}

func GenerateArticleText(article *models.Article) {
	defer wg.Done()

	text, _ := ExtractText(article.URL)
	if text != "" {
		article.Text = text

		tags, err := extractor.ExtractTags(text, 3)
		if err == nil {
			article.Tags = tags
		}
	}
}

//	Construct the Article list (TopStories struct).
//	Each Article in the list will only contain the URL, Section, and Title
//	after this call. These are the value returned from NYTimes API.
func (a *API) GenerateArticles() error {
	if a.url == "" {
		a.generateURL()
	}

	resp, err := http.Get(a.url)

	if err != nil {
		return err
	}
	defer func() {
		err = resp.Body.Close()
	}()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &a.TopStories)
	if err != nil {
		return err
	}

	a.FilterBySections()

	// Extract text from URL
	for i := range a.TopStories.Articles {
		wg.Add(1)
		go GenerateArticleText(&a.TopStories.Articles[i])
	}

	wg.Wait()

	// Filter out the node that is interactive ~= article.text == ""
	var articleWithText []models.Article
	for i := range a.TopStories.Articles {
		if a.TopStories.Articles[i].Text != "" {
			articleWithText = append(articleWithText, a.TopStories.Articles[i])
		}
	}

	a.TopStories.Articles = articleWithText

	return err
}

// A Get-Set style function to exposes the the articles array
// through interface
func (a *API) GetArticles() []models.Article {
	return a.TopStories.Articles
}
