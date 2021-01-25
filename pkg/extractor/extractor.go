package extractor

import (
	"github.com/jdkato/prose/v2"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

var (
	extraSubString = []string{
		"mr.",
		"ms.",
		"mrs.",
	}

	commonString = []string{ // Used to filter out the common tags
		"congress",
		"state",
		"department",
		"north",
		"south",
		"east",
		"west",
	}
)

func ExtractTextFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer func() {
		err = resp.Body.Close()
	}()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, nil
}

// Recursively extract the <p> tag in HTML string.
func ExtractText(s string) (string, error) {
	var paragraph string

	doc, err := html.Parse(strings.NewReader(s))

	if err != nil {
		return "", err
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode && n.Parent.Data == "p" {
			paragraph = paragraph + n.Data
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return paragraph, nil
}

func normaliseTags(tags []string) ([]string, error) {
	/*
		1. lowercase everything
		2. Remove all the "extra bits" in the tags, such as titles (Mr., Ms., ...)
		3. Remove common words such as Congress, State, Department, North, South, ...
		4. Trim trailing white space and some random punctuation
		5. Remove duplicate tags that might caused by all the operations above
	*/

	// 1. lowercase tags
	for i := range tags {
		tags[i] = strings.ToLower(tags[i])
	}

	// 2. Remove extra sub-string
	for i := range tags {
		for _, s := range extraSubString {
			tags[i] = strings.ReplaceAll(tags[i], s, "")
		}
	}

	var newTags []string

	isStringInside := func(s string, list []string) bool {
		for _, sToCheck := range list {
			if s == sToCheck {
				return true
			}
		}
		return false
	}

	// 3. Remove common words
	for _, tag := range tags {
		if !(isStringInside(tag, commonString)) {
			newTags = append(newTags, tag)
		}
	}

	// 4. Trimming
	for i := range newTags {
		newTags[i] = strings.Trim(newTags[i], " !@#$%^&*(),./<>?;':{}[]|\\\"-=_+~`")
	}

	// 5. Extract unique tags
	uniqueTagsMap := make(map[string]struct{})
	for _, tag := range newTags {
		uniqueTagsMap[tag] = struct{}{}
	}

	var uniqueTags []string
	for key := range uniqueTagsMap {
		uniqueTags = append(uniqueTags, key)
	}

	return uniqueTags, nil
}

func ExtractTags(text string, numberOfTags int) ([]string, error) {
	// Use prose package to extract the tags

	// After prose parses the text, we extract the unique label of each entity
	// and records the number of its occurrence
	doc, err := prose.NewDocument(text)
	if err != nil {
		return nil, err
	}

	uniqueLabel := make(map[string]int)
	for _, ent := range doc.Entities() {
		uniqueLabel[ent.Text] = uniqueLabel[ent.Text] + 1
	}

	// Since uniqueLabel is not sorted (it is a map), we need to sort it
	// by transform it into slice and then sort the slice
	type label struct {
		Name  string
		Count int
	}

	var sortedUniqueLabel []label
	for key, value := range uniqueLabel {
		sortedUniqueLabel = append(sortedUniqueLabel, label{key, value})
	}

	sort.Slice(sortedUniqueLabel, func(i, j int) bool {
		return sortedUniqueLabel[i].Count > sortedUniqueLabel[j].Count
	})

	// What we really interested in is the sorted label in descending order
	// Called the slice of string "tags" to differentiate with sortedUniqueLabel
	var tags []string
	for i, label := range sortedUniqueLabel {
		if i < numberOfTags {
			tags = append(tags, label.Name)
		} else {
			break
		}
	}

	return normaliseTags(tags)
}
