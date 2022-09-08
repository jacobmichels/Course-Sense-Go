package config

type Config struct {
	Firestore struct {
		ProjectID           string `mapstructure:"project_id"`
		CredentialsFilePath string `mapstructure:"credentials_file"`
		CollectionID        string `mapstructure:"collection_id"`
	} `mapstructure:"firestore"`
	Twilio struct {
		AccountSID  string `mapstructure:"account_sid"`
		AuthToken   string `mapstructure:"auth_token"`
		PhoneNumber string `mapstructure:"phone_number"`
	} `mapstructure:"twilio"`
	Smtp struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		From     string `mapstructure:"from"`
	} `mapstructure:"smtp"`
}
