package types

type User struct {
	Id          string       `bson:"_id" json:"id"`
	Name        string       `bson:"nm" json:"name"`
	UserName    string       `bson:"usr_nm" json:"user_name"`
	Password    string       `bson:"pwd" json:"password"`
	Email       string       `bson:"em" json:"email"`
	TwitterUser *TwitterUser `bson:"twt" json:"twitter_user"`
}

type ProfileChanges struct {
	Name     *string `bson:"nm,omitempty" json:"name,omitempty"`
	UserName *string `bson:"usr_nm,omitempty" json:"user_name,omitempty"`
	Password *string `bson:"pwd,omitempty" json:"password,omitempty"`
}
