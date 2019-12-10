package nlp

import (
	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
	language "cloud.google.com/go/language/apiv1"
	"context"
	"fmt"
	// "fmt"
	langpb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"log"
	"regexp"
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

func FindOperatorPhrase(words *[]*t.Word) (opPhrase *t.OperatorPhrase, leftMetric *t.Metric) {
	for _, action := range findActions(words) {
		opPhrase, leftMetric = buildOperatorPhrase(words, action)
		if opPhrase != nil {
			break
		}
	}
	return opPhrase, leftMetric
}

func FindLeftPlayerExpr(words *[]*t.Word, opPhrase *t.OperatorPhrase, leftMetric *t.Metric) (leftPlayerExpr *t.PlayerExpression) {
	// Find Player
	groupedNouns := t.FindGroupedWords(words, opPhrase.ActionWord.Index, []string{"NOUN"}, []string{leftMetric.Word.Text})
	for _, groupedNoun := range groupedNouns {
		player := db.SearchPlayerByName(t.JoinedWordGroup(groupedNoun, true))
		if player != nil {
			leftPlayerExpr = &t.PlayerExpression{
				Player: *player,
				Metric: leftMetric,
			}
			break
		}
	}
	// Find Event Time
	actionChildren := t.FindGroupedWords(words, opPhrase.ActionWord.Index, []string{}, []string{leftMetric.Word.Text})
	for _, a := range actionChildren {
		if isEventTimeLemma(a[0].Lemma) {
			remaining := a[1:len(a)]
			leftPlayerExpr.EventTime = &t.EventTime{
				Word:      *a[0],
				Modifiers: t.WordsLemmas(&remaining),
			}
		}
	}

	return leftPlayerExpr
}

func FindRightPlayerExpr(words *[]*t.Word, opPhrase *t.OperatorPhrase, leftMetric *t.Metric) *t.PlayerExpression {
	exclTxt := append(leftMetric.Modifiers, opPhrase.OperatorWord.Text)
	children := t.FindGroupedWords(words, leftMetric.Word.Index, []string{}, exclTxt)
	for _, c := range children {
		nouns := t.FilterWordsByTag(&c, "NOUN")
		player := db.SearchPlayerByName(t.JoinedWordGroup(nouns, true))
		if player != nil {
			return &t.PlayerExpression{Player: *player}
		}
	}
	return nil
}

// nlp helpers

func findActions(words *[]*t.Word) (actionWords []*t.Word) {
	verbs := t.FindWords(words, -1, []string{"VERB"}, []string{})
	for _, v := range verbs {
		if isActionLemma(v.Lemma) {
			vChildren := t.FindWords(words, v.Index, []string{"NOUN"}, []string{})
			if len(vChildren) > 0 {
				actionWords = append(actionWords, v)
			}
		}
	}
	return actionWords
}

func buildOperatorPhrase(words *[]*t.Word, action *t.Word) (opPhrase *t.OperatorPhrase, metric *t.Metric) {
	nouns := t.FindWords(words, action.Index, []string{"NOUN", "VERB"}, []string{})
	for _, noun := range nouns {
		if isMetricLemma(noun.Lemma) {
			adjs := t.FindWords(words, noun.Index, []string{"ADJ", "NOUN"}, []string{})
			modWords := t.FindWords(words, noun.Index, []string{}, []string{})
			metricMods := []string{}
			for _, m := range modWords {
				if isMetricModText(m.Text) {
					metricMods = append(metricMods, m.Text)
				}
			}
			metric = &t.Metric{
				Word:      *noun,
				Modifiers: metricMods,
			}
			for _, adj := range adjs {
				fmt.Println("adj", *adj)
				opPhrase = &t.OperatorPhrase{
					OperatorWord: *adj,
					ActionWord:   *action,
				}
				return opPhrase, metric
			}
		}
	}
	return nil, nil
}

// helpers

func RemoveReservedTwitterWords(text string) (result string) {
	var handleRgx = regexp.MustCompile(`\@[^\s]*`)
	var hashRgx = regexp.MustCompile(`\#[^\s]*`)
	result = handleRgx.ReplaceAllString(text, " ")
	result = hashRgx.ReplaceAllString(result, " ")
	return result
}

func isActionLemma(str string) bool {
	if str == "score" || str == "have" || str == "gain" {
		return true
	} else {
		return false
	}
}

func isMetricLemma(str string) bool {
	if str == "point" || str == "pt" || str == "yard" || str == "yd" || str == "touchdown" || str == "td" {
		return true
	} else {
		return false
	}
}

func isMetricModText(str string) bool {
	if str == "ppr" || str == "0.5ppr" || str == ".5ppr" {
		return true
	} else {
		return false
	}
}

func isEventTimeLemma(str string) bool {
	if str == "week" {
		return true
	} else {
		return false
	}
}

func isEventTimeModText(str string) bool {
	if str == "this" {
		return true
	} else {
		return false
	}
}

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
		}
		fmt.Println("Built word ", word.Text, word.Index, word.DependencyEdge.HeadTokenIndex, word)
		allWords = append(allWords, &word)
	}
	return allWords
}
