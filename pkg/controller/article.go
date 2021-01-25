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

func (a *Articles) GetArticles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		allArticles, err := a.db.AllArticles()
		must(err)

		section := r.FormValue("section")
		tag := r.FormValue("tag")

		var filteredArticles []models.Article
		for i := range allArticles {
			if section != allArticles[i].Section {
				continue
			}

			if tag != "" {
				for j := range allArticles[i].Tags {
					if tag == allArticles[i].Tags[j] {
						filteredArticles = append(filteredArticles, allArticles[i])
						break
					}
				}
			} else {
				filteredArticles = append(filteredArticles, allArticles[i])
			}
		}

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "    ")
		err = encoder.Encode(&filteredArticles)
		if err != nil {
			_, _ = fmt.Fprintf(w, "Sorry! Internal error. If you can tell me about this, it would be great!")
		}

	default:
		_, _ = fmt.Fprintf(w, "The request is not supported.")
	}
}

func (a *Articles) GetTags(w http.ResponseWriter, r *http.Request) {
	uniqueTags := make(map[string]struct{})

	allArticles, err := a.db.AllArticles()
	must(err)

	for i := range allArticles {
		for j := range allArticles[i].Tags {
			uniqueTags[allArticles[i].Tags[j]] = struct{}{}
		}
	}

	tags := make([]string, 0, len(uniqueTags)) // Create a slice with length 0 and capacity = len(uniqueTags)
	for tag := range uniqueTags {
		tags = append(tags, tag)
	}

	err = json.NewEncoder(w).Encode(&tags)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Sorry! Internal error. If you can tell me about this, it would be great!")
	}
}

func (a *Articles) GetSections(w http.ResponseWriter, r *http.Request) {
	uniqueSections := make(map[string]struct{})

	allArticles, err := a.db.AllArticles()
	must(err)

	for i := range allArticles {
		uniqueSections[allArticles[i].Section] = struct{}{}
	}

	sections := make([]string, 0, len(uniqueSections))
	for section := range uniqueSections {
		sections = append(sections, section)
	}

	err = json.NewEncoder(w).Encode(&sections)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Sorry! Internal error. If you can tell me about this, it would be great!")
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
