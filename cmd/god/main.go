// cmd/god/main.go
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"

	dbpkg "github.com/aabbtree77/schatzhauser/db"
	"github.com/aabbtree77/schatzhauser/internal/config"
)

// usage prints small help for subcommands.
func usage() {
	fmt.Println(`god — administrative CLI to manage users (create, delete, promote, demote, list)

Usage:
  god create   --username alice --password secret [--ip 1.2.3.4] [--role admin|user]
  god delete   --username alice
  god promote  --username alice
  god demote   --username alice
  god list

Examples:
  god create --username admin --password hunter2 --role admin
  god promote --username alice
  god list
`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	sub := os.Args[1]

	// Load config to find DB path
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := sql.Open("sqlite3", cfg.DBPath+"?_foreign_keys=on")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// sanity ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	store := dbpkg.NewStore(db)

	switch sub {
	case "create":
		cmd := flag.NewFlagSet("create", flag.ExitOnError)
		username := cmd.String("username", "", "username (required)")
		password := cmd.String("password", "", "password (required)")
		ip := cmd.String("ip", "", "ip (optional)")
		role := cmd.String("role", "user", "role: admin or user")
		cmd.Parse(os.Args[2:])

		if *username == "" || *password == "" {
			cmd.Usage()
			os.Exit(2)
		}
		roleVal := strings.ToLower(*role)
		if roleVal != "admin" && roleVal != "user" {
			log.Fatalf("invalid role: %s", *role)
		}

		// bcrypt hash
		pwHash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("hashing password: %v", err)
		}

		ctx := context.Background()
		// use CreateUserWithRole (sqlc-generated) that expects: username, password_hash, ip, role
		user, err := store.CreateUserWithRole(ctx, dbpkg.CreateUserWithRoleParams{
			Username:     *username,
			PasswordHash: string(pwHash),
			Ip:           *ip,
			Role:         roleVal,
		})

		if err != nil {
			log.Fatalf("create user: %v", err)
		}
		fmt.Printf(
			"created user: id=%d username=%s role=%s created_at=%-25v\n",
			user.ID, user.Username, user.Role,
			formatNullTime(user.CreatedAt),
		)

	case "delete":
		cmd := flag.NewFlagSet("delete", flag.ExitOnError)
		username := cmd.String("username", "", "username (required)")
		cmd.Parse(os.Args[2:])
		if *username == "" {
			cmd.Usage()
			os.Exit(2)
		}
		ctx := context.Background()
		if err := store.DeleteUserByUsername(ctx, *username); err != nil {
			log.Fatalf("delete user: %v", err)
		}
		fmt.Printf("deleted user %s\n", *username)

	case "promote":
		cmd := flag.NewFlagSet("promote", flag.ExitOnError)
		username := cmd.String("username", "", "username (required)")
		cmd.Parse(os.Args[2:])
		if *username == "" {
			cmd.Usage()
			os.Exit(2)
		}
		ctx := context.Background()
		u, err := store.UpdateUserRole(ctx, dbpkg.UpdateUserRoleParams{
			Role:     "admin",
			Username: *username,
		})

		if err != nil {
			log.Fatalf("promote: %v", err)
		}
		fmt.Printf("promoted user: id=%d username=%s role=%s\n", u.ID, u.Username, u.Role)

	case "demote":
		cmd := flag.NewFlagSet("demote", flag.ExitOnError)
		username := cmd.String("username", "", "username (required)")
		cmd.Parse(os.Args[2:])
		if *username == "" {
			cmd.Usage()
			os.Exit(2)
		}
		ctx := context.Background()
		u, err := store.UpdateUserRole(ctx, dbpkg.UpdateUserRoleParams{
			Role:     "user",
			Username: *username,
		})

		if err != nil {
			log.Fatalf("demote: %v", err)
		}
		fmt.Printf("demoted user: id=%d username=%s role=%s\n", u.ID, u.Username, u.Role)

	case "list":
		// no flags
		ctx := context.Background()
		users, err := store.ListUsers(ctx)
		if err != nil {
			log.Fatalf("list users: %v", err)
		}
		fmt.Printf("%-5s %-20s %-8s %-25s\n", "ID", "USERNAME", "ROLE", "CREATED_AT")
		for _, u := range users {
			fmt.Printf(
				"%-5d %-20s %-8s %-25v\n",
				u.ID,
				u.Username,
				u.Role,
				formatNullTime(u.CreatedAt))
		}

	case "help", "-h", "--help":
		usage()

	default:
		fmt.Printf("unknown subcommand: %s\n\n", sub)
		usage()
		os.Exit(2)
	}
}

/*
fmt.Printf with %s on SQLITE value created_at DATETIME DEFAULT CURRENT_TIMESTAMP
leads to

139   admin                admin    {2025-12-09 01:14:05 +0000 UTC %!s(bool=true                     )}

# This helper turns it into

139   admin                admin    2025-12-09T01:14:05Z
*/
func formatNullTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}
