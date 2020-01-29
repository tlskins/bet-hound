package gql

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"bet-hound/cmd/betting"
	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	mw "bet-hound/cmd/gql/server/middleware"
	"bet-hound/cmd/scraper"
	"bet-hound/cmd/types"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type UserObserver struct {
	Notifications chan *types.Notification
	Profile       chan *types.User
}

type resolver struct {
	Rooms         map[string]*types.Chatroom
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
			Rooms:         map[string]*types.Chatroom{},
			UserObservers: map[string]*UserObserver{},
			RotoArticles:  map[string][]*types.RotoArticle{},
		},
	}
}

type mutationResolver struct{ *resolver }

func (r *mutationResolver) SignOut(ctx context.Context) (bool, error) {
	authPointer := ctx.Value(mw.AuthContextKey("userID")).(*mw.AuthResponseWriter)
	return authPointer.DeleteSession(env.AppHost()), nil
}
func (r *mutationResolver) ViewProfile(ctx context.Context) (*types.User, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return db.ViewUserProfile(user.Id)
}
func (r *mutationResolver) UpdateUser(ctx context.Context, changes types.ProfileChanges) (*types.User, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return db.UpdateUserProfile(user.Id, &changes)
}
func (r *mutationResolver) CreateBet(ctx context.Context, changes types.BetChanges) (bet *types.Bet, err error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}
	sttgs, err := leagueFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bet, _, err = betting.CreateBet(user, &changes, sttgs)
	if err != nil {
		return nil, err
	}
	// push notifications if online
	users, err := db.FindUserByIds([]string{bet.Proposer.Id, bet.Recipient.Id})
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
	// push notifications if online
	users, err := db.FindUserByIds([]string{bet.Proposer.Id, bet.Recipient.Id})
	for _, user := range users {
		r.pushUserProfileNotification(user)
	}

	return true, nil
}
func (r *mutationResolver) Post(ctx context.Context, text string, username string, roomName string) (*types.Message, error) {
	r.mu.Lock()
	room := r.Rooms[roomName]
	if room == nil {
		room = &types.Chatroom{
			Name: roomName,
			Observers: map[string]struct {
				Username string
				Message  chan *types.Message
			}{},
		}
		r.Rooms[roomName] = room
	}
	r.mu.Unlock()

	message := types.Message{
		ID:        randString(8),
		CreatedAt: time.Now(),
		Text:      text,
		CreatedBy: username,
	}

	room.Messages = append(room.Messages, message)
	r.mu.Lock()
	for _, observer := range room.Observers {
		if observer.Username == "" || observer.Username == message.CreatedBy {
			observer.Message <- &message
		}
	}
	r.mu.Unlock()
	return &message, nil
}

func (r *mutationResolver) PostRotoArticle(ctx context.Context) (*types.RotoArticle, error) {
	fmt.Println("mutationResolver.PostRotoArticle... userObservers:", r.UserObservers)
	if len(r.UserObservers) == 0 {
		return nil, nil
	}
	articles, err := scraper.RotoNflArticles(10)
	if err != nil {
		return nil, err
	}
	last := articles[0] // last article is first in array
	if last == nil || last.Title == r.LastRotoTitle {
		return nil, nil
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

	return last, nil
}

type queryResolver struct{ *resolver }

func (r *queryResolver) SignIn(ctx context.Context, userName string, password string) (user *types.User, err error) {
	if user, err = db.SignInUser(userName, password); err == nil {
		authPointer := ctx.Value(mw.AuthContextKey("userID")).(*mw.AuthResponseWriter)
		authPointer.SetSession(env.AppHost(), user.Id)
		return
	}
	return nil, fmt.Errorf("Invalid user name or password")
}
func (r *queryResolver) LeagueSettings(ctx context.Context, id string) (*types.LeagueSettings, error) {
	return leagueFromContext(ctx)
}
func (r *queryResolver) Bets(ctx context.Context) ([]*types.Bet, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return []*types.Bet{}, err
	}

	return db.Bets(user.Id)
}
func (r *queryResolver) CurrentBets(ctx context.Context) ([]*types.Bet, error) {
	return db.CurrentBets()
}
func (r *queryResolver) Bet(ctx context.Context, id string) (*types.Bet, error) {
	return db.FindBetById(id)
}
func (r *queryResolver) FindGames(ctx context.Context, team *string, gameTime *time.Time, week *int, year *int) ([]*types.Game, error) {
	return db.SearchGames(team, gameTime, week, year, 10)
}
func (r *queryResolver) FindPlayers(ctx context.Context, name *string, team *string, position *string, withGame *bool) ([]*types.Player, error) {
	if withGame != nil && *withGame {
		settings, err := leagueFromContext(ctx)
		if err != nil {
			return []*types.Player{}, err
		}
		return db.SearchPlayersWithGame(settings, name, team, position, 10)
	} else {
		return db.SearchPlayers(name, team, position, 10)
	}
}
func (r *queryResolver) FindUsers(ctx context.Context, search string) ([]*types.User, error) {
	return db.FindUser(search, 10)
}
func (r *queryResolver) Room(ctx context.Context, name string) (*types.Chatroom, error) {
	r.mu.Lock()
	room := r.Rooms[name]
	if room == nil {
		room = &types.Chatroom{
			Name: name,
			Observers: map[string]struct {
				Username string
				Message  chan *types.Message
			}{},
		}
		r.Rooms[name] = room
	}
	r.mu.Unlock()

	return room, nil
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
	lgPointer := ctx.Value(mw.LgContextKey("league")).(*types.LeagueSettings)
	return db.GetCurrentGames(lgPointer)
}

type subscriptionResolver struct{ *resolver }

func (r *subscriptionResolver) MessageAdded(ctx context.Context, roomName string) (<-chan *types.Message, error) {
	r.mu.Lock()
	room := r.Rooms[roomName]
	if room == nil {
		room = &types.Chatroom{
			Name: roomName,
			Observers: map[string]struct {
				Username string
				Message  chan *types.Message
			}{},
		}
		r.Rooms[roomName] = room
	}
	r.mu.Unlock()

	id := randString(8)
	events := make(chan *types.Message, 1)

	go func() {
		<-ctx.Done()
		r.mu.Lock()
		delete(room.Observers, id)
		r.mu.Unlock()
	}()

	r.mu.Lock()
	room.Observers[id] = struct {
		Username string
		Message  chan *types.Message
	}{Username: getUsername(ctx), Message: events}
	r.mu.Unlock()

	return events, nil
}
func (r *subscriptionResolver) SubscribeNotifications(ctx context.Context) (<-chan *types.Notification, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}
	events := make(chan *types.Notification, 1)

	r.mu.Lock()
	r.UserObservers[user.Id] = &UserObserver{Notifications: events}
	r.mu.Unlock()

	return events, nil
}
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
	fmt.Println("pushing user profile note:", newProf.Id, r.UserObservers)
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
	fmt.Println("post pushing user profile note:", newProf.Id)
}

func (r *mutationResolver) pushUserNotification(userId string, note *types.Notification) {
	fmt.Println("pushing user note:", r.UserObservers)
	if r.UserObservers[userId] == nil {
		return
	}

	r.mu.Lock()
	select {
	case r.UserObservers[userId].Notifications <- note:
	case <-time.After(3 * time.Second):
		fmt.Println("push user notification timeout!")
	}
	r.mu.Unlock()
	fmt.Println("post pushing user note:", userId)
}

func getUsername(ctx context.Context) string {
	if username, ok := ctx.Value("username").(string); ok {
		return username
	}
	return ""
}

func userFromContext(ctx context.Context) (*types.User, error) {
	authPointer := ctx.Value(mw.AuthContextKey("userID")).(*mw.AuthResponseWriter)
	if users, err := db.FindUserByIds([]string{authPointer.UserId}); err == nil && len(users) > 0 {
		return users[0], nil
	}
	return nil, fmt.Errorf("Access denied")
}

func leagueFromContext(ctx context.Context) (*types.LeagueSettings, error) {
	lgPointer := ctx.Value(mw.LgContextKey("league")).(*types.LeagueSettings)
	return lgPointer, nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
