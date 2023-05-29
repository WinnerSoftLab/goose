package goose

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
)

// Deprecated: VERSION will no longer be supported in v4.
const VERSION = "v3.2.0"

var (
	minVersion      = int64(0)
	maxVersion      = int64((1 << 63) - 1)
	timestampFormat = "20060102150405"
	verbose         = false

	// base fs to lookup migrations
	baseFS             fs.FS = osFS{}
	ErrInvalidArgument       = fmt.Errorf("invalid argument, expected pairs separeded by equal sign 'key=value'")
)

// SetVerbose set the goose verbosity mode
func SetVerbose(v bool) {
	verbose = v
}

// SetBaseFS sets a base FS to discover migrations. It can be used with 'embed' package.
// Calling with 'nil' argument leads to default behaviour: discovering migrations from os filesystem.
// Note that modifying operations like Create will use os filesystem anyway.
func SetBaseFS(fsys fs.FS) {
	if fsys == nil {
		fsys = osFS{}
	}

	baseFS = fsys
}

// Run runs a goose command.
func Run(command string, db *sql.DB, dir string, args ...string) error {
	return RunCtx(context.Background(), command, db, dir, args...)
}

// RunWithOptions runs a goose command with options.
func RunWithOptions(command string, db *sql.DB, dir string, args []string, options ...OptionsFunc) error {
	return RunWithOptionsCtx(context.Background(), command, db, dir, args, options...)
}

// RunCtx runs a goose command.
func RunCtx(ctx context.Context, command string, db *sql.DB, dir string, args ...string) error {
	return run(ctx, command, db, dir, args)
}

// RunWithOptionsCtx runs a goose command with options.
func RunWithOptionsCtx(ctx context.Context, command string, db *sql.DB, dir string, args []string, options ...OptionsFunc) error {
	return run(ctx, command, db, dir, args, options...)
}

func run(ctx context.Context, command string, db *sql.DB, dir string, args []string, options ...OptionsFunc) error {
	switch command {
	case "up":
		if err := UpCtx(ctx, db, dir, options...); err != nil {
			return err
		}
	case "up-by-one":
		if err := UpByOneCtx(ctx, db, dir, options...); err != nil {
			return err
		}
	case "up-to":
		if len(args) == 0 {
			return fmt.Errorf("up-to must be of form: goose [OPTIONS] DRIVER DBSTRING up-to VERSION")
		}

		version, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("version must be a number (got '%s')", args[0])
		}
		if err := UpToCtx(ctx, db, dir, version, options...); err != nil {
			return err
		}
	case "create":
		if len(args) == 0 {
			return fmt.Errorf("create must be of form: goose [OPTIONS] DRIVER DBSTRING create NAME [go|sql]")
		}

		migrationType := "go"
		if len(args) >= 2 {
			migrationType = args[1]
		}
		params, err := makeParams(args)
		if err != nil {
			return err
		}
		if err = Create(db, dir, args[0], migrationType, params, options...); err != nil {
			return err
		}
	case "down":
		if err := DownCtx(ctx, db, dir, options...); err != nil {
			return err
		}
	case "down-to":
		if len(args) == 0 {
			return fmt.Errorf("down-to must be of form: goose [OPTIONS] DRIVER DBSTRING down-to VERSION")
		}

		version, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("version must be a number (got '%s')", args[0])
		}
		if err := DownToCtx(ctx, db, dir, version, options...); err != nil {
			return err
		}
	case "fix":
		if err := Fix(dir); err != nil {
			return err
		}
	case "redo":
		if err := RedoCtx(ctx, db, dir, options...); err != nil {
			return err
		}
	case "reset":
		if err := ResetCtx(ctx, db, dir, options...); err != nil {
			return err
		}
	case "status":
		if err := StatusCtx(ctx, db, dir, options...); err != nil {
			return err
		}
	case "version":
		if err := VersionCtx(ctx, db, dir, options...); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%q: no such command", command)
	}
	return nil
}

func makeParams(args []string) (map[string]string, error) {
	// forces specifying migration type to use key value pairs.
	// TBD: allow optional migration type
	if len(args) < 3 {
		return nil, nil
	}

	pairs := args[2:]

	result := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		values := strings.SplitN(pair, "=", 2)
		if len(values) != 2 {
			return nil, ErrInvalidArgument
		}
		result[values[0]] = values[1]
	}

	return result, nil
}
