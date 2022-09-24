package configurations

import (
	"os"

	"github.com/spf13/viper"
)

// Load the configuration file from the specified path
func Load[T any](path string, config *T) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	v := viper.New()
	v.SetConfigType("yaml")

	if err = v.ReadConfig(file); err != nil {
		return err
	}

	if err = v.Unmarshal(config); err != nil {
		return err
	}

	return nil
}
