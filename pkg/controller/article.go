package controller

import (
	"encoding/json"
	"fmt"
	"github.com/vitsensei/infogrid/pkg/models"
	"github.com/vitsensei/infogrid/pkg/textrank"
	"github.com/vitsensei/infogrid/pkg/views/articles"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

type API interface {
	GenerateArticles() error
	GetArticles() []models.ArticleInterface
}

func NewArticleController(db *models.ArticleDB, v *articles.View, api ...API) Articles {
	return Articles{
		apis:        api,
		db:          db,
		ArticleView: v,
	}
}

type Articles struct {
	apis []API
	tags []string
	db   *models.ArticleDB

	ArticleView *articles.View
}

// CaptureArticles will be called by RunPeriodicCapture at every constant
// time period. This is exported for debugging purposes in main.go
func (a *Articles) CaptureArticles() {
	var as []models.ArticleInterface

	// Get the articles from selected sections
	// and then get the summarised version of the text
	for _, api := range a.apis {
		err := api.GenerateArticles()
		if err == nil {
			for _, article := range api.GetArticles() {
				_, err = a.db.ByURL(article.GetURL())
				if err == mongo.ErrNoDocuments {
					if article.GetSummarised() == "" {
						t, err := textrank.NewText(article.GetText(), nil)
						if err == nil {
							summarisedText := t.Summarise(0.05)
							article.SetSummarised(summarisedText)

						}
					}
					as = append(as, article)
				}
			}
		}
	}

	// Insert the article into the database
	for _, article := range as {
		_ = a.db.InsertArticle(article)
	}
}

func (a *Articles) CaptureTags() {
	as, err := a.db.AllArticles()
	if err != nil {
		return
	}

	uniqueTags := make(map[string]struct{})
	for _, a := range as {
		for _, tag := range a.Tags {
			uniqueTags[tag] = struct{}{}
		}
	}

	var tags []string
	for key := range uniqueTags {
		tags = append(tags, key)
	}

	a.tags = tags

}

// Continuously running until the program exits.
func (a *Articles) RunPeriodicCapture(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Hour)
	start := time.Now()
	a.CaptureArticles()
	a.CaptureTags()
	fmt.Println("New articles captured after", time.Since(start))
	go func() {
		for {
			select {
			case <-ticker.C:
				start := time.Now()
				a.CaptureArticles()
				a.CaptureTags()
				fmt.Println("New articles captured after", time.Since(start))
			}
		}
	}()

	select {}
}

func (a *Articles) ShowArticles(w http.ResponseWriter, r *http.Request) {
	as, err := a.db.AllArticles()
	must(err)

	type Data struct {
		Tags     []string
		Articles []models.Article
	}

	data := Data{a.tags, as}

	err = a.ArticleView.Render(w, data)
	must(err)

}

func (a *Articles) ShowArticlesJSON(w http.ResponseWriter, r *http.Request) {
	as, err := a.db.AllArticles()
	must(err)

	type Data struct {
		Tags     []string
		Articles []models.Article
	}

	data := Data{a.tags, as}

	switch r.Method {
	case "GET":
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "    ")
		err = encoder.Encode(data)
		//err = json.NewEncoder(w).Encode(data)
		if err != nil {
			_, _ = fmt.Fprintf(w, "Sorry! Internal error. If you can tell me about this, it would be great!")
		}
	default:
		_, _ = fmt.Fprintf(w, "The request is not supported.")
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
