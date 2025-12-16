## Architecture

TD: remove academisms/fluff.

- ./cmd - entrance points to `server` and `god` (independent user management cli).

- ./data/data.db - SQlite db, single file, should be easier to vps and backup.

- ./db - mostly sqlc generated, except store.go and migrations.go, ask AI.

- ./internal

  - config/config.go - default params, validation, fatal failures.

  - ./handlers

    - register, login, profile, logout.

    - domain_helpers.go related to db, sessions, not json/http stuff.

  - httpx - x for extras, json/http helpers used by both: handlers and guards.

  - protect - middleware: (i) guards which are easily called by any handler in any sequence, and (ii) domain-level guards such as max account per ip limiter which is not so easy to abstract and chain due to db transactions.

  - server/routes.go - this is where all the parameters come from main.go and config/config.go and middleware is assembled, and then passed to each handler.

  - config.toml - all the parameters whose default values are in config/config.go via config structs duplicating the main ones (simple Go, no builders, no Rob Pike's option structs).

These rules eliminate 80 percent of the mess:

- A guard (protector's common type/interface) is never nil, disabled means inactive, not absent.

- Defaults + validation + fatal errors live in config, no paranoid checks elsewhere.

- Handlers never mess with a guard, they only run a fully formed guard or a sequence of them, defined and instantiated for a specific handler inside ./internal/server/routes.go.

- A guard checks if everything is alright and returns true, or writes an HTTP response and returns false. There is no nil, error handling, panic, and exit maze.

- PoW is headers-only early return, no fallbacks to json body, no need to read it.

## More about Middleware (Guards)

This is the code which runs inside a handler before business logic. It is **stateless, synchronous, and in-memory**. To chain/execute them in sequence we put them under a common type `protect.Guard`:

```go
type Guard interface {
    Check(w http.ResponseWriter, r *http.Request) bool
}
```

This lives inside ./internal/protect to break a cycle between ./internal/server/routes.go and ./internal/handlers.

The rest is just Go code. Inside a handler an active guard will emit an HTTP response and return false. The handler exits before business logic via return:

```go
for _, g := range h.Guards {
    if !g.Check(w, r) {
        return
    }
}
```

See ./internal/handlers/register.go as an example.

You will find the following tested guards inside ./internal/protect:

- ip_rate_guard.go – HTTP request rate per ip limiting.

- body_size_guard.go – request body size cap.

- pow_guard.go – optional [Proof-of-Work](docs/proof_of_work.md) challenge.

./internal/handlers/register.go also includes the Maximal Accounts per IP limiter,

```go
limiter := protect.NewAccountPerIPLimiter(
		h.AccountPerIPLimiter,
		txStore.CountUsersByIP,
	)
```

It needs to access db via txStore, which runs a transaction controlled by the api/register handler. Treat those stateful guards as a handler's business logic, more details in [Accounts per IP](docs/accounts_per_ip.md) and ./internal/handlers/register.go.

This way we cover (more or less) complex Ruby Rack machinery with basic typed Go code.
