package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"

	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/util/errors"
)

// ConfigEnv specifies the env variable which contains the path to a config file.
// To use a different env variable, change this in an init block in your app.
var ConfigEnv = "CONFIG_FILE"

// Module for config package.
type Module struct {
	Logger *logger.Module

	Byte       []byte
	Env        map[string]string
	ConfigPath string

	// TestConfigPath can be to load a specific config file in tests.
	// Alternatively, you can set Byte directly to the desired config
	// file contents.
	TestConfigPath string

	configDefs []reflect.Type

	DisableChdir bool
}

// Init implements the module interface method
func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.configDefs = []reflect.Type{}
		switch {
		case m.ConfigPath != "":
		// do nothing
		case os.Getenv(ConfigEnv) != "":
			m.ConfigPath = os.Getenv(ConfigEnv)
		default:
			m.ConfigPath = "config.json"
		}

		configAbs, _ := filepath.Abs(m.ConfigPath)
		m.Logger.Infof("config: %s", configAbs)

		// we only return an error if production, which
		// allows this to pass for env=test
		err := m.LoadConfig(m.ConfigPath)
		if err != nil {
			return errIfProduction(c, err)
		}

		if !m.DisableChdir {
			return m.chdirConfigPath()
		}
		return nil
	}
	c.SetupTest = func() {
		// this is a hack to have a safe default value
		// for m.Byte if we haven't loaded any config
		if len(m.Byte) == 0 {
			m.Byte = []byte(`{}`)
		}
		if m.TestConfigPath == "" {
			return
		}
		err := m.LoadConfig(m.TestConfigPath)
		if err != nil {
			c.Fatal(err)
		}
	}
}

// LoadConfig loads the config json file from the given path
func (m *Module) LoadConfig(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			m.Byte = []byte(`{}`)
			return nil
		}
		return err
	}
	m.Byte, err = os.ReadFile(path)
	if err != nil {
		return err
	}
	return nil
}

// ReadConfig json-decodes the config file bytes into i, which should be a pointer
// to a struct.
func (m *Module) ReadConfig(i interface{}) error {
	m.configDefs = append(m.configDefs, reflect.TypeOf(i))
	return json.Unmarshal(m.Byte, i)
}

// Getenv reads and caches env variable
func (m *Module) Getenv(key string) string {
	if _, ok := m.Env[key]; !ok {
		m.Env[key] = os.Getenv(key)
	}
	return m.Env[key]
}

func errIfProduction(c *service.Config, err error) error {
	if c.Env().IsHosted() {
		return err
	}
	service.BootPrintln(err)
	return nil
}

func (m *Module) ResourcePath(resource string) (string, error) {
	configDir, err := filepath.Abs(filepath.Dir(m.ConfigPath))
	if err != nil {
		return "", errors.Wrap(err)
	}
	return filepath.Join(configDir, resource), nil
}

func (m *Module) Resource(resource string) ([]byte, error) {
	p, err := m.ResourcePath(resource)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return data, nil
}

func (m *Module) chdirConfigPath() error {
	absConfigPath, err := filepath.Abs(m.ConfigPath)
	if err != nil {
		return err
	}

	configDir := filepath.Dir(absConfigPath)
	err = os.Chdir(configDir)
	if err != nil {
		return err
	}
	m.Logger.Infof("config: chdir %s", configDir)
	return nil
}
