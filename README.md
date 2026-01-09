<p align="center">
  <img src="docs/schatzhauser-banner.png" alt="Schatzhauser" style="width: 90%; height: auto;" />
</p>

<p align="center">
  <em>Authenticated bot-proof minimal JSON API.</em>
</p>

This is a minimal Go JSON API server to test ideas to fight bots, in the context of an authenticated web app.
It is a cookie-based session auth HTTP server with JSON payloads inside HTTP request bodies. The client is programmatic, either direct curl, or ./tests, no frontend. There is also a CLI to clean up the DB.

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

[RealWorld (Conduit)](https://github.com/gothinkster/realworld)

This repo includes a huge number of medium.com clones. They are the most archetypical web apps with auth, users, posts, and comments. Most of them are extremely over architected: go kit, aws dynamoDB, hexagonal architecture, OpenAPI, JWTs... all the wrong ideas there ;).

## Notes

- If you plan to add browser/frontend to this code, don't do that, better start from scratch, or use, cough cough, some framework. This is just to play with rate limiters and AI. SQLite with sqlc has turned out to be a brilliant idea, something to reuse.

- It is possible to get a reliable no-nonsense Go backend serving HTML/CSS/Js with auth. Server-first (SSR), as little JSON and Js as possible, with proper HTTP status codes and redirects (not entirely clear if this is better than JSON APIs or needed at all). Make everything a flat list of feature folders with a single route.go mapping links to folders via net/http. However, this only leads to a clean backend that does not do much. It is a mistake to assume that frontend will be just a bunch of HTML/Js files one per feature folder, which AI will write in no time. Frontend is where most of the work resides.

- The main reason people go for React and Next.js is libraries like shadcn where one can copy/paste components as code, not just styled HTML scaffold like DaisyUI. Forms have logic, async, they are tricky, not just "HTML form here and there". Ignore the multipart form parsing and touch HTTP bodies as streams in a few wrong places on the Go end, and you will hit some spectacular EOF heisenbugs there.

- HTTP is brittle. It will touch Go, HTML, HTML forms, Js in all sorts of weird ways and the codes will seldom show the most important part, who sends what and where. This is what the founding fathers did not do well.

  ```js
  fetch("/pow/challenge"...)

  fetch(form.action...)

  window.location.href = res.url;

  http.Redirect(w, r, "/login?registered=1", http.StatusSeeOther)

  <form method="POST" action="/actions/register" novalidate></form>

  w.Header().Set("Cache-Control", "no-store") w.Header().Set("Content-Type",
  "application/json") \_ = json.NewEncoder(w).Encode(resp)

  <script src="/js/pow.js" defer></script>
  ```

  A compiler won't help much here. Need to develop some intuition what sends where. F12 is guaranteed.

- Another trap is to assume a classical SSR approach with "React islands" and minimal React as advanced HTML (no SPA router and magic). Potentially using esbuild with Makefile to turn .jsx to .js, removing vite and such build systems which start to dictate architecture and conflict with the Go build system. Eventually all one achieves this way is just rebuilding another vite. Federating those .jsx files is a massive time sink. One self-sufficient .jsx file per feature folder won't work. Its .js will have to include the whole React runtime reaching 200KB in prod, and then each feature folder will multiply this into megabytes. This is why federation/vite is needed.

- First contentful paint, SEO, and similar discussions ditching SPAs are absolute nonsense, premature optimizations at best.

- If we start with React, SPA, JSON APIs, the question becomes, do we actually need Go, does it add value or becomes a friction? TypeScript and Node.js are Disneyland for sure, but we can also use TypeScript just as Go to write a clean backend and vastly reduce friction of having to jump between two languages. This is the main reason Go is so little used in building web apps. A vastly better runtime (compilation to binary to begin with), but unfortunately React holds the cards, paraphrasing DJT ;).
