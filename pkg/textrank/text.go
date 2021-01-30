package textrank

import (
	"fmt"
	"github.com/jdkato/prose/v2"
	"github.com/vitsensei/infogrid/pkg/graph"
	"math"
	"regexp"
	"sort"
	"strings"
)

var punctuationMarks = map[string]struct{}{
	".":  {},
	"?":  {},
	"!":  {},
	",":  {},
	";":  {},
	":":  {},
	"-":  {},
	"[":  {},
	"]":  {},
	"{":  {},
	"}":  {},
	"(":  {},
	")":  {},
	"'":  {},
	"\"": {},
}

// Sentence struct is part of Text struct, and it is the unit of Text Rank
type Sentence struct {
	Text string // Almost original text, without all the '\n', '\t',
	// or double space between words
	NormalisedText string   // Text after apply normaliseSentence
	Words          []string // Set of unique words representing NormalisedText
	numberOfWords  int      // Number of words (not unique) in this sentences
	Score          float64  // The score used to rank most relevant sentences
}

// The function calculate the number of overlapping words
// We assume Words is sorted in lexicographic order
func (s *Sentence) findSimilarity(anotherS *Sentence) float64 {
	// Step 1: Find number of overlapping words
	originalInd := 0
	anotherInd := 0

	similarity := 0
	for originalInd < len(s.Words) && anotherInd < len(anotherS.Words) {
		if s.Words[originalInd] < anotherS.Words[anotherInd] {
			originalInd++
		} else if s.Words[originalInd] > anotherS.Words[anotherInd] {
			anotherInd++
		} else {
			similarity++

			originalInd++
			anotherInd++
		}
	}

	// Step 2: Use the equation from https://web.eecs.umich.edu/~mihalcea/papers/mihalcea.emnlp04.pdf
	return float64(similarity) / (math.Log(float64(len(s.Words))) + math.Log(float64(len(anotherS.Words))))
}

type Text struct {
	Text          string            // Raw, original text
	numberOfWords int               // Number of words in this text, equal to sum of number of words of all sentences
	lemmaDict     map[string]string // lemmatization list, used in normalising sentence
	doc           *prose.Document   // used for sentences segmentation, might be replaced in the future
	Sentences     []Sentence        // Represent a sentence in a text
	graph         graph.Graph       // Represent the connected graph of Sentences
	windowSize    int               // A window size used for keywords extraction (not yet implemented)
	dampingFactor float64           // the value "d" in section 2.2 (https://web.eecs.umich.edu/~mihalcea/papers/mihalcea.emnlp04.pdf)
	maxIterations int               // Max iteration to calculate sentence's score
	threshold     float64           // The minimum difference between this score and last score of sentences
	// before the calculation stops.
}

// Iterate through all sentences and call findSimilarity.
// The similarity will be the weight of each edge
func (t *Text) findSimilarities() {
	for nodeInd := range t.graph.Nodes {
		node := &t.graph.Nodes[nodeInd]
		for neighborID, weight := range node.Neighbors {
			if weight != -1 {
				continue
			}
			similarity := t.Sentences[node.ID].findSimilarity(&t.Sentences[neighborID])
			node.Neighbors[neighborID] = similarity
			t.graph.Nodes[neighborID].Neighbors[node.ID] = similarity
		}
	}
}

// The TextRank algorithm.
// Calculate the score for each sentence until converge.
func (t *Text) doRanking() {
	newScores := make(map[int]float64)
	iterCount := 0

	for iterCount < t.maxIterations {
		iterCount++
		isContinue := false

		for i := range t.graph.Nodes {
			node := &t.graph.Nodes[i]

			currentScore := 0.0
			for neighborID, weight := range node.Neighbors {
				if t.graph.Nodes[neighborID].Value == 0 {
					continue
				}

				currentScore += weight / (t.graph.Nodes[neighborID].Value) * t.Sentences[neighborID].Score
			}
			currentScore = currentScore*t.dampingFactor + (1 - t.dampingFactor)

			delta := t.Sentences[node.ID].Score - currentScore

			if delta > t.threshold {
				isContinue = true
			}

			newScores[node.ID] = currentScore
		}

		for id, score := range newScores {
			t.Sentences[id].Score = score
		}

		if !isContinue {
			break
		}

	}
}

type Sentences struct {
	sentences []Sentence
	ind       []int
}

func (s Sentences) Len() int {
	return len(s.sentences)
}

func (s Sentences) Less(i, j int) bool {
	return s.sentences[i].Score < s.sentences[j].Score
}

func (s Sentences) Swap(i, j int) {
	s.sentences[i], s.sentences[j] = s.sentences[j], s.sentences[i]
	s.ind[i], s.ind[j] = s.ind[j], s.ind[i]
}

// Two tasks:
// 	1. Find the top N sentences with the highest score.
//  2. From that N sentences, order them by position in the text.
func (t *Text) Summarise(percentage float64) string {
	numberOfWords := int(float64(t.numberOfWords) * percentage)
	if numberOfWords < 1 {
		numberOfWords = 1
	}

	s := Sentences{
		sentences: t.Sentences,
		ind:       make([]int, len(t.Sentences)),
	}
	for i := range s.ind {
		s.ind[i] = i
	}

	sort.Sort(sort.Reverse(s))

	var topInd []int

	totalWords := 0
	for i := range s.sentences {
		if totalWords < numberOfWords {
			totalWords += s.sentences[i].numberOfWords
			topInd = append(topInd, i)
		} else {
			break
		}
	}

	sort.Sort(sort.IntSlice(topInd))

	summarisedText := ""
	for i, ind := range topInd {
		if i == 0 {
			summarisedText += t.Sentences[ind].Text
		} else {
			summarisedText = summarisedText + " " + t.Sentences[ind].Text
		}
	}

	return summarisedText
}

// As part of the equation calculating the score, we need to find the weight sum of all edge for each node
// Since the weight never changes, we will calculate the total weight once before running the ranking algorithm
// to save time. This function will be called in NewText
func totalNeighborWeight(g graph.Graph, id int) float64 {
	totalWeight := 0.0

	node := &g.Nodes[id]
	for _, weight := range node.Neighbors {
		totalWeight += weight
	}

	return totalWeight
}

// Sentence needs to be normalise to increase accuracy.
// This function will include all the normalisation method such as lower case, lemmatization, etc.
func normaliseSentence(sentence string, lemmaDict map[string]string) string {
	newSentence := ""

	// Remove all punctuation
	for _, c := range sentence {
		_, ok := punctuationMarks[string(c)]
		if !ok {
			newSentence = newSentence + string(c)
		}
	}

	// Lower case
	newSentence = strings.ToLower(newSentence)

	// TODO: Change all contraction ('s, 've, 're, etc.) to its original form

	// Lemmatization
	if lemmaDict != nil {
		word := ""
		newerSentence := ""
		for i, c := range newSentence {
			if (string(c) != " ") && (i != (len(newSentence) - 1)) {
				word = word + string(c)
			} else {
				if i == (len(newSentence) - 1) {
					word = word + string(c)
				}

				value, ok := lemmaDict[word]
				if !ok {
					value = word
				}

				if len(newerSentence) == 0 {
					newerSentence = newerSentence + value
				} else {
					newerSentence = newerSentence + " " + value
				}
				word = ""
			}
		}

		newSentence = newerSentence
	}

	return newSentence
}

/*

	NewText return a Text struct. The inputs for this function are:
		- text: Text extract from the news agency.
		- lemmaDict: Lemmatization list, if nil will try to read from pkg/textrank/lemmatization_list

*/

func NewText(text string, lemmaDict map[string]string) (*Text, error) {
	// set some basic configuration
	newText := Text{
		Text:          text,
		windowSize:    2,
		dampingFactor: 0.85,
		maxIterations: 30,
		threshold:     0.0001,
	}
	if lemmaDict != nil {
		newText.lemmaDict = lemmaDict
	} else {
		lemmaDict, err := ParseLemmatization()
		if err == nil {
			newText.lemmaDict = lemmaDict
		}
	}

	// Create prose document to tokenize the sentences
	var err error
	newText.doc, err = prose.NewDocument(newText.Text)
	if err != nil {
		return nil, err
	}

	nSentences := len(newText.doc.Sentences())

	// For summarisation, any sentences can be linked together based on its similarity.
	// To simplify this, we can consider one node is connected to all other node.
	// Any neighbor has no similarity to the current node will have a weight = 0.
	var neighbors []int
	for i := 0; i < nSentences; i++ {
		neighbors = append(neighbors, i)
	}

	for i, s := range newText.doc.Sentences() {
		// Cleaning up the sentences. This is important because it is required for display later on.
		// and therefore cannot be put into normaliseSentence sentence.
		// Note: s.Text is used for display, instead of s.NormalisedText
		// Remove tabs and end of lines
		cleanText := ""
		for _, c := range s.Text {
			if (c != '\t') && (c != '\n') {
				cleanText = cleanText + string(c)
			}
		}

		// Remove double space between words
		space := regexp.MustCompile(`\s+`)
		cleanText = space.ReplaceAllString(cleanText, " ")

		newSentence := Sentence{
			Text:           cleanText,
			NormalisedText: normaliseSentence(s.Text, newText.lemmaDict),
			Score:          1, // A default value, does not really matter since it will be calculated
			// again in findSimilarity
		}
		// Generate unique set of words that represent NormalisedText
		newSentence.Words, newSentence.numberOfWords = tokenizeSentenceToWords(newSentence.NormalisedText, "sort")

		newText.Sentences = append(newText.Sentences, newSentence)
		newText.graph.AddNode(i, -1, neighbors...) // in AddNode, neighbor that has the same
		// ID with current node ID will be ignored. ID = -1 will automatically assign ID for the
		// current node.
	}

	newText.findSimilarities()

	newText.numberOfWords = 0
	for i := range newText.Sentences {
		newText.numberOfWords += newText.Sentences[i].numberOfWords
	}

	// Find total weight for neighbors to optimise computation time
	for i := range newText.graph.Nodes {
		newText.graph.Nodes[i].Value = totalNeighborWeight(newText.graph, i)
	}

	newText.doRanking()

	return &newText, nil
}

// Simple tokenize words for sentence algorithm, based in character space ' '
func tokenizeSentenceToWords(sentence string, opts ...string) ([]string, int) {
	words := make(map[string]struct{})
	var word string
	numberOfWords := 0

	for i, c := range sentence {
		if c == ' ' {
			words[word] = struct{}{}
			word = ""
			numberOfWords++
		} else {
			word = word + string(c)
		}

		if i == (len(sentence) - 1) {
			words[word] = struct{}{}
		}
	}

	var uniqueWords []string
	for key := range words {
		uniqueWords = append(uniqueWords, key)
	}

	for _, o := range opts {
		if o == "sort" {
			// Sort the sentences in lexicographical order
			sort.Strings(uniqueWords)
		}
	}

	return uniqueWords, numberOfWords
}

func (t *Text) PrintGraph() {
	for i, node := range t.graph.Nodes {
		fmt.Println("Node number", i, "with score:", t.Sentences[node.ID].Score)
		//fmt.Println("Node number", i, "with score:", t.Sentences[node.ID].Score, "has the following words:", t.Sentences[node.ID].Words)
		//for j, similarity := range node.Neighbors {
		//	fmt.Println("\tneighbor", j, "has similarity:", similarity, "with words:", t.Sentences[j].Words)
		//}

	}
}
