package vendingmachine

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host         string `required:"true" split_words:"true"`
	Port         string `required:"true" split_words:"true"`
	Secret       string `required:"true" split_words:"true"`
	DatabaseName string `required:"true" split_words:"true"`
	DatabaseURI  string `required:"true" split_words:"true"`
}

func loadEnvironment(filename string) error {
	err := godotenv.Load()
	// handle if .env file does not exist, this is OK
	if os.IsNotExist(err) {
		return nil
	}

	return err
}

func LoadConfiguration(filename string) (*Config, error) {
	err := loadEnvironment(filename)
	if err != nil {
		return nil, err
	}

	config := new(Config)
	err = envconfig.Process("VENDOR_MACHINE", config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
