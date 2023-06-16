package config

type Config struct {
	qrlNodeConfig *QRLNodeConfig
	mongoDBConfig *MongoDBConfig

	ReOrgLimit uint64
}

type QRLNodeConfig struct {
	IP            string
	PublicAPIPort uint16
}

type MongoDBConfig struct {
	DBName   string
	Host     string
	Port     uint16
	Username string
	Password string
}

func GetConfig() *Config {
	c := &Config{
		qrlNodeConfig: &QRLNodeConfig{
			IP:            "127.0.0.1", // IP address of Python QRL node with PublicAPI support
			PublicAPIPort: 19009,
		},
		mongoDBConfig: &MongoDBConfig{
			DBName:   "QRLTokenIndexer",
			Host:     "127.0.0.1",
			Port:     27017, // Default MongoDB port
			Username: "",
			Password: "",
		},
		ReOrgLimit: 350,
	}
	return c
}

func (c *Config) GetQRLNodeConfig() *QRLNodeConfig {
	return c.qrlNodeConfig
}

func (c *Config) GetMongoDBConfig() *MongoDBConfig {
	return c.mongoDBConfig
}
