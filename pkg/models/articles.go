package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sort"
	"time"
)

type ArticleDB struct {
	ctx        context.Context
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func NewDB() *ArticleDB {
	return &ArticleDB{}
}

func (adb *ArticleDB) Init(uri string) error {
	var err error
	adb.client, err = mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	adb.ctx = context.Background()
	err = adb.client.Connect(adb.ctx)
	if err != nil {
		return err
	}

	adb.database = adb.client.Database("info_grid")
	adb.collection = adb.database.Collection("articles")

	return nil
}

func (adb *ArticleDB) Close() error {
	return adb.client.Disconnect(adb.ctx)
}

// Remove all documents in "articles" collection by
// dropping the collection and create it again.
func (adb *ArticleDB) DestructiveReset() error {
	err := adb.collection.Drop(adb.ctx)
	if err != nil {
		return err
	}

	adb.collection = adb.database.Collection("articles")

	return nil
}

// The document that goes into the (mongo) database.
type Article struct {
	URL            string   `bson:"url,omitempty" json:"url"`
	Title          string   `bson:"title,omitempty" json:"title"`
	Section        string   `bson:"section,omitempty" json:"section"`
	PublishedDate  string   `bson:"date_created,omitempty" json:"published_date"`
	Text           string   `bson:"text,omitempty" json:"-"`
	SummarisedText string   `bson:"summarised_text,omitempty"`
	Tags           []string `bson:"tags,omitempty"`
}

// Insert an article/document into the mongo database
func (adb *ArticleDB) InsertArticle(a Article) error {
	_, err := adb.collection.InsertOne(adb.ctx, a)
	if err != nil {
		return err
	}

	return nil
}

func (adb *ArticleDB) AllArticles() ([]Article, error) {
	c, err := adb.collection.Find(adb.ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var articles Articles
	err = c.All(adb.ctx, &articles)
	if err != nil {
		return nil, err
	}

	sort.Sort(articles)
	return articles, nil

}

// ByURL is used in controller packages to check for existing
// article, assuming the URL of the article never changed and each article
// has an unique url.
func (adb *ArticleDB) ByURL(url string) (*Article, error) {
	filter := bson.M{"url": url}

	var article Article
	err := adb.collection.FindOne(adb.ctx, filter).Decode(&article)

	if err != nil {
		return nil, err
	}

	return &article, nil
}

// Query the articles by tags and sections
func (adb *ArticleDB) BySectionsAndTags(sections []string, tags []string) ([]Article, error) {
	var filter bson.M

	if len(sections) > 0 && len(tags) > 0 {
		filter = bson.M{
			"section": bson.M{
				"$in": sections,
			},
			"tags": bson.M{
				"$all": tags,
			},
		}
	} else if len(sections) > 0 {
		filter = bson.M{
			"section": bson.M{
				"$in": sections,
			},
		}
	} else {
		filter = bson.M{
			"tags": bson.M{
				"$all": tags,
			},
		}
	}

	var articles []Article
	c, err := adb.collection.Find(adb.ctx, filter)

	if err != nil {
		return nil, err
	}

	err = c.All(adb.ctx, &articles)
	if err != nil {
		return nil, err
	}

	return articles, nil
}

// Query the articles by sections
func (adb *ArticleDB) BySections(sections []string) ([]Article, error) {
	return adb.BySectionsAndTags(sections, []string{})
}

// Query the articles by tags
func (adb *ArticleDB) ByTags(tags []string) ([]Article, error) {
	return adb.BySectionsAndTags([]string{}, tags)
}

type Articles []Article

func (as Articles) Len() int {
	return len(as)
}

func (as Articles) Less(i, j int) bool {
	layout := "2006-01-02 15:04:05 -0700 MST"
	time1, err := time.Parse(layout, as[i].PublishedDate)
	if err != nil {
		return true
	}

	time2, err := time.Parse(layout, as[j].PublishedDate)
	if err != nil {
		return true
	}

	return time1.Before(time2)
}

func (as Articles) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

// Delete old articles
func (adb *ArticleDB) CleanOldArticles(numberOfArticles int, logger *log.Logger) {
	articles, err := adb.AllArticles()
	if err != nil {
		return
	}

	if len(articles) < numberOfArticles {
		return
	}

	layout := "2006-01-02 15:04:05 -0700 MST"
	for i := range articles {
		publishedDate, err := time.Parse(layout, articles[i].PublishedDate)
		if err != nil {
			logger.Println(err)
			continue
		}

		if time.Since(publishedDate).Hours() > 72 {
			_, err = adb.collection.DeleteOne(adb.ctx, bson.M{"url": articles[i].URL})
			if err != nil {
				logger.Println("[ERROR] Fail to delete article with title", articles[i].Title)
			} else {
				logger.Println("[INFO] Delete article with title", articles[i].Title)
			}
		} else {
			// Since adb.AllArticles() return articles sorted by their published date (old to new),
			// we can just break the loop when first encounter articles that is less than 3 days old.
			break
		}
	}
}
