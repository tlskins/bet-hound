package nlp

import (
	// "bet-hound/cmd/db"
	t "bet-hound/cmd/types"
	language "cloud.google.com/go/language/apiv1"
	"context"
	"fmt"
	langpb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"log"
	"regexp"
	"strconv"
	"strings"
)

func ParseText(text string) (allWords []*t.Word) {
	ctx := context.Background()
	lc, err := language.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to create language client: %s", err)
	}
	resp, err := lc.AnalyzeSyntax(ctx, buildSyntaxRequest(text))
	if err != nil {
		log.Fatalf("failed to analyze syntax: %s", err)
	}

	return buildWords(resp)
}

func WordsLemmas(words *[]*t.Word) (results []string) {
	for _, w := range *words {
		results = append(results, w.Lemma)
	}
	return results
}

func matchSearchWord(w *t.Word, hdIdx, stIdx, endIdx int, tags, btCmps []string) bool {
	// head index match
	idxMatch := hdIdx == -1 || w.DependencyEdge.HeadTokenIndex == hdIdx
	if !idxMatch {
		return false
	}

	// match start - end idx
	idx := w.Index
	if (stIdx != -1 && stIdx >= idx) || (endIdx != -1 && endIdx <= idx) {
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
			return false
		}
	}

	return true
}

func SearchLastParent(words *[]*t.Word, idx, stIdx, endIdx int, tags, btCmps []string) (word *t.Word) {
	tgt := (*words)[idx]
	selfMatch := matchSearchWord(tgt, -1, stIdx, endIdx, tags, btCmps)
	var parentMatch *t.Word
	if tgt.DependencyEdge.HeadTokenIndex != tgt.Index {
		parentMatch = SearchLastParent(words, tgt.DependencyEdge.HeadTokenIndex, stIdx, endIdx, tags, btCmps)
	}

	if parentMatch != nil {
		return parentMatch
	} else if selfMatch {
		return tgt
	} else {
		return nil
	}
}

func SearchChildren(words *[]*t.Word, idx, stIdx, endIdx int, tags, btCmps []string) (results []*t.Word) {
	word := (*words)[idx]
	if matchSearchWord(word, -1, stIdx, endIdx, tags, btCmps) {
		results = append(results, word)
	}

	for _, wc := range word.DependencyEdge.ChildTokenIndices {
		recurse := SearchChildren(words, wc, stIdx, endIdx, tags, btCmps)
		results = append(results, recurse...)
	}
	return results
}

func SearchFirstChild(words *[]*t.Word, hdIdx, stIdx, endIdx int, tags, btCmps []string) (word *t.Word) {
	head := (*words)[hdIdx]
	if len(head.DependencyEdge.ChildTokenIndices) == 0 || head.DependencyEdge.HeadTokenIndex == head.Index {
		return nil
	}

	for _, cIdx := range head.DependencyEdge.ChildTokenIndices {
		child := (*words)[cIdx]
		if matchSearchWord(child, -1, stIdx, endIdx, tags, btCmps) {
			return child
		} else {
			recurse := SearchFirstChild(words, cIdx, stIdx, endIdx, tags, btCmps)
			if recurse != nil {
				return recurse
			}
		}
	}
	return nil
}

func FindPlayerWords(words *[]*t.Word) (playerWords [][]*t.Word) {
	var temp *[]*t.Word
	for i, word := range *words {
		isPlayerWord := word.PartOfSpeech.Tag == "NOUN" && word.BetComponent == ""
		isLastWord := i == len(*words)-1
		if (!isPlayerWord || isLastWord) && temp != nil && len(*temp) > 0 {
			playerWords = append(playerWords, *temp)
			temp = nil
		} else if isPlayerWord {
			if temp == nil {
				temp = &[]*t.Word{word}
			} else {
				*temp = append(*temp, word)
			}
		}
	}

	return playerWords
}

func CalcBetComponent(lemma string) string {
	_, floatErr := strconv.ParseFloat(lemma, 64) // no err means it is a float
	if lemma == "score" || lemma == "have" || lemma == "gain" {
		return "ACTION"
	} else if lemma == "more" || lemma == "few" || lemma == "less" {
		return "OPERATOR"
	} else if lemma == "and" || lemma == "," {
		return "SUB_OPERATOR"
	} else if lemma == "than" {
		return "DELIMITER"
	} else if lemma == "point" || lemma == "pt" || lemma == "yard" || lemma == "yd" || lemma == "touchdown" || lemma == "td" {
		return "METRIC"
	} else if lemma == "ppr" || lemma == "standard" || lemma == "std" || lemma == "0.5ppr" || lemma == ".5ppr" || floatErr == nil {
		return "METRIC_MOD"
	} else if lemma == "week" {
		return "EVENT_TIME"
	} else if lemma == "this" {
		return "EVENT_TIME_MOD"
	} else {
		return ""
	}
}

func RemoveReservedTwitterWords(text string) (result string) {
	var handleRgx = regexp.MustCompile(`\@[^\s]*`)
	var hashRgx = regexp.MustCompile(`\#[^\s]*`)
	result = handleRgx.ReplaceAllString(text, " ")
	result = hashRgx.ReplaceAllString(result, " ")
	return result
}

// helpers

func buildSyntaxRequest(text string) *langpb.AnalyzeSyntaxRequest {
	return &langpb.AnalyzeSyntaxRequest{
		Document: &langpb.Document{
			Type: langpb.Document_PLAIN_TEXT,
			Source: &langpb.Document_Content{
				Content: text,
			},
		},
		EncodingType: langpb.EncodingType_UTF8,
	}
}

func buildWords(resp *langpb.AnalyzeSyntaxResponse) (allWords []*t.Word) {
	for i, token := range resp.Tokens {
		pos := token.PartOfSpeech.Tag.String()
		word := t.Word{
			Text:  token.Text.Content,
			Lemma: token.Lemma,
			Index: i,
			PartOfSpeech: t.PartOfSpeech{
				Tag:    pos,
				Proper: strings.TrimSpace(token.PartOfSpeech.Proper.String()),
				Case:   token.PartOfSpeech.Case.String(),
				Person: token.PartOfSpeech.Person.String(),
				Mood:   token.PartOfSpeech.Mood.String(),
				Tense:  token.PartOfSpeech.Tense.String(),
			},
			DependencyEdge: t.DependencyEdge{
				Label:          token.DependencyEdge.Label.String(),
				HeadTokenIndex: int(token.DependencyEdge.HeadTokenIndex),
			},
			BetComponent: CalcBetComponent(token.Lemma),
		}
		allWords = append(allWords, &word)
	}

	// Add children idx
	for _, word := range allWords {
		tgtIdx := word.DependencyEdge.HeadTokenIndex
		if tgtIdx < len(allWords) && tgtIdx != word.Index {
			tgt := allWords[tgtIdx]
			children := tgt.DependencyEdge.ChildTokenIndices
			tgt.DependencyEdge.ChildTokenIndices = append(children, word.Index)
		}
	}

	for _, word := range allWords {
		fmt.Println("Built word ", word.Text, word.Index, word.DependencyEdge.HeadTokenIndex, word.BetComponent, word.DependencyEdge.ChildTokenIndices, word)
	}

	return allWords
}
