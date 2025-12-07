## schatzhauser

This is a minimal RESTful API server (backend) in Go:

- dev-centric and AI-friendly:

  - import "net/http", SQLite, sqlc, no fragile 3rd party,
  - "middleware" is just Go inside a request handler,
  - no javaisms, no lambda calculus: simple config structs,
  - self-sufficient tests as external to the server Go programs (no import "testing").

- username/passwd auth with session cookies,

- maximal request rate per IP (fixed window, in memory, non-leaking memory),

- maximal account number per IP (persistent in SQLite),

- [Mat Ryer's graceful ctrl+C](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/).

More to come, read below.

## Motivation

The ideal is [PocketBase](https://pocketbase.io/docs/authentication/), aiming for something even simpler here: no GUIs, no emails, no multiple auth providers, more focus on reliability and dev-centric usability.

The PocketBase revolution:

```
systemctl enable myapp
systemctl start myapp
journalctl -u myapp -f
```

myapp is a Go binary which updates data.db on the same VPS, plain Linux. No devops yaml "application gateway ingress controllers" from hell, no Js/Ts bundlers and Next.js. If you are scaling, you are on the wrong side of history.

```bash
git clone https://github.com/aabbtree77/schatzhauser.git
cd schatzhauser
go mod init github.com/aabbtree77/schatzhauser
go get github.com/mattn/go-sqlite3
go get golang.org/x/crypto/bcrypt
go get github.com/sqlc-dev/sqlc
go get github.com/BurntSushi/toml
sqlc generate
mkdir -p data
go build -o myapp .
```

## Tests

Terminal 1:

```bash
./myapp
time=2025-12-02T22:41:03.389+02:00 level=INFO msg="starting schatzhauser" debug=true
time=2025-12-02T22:41:03.389+02:00 level=INFO msg="listening on :8080"
^Ctime=2025-12-02T22:41:40.035+02:00 level=INFO msg="shutting down"
```

Terminal 2:

```bash
# register
curl -i -X POST -H 'Content-Type: application/json' -d '{"username":"u1","password":"p1"}' http://localhost:8080/register

# login (save cookie)
curl -i -c cookiejar.txt -X POST -H 'Content-Type: application/json' -d '{"username":"u1","password":"p1"}' http://localhost:8080/login

# profile (send cookie)
curl -i -b cookiejar.txt http://localhost:8080/profile

# logout
curl -i -b cookiejar.txt -X POST http://localhost:8080/logout
```

Further tests:

```bash
go run ./tests/profile
Cookies after login: [schatz_sess=267e63d4c1d67ab173ba385cb53929a1db5a80e542f0cae1065209c622c2891e]
PASS: profile with cookie: status=200, body={"status":"ok","user":{"created":{"Time":"2025-12-02T21:31:43Z","Valid":true},"id":19,"username":"profile_test_1764711103328518799"}}

PASS: profile without cookie: got 401 as expected
=== SUMMARY: 1/1 passed ===
```

```
go run ./tests/register
go run ./tests/login
go run ./tests/profile
go run ./tests/logout
go run ./tests/req_rate_per_ip
go run ./tests/account_rate_per_ip
```

The IP rate limiter is a simple-looking fixed window counter, but it is already the second version as the first one leaked memory. Tricky...

Ask AI to write an industrial grade IP rate limiter, but bear in mind the codes which are hard to understand will be even harder to debug, so I follow KISS here.

## More to Come

- [x] Maximal number of registered accounts per IP.

- [ ] Admin to manage users (role variable with create/delete capability).

- [ ] Proof of work to make spam not viable economically.

- [ ] HTTP request body size limiter.

- [ ] Black listing IPs.

- [ ] sqlc workflow to do just about anything.

- [ ] HTTPS with Caddy, tests on real VPS or locally with ngrok.

- [ ] Harmonization, versioning, docs, use, promotion.

- [ ] The ultimate goal is to build useful reliable services, bring back the joy of programming.

## References

[How We Went All In on sqlc/pgx for Postgres + Go](https://brandur.org/sqlc)

[How We Went All In on sqlc... on HN](https://news.ycombinator.com/item?id=28462162)

[Pocketbase – open-source realtime back end in 1 file on HN](https://news.ycombinator.com/item?id=46075320)

[PocketBase: FLOSS/fund sponsorship and UI rewrite #7287](https://github.com/pocketbase/pocketbase/discussions/7287)
