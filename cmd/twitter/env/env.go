package env

import (
	"github.com/spf13/viper"
)

type environment struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessTokenKey    string
	AccessTokenSecret string
	WebhookEnv        string
	AppUrl            string
	LogPath           string
}

var E = &environment{}

func Init(configFile, configPath string) error {
	viper.SetConfigName(configFile)
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	E.ConsumerKey = viper.GetString("consumer_key")
	E.ConsumerSecret = viper.GetString("consumer_secret")
	E.AccessTokenKey = viper.GetString("access_token_key")
	E.AccessTokenSecret = viper.GetString("access_token_secret")
	E.WebhookEnv = viper.GetString("webhook_env")
	E.AppUrl = viper.GetString("app_url")
	E.LogPath = viper.GetString("log_path")

	return nil
}
