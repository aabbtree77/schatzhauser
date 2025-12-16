## Maximal Account Number per IP

### Modification No. 1

Counting users per ip directly turns out to be not a good idea:

```sql
-- name: CountUsersByIP :one
SELECT COUNT(*) FROM users
WHERE ip = ?;
```

The SQL code above will also count deleted users (unless hard-deleted) and spam accounts on the same IP as the latter may change (VPN, NAT). This is a rough fix:

```sql
SELECT COUNT(*) FROM users
WHERE ip = ?
AND created_at >= datetime('now', '-30 days');
```

The actual limit is on the account number per IP per 30 days.

### Modification No. 2

There is an edge case with the race condition. Two /register requests from the same IP arrive at nearly the same time.

Both do:

SELECT COUNT(users WHERE ip = X).

See count = 6 (limit is 7).

Both proceed.

Both INSERT user.

Result: 8 accounts from one IP, limit violated.

Why it happens:

The count check and the insert are two separate statements.

Default isolation allows both transactions to see the same snapshot.

The DB is doing exactly what is asked, but not what is meant.

The solution is to create a dummy query for sqlc:

```sql
-- name: TouchUsersTable :exec
UPDATE users SET ip = ip WHERE ip = ?;
```

leading to this Go code:

```go
if err := txStore.TouchUsersTable(r.Context(), ip); err != nil {
    httpx.InternalError(w, "cannot lock users table")
    return
}
```

It updates zero or more rows, changes nothing, **forces a write lock**.

With that writer lock, when two users register at the same time:

Request A reaches TouchUsersTable -> gets writer lock.

Request B reaches TouchUsersTable -> blocks.

A counts, inserts, commits.

B resumes, recounts, now sees **updated state**, and discards user creation as the limit 7 is already reached.

SQLite internally serializes writers. Readers may still run, writers wait their turn.
