// Package dotenv provides a tiny, predictable loader for .env files.
//
// It reads simple KEY=VALUE lines, ignoring blank lines and lines starting
// with '#'. Optional single or double quotes around values are trimmed.
// When multiple paths are provided, later ones override earlier ones. If a
// provided path is a directory, ".env" is joined to it.
package dotenv

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"strings"
)

type Options struct {
	Paths  []string
	RootFs fs.FS
	Logger Logger
}

type Option func(*Options)

// WithPaths sets candidate paths (files or directories) to read, in order.
// If a path is a directory, ".env" is joined. When multiple paths are
// provided, later ones override earlier ones.
func WithPaths(paths ...string) Option {
	return func(o *Options) {
		o.Paths = paths
	}
}

// WithFs sets the filesystem root used to open paths. When not provided, the
// current directory is used via os.OpenRoot(".").FS()
func WithFs(rootFs fs.FS) Option {
	return func(o *Options) {
		o.RootFs = rootFs
	}
}

// Logger is a minimal logger used by Load for informational and warning
// messages. Bring your own implementation; a no-op logger is used by default.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
}

var _ Logger = &slog.Logger{}

// WithLogger sets a custom logger implementation used during loading.
func WithLogger(l Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

func validateOptions(opts Options) error {
	if len(opts.Paths) == 0 {
		return fmt.Errorf("should provide at least a single path")
	}
	if opts.RootFs == nil {
		return fmt.Errorf("should provide root fs")
	}
	if opts.Logger == nil {
		return fmt.Errorf("logger should be provided")
	}
	return nil
}

// Load reads .env entries from the configured paths and exports them via
// os.Setenv. It is safe to call multiple times; later files override
// earlier ones according to the provided paths.
// Not found paths will be ignored and logged.
func Load(userOptions ...Option) error {
	opts := Options{
		Paths:  []string{"."},
		Logger: nopLogger{},
	}
	for _, userOption := range userOptions {
		userOption(&opts)
	}

	if opts.RootFs == nil {
		root, err := os.OpenRoot(".")
		if err != nil {
			return fmt.Errorf("failed to create fs.FS from current directory: %w", err)
		}
		opts.RootFs = root.FS()
	}

	if err := validateOptions(opts); err != nil {
		return fmt.Errorf("can export .env file with these options: %w", err)
	}

	return load(opts)
}

func load(opts Options) error {
	for _, p := range opts.Paths {
		info, err := fs.Stat(opts.RootFs, p)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				opts.Logger.Warn("path not found", "path", p)
				continue
			}
			return fmt.Errorf("stat %s: %w", p, err)
		}

		var envPath string
		if info.IsDir() {
			envPath = path.Join(p, ".env")
			opts.Logger.Info("directory detected; joining dotenv", "path", p, "dotenv", envPath)
		} else {
			envPath = p
		}

		if _, err := fs.Stat(opts.RootFs, envPath); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				opts.Logger.Warn("dotenv not found", "path", envPath)
				continue
			}
			return fmt.Errorf("stat %s: %w", envPath, err)
		}

		err = processFile(opts.RootFs, envPath, func(f fs.File) error {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				eq := strings.IndexByte(line, '=')
				if eq <= 0 {
					continue
				}
				key := strings.TrimSpace(line[:eq])
				val := strings.TrimSpace(line[eq+1:])

				if len(val) >= 2 {
					if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
						val = val[1 : len(val)-1]
					}
				}

				if key != "" {
					if err := os.Setenv(key, val); err != nil {
						return fmt.Errorf("setenv %s: %w", key, err)
					}
				}
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("read %s: %w", envPath, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func processFile(rootFs fs.FS, path string, processorFn func(f fs.File) error) error {
	f, err := rootFs.Open(path)
	if err != nil {
		return fmt.Errorf("open %q: %w", path, err)
	}

	err = processorFn(f)
	closeErr := f.Close()
	if err != nil {
		if closeErr != nil {
			err = errors.Join(err, fmt.Errorf("also failed to close file %q: %w", path, err))
		}
		return err
	}

	return nil
}

type nopLogger struct{}

func (nopLogger) Info(string, ...any) {}
func (nopLogger) Warn(string, ...any) {}
