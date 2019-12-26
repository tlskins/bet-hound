package gql

import (
	"bet-hound/cmd/db"
	"bet-hound/cmd/types"
	"context"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Bets(ctx context.Context) ([]*types.Bet, error) {
	return db.AllBets(), nil
}
func (r *queryResolver) Bet(ctx context.Context, id string) (*types.Bet, error) {
	return db.FindBetById(id)
}
