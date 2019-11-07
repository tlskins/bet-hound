package types

type Tweet struct {
	Id                   int64  `bson:"_id" json:"id"`
	IdStr                string `bson:"id_str" json:"id_str"`
	FullText             string `bson:"full_txt" json:"full_text"`
	InReplyToStatusId    int64  `bson:"rply_to_id" json:"in_reply_to_status_id"`
	InReplyToStatusIdStr string `bson:"rply_to_id_str" json:"in_reply_to_status_id_str"`
	InReplyToScreenName  string `bson:"rply_to_sn" json:"in_reply_to_screen_name"`
	User                 User   `bson:"usr" json:"user"`
}

type Entities struct {
	Hashtags     []string      `bson:"hashtgs" json:"hashtags"`
	Symbols      []string      `bson:"symb" json:"symbols"`
	UserMentions []UserMention `bson:"usr_mtns" json:"user_mentions"`
	Urls         []string      `bson:"urls" json:"urls"`
}

type UserMention struct {
	ScreenName string  `bson:"scrn_nm" json:"screen_name"`
	Name       string  `bson:"nm" json:"name"`
	Id         int64   `bson:"_id" json:"id"`
	IdStr      string  `bson:"id_str" json:"id_str"`
	Indices    []int64 `bson:"indices" json:"indices"`
}

type User struct {
	Id         int64  `bson:"_id" json:"id"`
	IdStr      string `bson:"id_str" json:"id_str"`
	Name       string `bson:"nm" json:"name"`
	ScreenName string `json:"screen_name"`
}
