package db

import (
	t "bet-hound/cmd/types"
)

func SearchSubjects(settings *t.LeagueSettings, search string) (subjects []t.SubjectUnion, err error) {
	players, err := SearchPlayersWithGame(settings, &search, nil, nil, 5)
	if err != nil {
		return
	}
	teams, err := SearchTeamsWithGame(settings, &search, &search, 5)
	if err != nil {
		return
	}

	for _, player := range players {
		var s t.SubjectUnion = *player
		subjects = append(subjects, s)
	}
	for _, team := range teams {
		var s t.SubjectUnion = *team
		subjects = append(subjects, s)
	}

	return
}
