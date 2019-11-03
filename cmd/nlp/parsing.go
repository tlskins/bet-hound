package nlp

import (
	t "bet-hound/cmd/nlp/types"
	language "cloud.google.com/go/language/apiv1"
	"context"
	"fmt"
	langpb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"log"
	"strings"
)

func ParseText(text string) (groupedNouns []*t.Word, groupedVerbs []*t.Word) {
	ctx := context.Background()
	lc, err := language.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to create language client: %s", err)
	}
	resp, err := lc.AnalyzeSyntax(ctx, buildSyntaxRequest(text))
	if err != nil {
		log.Fatalf("failed to analyze syntax: %s", err)
	}

	nouns, verbs, adjs, dets := buildWords(resp)
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
	fmt.Println("Dets:")
	for _, det := range dets {
		fmt.Println(det, det.DependencyEdge, det.PartOfSpeech)
	}

	fmt.Println("groupedNouns:")
	groupedNouns = groupWords(nouns, nouns, dets)
	for _, noun := range groupedNouns {
		fmt.Println(noun, noun.AllLemmas())
	}
	fmt.Println("groupedVerbs:")
	groupedVerbs = groupWords(verbs, nouns, adjs)
	for _, verb := range groupedVerbs {
		fmt.Println(verb, verb.AllLemmas())
	}

	return groupedNouns, groupedVerbs
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

func findWord(words []*t.Word, index int) *t.Word {
	for _, w := range words {
		if w.Index == index {
			return w
		}
	}
	return nil
}

func groupWords(parents []*t.Word, children ...[]*t.Word) (grouped []*t.Word) {
	for _, parent := range parents {
		for _, child := range children {
			for _, word := range child {
				if word.DependencyEdge.HeadTokenIndex == parent.Index {
					*parent.Dependents = append(*parent.Dependents, word)
					if findWord(grouped, parent.Index) == nil {
						grouped = append(grouped, parent)
					}
				}
			}
		}
	}

	return grouped
}

func buildWords(resp *langpb.AnalyzeSyntaxResponse) (nouns []*t.Word, verbs []*t.Word, adjs []*t.Word, dets []*t.Word) {
	for i, token := range resp.Tokens {
		fmt.Println(token)
		pos := token.PartOfSpeech.Tag.String()
		if pos == "VERB" || pos == "NOUN" || pos == "ADJ" || pos == "DET" {
			word := t.Word{
				Text:  token.Text.Content,
				Lemma: token.Lemma,
				Index: i,
				PartOfSpeech: &t.PartOfSpeech{
					Tag:    pos,
					Proper: strings.TrimSpace(token.PartOfSpeech.Proper.String()),
					Case:   token.PartOfSpeech.Case.String(),
					Person: token.PartOfSpeech.Person.String(),
					Mood:   token.PartOfSpeech.Mood.String(),
					Tense:  token.PartOfSpeech.Tense.String(),
				},
				DependencyEdge: &t.DependencyEdge{
					Label:          token.DependencyEdge.Label.String(),
					HeadTokenIndex: int(token.DependencyEdge.HeadTokenIndex),
				},
				Dependents: &[]*t.Word{},
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
			if pos == "DET" {
				dets = append(dets, &word)
			}
		}
	}

	return nouns, verbs, adjs, dets
}
