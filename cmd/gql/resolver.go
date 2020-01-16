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
	"bet-hound/cmd/gql/server/auth"
	"bet-hound/cmd/scraper"
	"bet-hound/cmd/types"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type resolver struct {
	Rooms        map[string]*types.Chatroom
	RotoObserver *types.RotoObserver
	RotoArticles map[string][]*types.RotoArticle
	mu           sync.Mutex
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
			Rooms:        map[string]*types.Chatroom{},
			RotoObserver: &types.RotoObserver{},
			RotoArticles: map[string][]*types.RotoArticle{},
		},
	}
}

func getUsername(ctx context.Context) string {
	if username, ok := ctx.Value("username").(string); ok {
		return username
	}
	return ""
}

func UserFromContext(ctx context.Context) (*types.User, error) {
	authPointer := ctx.Value(auth.ContextKey("userID")).(*auth.AuthResponseWriter)
	if user, err := db.FindUserById(authPointer.UserId); err == nil {
		return user, nil
	}
	return nil, fmt.Errorf("Access denied")
}

type mutationResolver struct{ *resolver }

func (r *mutationResolver) SignOut(ctx context.Context) (bool, error) {
	authPointer := ctx.Value(auth.ContextKey("userID")).(*auth.AuthResponseWriter)
	return authPointer.DeleteSession(env.AppUrl()), nil
}
func (r *mutationResolver) CreateBet(ctx context.Context, changes types.BetChanges) (bet *types.Bet, err error) {
	user, err := UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return betting.CreateBet(user, changes)
}
func (r *mutationResolver) AcceptBet(ctx context.Context, id string, accept bool) (bool, error) {
	user, err := UserFromContext(ctx)
	if err != nil {
		return false, err
	}

	return betting.AcceptBet(user, id, accept)
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

// need to change this nfl specific
func (r *mutationResolver) PostRotoArticle(ctx context.Context) (*types.RotoArticle, error) {
	if len(r.RotoObserver.Observers) == 0 {
		return nil, nil
	}
	articles, err := scraper.RotoNflArticles(10)
	if err != nil {
		return nil, err
	}
	last := articles[0] // last article is first in array
	if last == nil || last.Title == r.RotoObserver.Title {
		return nil, nil
	}

	r.mu.Lock()
	r.RotoArticles["nfl"] = articles
	r.RotoObserver.Title = last.Title
	for _, observer := range r.RotoObserver.Observers {
		observer <- last
	}
	r.mu.Unlock()

	return last, nil
}

type queryResolver struct{ *resolver }

func (r *queryResolver) SignIn(ctx context.Context, userName string, password string) (user *types.User, err error) {
	if user, err = db.SignInUser(userName, password); err == nil {
		authPointer := ctx.Value(auth.ContextKey("userID")).(*auth.AuthResponseWriter)
		authPointer.SetSession(env.AppUrl(), user.Id)
		return
	}
	return nil, fmt.Errorf("Invalid user name or password")
}
func (r *queryResolver) LeagueSettings(ctx context.Context, id string) (*types.LeagueSettings, error) {
	return db.GetLeagueSettings(id)
}
func (r *queryResolver) Bets(ctx context.Context) ([]*types.Bet, error) {
	user, err := UserFromContext(ctx)
	if err != nil {
		return []*types.Bet{}, err
	}

	return db.Bets(user.Id), nil
}
func (r *queryResolver) Bet(ctx context.Context, id string) (*types.Bet, error) {
	return db.FindBetById(id)
}
func (r *queryResolver) FindGames(ctx context.Context, team *string, gameTime *time.Time, week *int, year *int) ([]*types.Game, error) {
	return db.SearchGames(team, gameTime, week, year, 10)
}
func (r *queryResolver) FindPlayers(ctx context.Context, name *string, team *string, position *string, withGame *bool) ([]*types.Player, error) {
	if withGame != nil && *withGame {
		return db.SearchPlayersWithGame(name, team, position, 10)
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

func (r *subscriptionResolver) RotoArticleAdded(ctx context.Context) (<-chan *types.RotoArticle, error) {
	events := make(chan *types.RotoArticle, 1)

	r.mu.Lock()
	r.RotoObserver.Observers = append(r.RotoObserver.Observers, events)
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
