package types

import (
	"fmt"
	"strings"
)

type BetStatus int

const (
	BetStatusPendingProposer BetStatus = iota
	BetStatusPendingRecipient
	BetStatusAccepted
	BetStatusFinal
)

func BetStatusFromString(s string) BetStatus {
	return map[string]BetStatus{
		"Pending Proposer":  BetStatusPendingProposer,
		"Pending Recipient": BetStatusPendingRecipient,
		"Accepted":          BetStatusAccepted,
		"Final":             BetStatusFinal,
	}[s]
}
func (s BetStatus) String() string {
	return map[BetStatus]string{
		BetStatusPendingProposer:  "Pending Proposer",
		BetStatusPendingRecipient: "Pending Recipient",
		BetStatusAccepted:         "Accepted",
		BetStatusFinal:            "Final",
	}[s]
}

type Bet struct {
	Id                    *string       `bson:"_id" json:"id"`
	Fk                    *string       `bson:"fk,omitempty" json:"fk"`
	ActionPhrase          *Phrase       `bson:"act_phrs,omitempty" json:"action_phrase"`
	MetricPhrase          *MetricPhrase `bson:"met_phrs,omitempty" json:"metric_phrase"`
	ProposerSourcePhrase  *Phrase       `bson:"p_src_phrs,omitempty" json:"proposer_source_phrase"`
	RecipientSourcePhrase *Phrase       `bson:"r_src_phrs,omitempty" json:"recipient_source_phrase"`
	BetStatus             BetStatus     `bson:"status" json:"bet_status"`
	Proposer              *User         `bson:"proposer,omitempty" json:"proposer"`
	Recipient             *User         `bson:"recipient,omitempty" json:"recipient"`
	ProposerCheckTweetId  *string       `bson:"pchk_tweet_id,omitempty" json:"proposer_check_tweet_id"`
	RecipientCheckTweetId *string       `bson:"rchk_tweet_id,omitempty" json:"recipient_check_tweet_id"`
}

func (b *Bet) Response() (txt string) {
	if b.BetStatus.String() == "Pending Proposer" {
		return fmt.Sprintf("%s%s Is this correct: \"%s\" ? Reply \"Yes\"", "@", b.Proposer.ScreenName, b.Text())
	} else if b.BetStatus.String() == "Pending Recipient" {
		return fmt.Sprintf("%s%s Do you accept this bet? : \"%s\" Reply \"Yes\"", "@", b.Recipient.ScreenName, b.Text())
	}
	return "Pending game results..."
}

func (b *Bet) Text() (txt string) {
	pSrc := *b.ProposerSourcePhrase.SourceDesc()
	action := strings.ToLower(b.ActionPhrase.Word.Text)
	comp := strings.ToLower(b.MetricPhrase.OperatorWord.Lemma)
	modsArr := []string{}
	for _, m := range b.MetricPhrase.ModifierWords {
		modsArr = append(modsArr, strings.ToLower(m.Lemma))
	}
	subjMods := strings.Join(modsArr, " ")
	subject := strings.ToLower(b.MetricPhrase.Word.Text)
	subject = strings.TrimSpace(subjMods + subject)
	rSrc := *b.RecipientSourcePhrase.SourceDesc()

	return strings.Join([]string{pSrc, action, comp, subject, rSrc}, " ")
}
