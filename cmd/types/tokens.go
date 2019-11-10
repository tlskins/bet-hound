package types

import (
	"fmt"
	// "strings"
)

// type Phrase struct {
// 	Word     *Word   `bson:"word,omitempty" json:"word"`
// 	Source   *Source `bson:"src,omitempty" json:"source"`
// 	HomeGame *Game   `bson:"h_gm,omitempty" json:"home_game"`
// 	AwayGame *Game   `bson:"a_gm,omitempty" json:"away_game"`
// }

// type MetricPhrase struct {
// 	Word          *Word   `bson:"word,omitempty" json:"word"`
// 	OperatorWord  *Word   `bson:"op_word,omitempty" json:"operator_word"`
// 	ModifierWords []*Word `bson:"mod_words,omitempty" json:"modifier_words"`
// }

type OperatorPhrase struct {
	MetricWord      Word   `bson:"m_word" json:"metric_word"`
	MetricModifiers []Word `bson:"m_mods" json:"metric_modifiers"`
	OperatorWord    Word   `bson:"op_word" json:"operator_word"`
	ActionWord      Word   `bson:"a_word" json:"action_word"`
}

func (p OperatorPhrase) Text() (desc string) {
	metric := p.MetricWord.Text
	for _, m := range p.MetricModifiers {
		metric = m.Text + " " + metric
	}
	return fmt.Sprintf("%s %s %s", p.ActionWord.Text, p.OperatorWord.Text, metric)
}

// func (p Phrase) Game() *Game {
// 	if p.HomeGame != nil {
// 		return p.HomeGame
// 	} else if p.AwayGame != nil {
// 		return p.AwayGame
// 	}
// 	return nil
// }

// func (p Phrase) SourceDesc() (desc *string) {
// 	if p.Source == nil || p.Game() == nil {
// 		return nil
// 	}
// 	fName := (*p.Source.FirstName)[:1]
// 	lName := *p.Source.LastName
// 	pos := *p.Source.Position
// 	tm := *p.Source.TeamShort
// 	srcTeamFk := p.Source.TeamFk
// 	gm := p.Game()
// 	var vsTeam string
// 	if gm.HomeTeamFk == srcTeamFk {
// 		vsTeam = *gm.AwayTeamName
// 	} else {
// 		vsTeam = *gm.HomeTeamName
// 	}
// 	result := fName + "." + lName + " (" + tm + "-" + pos + ")" + " vs " + vsTeam
// 	return &result
// }

// func descendentLemmas(word *Word) (lemmas []string) {
// 	lemmas = append(lemmas, word.Lemma)
// 	if word.Children != nil {
// 		for _, child := range *word.Children {
// 			lemmas = append(lemmas, child.Lemma)
// 		}
// 	}
// 	return lemmas
// }

// func (p *Phrase) AllLemmas() []string {
// 	return descendentLemmas(p.Word)
// }

// func (m *MetricPhrase) AllLemmas() []string {
// 	return descendentLemmas(m.Word)
// }

// func descendentText(word *Word) (text []string) {
// 	text = append(text, word.Text)
// 	if word.Children != nil {
// 		for _, child := range *word.Children {
// 			text = append(text, child.Lemma)
// 		}
// 	}
// 	return text
// }

// func (p *Phrase) AllText() []string {
// 	return descendentText(p.Word)
// }

// func (m *MetricPhrase) AllText() []string {
// 	return descendentText(m.Word)
// }

func FindWords(words *[]*Word, hdIdx *int, tags *[]string, labels *[]string) *[]*Word {
	results := &[]*Word{}
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
			*results = append(*results, w)
		}
	}
	// Search down hiearchy recursively only if given a head token index
	if hdIdx != nil {
		recurseResults := []*Word{}
		for _, w := range *results {
			children := FindWords(words, &w.Index, tags, labels)
			recurseResults = append(recurseResults, *children...)
		}
		*results = append(*results, recurseResults...)
	}
	// if hdIdx != nil {
	// 	recurseResults := []*Word{}
	// 	for _, w := range *result {
	// 		children := FindWords(words, &w.Index, tags, labels)
	// 		recurseResults = append(recurseResults, *children...)
	// 	}
	// 	*result = append(*result, recurseResults...)
	// }
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

func FindWordByTxt(words []*Word, txt string) *Word {
	for _, w := range words {
		if w.Text == txt {
			return w
		}
	}
	return nil
}

func FindWordByIdx(words []*Word, idx int) *Word {
	for _, w := range words {
		if w.Index == idx {
			return w
		}
	}
	return nil
}

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
