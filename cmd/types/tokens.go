package types

import (
// "fmt"
// "strings"
)

func ReverseStrings(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

func WordsText(words *[]Word) (results []string) {
	for _, w := range *words {
		results = append(results, w.Text)
	}
	return results
}

func WordsLemmas(words *[]Word) (results []string) {
	for _, w := range *words {
		results = append(results, w.Lemma)
	}
	return results
}

func FindWords(words *[]*Word, hdIdx *int, tags *[]string, labels *[]string) *[]Word {
	results := &[]Word{}
	for _, w := range *words {
		idxMatch := hdIdx == nil || w.DependencyEdge.HeadTokenIndex == *hdIdx
		tagMatch := tags == nil
		if tags != nil {
			tagMatch = false
			for _, t := range *tags {
				if t == w.PartOfSpeech.Tag {
					tagMatch = true
					break
				}
			}
		}
		lblMatch := labels == nil
		if labels != nil {
			lblMatch = false
			for _, l := range *labels {
				if l == w.DependencyEdge.Label {
					lblMatch = true
					break
				}
			}
		}
		hdIdxMatch := hdIdx == nil
		if hdIdx != nil {
			wHdIdx := w.DependencyEdge.HeadTokenIndex
			// Words can be their own children
			hdIdxMatch = (wHdIdx == *hdIdx) && (w.Index != wHdIdx)
		}
		if idxMatch && tagMatch && lblMatch && hdIdxMatch {
			*results = append(*results, *w)
		}
	}
	// Search down hiearchy recursively only if given a head token index
	if hdIdx != nil {
		recurseResults := []Word{}
		for _, w := range *results {
			children := FindWords(words, &w.Index, tags, labels)
			recurseResults = append(recurseResults, *children...)
		}
		*results = append(*results, recurseResults...)
	}
	return results
}

type Word struct {
	Text           string         `bson:"txt,omitempty" json:"text"`
	Lemma          string         `bson:"lemma,omitempty" json:"lemma"`
	Index          int            `bson:"idx,omitempty" json:"index"`
	PartOfSpeech   PartOfSpeech   `bson:"pos,omitempty" json:"part_of_speech"`
	DependencyEdge DependencyEdge `bson:"dep_edge,omitempty" json:"dependency_edge"`
}

type PartOfSpeech struct {
	Tag    string `bson:"tag,omitempty" json:"tag"`
	Proper string `bson:"proper,omitempty" json:"proper"`
	Case   string `bson:"case,omitempty" json:"case"`
	Person string `bson:"person,omitempty" json:"person"`
	Mood   string `bson:"mood,omitempty" json:"mood"`
	Tense  string `bson:"tense,omitempty" json:"tense"`
}

type DependencyEdge struct {
	Label          string `bson:"label,omitempty" json:"label"`
	HeadTokenIndex int    `bson:"hd_tkn_idx,omitempty" json:"head_token_index"`
}

// func FindWordByTxt(words []Word, txt string) *Word {
// 	for _, w := range words {
// 		if w.Text == txt {
// 			return &w
// 		}
// 	}
// 	return nil
// }

// func FindWordByIdx(words []*Word, idx int) *Word {
// 	for _, w := range words {
// 		if w.Index == idx {
// 			return &w
// 		}
// 	}
// 	return nil
// }

// func FindPhraseByIdx(phrases []*Phrase, index int) *Phrase {
// 	for _, p := range phrases {
// 		if p.Word.Index == index {
// 			return p
// 		}
// 	}
// 	return nil
// }

// func findPhraseByWordTxt(phrases []*Phrase, txt string) *Phrase {
// 	for _, p := range phrases {
// 		if p.Word.Text == txt {
// 			return p
// 		}
// 	}
// 	return nil
// }
