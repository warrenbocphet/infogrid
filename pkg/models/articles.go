package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	URL            string   `bson:"url,omitempty"`
	Title          string   `bson:"title,omitempty"`
	Section        string   `bson:"section,omitempty"`
	DateCreated    string   `bson:"date_created,omitempty"`
	Text           string   `bson:"text,omitempty" json:"-"`
	SummarisedText string   `bson:"summarised_text,omitempty"`
	Tags           []string `bson:"tags,omitempty"`
}

// The interface for all articles.
// We need this interface instead of using a struct from other package
// there are more than one news source and each agency might have different
// structure for their article.
type ArticleInterface interface {
	GetURL() string
	GetTitle() string
	GetSection() string
	GetDateCreated() string
	SetText(string)
	GetText() string
	SetSummarised(string)
	GetSummarised() string
	GetTags() []string
}

// Insert an article/document into the mongo database
func (adb *ArticleDB) InsertArticle(a ArticleInterface) error {
	url := a.GetURL()
	title := a.GetTitle()
	section := a.GetSection()
	text := a.GetText()
	summarisedText := a.GetSummarised()
	dateCreated := a.GetDateCreated()
	tags := a.GetTags()

	article := Article{
		URL:            url,
		Title:          title,
		Section:        section,
		Text:           text,
		SummarisedText: summarisedText,
		DateCreated:    dateCreated,
		Tags:           tags,
	}

	_, err := adb.collection.InsertOne(adb.ctx, article)
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

	var articles []Article
	err = c.All(adb.ctx, &articles)
	if err != nil {
		return nil, err
	}

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
