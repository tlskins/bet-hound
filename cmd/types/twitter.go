package types

type Tweet struct {
	Id                   int64       `bson:"_id" json:"id"`
	IdStr                string      `bson:"id_str" json:"id_str"`
	Text                 *string     `bson:"txt,omitempty" json:"text"`
	FullText             *string     `bson:"full_txt,omitempty" json:"full_text"`
	InReplyToStatusId    int64       `bson:"rply_to_id" json:"in_reply_to_status_id"`
	InReplyToStatusIdStr string      `bson:"rply_to_id_str" json:"in_reply_to_status_id_str"`
	InReplyToUserIdStr   string      `bson:"rply_to_usr_id_str" json:"in_reply_to_user_id_str"`
	InReplyToScreenName  string      `bson:"rply_to_sn" json:"in_reply_to_screen_name"`
	TwitterUser          TwitterUser `bson:"usr" json:"user"`
	Entities             Entities    `bson:"entities" json:"entities"`
}

func (t *Tweet) GetText() (text string) {
	if t.FullText != nil {
		return *t.FullText
	} else if t.Text != nil {
		return *t.Text
	} else {
		return ""
	}
}

func (t *Tweet) Recipients(botHandle string) (handles []TwitterUser) {
	for _, user := range t.Entities.UserMentions {
		if user.ScreenName != botHandle {
			handles = append(handles, user)
		}
	}
	return handles
}

type Entities struct {
	Hashtags     []string      `bson:"hashtgs" json:"hashtags"`
	Symbols      []string      `bson:"symb" json:"symbols"`
	UserMentions []TwitterUser `bson:"usr_mtns" json:"user_mentions"`
	// Urls         []string `bson:"urls" json:"urls"`
}

type TwitterUser struct {
	Id         int64    `bson:"_id" json:"id"`
	ScreenName string   `bson:"scrn_nm" json:"screen_name"`
	Name       string   `bson:"nm" json:"name"`
	IdStr      string   `bson:"id_str" json:"id_str"`
	Indices    *[]int64 `bson:"indices,omitempty" json:"indices"`
}

type WebhookLoad struct {
	UserId           string  `json:"for_user_id"`
	TweetCreateEvent []Tweet `json:"tweet_create_events"`
}
