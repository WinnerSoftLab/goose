package goose

import (
	"context"
	"database/sql"
)

// Redo rolls back the most recently applied migration, then runs it again.
func Redo(db *sql.DB, dir string, opts ...OptionsFunc) error {
	return RedoCtx(context.Background(), db, dir, opts...)
}

func RedoCtx(ctx context.Context, db *sql.DB, dir string, opts ...OptionsFunc) error {
	option := &options{}
	for _, f := range opts {
		f(option)
	}
	migrations, err := CollectMigrations(dir, minVersion, maxVersion)
	if err != nil {
		return err
	}
	var (
		currentVersion int64
	)
	if option.noVersioning {
		if len(migrations) == 0 {
			return nil
		}
		currentVersion = migrations[len(migrations)-1].Version
	} else {
		if currentVersion, err = GetDBVersionCtx(ctx, db); err != nil {
			return err
		}
	}

	current, err := migrations.Current(currentVersion)
	if err != nil {
		return err
	}
	current.noVersioning = option.noVersioning

	if err := current.DownCtx(ctx, db); err != nil {
		return err
	}
	if err := current.UpCtx(ctx, db); err != nil {
		return err
	}
	return nil
}
