package dotenv_test

import (
	"fmt"
	"os"
	"testing/fstest"

	"github.com/pechorka/dotenv"
)

// Example shows the simplest way to load variables from a directory.
func Example() {
	fs := fstest.MapFS{
		"app/.env": &fstest.MapFile{Data: []byte("GREETING=hello\n")},
	}

	_ = os.Unsetenv("GREETING")
	_ = dotenv.Load(
		dotenv.WithPaths("app"), // directory -> "app/.env"
		dotenv.WithFs(fs),
	)

	fmt.Println(os.Getenv("GREETING"))
	// Output: hello
}

// Example_withOverride demonstrates that later paths override earlier ones.
func Example_withOverride() {
	fs := fstest.MapFS{
		"a/.env": &fstest.MapFile{Data: []byte("KEY=one\n")},
		"b/.env": &fstest.MapFile{Data: []byte("KEY=two\n")},
	}

	_ = os.Unsetenv("KEY")
	_ = dotenv.Load(
		dotenv.WithPaths("a", "b"), // values from b override a
		dotenv.WithFs(fs),
	)

	fmt.Println(os.Getenv("KEY"))
	// Output: two
}
