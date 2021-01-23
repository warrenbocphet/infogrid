package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/vitsensei/infogrid/pkg/controller"
	"github.com/vitsensei/infogrid/pkg/models"
	"github.com/vitsensei/infogrid/pkg/nytimes"
	"github.com/vitsensei/infogrid/pkg/views/articles"
	"net/http"
	"os"
	"time"
)


var (
	//mongoURI = "mongodb://localhost:27017/"
	mongoURI = "mongodb://" + os.Getenv("MONGO_HST") + ":" + os.Getenv("MONGO_PRT") + "/"
)

func main() {
	// Create Database
	adb := models.NewDB()
	err := adb.Init(mongoURI)
	defer adb.Close()
	must(err)

	err = adb.DestructiveReset()
	must(err)
	fmt.Println("Finished deleting previous database")
	// Create API and controller
	nytimesAPI := nytimes.NewAPI()
	must(err)

	views := articles.NewView("display", "articles/simple_display")

	ac := controller.NewArticleController(adb, views, nytimesAPI)

	go ac.RunPeriodicCapture(4)

	// Create router
	r := mux.NewRouter()
	r.HandleFunc("/", ac.ShowArticles)
	r.HandleFunc("/articles", ac.ShowArticlesJSON)
	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	err = srv.ListenAndServe()
	must(err)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
