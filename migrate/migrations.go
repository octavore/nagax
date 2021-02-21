package migrate

import (
	"io/fs"
	"net/http"
	"path/filepath"

	migrate "github.com/rubenv/sql-migrate"
)

type (
	assetFunc    func(path string) ([]byte, error)
	assetDirFunc func(path string) ([]string, error)
)

// SetMigrationSource sets the migration source, for compatibility with
// embedded file assets.
func (m *Module) SetMigrationSource(asset assetFunc, assetDir assetDirFunc, dir string) {
	m.migrationSource = &migrate.AssetMigrationSource{
		Asset:    asset,
		AssetDir: assetDir,
		Dir:      dir,
	}
}

// SetMigrationSource sets the migration source, for compatibility with
// embedded file assets.
func (m *Module) SetMigrationFS(f fs.FS) {
	m.migrationSource = &migrate.HttpFileSystemMigrationSource{
		FileSystem: http.FS(f),
	}
}

// getMigrationSource returns the m.migrationSource if set, otherwise
// it defaults by reading from the MigrationsDir specified in
func (m *Module) getMigrationSource() (migrate.MigrationSource, error) {
	if m.migrationSource != nil {
		return m.migrationSource, nil
	}
	configPath, err := filepath.Abs(m.Config.ConfigPath)
	if err != nil {
		return nil, err
	}
	migrationPath := filepath.Join(filepath.Dir(configPath), m.config.MigrationsDir)
	return migrate.FileMigrationSource{Dir: migrationPath}, nil
}
