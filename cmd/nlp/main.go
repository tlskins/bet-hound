package main

import (
	language "cloud.google.com/go/language/apiv1"
	"context"
	"fmt"
	langpb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"log"
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
		fmt.Println(noun)
	}
	fmt.Println("Verbs:")
	for _, verb := range verbs {
		fmt.Println(verb)
	}
	fmt.Println("Adjs:")
	for _, adj := range adjs {
		fmt.Println(adj)
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
					Proper: t.PartOfSpeech.Proper.String(),
					Case:   t.PartOfSpeech.Case.String(),
					Person: t.PartOfSpeech.Person.String(),
					Mood:   t.PartOfSpeech.Mood.String(),
					Tense:  t.PartOfSpeech.Tense.String(),
				},
				DependencyEdge: &DependencyEdge{
					Label:          t.DependencyEdge.Label.String(),
					HeadTokenIndex: int(t.DependencyEdge.HeadTokenIndex),
				},
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
