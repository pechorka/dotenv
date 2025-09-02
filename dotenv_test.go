package dotenv

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func Test_load(t *testing.T) {
	t.Run("loads from directory path", func(t *testing.T) {
		fs := fstest.MapFS{
			"dir/.env": &fstest.MapFile{Data: []byte(`FOO=bar
BAZ=qux
`)},
		}
		os.Unsetenv("FOO")
		os.Unsetenv("BAZ")

		err := Load(WithPaths("dir"), WithFs(fs))
		assertNoError(t, err)
		assertEqual(t, os.Getenv("FOO"), "bar")
		assertEqual(t, os.Getenv("BAZ"), "qux")
	})

	t.Run("skips missing .env in directory (logs)", func(t *testing.T) {
		fs := fstest.MapFS{}
		err := Load(WithPaths("missing"), WithFs(fs))
		assertNoError(t, err)
	})

	t.Run("later paths override earlier", func(t *testing.T) {
		fs := fstest.MapFS{
			"a/.env": &fstest.MapFile{Data: []byte(`KEY=1
`)},
			"b/.env": &fstest.MapFile{Data: []byte(`KEY=2
`)},
		}
		os.Unsetenv("KEY")
		err := Load(WithPaths("a", "b"), WithFs(fs))
		assertNoError(t, err)
		assertEqual(t, os.Getenv("KEY"), "2")
	})

	t.Run("handles comments and quotes", func(t *testing.T) {
		fs := fstest.MapFS{
			"dir/.env": &fstest.MapFile{Data: []byte(`# comment
HELLO=world
NAME="John Doe"
TITLE=' Sr Dev '
   # another comment
`)},
		}
		os.Unsetenv("HELLO")
		os.Unsetenv("NAME")
		os.Unsetenv("TITLE")

		err := Load(WithPaths("dir"), WithFs(fs))
		assertNoError(t, err)
		assertEqual(t, os.Getenv("HELLO"), "world")
		assertEqual(t, os.Getenv("NAME"), "John Doe")
		assertEqual(t, os.Getenv("TITLE"), " Sr Dev ")
	})

	t.Run("loads from explicit file path", func(t *testing.T) {
		fs := fstest.MapFS{
			"p/.env": &fstest.MapFile{Data: []byte(`X=1
`)},
		}
		os.Unsetenv("X")
		err := Load(WithPaths("p/.env"), WithFs(fs))
		assertNoError(t, err)
		assertEqual(t, os.Getenv("X"), "1")
	})

	t.Run("logger reports joins and not-found", func(t *testing.T) {
		fs := fstest.MapFS{
			// Only second directory has dotenv
			"a/.env": &fstest.MapFile{Data: []byte(`A=1
`)},
		}
		lg := &testLogger{}

		os.Unsetenv("A")
		err := Load(WithPaths("missing", "a", "b"), WithFs(fs), WithLogger(lg))
		assertNoError(t, err)
		assertEqual(t, os.Getenv("A"), "1")

		out := lg.String()
		if !(strings.Contains(out, "path not found") || strings.Contains(out, "directory detected; joining dotenv")) {
			t.Fatalf("expected logs about path not found or join; got: %q", out)
		}
	})
}

type testLogger struct{ bytes.Buffer }

func (l *testLogger) log(msg string, args ...any) {
	l.WriteString(msg)
	if len(args) > 0 {
		l.WriteString(" ")
		l.WriteString(fmt.Sprint(args...))
	}
	l.WriteString("\n")
}
func (l *testLogger) Info(msg string, args ...any) { l.log(msg, args...) }
func (l *testLogger) Warn(msg string, args ...any) { l.log(msg, args...) }

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("got=%v, want=%v", got, want)
	}
}
