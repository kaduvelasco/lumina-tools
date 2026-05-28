# Go Development — Lumina Standard

Go development standards for the Lumina ecosystem.
Focused on idiomatic code, robust error handling, and safe concurrency.
Use for CLI tools, backend services, and system utilities.

---

## Language

| Context | Language |
|---|---|
| Responses to the user | Brazilian Portuguese (pt-BR) |
| Code comments | English |

---

## Project Structure

Standard layout for Go projects:

```text
cmd/
  myapp/
    main.go           # Entry point — calls app.Run()
internal/
  app/                # Wiring: arg dispatch and top-level coordination
  config/             # Configuration load/save
  executor/           # Single point of privilege escalation (sudo)
  domain/             # One package per domain area
pkg/                  # Public packages safe to import externally
assets/               # Embedded files (go:embed)
go.mod
go.sum
Makefile
```

**Key invariants:**
- `main.go` does nothing except parse args and call into `internal/app`.
- No domain package calls `sudo` directly — all privileged commands go through `executor`.
- Embedded assets require `all:` prefix when the tree contains dot-files.

---

## Golden Rules

**1. Make the zero value useful.**

```go
// Good: ready to use without initialization
var buf bytes.Buffer
buf.WriteString("hello")

// Bad: nil map panics
type Bad struct{ counts map[string]int }
```

**2. Accept interfaces, return structs.**

```go
// Good
func Process(r io.Reader) (*Result, error) { ... }

// Bad: hides implementation, forces callers to use interface
func Process(r io.Reader) (io.Reader, error) { ... }
```

**3. Return early — keep the happy path unindented.**

```go
// Good
if err != nil {
    return fmt.Errorf("load config: %w", err)
}
// continue happy path...

// Bad: deeply nested happy path
if err == nil {
    // ... 30 lines of logic
}
```

**4. Never ignore errors.**

```go
// Bad
result, _ := doSomething()

// Good
result, err := doSomething()
if err != nil {
    return err
}
```

**5. Avoid package-level mutable state.**

```go
// Bad: global mutable DB
var db *sql.DB

// Good: inject dependencies
type Server struct{ db *sql.DB }
func NewServer(db *sql.DB) *Server { return &Server{db: db} }
```

---

## Error Handling

### Wrap errors with context

```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("load config %s: %w", path, err)
    }
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse config %s: %w", path, err)
    }
    return &cfg, nil
}
```

### Sentinel errors and custom types

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
)

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}
```

### Inspect errors with errors.Is / errors.As

```go
if errors.Is(err, sql.ErrNoRows) {
    return nil, ErrNotFound
}

var ve *ValidationError
if errors.As(err, &ve) {
    fmt.Fprintf(stderr, "field %s: %s\n", ve.Field, ve.Message)
}
```

---

## Concurrency

### Create contexts with timeout or cancellation

Always derive contexts from a parent — never use `context.Background()` deep in a call chain.

```go
// Timeout — cancels automatically after the deadline
ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
defer cancel()

// Cancellation — caller controls when to stop
ctx, cancel := context.WithCancel(parentCtx)
defer cancel()
```

`defer cancel()` is always required — even when the context times out on its own, calling cancel releases the associated resources immediately.

### Always pass context as the first parameter

```go
func Fetch(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetch %s: %w", url, err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("fetch %s: unexpected status %d", url, resp.StatusCode)
    }
    return io.ReadAll(resp.Body)
}
```

### Worker pool

```go
func WorkerPool(ctx context.Context, jobs <-chan Job, results chan<- Result, n int) {
    var wg sync.WaitGroup
    for range n {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- process(ctx, job)
            }
        }()
    }
    wg.Wait()
    close(results)
}
```

### Coordinated goroutines with errgroup

```go
import "golang.org/x/sync/errgroup"

func FetchAll(ctx context.Context, urls []string) ([][]byte, error) {
    g, ctx := errgroup.WithContext(ctx)
    results := make([][]byte, len(urls))
    for i, url := range urls {
        g.Go(func() error {
            data, err := Fetch(ctx, url)
            if err != nil {
                return err
            }
            results[i] = data
            return nil
        })
    }
    return results, g.Wait()
}
```

### Avoid goroutine leaks

```go
// Bad: blocks forever if no receiver
func leaky(url string) <-chan []byte {
    ch := make(chan []byte)
    go func() { ch <- fetch(url) }()
    return ch
}

// Good: buffered channel + context cancellation
func safe(ctx context.Context, url string) <-chan []byte {
    ch := make(chan []byte, 1)
    go func() {
        data, err := Fetch(ctx, url)
        if err != nil {
            return
        }
        select {
        case ch <- data:
        case <-ctx.Done():
        }
    }()
    return ch
}
```

---

## Interface Design

### Keep interfaces small

```go
// Good: single-method interfaces compose well
type Storer interface{ Store(key string, val []byte) error }
type Loader interface{ Load(key string) ([]byte, error) }
type Cache interface {
    Storer
    Loader
}
```

### Define interfaces in the consumer package

```go
// In the service package — not in the storage package
package service

type UserStore interface {
    GetUser(id string) (*User, error)
    SaveUser(user *User) error
}

type Service struct{ store UserStore }
```

### Functional options for configurable types

```go
type Server struct {
    addr    string
    timeout time.Duration
}

type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) { s.timeout = d }
}

func NewServer(addr string, opts ...Option) *Server {
    s := &Server{addr: addr, timeout: 30 * time.Second}
    for _, o := range opts {
        o(s)
    }
    return s
}
```

---

## Performance

### Preallocate slices when size is known

```go
// Bad: grows slice multiple times
var results []Result
for _, item := range items {
    results = append(results, process(item))
}

// Good: single allocation
results := make([]Result, 0, len(items))
for _, item := range items {
    results = append(results, process(item))
}
```

### Use strings.Builder for concatenation in loops

```go
// Bad: O(n²) allocations
var s string
for _, p := range parts { s += p }

// Good
var sb strings.Builder
for _, p := range parts { sb.WriteString(p) }
return sb.String()
```

### Reuse allocations with sync.Pool

```go
var pool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func handle(data []byte) string {
    buf := pool.Get().(*bytes.Buffer)
    defer func() { buf.Reset(); pool.Put(buf) }()
    buf.Write(data)
    return buf.String() // String copies the bytes — safe to return after defer returns buf to pool
}
```

Never return `buf.Bytes()` — the slice points into the buffer's internal array, which may be overwritten once the buffer is returned to the pool. Use `buf.String()` or copy explicitly.

---

## Testing

### Table-driven tests

Use `t.Run` with a slice of test cases. Keep the struct definition inline.

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 1, 2, 3},
        {"zero", 0, 0, 0},
        {"negative", -1, -2, -3},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Add(tt.a, tt.b); got != tt.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

Use `t.Helper()` in shared assertion helpers so failure lines point to the caller, not the helper:

```go
func assertEqual[T comparable](t *testing.T, got, want T) {
    t.Helper()
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

---

## Anti-Patterns

| Anti-pattern | Why |
|---|---|
| Naked returns in long functions | Impossible to tell what is being returned |
| `panic` for control flow | Crashes the program; use errors instead |
| Context stored in a struct field | Context belongs as a function parameter |
| Mixed value and pointer receivers | Inconsistent; pick one per type |
| Separate declaration and immediate assignment without reason | Use `:=` directly unless preserving exit code |
| Importing a package only for side effects without a comment | Silent `init()` calls are hard to trace |

---

## Quality

```bash
go build ./...
go test -race ./...
go vet ./...
golangci-lint run
gofmt -w .
go mod tidy
```

- **gofmt** — non-negotiable; run before every commit.
- **go vet** — catches correctness issues gofmt does not.
- **golangci-lint** — enable at minimum: `errcheck`, `govet`, `staticcheck`, `unused`, `gofmt`.
- **-race** — always run the test suite with the race detector.
