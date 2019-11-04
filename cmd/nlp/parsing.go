package nlp

import (
	t "bet-hound/cmd/types"
	language "cloud.google.com/go/language/apiv1"
	"context"
	"fmt"
	langpb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"log"
	"strings"
)

func ParseText(text string) (nounPhrases []*t.Phrase, verbPhrases []*t.Phrase, allWords []*t.Word) {
	ctx := context.Background()
	lc, err := language.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to create language client: %s", err)
	}
	resp, err := lc.AnalyzeSyntax(ctx, buildSyntaxRequest(text))
	if err != nil {
		log.Fatalf("failed to analyze syntax: %s", err)
	}

	nouns, verbs, adjs, dets, allWords := buildWords(resp)
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

	fmt.Println("nounPhrases:")
	nounPhrases = groupPhrases(nouns, nouns, dets)
	for _, noun := range nounPhrases {
		fmt.Println(noun, noun.AllLemmas())
	}
	fmt.Println("verbPhrases:")
	verbPhrases = groupPhrases(verbs, nouns, adjs)
	for _, verb := range verbPhrases {
		fmt.Println(verb, verb.AllLemmas())
	}

	return nounPhrases, verbPhrases, allWords
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

func groupPhrases(parents []*t.Word, children ...[]*t.Word) (phrases []*t.Phrase) {
	for _, parent := range parents {
		for _, child := range children {
			for _, word := range child {
				if word.DependencyEdge.HeadTokenIndex == parent.Index {
					phrase := t.FindPhraseByIdx(phrases, parent.Index)
					if phrase == nil {
						phrases = append(phrases, &t.Phrase{Word: parent})
					}
				}
			}
		}
	}

	return phrases
}

func buildWords(resp *langpb.AnalyzeSyntaxResponse) (nouns, verbs, adjs, dets, allWords []*t.Word) {
	for i, token := range resp.Tokens {
		fmt.Println(token)
		// tokens = append(tokens, token)
		pos := token.PartOfSpeech.Tag.String()
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
		}

		allWords = append(allWords, &word)
		if pos == "VERB" || pos == "NOUN" || pos == "ADJ" || pos == "DET" {
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

	for _, word := range allWords {
		fmt.Println(word.Index, word.DependencyEdge.HeadTokenIndex)
	}

	// build word hiearchy
	for _, word := range allWords {
		w := t.FindWordByIdx(allWords, word.DependencyEdge.HeadTokenIndex)
		if w == nil {
			panic("Bad hierarchy!")
		}
		word.Parent = w
		if w.Children == nil {
			w.Children = &[]*t.Word{}
		}
		*w.Children = append(*w.Children, word)
	}

	return nouns, verbs, adjs, dets, allWords
}
