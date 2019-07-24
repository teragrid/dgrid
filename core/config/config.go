package config

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	cmn "github.com/teragrid/dgrid/pkg/common"
)

const (
	// DefaultDirPerm is the default permissions used when creating directories.
	DefaultDirPerm = 0700

	// ConfigHome is the main location to look for configuration
	ConfigHome = ".dgrid"
)

// Config encapsulates config (league or resource) tree
type Config interface {
	// Default() returns the default config for the initialization
	Default() *Config

	// Validate performs basic validation (checking param bounds, etc.) and
	// returns an error if any check fails.
	Validate() error
}

// DefaultConfig returns the default config of a resource config
func DefaultConfig(config Config) Config {
	return config.Default()
}

// Manager provides access to the resource config
type Manager interface {
	// GetChannelConfig defines methods that are related to channel configuration
	GetLeagueConfig(leagueID string) Config
	// UpdateLeagueConfig attemps to submit a update propose in a given league
	UpdateLeagueConfig(leagueID string, config *Config) (*Config, error)
}

// DefaultConfigTemplate loads default template for the Base and Regular
// by the init command
func DefaultConfigTemplate(configTemplateFile string) *template.Template {
	var err error
	if configTemplate, err = template.New("configFileTemplate").Parse(
		configTemplateFile); err != nil {
		panic(err)
	}
	return configTemplate
}

// EnsureRoot creates the root, config, and data directories if they don't exist,
// and panics if it fails.
func EnsureRoot(rootDir string, leagueID string) {
	if err := cmn.EnsureDir(rootDir, DefaultDirPerm); err != nil {
		cmn.PanicSanity(err.Error())
	}
	if err := cmn.EnsureDir(filepath.Join(rootDir, leagueID, defaultConfigDir),
		DefaultDirPerm); err != nil {
		cmn.PanicSanity(err.Error())
	}
	if err := cmn.EnsureDir(filepath.Join(rootDir, leagueID, defaultDataDir),
		DefaultDirPerm); err != nil {
		cmn.PanicSanity(err.Error())
	}

	configFilePath := filepath.Join(rootDir, leagueID,
		defaultConfigDir, defaultConfigFile)

	// Write default config file if missing.
	// 1. The default config file for the Base
	if !cmn.FileExists(configFilePath) {
		writeDefaultConfigFile(configFilePath, leagueID)
	}
	// 2. The default config file for the first regular league
}

// writeDefaultConfigFile should probably be called by cmd/tgrid/commands/init.go
// alongside the writing of the genesis.json and consensus.json
func writeDefaultConfigFile(configFilePath string, leagueType LeagueType) {
	cfg = NewLeagueConfig(leagueType)
	configTemplate = LeagueConfigFileTemplate(leagueType)
	WriteConfigFile(configTemplate, configFilePath, cfg)
}

// WriteConfigFile renders config using the template and writes it to configFilePath.
func WriteConfigFile(configTemplate, configFilePath string, config *Config) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	cmn.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

// AddConfigPath adds a config path for searching configuration
func AddConfigPath(v *viper.Viper, p string) {
	if v != nil {
		v.AddConfigPath(p)
	} else {
		viper.AddConfigPath(p)
	}
}

// TranslatePath translates a relative path into a fully qualified path relative to the config
// file that specified it.  Absolute paths are passed unscathed.
func TranslatePath(base, p string) string {
	if filepath.IsAbs(p) {
		return p
	}

	return filepath.Join(base, p)
}

// TranslatePathInPlace translates a relative path into a fully qualified path in-place (updating the
// pointer) relative to the config file that specified it.  Absolute paths are
// passed unscathed.
func TranslatePathInPlace(base string, p *string) {
	*p = TranslatePath(base, *p)
}

// GetPath allows configuration strings that specify a (config-file) relative path
//
// For example: Assume our config is located in /etc/teragrid/dgrid/core.yaml with
// a key "consensus.configPath" = "consensus/config.yaml".
//
// This function will return:
//      GetPath("consensus.configPath") -> /etc/teragrid/dgrid/consensus/config.yaml
//
//----------------------------------------------------------------------------------
func GetPath(key string) string {
	p := viper.GetString(key)
	if p == "" {
		return ""
	}

	return TranslatePath(filepath.Dir(viper.ConfigFileUsed()), p)
}

// InitViper performs basic initialization of our viper-based configuration layer.
// Primary thrust is to establish the paths that should be consulted to find
// the configuration we need.  If v == nil, we will initialize the global
// Viper instance
//----------------------------------------------------------------------------------
func InitViper(v *viper.Viper, configName string) error {
	var altPath = os.Getenv("DGRID_CFG_PATH")
	if altPath != "" {
		// If the user has overridden the path with an envvar, its the only path
		// we will consider

		if !dirExists(altPath) {
			return fmt.Errorf("DGRID_CFG_PATH %s does not exist", altPath)
		}

		AddConfigPath(v, altPath)
	} else {
		// If we get here, we should use the default paths in priority order:
		//
		// *) CWD
		// *) /etc/teragrid/dgrid

		// CWD
		AddConfigPath(v, "./")

		// And finally, the official path
		if dirExists(ConfigHome) {
			AddConfigPath(v, ConfigHome)
		}
	}

	// Now set the configuration file.
	if v != nil {
		v.SetConfigName(configName)
	} else {
		viper.SetConfigName(configName)
	}

	return nil
}

// Rootify makes config creation independent of root dir
func Rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
