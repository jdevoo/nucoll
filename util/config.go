package util

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// TwitterConfig stores oauth2 client credential grants
type TwitterConfig struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}

//type PinterestConfig struct {
//	PinterestID string `json:"clientId"`
//	PinterestSecret string `json:"clientSecret"`
//}

// NucollConfig holds access details to all supported SNSes
type NucollConfig struct {
	TwitterConfig
	//	PinterestConfig
}

// ReadConfig from .nucoll in home directory
func ReadConfig() (*NucollConfig, error) {
	var config NucollConfig

	configDir, err := DotNucollPath()
	if err != nil {
		return nil, err
	}
	file, err := ioutil.ReadFile(filepath.Join(configDir, "."+filepath.Base(os.Args[0])))
	if err != nil {
		return &NucollConfig{}, nil
	}

	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// WriteConfig to .nucoll file in home directory
func WriteConfig(config *NucollConfig) error {
	configDir, err := DotNucollPath()
	if err != nil {
		return err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(configDir, "."+filepath.Base(os.Args[0])), configJSON, 0600)
	if err != nil {
		return err
	}
	log.Printf("credentials stored in %s\n", filepath.Join(configDir, "."+filepath.Base(os.Args[0])))

	return nil
}
