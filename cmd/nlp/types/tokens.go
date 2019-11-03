package types

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

type Word struct {
	Text           string
	Lemma          string
	Index          int
	PartOfSpeech   *PartOfSpeech
	DependencyEdge *DependencyEdge
	Dependents     *[]*Word
}

func (w *Word) AllLemmas() (lemmas []string) {
	lemmas = append(lemmas, w.Lemma)
	if w.Dependents == nil {
		return lemmas
	}
	for _, d := range *w.Dependents {
		lemmas = append(lemmas, d.AllLemmas()...)
	}
	return lemmas
}
