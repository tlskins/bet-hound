package db

import (
	t "bet-hound/cmd/types"
)

func SearchSubjects(search string) (subjects []t.SubjectUnion, err error) {
	players, err := SearchPlayersWithGame(&search, nil, nil, 5)
	if err != nil {
		return
	}
	teams, err := SearchTeamsWithGame(&search, &search, 5)
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
