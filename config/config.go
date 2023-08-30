package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/cpendery/wock/hosts"
)

const wockConfigName = ".wock.json"

type config struct {
	Aliases []alias `json:"aliases,omitempty"`
}

type alias struct {
	Alias     string `json:"alias"`
	Host      string `json:"host"`
	Directory string `json:"directory"`
}

var WockConfig config

func init() {
	if err := LoadConfig(); err != nil {
		slog.Error("invalid wock config", slog.String("error", err.Error()))
	}
}

func LoadConfig() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to load working directory: %w", err)
	}
	configLocation := filepath.Join(wd, wockConfigName)
	if _, err := os.Stat(configLocation); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	configFile, err := os.Open(configLocation)
	if err != nil {
		return fmt.Errorf("unable to open config file: %w", err)
	}
	configData, err := io.ReadAll(configFile)
	if err != nil {
		return fmt.Errorf("unable to read config file: %w", err)
	}
	if err := json.Unmarshal(configData, &WockConfig); err != nil {
		return fmt.Errorf("unable to unmarshal config: %w", err)
	}
	return validateConfig()
}

func IsValidAlias(alias string) bool {
	for _, aliasItem := range WockConfig.Aliases {
		if strings.EqualFold(aliasItem.Alias, alias) {
			return true
		}
	}
	return false
}

func IsValidDirectory(userInput string) (*string, error) {
	var dir string
	if filepath.IsAbs(userInput) {
		dir = userInput
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("unable to check working directory: %w", err)
		}
		dir = filepath.Join(wd, userInput)
	}

	fileinfo, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("unable to serve %s as it doesn't exist", dir)
	} else if err != nil {
		return nil, fmt.Errorf("unable to validate directory exists: %w", err)
	} else if !fileinfo.IsDir() {
		return nil, fmt.Errorf("unable to serve %s as it isn't a directory", dir)
	} else {
		return &dir, nil
	}
}

func GetAlias(alias string) (string, string) {
	for _, aliasItem := range WockConfig.Aliases {
		if strings.EqualFold(aliasItem.Alias, alias) {
			return aliasItem.Host, aliasItem.Directory
		}
	}
	log.Fatalln("invalid alias ", alias)
	return "", ""
}

func validateConfig() error {
	for _, aliasItem := range WockConfig.Aliases {
		if _, err := IsValidDirectory(aliasItem.Directory); err != nil {
			return err
		} else if !hosts.IsValidHostname(aliasItem.Host) {
			return fmt.Errorf("invalid hostname '%s'", aliasItem.Host)
		}
	}
	return nil
}
