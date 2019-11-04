package types

type Bet struct {
	Id                    *string       `bson:"_id,omitempty" json:"id"`
	Fk                    *string       `bson:"fk,omitempty" json:"fk"`
	ActionPhrase          *Phrase       `bson:"act_phrs,omitempty" json:"action_phrase"`
	MetricPhrase          *MetricPhrase `bson:"met_phrs,omitempty" json:"metric_phrase"`
	ProposerSourcePhrase  *Phrase       `bson:"p_src_phrs,omitempty" json:"proposer_source_phrase"`
	RecipientSourcePhrase *Phrase       `bson:"r_src_phrs,omitempty" json:"recipient_source_phrase"`
}
