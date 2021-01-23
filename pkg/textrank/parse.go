package textrank

import (
	"bufio"
	"os"
)

func ParseLemmatization() (map[string]string, error) {
	file, err := os.Open("./pkg/textrank/lemmatization_list")
	if err != nil {
		return nil, err
	}

	lemDict := make(map[string]string)
	var newLine string
	var lemma string
	var token string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		newLine = scanner.Text()
		if len(newLine) == 0 {
			continue
		}
		i := 0
		for newLine[i] != '\t' {
			lemma = lemma + string(newLine[i])
			i++
		}
		i++
		for i < len(newLine) {
			token = token + string(newLine[i])
			i++
		}

		lemDict[token] = lemma
		token = ""
		lemma = ""
	}
	if scanner.Err() != nil {
		return nil, err
	}

	return lemDict, nil
}
