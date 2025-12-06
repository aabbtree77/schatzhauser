## schatzhauser

This is a minimal RESTful API server (backend) in Go:

- the Go stdlib router, no chi,

- username and password authentication with session cookies,

- request rate limiter per IP to fight evil,

- "middleware" is just Go inside a request handler, Do Repeat Yourself,

- a builder to store default params**,** dropped config structs and functional options: [1](https://www.reddit.com/r/golang/comments/5ky6sf/the_functional_options_pattern/), [2](https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html), [3](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis), [4](https://www.youtube.com/watch?v=MDy7JQN5MN4), but discarded both,

- [Mat Ryer's graceful ctrl+C shutdown](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/).

## Motivation

The ideal is [PocketBase](https://pocketbase.io/docs/authentication/), but aiming for something smaller, easier to use, and more robust.

```
systemctl enable myapp
systemctl start myapp
journalctl -u myapp -f
```

myapp is just a single Go binary file which reads and updates the SQLite file data.db on the same VPS, a few commands, plain Linux. No devops yaml "application gateway ingress controllers" from hell.

If you are into scaling, you are on the wrong side of history.

Almost no 3rd party, except these:

```bash
go mod init github.com/aabbtree77/schatzhauser
go get github.com/mattn/go-sqlite3
go get golang.org/x/crypto/bcrypt
go get github.com/sqlc-dev/sqlc
go get github.com/BurntSushi/toml
```

## Details

Terminal 1:

```bash
sqlc generate
go build -o schatzhauser .
mkdir -p data
./schatzhauser
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
go run ./tests/ip_rate_limit
```

Comment out the IP rate limiters when testing anything but `ip_rate_limit`. In the latter, make sure config.toml [ip_rate_limiter.login] section params match the ones at the start of func main() inside tests/ip_rate_limit/main.go.

The IP rate limiter is a simple-looking fixed window counter, but it is already the second version as the first one leaked memory. Tricky...

Ask AI to write an industrial grade IP rate limiter, but bear in mind those codes are hard to understand and debug, so I follow KISS here.

## Coming Soon

- [ ] Rate limiter per user.

- [ ] HTTP request body size limits.

- [ ] Session expiry.

- [ ] IP bans.

## Some Food for Thought

- A long time Go proponent Anthony GG uses [Remix with Supabase](https://www.youtube.com/watch?v=rlJx5f5OlYA&t=791s) in his latest projects. Go is still somewhere on the server side, but not with DB and VPS all the way, and no PocketBase either.

- Ultimately, the goal is to build useful no nonsense services, like eBay, [vinted.lt](https://www.vinted.lt/), [barbora.lt](https://barbora.lt/)...

- Js/Ts apps slowly leak memory, see [this case by Web Dev Cody](https://youtu.be/gNDBwxeBrF4?t=176), but nobody cares.

- Even Rob Pike does ["Java design patterns"](https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html) at times, but he does have [better stuff](https://www.youtube.com/watch?v=oV9rvDllKEg&t=327s).
