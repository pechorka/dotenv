# dotenv

Tiny, no-frills `.env` loader for Go.
It reads simple `KEY=VALUE` pairs from one or more paths and exports them to the process environment.
Later paths override earlier ones. It purposely avoids clever features to stay predictable.

## Philosophy

- Keep it simple: parse `KEY=VALUE`, ignore blanks and `#` comments, trim optional single/double quotes.
- Be explicit: you choose the paths to read; directories imply `path/.env`.
- No kitchen sink: no interpolation/expansion, no type casting, no surprises.

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/pechorka/dotenv"
)

func main() {
    // Loads from current dir's .env by default.
    if err := dotenv.Load(); err != nil {
        log.Fatal(err)
    }

    fmt.Println("HELLO=", os.Getenv("HELLO"))
}
```

See GoDoc for details, options, and additional examples.

## License

MIT

