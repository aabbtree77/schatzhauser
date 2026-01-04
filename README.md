<p align="center">
  <img src="docs/schatzhauser-banner.png" alt="Schatzhauser" style="width: 90%; height: auto;" />
</p>

<p align="center">
  <em>Authenticated bot-proof JSON API, but still KISS-compatible.</em>
</p>

This is a minimal Go JSON API server to test ideas to fight bots, in the context of an authenticated web app.
MIT licensed: self-educate, steal it, reuse it in your own conquest of the world. 

A simple CRUD app is a modern day QFT or string theory. Extremely overcrowded spaces, everybody knows everything, yet very little progress with auth, ddos, reliability, hosting, payments, "components". One ideal is PocketBase, but I want something smaller, more focused and reliable. Ultimately, a concrete running app serving people, not a framework with multiple connectors and adapters.

## API

```bash
## API Usage

### Register
curl -i -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"u1","password":"p1"}' \
  http://localhost:8080/api/register

### Login (save cookie)
curl -i -c cookiejar.txt \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"u1","password":"p1"}' \
  http://localhost:8080/api/login

### Profile (authenticated)
curl -i -b cookiejar.txt \
  http://localhost:8080/api/profile

### Logout
curl -i -b cookiejar.txt \
  -X POST \
  http://localhost:8080/api/logout

``` 

HTTP vs JSON can get confusing, but this is a classical cookie-based session auth HTTP server with JSON payloads inside HTTP request bodies. No JWTs and no SPA (4), but also not enough redirects or textbook HTTP (1). Also more HTTP-first than just JSON (so roughly 2, not enough focus on sending data to be 3):

1. HTTP + cookies + HTML

2. HTTP + cookies + JSON ← you are here

3. HTTP + cookies + HTML + JSON (hybrid)

4. HTTP + tokens + JSON + SPA (complexity cliff)

The client is programmatic, either direct curl, or ./tests, no frontend. There is also CLI to clean up DB.

Instruments against the bots:

- maximal request rate per IP (fixed window, in memory),

- maximal request body size, 

- [proof of work (PoW)](docs/proof_of_work.md).

- [maximal account number per IP](docs/accounts_per_ip.md) (persistent in SQLite).

Delve into ./tests and docs for the tricky bits. See [Architecture](docs/architecture.md) for more details.

## Setup/Workflow

Clone, cd, and run `make all` which should create two binaries inside ./bin: server and god.

First time (no DB):

```bash
mkdir data && touch data/data.db
sqlc generate
```

After modifying Go code:

```bash
sqlc generate
make
```

After adding a new migration file to db/migrations:

```bash
sqlc generate
make
./bin/server
```

./bin/server is a compiled .cmd/main.go which executes migrations via ./db/migrations.go. You must start/restart the server for migrations to take place.

Any change to DB takes place by creating a new migration file inside ./db/migrations.

The first one, 001_init.sql sets up the schema, so I keep the folder ./db/schema empty not to duplicate stuff. Once a migration takes place by starting ./bin/server, there is no way of modifying these files if you want to do things correctly with a running DB. There is no rolling back, deleting files, this is the SQL world.

Any modification takes place by adding a new migration which can add a variable or delete it via a new transaction which copies and recreates the whole table.

`sqlc generate` is a static tool to generate Go files once a migration file is added, and/or new queries are added inside ./db/queries.sql. It does not execute migrations, it only generates Go inside ./db.

Run cli tool `god` to manage users, SQLite allows that irrespectively whether server is running or not.

```bash
./bin/god user set --username alice --password passw0
./bin/server
time=2025-12-02T22:41:03.389+02:00 level=INFO msg="starting server" debug=true
time=2025-12-02T22:41:03.389+02:00 level=INFO msg="listening on :8080"
^Ctime=2025-12-02T22:41:40.035+02:00 level=INFO msg="shutting down"
```

Adjust config.toml as you wish, but the tests will barf about the right values. The Go code uses defaults where needed if the values are wrong or omitted inside config.toml.

## God Mode

Use god to create/update (set) users, adjust (set) roles, delete users. It is a minimal CLI app which opens the same SQLite database (DB) file directly. It uses the same schema and sqlc queries.

SQLite supports multiple processes safely (file locking handles it), with one caveat. If the server is actively writing at the same moment, SQLite may briefly lock the DB. In that case god will just get a transient error; rerun is fine. So one can run god while the server is up or down.

```bash
# create admins
./bin/god user set --username test_user0 --role admin --password hunter2
./bin/god user set --username salomeja --role admin --password neris

# create regular users
./bin/god user set --username test_user1 --password pass1
./bin/god user set --username test_user2 --password pass2

# get user profile
./bin/god user get salomeja

# list users
./bin/god users list

# create/update user fields including role promotion/demotion and password rotation
./bin/god user set --username test_user0 --role user --password hunter2
./bin/god user set --username test_user1 --role admin --ip 1.2.3.4
./bin/god user set --username test_user0 --password newpass

# delete single user
./bin/god user delete --username test_user1

# bulk delete by prefix
./bin/god users delete --prefix test_

# bulk delete by creation date
./bin/god users delete --created-between 2025-12-02 2025-12-05
```

## Tests

Start the server, open another terminal and run one of these:

```bash
go run ./tests/register
go run ./tests/login
go run ./tests/profile
go run ./tests/logout
go run ./tests/req_rate_per_ip
go run ./tests/account_rate_per_ip
go run ./tests/req_body_size
go run ./tests/pow_register
```

## Some References

[Kyle Conroy Gray: Introducing sqlc - Compile SQL queries to type-safe Go (2019)](https://conroy.org/introducing-sqlc)

sqlc is the way to deal with SQL in Go. No extra DSLs of ORMs, no SQL strings scattered in Go. Write SQL in SQL/AI, generate a succinct precise Go function per query with sqlc, and use that in Go. No magic objects, only plain functions.

[Mat Ryer: How I write HTTP services in Go after 13 years (2024) on HN](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/)

Two useful things to take away: (i) graceful shutdown with signal.NotifyContext, and (ii) reducing application startup time with sync.Once. The first one is in ./cmd/server/main.go.

[Rob Pike: Self-referential functions and the design of options (2014)](https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html)

Do "builder design pattern" instead if you must, at least it is some sort of a "standard". I prefer plain Go, see ./internal/config/config.go.

## Notes

I am building a web app, some sort of a 3rd-party-free bulletin board, and releasing its stages as separate modules. 

The code such as this one only serves two purposes: (i) to assess/monitor/landmark progress, and (ii) release something self-sufficient of use to others/later me. See docs where some pain point or design decisions are discussed.

There is a bewildering number of ways to implement web apps. See [RealWorld (Conduit)](https://github.com/gothinkster/realworld) repo for a huge number of medium.com clones. They are the most archetypical web apps with auth, users, posts, and comments. Most of them are extremely over architected (go kit, aws dynamoDB, hexagonal architecture, OpenAPI, React Router, JWTs...), but they also provide concrete code for web apps which are complex enough to be interesting, and yet not too complex to be hopeless.

**I am not interested in how things work in general, I want to know how they are implemented.**

A classical router with session cookies such as the one exhibited here is the closest to KISS and YAGNI at the moment, I think, but bear in mind that everything is a trade off. Classics is easier to debug and understand, but there is so much bureaucracy and friction already that no wonder people are inventing endless shortcuts with services and components. 

This code is around 3KLOC which is about a few basic routes, albeit with auth and guards against bots. It already induces pain to remove some property/variable from SQL user data as there is too much dancing between SQL, Go, HTTP request headers and bodies, and tests.

Once the browser frontend is added, there emerges another friction layer between Js/Ts/React and Go which leads to the need to drop JSON or HTTP. To remove confusing "dual mode" communication and energy doubling needed on ./tests and the browser client. The fragility of net/http with streams and body parsing when processing forms becomes apparent, the issue of flickering and first paint... This is the story of my next repo.


