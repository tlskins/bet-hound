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
	Observer chan *types.Notification
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

func getUsername(ctx context.Context) string {
	if username, ok := ctx.Value("username").(string); ok {
		return username
	}
	return ""
}

func userFromContext(ctx context.Context) (*types.User, error) {
	authPointer := ctx.Value(mw.AuthContextKey("userID")).(*mw.AuthResponseWriter)
	if user, err := db.FindUserById(authPointer.UserId); err == nil {
		return user, nil
	}
	return nil, fmt.Errorf("Access denied")
}

func leagueFromContext(ctx context.Context) (*types.LeagueSettings, error) {
	lgPointer := ctx.Value(mw.LgContextKey("league")).(*types.LeagueSettings)
	return lgPointer, nil
}

type mutationResolver struct{ *resolver }

func (r *mutationResolver) SignOut(ctx context.Context) (bool, error) {
	authPointer := ctx.Value(mw.AuthContextKey("userID")).(*mw.AuthResponseWriter)
	return authPointer.DeleteSession(env.AppHost()), nil
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
	bet, note, err := betting.CreateBet(user, &changes, sttgs)
	if err != nil {
		return nil, err
	}
	// push notifications if online
	for _, userId := range []string{bet.Proposer.Id, bet.Recipient.Id} {
		if r.UserObservers[userId] != nil {
			r.mu.Lock()
			r.UserObservers[userId].Observer <- note
			r.mu.Unlock()
		}
	}

	return
}
func (r *mutationResolver) AcceptBet(ctx context.Context, id string, accept bool) (bool, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return false, err
	}
	bet, note, err := betting.AcceptBet(user, id, accept)
	if err != nil {
		return false, err
	}
	// push notifications if online
	for _, userId := range []string{bet.Proposer.Id, bet.Recipient.Id} {
		if r.UserObservers[userId] != nil {
			r.mu.Lock()
			r.UserObservers[userId].Observer <- note
			r.mu.Unlock()
		}
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
	for _, userObserver := range r.UserObservers {
		userObserver.Observer <- &types.Notification{
			Title:   last.Title,
			Type:    "RotoAlert",
			SentAt:  time.Now(),
			Message: last.Article,
		}
	}
	r.mu.Unlock()

	return last, nil
}

// not in use...
func (r *mutationResolver) PostUserNotification(ctx context.Context, userId string, sentAt time.Time, title string, typeArg string, message *string) (*types.Notification, error) {
	fmt.Println("mutationResolver.PostUserNotification...")
	if len(r.UserObservers) == 0 {
		return nil, nil
	}

	sent := false
	note := &types.Notification{
		Title:  title,
		Type:   typeArg,
		SentAt: sentAt,
	}
	r.mu.Lock()
	if message != nil {
		note.Message = *message
	}
	if r.UserObservers[userId] != nil {
		r.UserObservers[userId].Observer <- note
		sent = true
	}
	r.mu.Unlock()

	if sent {
		return note, nil
	} else {
		return nil, nil
	}
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

func (r *subscriptionResolver) UserNotification(ctx context.Context) (<-chan *types.Notification, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}
	events := make(chan *types.Notification, 1)

	r.mu.Lock()
	r.UserObservers[user.Id] = &UserObserver{Observer: events}
	r.mu.Unlock()

	return events, nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
