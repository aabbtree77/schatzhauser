## schatzhauser

This is a RESTful API server (backend) in Go. The ideal is [PocketBase](https://pocketbase.io/docs/authentication/), aiming for something smaller here: no GUI/TUIs (they do not scale), no emails, no multiple auth providers, focus on reliability and dev-centric usability.

- [x] username/passwd auth with session cookies,

- [x] maximal request rate per IP (fixed window, in memory),

- [x] maximal account number per IP (persistent in SQLite),

- [x] god mode to create, update, list, delete users and change roles,

- [x] tests as examples (independent Go programs, no import "testing", no faking).

The code uses only Go stdlib for routing, SQLite, sqlc v2 with no SQL in the Go code. Middleware is just Go inside a request handler.

WIP...

## Setup

Clone, cd, and run `make all` which should create two binaries inside ./bin: server and god.

Rebuilding when modifying code:

```bash
make clean
make
./bin/server
time=2025-12-02T22:41:03.389+02:00 level=INFO msg="starting server" debug=true
time=2025-12-02T22:41:03.389+02:00 level=INFO msg="listening on :8080"
^Ctime=2025-12-02T22:41:40.035+02:00 level=INFO msg="shutting down"
```

Adjust config.toml to your wishes, but the tests will demand their right values.

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

## God Mode

Use god to create admin, adjust roles, cleanup users. It is a minimal CLI app which opens the same SQLite database (DB) file directly. It uses the same schema and sqlc queries.

SQLite supports multiple processes safely (file locking handles it), with one caveat. If the server is actively writing at the same moment, SQLite may briefly lock the DB. In that case god will just get a transient error; rerun is fine. So one can run god while the server is up or down.

```bash
./bin/god create --username admin --password hunter2 --role admin
./bin/god create --username alice --password s3cret --ip 203.0.113.7
./bin/god promote --username alice
./bin/god demote --username alice
./bin/god delete --username alice
./bin/god list
```

Coming soon:

```bash
./bin/god user get alice
./bin/god user set --username alice --role admin
./bin/god user rotate-password alice

./bin/god users delete --prefix test_
./bin/god users delete --created-between 2024-01-01 2024-02-01
```

## References

[How We Went All In on sqlc/pgx for Postgres + Go (2021)](https://brandur.org/sqlc)

[How We Went All In on sqlc... on HN](https://news.ycombinator.com/item?id=28462162)

[Pocketbase – open-source realtime back end in 1 file on HN](https://news.ycombinator.com/item?id=46075320)

[PocketBase: FLOSS/fund sponsorship and UI rewrite #7287](https://github.com/pocketbase/pocketbase/discussions/7287)

[Mat Ryer: How I write HTTP services in Go after 13 years (2024)](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/)
