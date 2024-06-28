package helper

import "fmt"

var (
	EnvToken     = "PECKER_TOKEN"
	EnvType      = "PECKER_TYPE"
	EnvLang      = "PECKER_LANG"
	EnvAPPID     = "PECKER_APPID"
	EnvAPISecret = "PECKER_API_SECRET"
	EnvAPIKey    = "PECKER_API_KEY"
)

type Options struct {
	Token     string
	Typ       string
	Lang      string
	AppID     string
	APISecret string
	APIKey    string
}

func NewOptions() *Options {
	return &Options{
		Token:     getEnvOrDefault(EnvToken, ""),
		Typ:       getEnvOrDefault(EnvType, "Spark"),
		Lang:      getEnvOrDefault(EnvLang, "Chinese"),
		AppID:     getEnvOrDefault(EnvAPPID, "e5bdf6ba"),
		APISecret: getEnvOrDefault(EnvAPISecret, "ZTMxZDJkODY0YjVkOWVlYjE2OTdjNGI2"),
		APIKey:    getEnvOrDefault(EnvAPIKey, "57fb33db4729a9590c11470e1872c5e4"),
	}
}

func (this *Options) GenerateHelpMessage() string {
	return fmt.Sprintf(`
You need at least three ENVs to run this plugin.
Set %s to specify your token.
Set %s to specify your token type, current type is: %s.
Set %s to specify the language, current language is: %s. Valid options like Chinese, French, Spain, etc.
`, EnvToken, EnvType, this.Typ, EnvLang, this.Lang)
}
