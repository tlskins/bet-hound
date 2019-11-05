package types

import (
	"strings"
)

type Bet struct {
	Id                    *string       `bson:"_id,omitempty" json:"id"`
	Fk                    *string       `bson:"fk,omitempty" json:"fk"`
	ActionPhrase          *Phrase       `bson:"act_phrs,omitempty" json:"action_phrase"`
	MetricPhrase          *MetricPhrase `bson:"met_phrs,omitempty" json:"metric_phrase"`
	ProposerSourcePhrase  *Phrase       `bson:"p_src_phrs,omitempty" json:"proposer_source_phrase"`
	RecipientSourcePhrase *Phrase       `bson:"r_src_phrs,omitempty" json:"recipient_source_phrase"`
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
