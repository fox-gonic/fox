package configurations

import (
	"os"

	"github.com/spf13/viper"
)

// Configurations instance type
type Configurations struct {
	*viper.Viper
}

// New configurations file
func New(path string) (*Configurations, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	v := viper.New()
	v.SetConfigType("yaml")

	if err = v.ReadConfig(file); err != nil {
		return nil, err
	}

	configurations := &Configurations{
		Viper: v,
	}

	return configurations, nil
}

// Parse config from file path
func Parse(path string, conf interface{}) error {

	configurations, err := New(path)
	if err != nil {
		return nil
	}

	return configurations.Unmarshal(conf)
}
