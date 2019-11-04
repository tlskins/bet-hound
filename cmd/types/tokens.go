package types

type Phrase struct {
	Word       *Word
	Dependents *[]*Word
	Source     *Source
}

type Word struct {
	Text           string
	Lemma          string
	Index          int
	PartOfSpeech   *PartOfSpeech
	DependencyEdge *DependencyEdge
	Parent         *Word
	Children       *[]*Word
}

func descendentLemmas(word *Word) (lemmas []string) {
	lemmas = append(lemmas, word.Lemma)
	if word.Children != nil {
		for _, child := range *word.Children {
			// children := descendentLemmas(child)
			// lemmas = append(lemmas, children...)
			lemmas = append(lemmas, child.Lemma)
		}
	}
	return lemmas
}

func (p *Phrase) AllLemmas() []string {
	return descendentLemmas(p.Word)
}

func descendentText(word *Word) (text []string) {
	text = append(text, word.Text)
	if word.Children != nil {
		for _, child := range *word.Children {
			// children := descendentText(child)
			// text = append(text, children...)
			text = append(text, child.Lemma)
		}
	}
	return text
}

func (p *Phrase) AllText() []string {
	return descendentText(p.Word)
}

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
