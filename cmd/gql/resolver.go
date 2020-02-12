package gql

import (
	"context"
	"fmt"
	"sync"
	"time"

	"bet-hound/cmd/betting"
	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/scraper"
	"bet-hound/cmd/types"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type UserObserver struct {
	Profile chan *types.User
}

type resolver struct {
	UserObservers map[string]*UserObserver
	RotoArticles  map[string][]*types.RotoArticle
	LastRotoTitle string
	mu            sync.Mutex
}

func (r *resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

func (r *resolver) Query() QueryResolver {
	return &queryResolver{r}
}

func (r *resolver) Subscription() SubscriptionResolver {
	return &subscriptionResolver{r}
}

func New() Config {
	return Config{
		Resolvers: &resolver{
			UserObservers: map[string]*UserObserver{},
			RotoArticles:  map[string][]*types.RotoArticle{},
		},
	}
}

type mutationResolver struct{ *resolver }

func (r *mutationResolver) SignOut(ctx context.Context) (bool, error) {
	authPointer := ctx.Value(AuthContextKey("userID")).(*AuthResponseWriter)
	return authPointer.DeleteSession(env.AppHost()), nil
}
func (r *mutationResolver) ViewProfile(ctx context.Context, sync bool) (*types.User, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if sync {
		return db.ViewUserProfile(user.Id)
	} else if users, err := db.FindUserByIds([]string{user.Id}); err != nil {
		return nil, err
	} else if len(users) > 0 {
		return users[0], nil
	}
	return nil, nil
}
func (r *mutationResolver) UpdateUser(ctx context.Context, changes types.ProfileChanges) (*types.User, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return db.UpdateUserProfile(user.Id, &changes)
}
func (r *mutationResolver) CreateBet(ctx context.Context, newBet types.NewBet) (bet *types.Bet, err error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bet, _, err = betting.CreateBet(user, &newBet)
	if err != nil {
		return nil, err
	}
	// push notifications if online
	userIds := []string{bet.Proposer.Id}
	if bet.Recipient != nil {
		userIds = append(userIds, bet.Recipient.Id)
	}
	users, err := db.FindUserByIds(userIds)
	for _, user := range users {
		r.pushUserProfileNotification(user)
	}

	return
}
func (r *mutationResolver) AcceptBet(ctx context.Context, id string, accept bool) (bool, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return false, err
	}
	bet, _, err := betting.AcceptBet(user, id, accept)
	if err != nil {
		return false, err
	}
	// push notifications
	users, err := db.FindUserByIds([]string{bet.Proposer.Id, bet.Recipient.Id})
	for _, user := range users {
		r.pushUserProfileNotification(user)
	}

	return true, err
}
func (r *mutationResolver) PostRotoArticle(ctx context.Context) (bool, error) {
	fmt.Println("mutationResolver.PostRotoArticle... userObservers:", r.UserObservers)
	if len(r.UserObservers) == 0 {
		return false, nil
	}
	articles, err := scraper.RotoNflArticles(10)
	if err != nil {
		return false, err
	}
	last := articles[0] // last article is first in array
	if last == nil || last.Title == r.LastRotoTitle {
		return false, nil
	}

	r.mu.Lock()
	r.RotoArticles["nfl"] = articles
	r.LastRotoTitle = last.Title
	note := &types.Notification{
		Title:   last.Title,
		Type:    "RotoAlert",
		SentAt:  time.Now(),
		Message: last.Article,
	}
	for _, userObserver := range r.UserObservers {
		userObserver.Profile <- &types.User{Notifications: []*types.Notification{note}}
	}
	r.mu.Unlock()

	return true, nil
}

type queryResolver struct{ *resolver }

func (r *queryResolver) SignIn(ctx context.Context, userName string, password string) (user *types.User, err error) {
	if user, err = db.SignInUser(userName, password); err == nil {
		authPointer := ctx.Value(AuthContextKey("userID")).(*AuthResponseWriter)
		authPointer.SetSession(env.AppHost(), user.Id)
		return
	}
	return nil, fmt.Errorf("Invalid user name or password")
}
func (r *queryResolver) Bets(ctx context.Context) (*types.BetsResponse, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return db.Bets(user.Id)
}
func (r *queryResolver) CurrentBets(ctx context.Context) (*types.BetsResponse, error) {
	return db.CurrentBets()
}
func (r *queryResolver) Bet(ctx context.Context, id string) (*types.Bet, error) {
	return db.FindBetById(id)
}
func (r *queryResolver) FindGames(ctx context.Context, team *string, gameTime *time.Time) ([]*types.Game, error) {
	return db.SearchGames(team, gameTime, 10)
}
func (r *queryResolver) FindPlayers(ctx context.Context, name *string, team *string, position *string) ([]*types.Player, error) {
	return db.SearchPlayersWithGame(name, team, position, 10)
}
func (r *queryResolver) FindUsers(ctx context.Context, search string) ([]*types.User, error) {
	return db.FindUser(search, 10)
}
func (r *queryResolver) CurrentRotoArticles(ctx context.Context, id string) (articles []*types.RotoArticle, err error) {
	articles = r.RotoArticles[id]
	if len(articles) == 0 {
		if articles, err = scraper.RotoNflArticles(10); err != nil {
			return
		}

		if len(articles) > 0 {
			r.mu.Lock()
			r.RotoArticles[id] = articles
			r.mu.Unlock()
		}
	}
	return articles, nil
}
func (r *queryResolver) CurrentGames(ctx context.Context) ([]*types.Game, error) {
	return db.GetCurrentGames()
}
func (r *queryResolver) SearchSubjects(ctx context.Context, search string) ([]types.SubjectUnion, error) {
	return db.SearchSubjects(search)
}
func (r *queryResolver) SearchBets(ctx context.Context, search string, userID, betStatus *string) ([]*types.Bet, error) {
	return db.SearchBets(search, userID, betStatus)
}
func (r *queryResolver) GetBetMaps(ctx context.Context, leagueId, betType *string) ([]*types.BetMap, error) {
	return db.GetBetMaps(leagueId, betType)
}
func (r *queryResolver) GetUser(ctx context.Context, userId string) (*types.User, error) {
	return db.FindUserById(userId)
}

type subscriptionResolver struct{ *resolver }

func (r *subscriptionResolver) SubscribeUserNotifications(ctx context.Context) (<-chan *types.User, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}
	events := make(chan *types.User, 1)

	r.mu.Lock()
	r.UserObservers[user.Id] = &UserObserver{Profile: events}
	r.mu.Unlock()

	return events, nil
}

// Helper functions

func (r *mutationResolver) pushUserProfileNotification(newProf *types.User) {
	if r.UserObservers[newProf.Id] == nil {
		return
	}

	r.mu.Lock()
	select {
	case r.UserObservers[newProf.Id].Profile <- newProf:
	case <-time.After(3 * time.Second):
		fmt.Println("push user profile timeout!")
	}
	r.mu.Unlock()
}

func userFromContext(ctx context.Context) (*types.User, error) {
	authPointer := ctx.Value(AuthContextKey("userID")).(*AuthResponseWriter)
	if users, err := db.FindUserByIds([]string{authPointer.UserId}); err == nil && len(users) > 0 {
		return users[0], nil
	}
	return nil, fmt.Errorf("Access denied")
}
