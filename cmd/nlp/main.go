package main

import (
	language "cloud.google.com/go/language/apiv1"
	"context"
	"fmt"
	langpb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"log"
	"strings"
)

const text = "bet you that tevin coleman scores more ppr points than matt breida this week"

type PartOfSpeech struct {
	Tag    string
	Proper string
	Case   string
	Person string
	Mood   string
	Tense  string
}

type DependencyEdge struct {
	Label          string
	HeadTokenIndex int
}

type Word struct {
	Text           string
	Lemma          string
	Index          int
	PartOfSpeech   *PartOfSpeech
	DependencyEdge *DependencyEdge
	Dependents     *[]*Word
}

func (w *Word) AllLemmas() (lemmas []string) {
	if w.Dependents == nil {
		return lemmas
	}
	for _, w := range *w.Dependents {
		lemmas = append(lemmas, w.Lemma)
	}
	return lemmas
}

func main() {
	ctx := context.Background()
	lc, err := language.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to create language client: %s", err)
	}
	resp, err := lc.AnalyzeSyntax(ctx, buildSyntaxRequest(text))
	if err != nil {
		log.Fatalf("failed to analyze syntax: %s", err)
	}

	nouns, verbs, adjs := buildWords(resp)
	fmt.Println("Nouns:")
	for _, noun := range nouns {
		fmt.Println(noun, noun.DependencyEdge, noun.PartOfSpeech)
	}
	fmt.Println("Verbs:")
	for _, verb := range verbs {
		fmt.Println(verb, verb.DependencyEdge, verb.PartOfSpeech)
	}
	fmt.Println("Adjs:")
	for _, adj := range adjs {
		fmt.Println(adj, adj.DependencyEdge, adj.PartOfSpeech)
	}

	fmt.Println("groupedNouns:")
	groupedNouns := groupNouns(nouns)

	for _, noun := range groupedNouns {
		fmt.Println(noun, *noun.Dependents, noun.AllLemmas())
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

func findWord(words []*Word, index int) *Word {
	for _, w := range words {
		if w.Index == index {
			return w
		}
	}
	return nil
}

func groupNouns(nouns []*Word) (groupedNouns []*Word) {
	for _, noun := range nouns {
		parent := findWord(nouns, noun.DependencyEdge.HeadTokenIndex)
		if parent != nil {
			*parent.Dependents = append(*parent.Dependents, noun)
			if findWord(groupedNouns, parent.Index) == nil {
				groupedNouns = append(groupedNouns, parent)
			}
		}
	}

	return groupedNouns
}

func buildWords(resp *langpb.AnalyzeSyntaxResponse) (nouns []*Word, verbs []*Word, adjs []*Word) {
	for i, t := range resp.Tokens {
		fmt.Println(t)
		pos := t.PartOfSpeech.Tag.String()
		if pos == "VERB" || pos == "NOUN" || pos == "ADJ" {
			word := Word{
				Text:  t.Text.Content,
				Lemma: t.Lemma,
				Index: i,
				PartOfSpeech: &PartOfSpeech{
					Tag:    pos,
					Proper: strings.TrimSpace(t.PartOfSpeech.Proper.String()),
					Case:   t.PartOfSpeech.Case.String(),
					Person: t.PartOfSpeech.Person.String(),
					Mood:   t.PartOfSpeech.Mood.String(),
					Tense:  t.PartOfSpeech.Tense.String(),
				},
				DependencyEdge: &DependencyEdge{
					Label:          t.DependencyEdge.Label.String(),
					HeadTokenIndex: int(t.DependencyEdge.HeadTokenIndex),
				},
				Dependents: &[]*Word{},
			}

			if pos == "VERB" {
				verbs = append(verbs, &word)
			}
			if pos == "NOUN" {
				nouns = append(nouns, &word)
			}
			if pos == "ADJ" {
				adjs = append(adjs, &word)
			}
		}
	}

	return nouns, verbs, adjs
}
