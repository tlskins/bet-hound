package types

type Phrase struct {
	Word     *Word   `bson:"word,omitempty" json:"word"`
	Source   *Source `bson:"src,omitempty" json:"source"`
	HomeGame *Game   `bson:"h_gm,omitempty" json:"home_game"`
	AwayGame *Game   `bson:"a_gm,omitempty" json:"away_game"`
}

type MetricPhrase struct {
	Word          *Word   `bson:"word,omitempty" json:"word"`
	OperatorWord  *Word   `bson:"op_word,omitempty" json:"operator_word"`
	ModifierWords []*Word `bson:"mod_words,omitempty" json:"modifier_words"`
}

func (p *Phrase) Game() *Game {
	if p.HomeGame != nil {
		return p.HomeGame
	} else if p.AwayGame != nil {
		return p.AwayGame
	}
	return nil
}

func (p *Phrase) SourceDesc() (desc *string) {
	if p.Source == nil || p.Game() == nil {
		return nil
	}
	fName := (*p.Source.FirstName)[:1]
	lName := *p.Source.LastName
	pos := *p.Source.Position
	tm := *p.Source.TeamShort
	srcTeam := *p.Source.TeamFk
	gm := *p.Game()
	var vsTeam string
	if *gm.HomeTeamFk == srcTeam {
		vsTeam = *gm.AwayTeamName
	} else {
		vsTeam = *gm.HomeTeamName
	}
	result := fName + "." + lName + " (" + tm + "-" + pos + ")" + " vs " + vsTeam
	return &result
}

type Word struct {
	Text           string          `bson:"txt,omitempty" json:"text"`
	Lemma          string          `bson:"lemma,omitempty" json:"lemma"`
	Index          int             `bson:"idx,omitempty" json:"index"`
	PartOfSpeech   *PartOfSpeech   `bson:"pos,omitempty" json:"part_of_speech"`
	DependencyEdge *DependencyEdge `bson:"dep_edge,omitempty" json:"dependency_edge"`
	Parent         *Word           `bson:"-" json:"parent"`
	Children       *[]*Word        `bson:"-" json:"children"`
}

func descendentLemmas(word *Word) (lemmas []string) {
	lemmas = append(lemmas, word.Lemma)
	if word.Children != nil {
		for _, child := range *word.Children {
			lemmas = append(lemmas, child.Lemma)
		}
	}
	return lemmas
}

func (p *Phrase) AllLemmas() []string {
	return descendentLemmas(p.Word)
}

func (m *MetricPhrase) AllLemmas() []string {
	return descendentLemmas(m.Word)
}

func descendentText(word *Word) (text []string) {
	text = append(text, word.Text)
	if word.Children != nil {
		for _, child := range *word.Children {
			text = append(text, child.Lemma)
		}
	}
	return text
}

func (p *Phrase) AllText() []string {
	return descendentText(p.Word)
}

func (m *MetricPhrase) AllText() []string {
	return descendentText(m.Word)
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
	Label          string `bson:"label,omitempty" json:"label"`
	HeadTokenIndex int    `bson:"hd_tkn_idx,omitempty" json:"head_token_index"`
}

func FindWordByTxt(words []*Word, txt string) *Word {
	for _, w := range words {
		if w.Text == txt {
			return w
		}
	}
	return nil
}

func FindWordByIdx(words []*Word, idx int) *Word {
	for _, w := range words {
		if w.Index == idx {
			return w
		}
	}
	return nil
}

func FindPhraseByIdx(phrases []*Phrase, index int) *Phrase {
	for _, p := range phrases {
		if p.Word.Index == index {
			return p
		}
	}
	return nil
}

func findPhraseByWordTxt(phrases []*Phrase, txt string) *Phrase {
	for _, p := range phrases {
		if p.Word.Text == txt {
			return p
		}
	}
	return nil
}
