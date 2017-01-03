package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/octavore/naga/service"
)

// ConfigEnv specifies the env variable which contains the path to a config file.
// To use a different env variable, change this in an init block in your app.
var ConfigEnv = "CONFIG_FILE"

// Module for config package.
type Module struct {
	Byte       []byte
	Env        map[string]string
	ConfigPath string

	// TestConfigPath can be to load a specific config file in tests.
	// Alternatively, you can set Byte directly to the desired config
	// file contents.
	TestConfigPath string
}

// Init implements the module interface method
func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		switch {
		case m.ConfigPath != "":
		// do nothing
		case os.Getenv(ConfigEnv) != "":
			m.ConfigPath = os.Getenv(ConfigEnv)
		default:
			m.ConfigPath = "config.json"
		}

		err := m.LoadConfig(m.ConfigPath)
		// we only return an error if production, which
		// allows this to pass for env=test
		if err != nil {
			return errIfProduction(c, err)
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
			panic(err)
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
	m.Byte, err = ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return nil
}

// ReadConfig json-decodes the config file bytes into i, which should be a pointer
// to a struct.
func (m *Module) ReadConfig(i interface{}) error {
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
	if c.Env().IsProduction() {
		return err
	}
	service.BootPrintln(err)
	return nil
}
