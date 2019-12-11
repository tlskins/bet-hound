package types

import (
// "fmt"
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
	idxMatch := hdIdx == -1 || w.DependencyEdge.HeadTokenIndex == hdIdx
	if !idxMatch {
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
			return false
		}
	}

	exclTxtMatch := len(exclTxt) == 0
	if !exclTxtMatch {
		for _, x := range exclTxt {
			if x == w.Text {
				return false
			}
		}
	}

	hdIdxMatch := hdIdx == -1
	if !hdIdxMatch {
		wHdIdx := w.DependencyEdge.HeadTokenIndex
		// Words can be their own children
		if !((wHdIdx == hdIdx) && (w.Index != wHdIdx)) {
			return false
		}
	}

	return true
}

func matchSearchWord(w *Word, hdIdx, stIdx, endIdx int, tags, btCmps []string) bool {
	// head index match
	idxMatch := hdIdx == -1 || w.DependencyEdge.HeadTokenIndex == hdIdx
	if !idxMatch {
		// fmt.Printf("fail idx match %s\n", w.Text)
		return false
	}

	// match start - end idx
	idx := w.Index
	if (stIdx != -1 && stIdx >= idx) || (endIdx != -1 && endIdx <= idx) {
		// fmt.Printf("fail st - end match %s\n", w.Text)
		return false
	}

	// match part of speech
	tagMatch := len(tags) == 0
	if !tagMatch {
		for _, t := range tags {
			if t == w.PartOfSpeech.Tag {
				tagMatch = true
				break
			}
		}
		if !tagMatch {
			// fmt.Printf("fail st - end match %s\n", w.Text)
			return false
		}
	}

	// match bet component
	cmpMatch := len(btCmps) == 0
	if !cmpMatch {
		for _, t := range btCmps {
			if t == w.BetComponent {
				cmpMatch = true
				break
			} else if t == "NONE" && w.BetComponent == "" {
				cmpMatch = true
				break
			}
		}
		if !cmpMatch {
			// fmt.Printf("fail cmp match %s\n", w.Text)
			return false
		}
	}

	return true
}

func SearchWords(words *[]*Word, hdIdx, stIdx, endIdx int, tags, btCmps []string) (results []*Word) {
	for _, w := range *words {
		if matchSearchWord(w, hdIdx, stIdx, endIdx, tags, btCmps) {
			results = append(results, w)
		}
	}

	// Search down hiearchy recursively only if given a head token index
	if hdIdx != -1 {
		recurseResults := []*Word{}
		for _, w := range results {
			children := SearchWords(words, w.Index, stIdx, endIdx, tags, btCmps)
			recurseResults = append(recurseResults, children...)
		}
		results = append(results, recurseResults...)
	}
	return results
}

func SearchGroupedWords(words *[]*Word, hdIdx, stIdx, endIdx int) (results [][]*Word) {
	word := (*words)[hdIdx]
	if len(word.DependencyEdge.ChildTokenIndices) == 0 || word.DependencyEdge.HeadTokenIndex == word.Index {
		return append(results, []*Word{word})
	}

	for _, cIdx := range word.DependencyEdge.ChildTokenIndices {
		if (stIdx == -1 || stIdx < hdIdx) && (endIdx == -1 || endIdx > hdIdx) {
			child := (*words)[cIdx]
			recurse := SearchGroupedWords(words, child.Index, stIdx, endIdx)
			for _, r := range recurse {
				r = append(r, word)
				results = append(results, r)
			}
		}
	}
	return results
}

// func SearchGroupedWords(words *[]*Word, hdIdx, stIdx, endIdx int, tags, btCmps []string) (results [][]*Word) {
// 	children := []*Word{}
// 	for _, w := range *words {
// 		if matchSearchWord(w, hdIdx, stIdx, endIdx, tags, btCmps) {
// 			results = append(results, []*Word{w})
// 		}
// 		if hdIdx != -1 && w.DependencyEdge.HeadTokenIndex == hdIdx {
// 			children = append(children, w)
// 		}
// 	}

// 	// Search down hiearchy recursively only if given a head token index
// 	if hdIdx != -1 {
// 		for _, idx := range children {
// 			child := (*words)[idx]
// 			cResults := SearchWords(words, child.Index, stIdx, endIdx, tags, btCmps)
// 			results[i] = append(w, cResults...)
// 		}
// 	}
// 	return results
// }

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
	BetComponent   string         `bson:"b_comp" json:"bet_component"`
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
	Label             string `bson:"label,omitempty" json:"label"`
	HeadTokenIndex    int    `bson:"hd_tkn_idx,omitempty" json:"head_token_index"`
	ChildTokenIndices []int  `bson:"ch_tkn_idxs,omitempty" json:"child_token_indices"`
}
