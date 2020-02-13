package db

import (
	"github.com/globalsign/mgo"

	"bet-hound/cmd/env"
)

func EnsureIndexes(db *mgo.Database) (err error) {
	// games indexes
	gDb := db.C(env.GamesCollection())
	if err = gDb.EnsureIndex(mgo.Index{
		Key:        []string{"wk", "yr"},
		Background: true,
	}); err != nil {
		return err
	}
	if err = gDb.EnsureIndexKey("a_team_fk"); err != nil {
		return err
	}
	if err = gDb.EnsureIndexKey("h_team_fk"); err != nil {
		return err
	}
	if err = gDb.EnsureIndexKey("fin"); err != nil {
		return err
	}

	// user indexes
	uDb := db.C(env.UsersCollection())
	if err = uDb.EnsureIndex(mgo.Index{
		Key:        []string{"nm", "usr_nm", "em"},
		Background: true,
	}); err != nil {
		return err
	}
	if err = uDb.EnsureIndex(mgo.Index{
		Key:        []string{"usr_nm", "twt.scrn_nm"},
		Unique:     true,
		DropDups:   true,
		Background: false,
	}); err != nil {
		return err
	}
	// if err = uDb.EnsureIndex(mgo.Index{
	// 	Key:        []string{"twt._id"},
	// 	Unique:     true,
	// 	DropDups:   true,
	// 	Background: false,
	// }); err != nil {
	// 	return err
	// }

	// players indexes
	pDb := db.C(env.PlayersCollection())
	if err = pDb.EnsureIndex(mgo.Index{
		Key:        []string{"name", "team_fk", "team_short", "pos"},
		Background: true,
	}); err != nil {
		return err
	}
	if err = pDb.EnsureIndexKey("fk"); err != nil {
		return err
	}
	if err = pDb.EnsureIndexKey("team_fk"); err != nil {
		return err
	}

	// bet indexes
	bDb := db.C(env.BetsCollection())
	if err = bDb.EnsureIndex(mgo.Index{
		Key:        []string{"final_at", "lg_id", "proposer", "recipient", "status", "rslt"},
		Background: true,
	}); err != nil {
		return err
	}
	if err = bDb.EnsureIndex(mgo.Index{
		Key:        []string{"acc_fk", "status"},
		Background: true,
	}); err != nil {
		return err
	}
	if err = bDb.EnsureIndexKey("crt_at"); err != nil {
		return err
	}

	// leader board indexes
	lbDb := db.C(env.BetsCollection())
	if err = lbDb.EnsureIndex(mgo.Index{
		Key:        []string{"st", "end", "fin", "lg_id"},
		Background: true,
	}); err != nil {
		return err
	}

	return nil
}
