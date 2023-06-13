package config

type Config struct {
	Database      Database
	Notifications Notifications
}

type Database struct {
	Type      string `mapstructure:"type"`
	Firestore Firestore
	SQLite    SQLite
}

type Firestore struct {
	ProjectID           string `mapstructure:"project_id"`
	CredentialsFile     string `mapstructure:"credentials_file"`
	SectionCollectionID string `mapstructure:"section_collection_id"`
	WatcherCollectionID string `mapstructure:"watcher_collection_id"`
}

type SQLite struct {
	ConnectionString string `mapstructure:"connection_string"`
}

type Notifications struct {
	EmailSmtp EmailSmtp
}

type EmailSmtp struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}
