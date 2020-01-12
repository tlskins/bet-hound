package types

type User struct {
	Id          string       `bson:"_id" json:"id"`
	Name        string       `bson:"nm" json:"user_name"`
	UserName    string       `bson:"usr_nm" json:"user_name"`
	Password    string       `bson:"pwd" json:"password"`
	Email       string       `bson:"en" json:"email"`
	TwitterUser *TwitterUser `bson:"twt" json:"twitter_user"`
}
