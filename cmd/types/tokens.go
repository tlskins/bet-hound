package types

import (
// "strings"
)

func ReverseStrings(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

func WordsText(words *[]*Word) (results []string) {
	for _, w := range *words {
		results = append(results, w.Text)
	}
	return results
}

func WordsLemmas(words *[]*Word) (results []string) {
	for _, w := range *words {
		results = append(results, w.Lemma)
	}
	return results
}

func FilterWordsByTag(words *[]*Word, tag string) (results []*Word) {
	for _, w := range *words {
		if w.PartOfSpeech.Tag == tag {
			results = append(results, w)
		}
	}
	return results
}

func matchWord(w *Word, hdIdx int, tags []string, exclTxt []string) bool {
	// fmt.Println("matching ", w.Text, w.DependencyEdge.HeadTokenIndex, hdIdx, w.PartOfSpeech.Tag, tags)
	idxMatch := hdIdx == -1 || w.DependencyEdge.HeadTokenIndex == hdIdx
	if !idxMatch {
		// fmt.Println("not id match")
		return false
	}

	tagMatch := len(tags) == 0
	if !tagMatch {
		for _, t := range tags {
			if t == w.PartOfSpeech.Tag {
				tagMatch = true
				break
			}
		}
		if !tagMatch {
			// fmt.Println("not tag match")
			return false
		}
	}

	exclTxtMatch := len(exclTxt) == 0
	if !exclTxtMatch {
		for _, x := range exclTxt {
			if x == w.Text {
				// fmt.Println("not exclTxtMatch match")
				return false
			}
		}
	}

	hdIdxMatch := hdIdx == -1
	if !hdIdxMatch {
		wHdIdx := w.DependencyEdge.HeadTokenIndex
		// Words can be their own children
		if !((wHdIdx == hdIdx) && (w.Index != wHdIdx)) {
			// fmt.Println("not child of match")
			return false
		}
	}

	return true
}

func FindWords(words *[]*Word, hdIdx int, tags []string, exclTxt []string) []*Word {
	results := []*Word{}
	for _, w := range *words {
		// fmt.Println("considering word", w.Text, w.Index)
		if matchWord(w, hdIdx, tags, exclTxt) {
			results = append(results, w)
		}
	}
	// Search down hiearchy recursively only if given a head token index
	if hdIdx != -1 {
		recurseResults := []*Word{}
		for _, w := range results {
			children := FindWords(words, w.Index, tags, exclTxt)
			recurseResults = append(recurseResults, children...)
		}
		results = append(results, recurseResults...)
	}
	return results
}

func FindGroupedWords(words *[]*Word, hdIdx int, tags []string, exclTxt []string) [][]*Word {
	results := [][]*Word{}
	for _, w := range *words {
		if matchWord(w, hdIdx, tags, exclTxt) {
			results = append(results, []*Word{w})
		}
	}
	// Search down hiearchy recursively only if given a head token index
	if hdIdx != -1 {
		for i, w := range results {
			children := FindWords(words, w[0].Index, tags, exclTxt)
			results[i] = append(w, children...)
		}
	}
	return results
}

func JoinedWordGroup(grouped []*Word, reverse bool) (result string) {
	for _, g := range grouped {
		if reverse {
			result = g.Text + " " + result
		} else {
			result = result + " " + g.Text
		}
	}
	return result
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
