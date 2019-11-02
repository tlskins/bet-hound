package main

import (
	"context"
	"log"
	// "strings"

	language "cloud.google.com/go/language/apiv1"

	"bet-hound/cmd/nlp/analyze"
)

func main() {
	// [START init]
	ctx := context.Background()
	client, err := language.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// [END init]

	text := "bet you that tevin coleman scores more ppr points than matt breida this week"

	analyze.PrintResp(analyze.AnalyzeSyntax(ctx, client, text))
}
