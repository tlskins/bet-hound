package nlp

import (
	"fmt"
	// "bet-hound/cmd/db"
	// "bet-hound/cmd/scraper"
	t "bet-hound/cmd/types"
	language "cloud.google.com/go/language/apiv1"
	"context"
	// "fmt"
	langpb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"log"
	"regexp"
	"strings"
)

// func ParseTweet(tweet *t.Tweet) (bet *t.Bet, err error) {
// 	tweetIdStr := tweet.IdStr
// 	msg := tweet.GetText()
// 	proposer := tweet.User
// 	recipients := tweet.Recipients()
// 	if len(recipients) == 0 {
// 		return bet, fmt.Errorf("No bet recipient found!")
// 	}
// 	recipient := recipients[0]
// 	bet, err = ParseNewText(msg, tweetIdStr, &proposer, &recipient)
// 	if err != nil {
// 		return bet, err
// 	}
// 	bet, err = db.UpsertBet(bet)
// 	if err != nil {
// 		return bet, err
// 	}
// 	// _, err = db.UpsertTweet(tweet)
// 	return bet, err
// }

// func ParseNewText(text, fk string, proposer, recipient *t.User) (bet *t.Bet, err error) {
// 	fmt.Println("Parsing new text", text)
// 	// Find noun and verb phrases
// 	nounPhrases, verbPhrases, _ := ParsePhrases(text)
// 	if len(nounPhrases) < 2 {
// 		return bet, fmt.Errorf("Not enough noun phrases found!")
// 	}
// 	if len(verbPhrases) < 1 {
// 		return bet, fmt.Errorf("Not enough verb phrases found!")
// 	}

// 	// Find sources for nouns phrases
// 	var sources []*t.Source
// 	var sourcePhrases []*t.Phrase
// 	for _, nounPhrase := range nounPhrases {
// 		// reverse text to get first name -> last name
// 		nounTxt := []string{}
// 		texts := nounPhrase.AllText()
// 		for i := len(texts) - 1; i >= 0; i-- {
// 			nounTxt = append(nounTxt, texts[i])
// 		}

// 		foundSrcs, err := db.SearchSourceByName(strings.Join(nounTxt, " "), 1)
// 		if err != nil {
// 			fmt.Println("search source by name err", err)
// 		}
// 		if len(foundSrcs) > 0 {
// 			nounPhrase.Source = &foundSrcs[0]
// 			sourcePhrases = append(sourcePhrases, nounPhrase)
// 			sources = append(sources, &foundSrcs[0])
// 		}
// 	}
// 	if len(sourcePhrases) < 2 {
// 		return bet, fmt.Errorf("Not enough sources found!")
// 	}

// 	// Find Metric
// 	var metricPhrase *t.MetricPhrase
// 	for _, n := range nounPhrases {
// 		nString := n.Word.Lemma
// 		isMetricStr := nString == "point" || nString == "pt" || nString == "yard" || nString == "yd" || nString == "touchdown" || nString == "td"
// 		if isMetricStr && n.Word.Children != nil && len(*n.Word.Children) > 1 {
// 			newMetricPhrase := t.MetricPhrase{Word: n.Word}
// 			for _, child := range *n.Word.Children {
// 				if child.Lemma == "more" || child.Lemma == "great" || child.Lemma == "less" || child.Lemma == "few" {
// 					newMetricPhrase.OperatorWord = child
// 				}
// 				if child.Text == "ppr" || child.Text == "0.5ppr" || child.Text == ".5ppr" {
// 					if newMetricPhrase.ModifierWords == nil {
// 						newMetricPhrase.ModifierWords = []*t.Word{}
// 					}
// 					newMetricPhrase.ModifierWords = append(newMetricPhrase.ModifierWords, child)
// 				}
// 			}
// 			if newMetricPhrase.OperatorWord != nil {
// 				metricPhrase = &newMetricPhrase
// 				break
// 			}
// 		}
// 	}
// 	if metricPhrase == nil {
// 		return bet, fmt.Errorf("Metric phrase not found!")
// 	}

// 	// Find Action
// 	var actionPhrase *t.Phrase
// 	for _, v := range verbPhrases {
// 		vString := v.Word.Lemma
// 		if vString == "score" || vString == "have" || vString == "gain" {
// 			for _, lemma := range v.AllLemmas() {
// 				if metricPhrase.Word.Lemma == lemma {
// 					actionPhrase = v
// 					break
// 				}
// 			}
// 		}
// 	}
// 	if actionPhrase == nil {
// 		return bet, fmt.Errorf("Action phrase not found!")
// 	}

// 	// Find Proposer Source
// 	var proposerSourcePhrase *t.Phrase
// 	for _, child := range *actionPhrase.Word.Children {
// 		for _, src := range sourcePhrases {
// 			if child.Text == src.Word.Text {
// 				proposerSourcePhrase = src
// 				break
// 			}
// 		}
// 	}
// 	if proposerSourcePhrase == nil {
// 		return bet, fmt.Errorf("Proposer source phrase not found!")
// 	}

// 	// Find Recipient Source
// 	var recipientSourcePhrase *t.Phrase
// 	for _, p := range nounPhrases {
// 		if p.Source != nil && p.Source != proposerSourcePhrase.Source {
// 			recipientSourcePhrase = p
// 			break
// 		}
// 	}
// 	// TODO : Calculate this through "breida" -> "than" -> "points"
// 	if proposerSourcePhrase == nil {
// 		return bet, fmt.Errorf("Recipient source phrase not found!")
// 	}

// 	fmt.Println("action word ", actionPhrase.AllLemmas())
// 	fmt.Println("metric word ", metricPhrase.AllLemmas())
// 	fmt.Println("proposer source ", *proposerSourcePhrase.Source.Name)
// 	fmt.Println("recipient source ", *recipientSourcePhrase.Source.Name)

// 	// Get game data
// 	allGames := scraper.ScrapeThisWeeksGames()
// 	for _, game := range allGames {
// 		if *proposerSourcePhrase.Source.TeamFk == *game.HomeTeamFk {
// 			proposerSourcePhrase.HomeGame = game
// 		} else if *proposerSourcePhrase.Source.TeamFk == *game.AwayTeamFk {
// 			proposerSourcePhrase.AwayGame = game
// 		}

// 		if *recipientSourcePhrase.Source.TeamFk == *game.HomeTeamFk {
// 			recipientSourcePhrase.HomeGame = game
// 		} else if *recipientSourcePhrase.Source.TeamFk == *game.AwayTeamFk {
// 			recipientSourcePhrase.AwayGame = game
// 		}
// 	}
// 	if proposerSourcePhrase.Game() == nil {
// 		return bet, fmt.Errorf("Proposer source game not found!")
// 	}
// 	if recipientSourcePhrase.Game() == nil {
// 		return bet, fmt.Errorf("Recipient source game not found!")
// 	}

// 	fmt.Println("proposer source game", *proposerSourcePhrase.Game().Name)
// 	fmt.Println("recipient source game", *recipientSourcePhrase.Game().Name)

// 	betStatus := t.BetStatusFromString("Pending Proposer")
// 	bet = &t.Bet{
// 		Fk:                    &fk,
// 		ActionPhrase:          actionPhrase,
// 		MetricPhrase:          metricPhrase,
// 		ProposerSourcePhrase:  proposerSourcePhrase,
// 		RecipientSourcePhrase: recipientSourcePhrase,
// 		BetStatus:             betStatus,
// 		Proposer:              proposer,
// 		Recipient:             recipient,
// 	}
// 	fmt.Println("new bet", bet)
// 	return bet, err
// }

// func ParsePhrases(text string) (nounPhrases []*t.Phrase, verbPhrases []*t.Phrase, allWords []*t.Word) {
// 	ctx := context.Background()
// 	lc, err := language.NewClient(ctx)
// 	if err != nil {
// 		log.Fatalf("failed to create language client: %s", err)
// 	}
// 	resp, err := lc.AnalyzeSyntax(ctx, buildSyntaxRequest(text))
// 	if err != nil {
// 		log.Fatalf("failed to analyze syntax: %s", err)
// 	}

// 	nouns, verbs, adjs, dets, allWords := buildWords(resp)
// 	fmt.Println("Nouns:")
// 	for _, noun := range nouns {
// 		fmt.Println(noun, noun.DependencyEdge, noun.PartOfSpeech)
// 	}
// 	fmt.Println("Verbs:")
// 	for _, verb := range verbs {
// 		fmt.Println(verb, verb.DependencyEdge, verb.PartOfSpeech)
// 	}
// 	fmt.Println("Adjs:")
// 	for _, adj := range adjs {
// 		fmt.Println(adj, adj.DependencyEdge, adj.PartOfSpeech)
// 	}
// 	fmt.Println("Dets:")
// 	for _, det := range dets {
// 		fmt.Println(det, det.DependencyEdge, det.PartOfSpeech)
// 	}

// 	fmt.Println("nounPhrases:")
// 	// nounPhrases = groupPhrases(nouns, nouns, dets)
// 	nounPhrases = []*t.Phrase{}
// 	for _, noun := range nouns {
// 		if t.FindWordByIdx(nouns, noun.DependencyEdge.HeadTokenIndex) == nil {
// 			nounPhrases = append(nounPhrases, &t.Phrase{Word: noun})
// 		}
// 	}
// 	for _, noun := range nounPhrases {
// 		fmt.Println(noun, noun.AllLemmas())
// 	}
// 	fmt.Println("verbPhrases:")
// 	verbPhrases = groupPhrases(verbs, nouns, adjs)
// 	for _, verb := range verbPhrases {
// 		fmt.Println(verb, verb.AllLemmas())
// 	}

// 	return nounPhrases, verbPhrases, allWords
// }

func RemoveReservedTwitterWords(text string) (result string) {
	var handleRgx = regexp.MustCompile(`\@[^\s]*`)
	var hashRgx = regexp.MustCompile(`\#[^\s]*`)
	result = handleRgx.ReplaceAllString(text, " ")
	result = hashRgx.ReplaceAllString(result, " ")
	return result
}

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

func FindActions(words []*t.Word) (actionWords []t.Word) {
	verbs := t.FindWords(&words, nil, &[]string{"VERB"}, nil)
	for _, v := range *verbs {
		if isActionLemma(v.Lemma) {
			vChildren := t.FindWords(&words, &v.Index, &[]string{"NOUN"}, nil)
			if len(*vChildren) > 0 {
				actionWords = append(actionWords, v)
			}
		}
	}
	return actionWords
}

func FindOperatorPhrases(words []*t.Word, actions []t.Word) (phrases []t.OperatorPhrase) {
	for _, action := range actions {
		nouns := t.FindWords(&words, &action.Index, &[]string{"NOUN"}, nil)
		for _, noun := range *nouns {
			if isMetricLemma(noun.Lemma) {
				adjs := t.FindWords(&words, &noun.Index, &[]string{"ADJ"}, nil)
				for _, adj := range *adjs {
					phrases = append(phrases, t.OperatorPhrase{
						OperatorWord: adj,
						ActionWord:   action,
					})
				}
			}
		}
	}
	return phrases
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

// func FindActionMetricOperator(words []*t.Word) (*t.Word, *t.Word, metric *t.Metric) {
// 	var actionWord, operatorWord t.Word
// 	verbs := t.FindWords(&words, nil, &[]string{"VERB"}, nil)
// 	for _, v := range *verbs {
// 		str := v.Lemma
// 		if str == "score" || str == "have" || str == "gain" {
// 			vChildren := t.FindWords(&words, &v.Index, &[]string{"NOUN"}, nil)
// 			if len(*vChildren) == 0 {
// 				continue
// 			}
// 			// found action
// 			fmt.Println("action: ", v.Lemma, t.WordsLemmas(vChildren))
// 			actionWord = &v
// 			for _, n := range *vChildren {
// 				str = n.Lemma
// 				if str == "point" || str == "pt" || str == "yard" || str == "yd" || str == "touchdown" || str == "td" {
// 					aChildren := t.FindWords(&words, &n.Index, &[]string{"NOUN", "ADJ"}, nil)
// 					if len(*aChildren) > 0 {
// 						// found metric
// 						metric = &t.Metric{
// 							Text:      n.Text,
// 							Lemma:     n.Lemma,
// 							Modifiers: []string{},
// 						}
// 						for _, a := range *aChildren {
// 							str = a.Lemma
// 							if str == "more" || str == "great" || str == "few" || str == "less" {
// 								// found operator
// 								operatorWord = &a
// 								fmt.Println("opword:", operatorWord.Lemma)
// 							} else if a.PartOfSpeech.Tag == "NOUN" || a.PartOfSpeech.Tag == "ADJ" {
// 								// found metric modifier
// 								metric.Modifiers = append(metric.Modifiers, a.Text)
// 							}
// 							fmt.Println("opword loop:", operatorWord)
// 						}

// 						if metric != nil && actionWord != nil && operatorWord != nil {
// 							fmt.Println("opword2:", operatorWord.Lemma)
// 							return actionWord, operatorWord, metric
// 						}
// 					}
// 				}

// 			}
// 		}
// 	}

// 	return nil, nil, nil
// }

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

// func groupPhrases(parents []*t.Word, children ...[]*t.Word) (phrases []*t.Phrase) {
// 	for _, parent := range parents {
// 		for _, child := range children {
// 			for _, word := range child {
// 				if word.DependencyEdge.HeadTokenIndex == parent.Index {
// 					phrase := t.FindPhraseByIdx(phrases, parent.Index)
// 					if phrase == nil {
// 						phrases = append(phrases, &t.Phrase{Word: parent})
// 					}
// 				}
// 			}
// 		}
// 	}

// 	return phrases
// }

func buildWords(resp *langpb.AnalyzeSyntaxResponse) (allWords []*t.Word) {
	for i, token := range resp.Tokens {
		// tokens = append(tokens, token)
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
