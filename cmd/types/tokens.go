package types

type Word struct {
	Text           string         `bson:"txt,omitempty" json:"text"`
	Lemma          string         `bson:"lemma,omitempty" json:"lemma"`
	Index          int            `bson:"idx,omitempty" json:"index"`
	PartOfSpeech   PartOfSpeech   `bson:"pos,omitempty" json:"part_of_speech"`
	DependencyEdge DependencyEdge `bson:"dep_edge,omitempty" json:"dependency_edge"`
	BetComponent   string         `bson:"b_comp" json:"bet_component"`
}

type PartOfSpeech struct {
	Tag    string `bson:"tag,omitempty" json:"tag"`
	Proper string `bson:"proper,omitempty" json:"proper"`
	Case   string `bson:"case,omitempty" json:"case"`
	Person string `bson:"person,omitempty" json:"person"`
	Mood   string `bson:"mood,omitempty" json:"mood"`
	Tense  string `bson:"tense,omitempty" json:"tense"`
}

type DependencyEdge struct {
	Label             string `bson:"label,omitempty" json:"label"`
	HeadTokenIndex    int    `bson:"hd_tkn_idx,omitempty" json:"head_token_index"`
	ChildTokenIndices []int  `bson:"ch_tkn_idxs,omitempty" json:"child_token_indices"`
}
