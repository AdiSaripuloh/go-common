package config

import (
	"strings"

	"github.com/spf13/viper"
)

// BindFromFile load config from filename then assign to destination
func BindFromFile(dest any, filename string, paths ...string) error {
	v := viper.New()

	v.SetConfigType("yaml")
	v.SetConfigName(filename)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	for _, path := range paths {
		v.AddConfigPath(path)
	}

	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	err = v.Unmarshal(dest)
	if err != nil {
		return err
	}

	return nil
}

// BindFromConsul load config from remote consul then assign to destination
func BindFromConsul(dest any, endPoint, path string) error {
	v := viper.New()

	v.SetConfigType("yaml")
	err := v.AddRemoteProvider("consul", endPoint, path)
	if err != nil {

	}
	err = v.ReadRemoteConfig()
	if err != nil {
		return err
	}

	err = v.Unmarshal(dest)
	if err != nil {
		return err
	}

	return nil
}
