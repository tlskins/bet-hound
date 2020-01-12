package gql

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	// "github.com/99designs/gqlgen/graphql"

	"bet-hound/cmd/betting"
	"bet-hound/cmd/db"
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

// type ContextKey string

// var (
// 	UserIDCtxKey = contextKey("userID")
// )

func New() Config {
	return Config{
		Resolvers: &resolver{
			Rooms:        map[string]*types.Chatroom{},
			RotoObserver: &types.RotoObserver{},
			RotoArticles: map[string][]*types.RotoArticle{},
		},
		// Directives: DirectiveRoot{
		// 	User: func(ctx context.Context, obj interface{}, next graphql.Resolver, username string) (res interface{}, err error) {
		// 		return next(context.WithValue(ctx, "username", username))
		// 	},
		// 	IsAuthenticated: func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		// 		ctxUserID := ctx.Value(UserIDCtxKey)
		// 		if ctxUserID != nil {
		// 			return next(ctx)
		// 		} else {
		// 			return nil, fmt.Errorf("Unauthorized")
		// 		}
		// 	},
		// },
	}
}

func getUsername(ctx context.Context) string {
	if username, ok := ctx.Value("username").(string); ok {
		return username
	}
	return ""
}

func UserFromContext(ctx context.Context) *types.User {
	// userId, err := ctx.Value(auth.ContextKey("userID")).(string)
	authPointer := ctx.Value(auth.ContextKey("userID")).(*auth.AuthResponseWriter)
	fmt.Println("userid in context", authPointer.UserId)
	if user, err := db.FindUserById(authPointer.UserId); err == nil {
		return user
	}
	return nil
}

type mutationResolver struct{ *resolver }

func (r *mutationResolver) SignIn(ctx context.Context, userName string, password string) (user *types.User, err error) {
	fmt.Println("resolver", userName, password)
	// return UserFromContext(ctx), nil
	user, err = db.SignInUser(userName, password)
	if err == nil {
		authPointer := ctx.Value(auth.ContextKey("userID")).(*auth.AuthResponseWriter)
		authPointer.SetSession(user.Id)
		return
	} else {
		return nil, fmt.Errorf("Invalid user name or password")
	}

}
func (r *mutationResolver) CreateBet(ctx context.Context, changes types.BetChanges) (bet *types.Bet, err error) {
	return betting.CreateBet(changes)
}
func (r *mutationResolver) UpdateBet(ctx context.Context, id string, changes types.BetChanges) (bet *types.Bet, err error) {
	return betting.UpdateBet(id, changes)
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

func (r *queryResolver) LeagueSettings(ctx context.Context, id string) (*types.LeagueSettings, error) {
	return db.GetLeagueSettings(id)
}
func (r *queryResolver) Bets(ctx context.Context) ([]*types.Bet, error) {
	if user := UserFromContext(ctx); user == nil {
		return nil, fmt.Errorf("Access denied")
	}

	return db.AllBets(), nil
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
