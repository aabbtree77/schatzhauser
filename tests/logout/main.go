package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"
)

func must(err error) {
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}

func registerAndLogin() (*http.Client, string) {
	username := fmt.Sprintf("logout_test_%d", time.Now().UnixNano())
	password := "secret123"

	// register
	payload := map[string]string{"username": username, "password": password}
	data, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:8080/register", "application/json", bytes.NewReader(data))
	must(err)
	resp.Body.Close()
	if resp.StatusCode != 201 && resp.StatusCode != 409 {
		fmt.Println("Failed to register user:", resp.StatusCode)
		os.Exit(1)
	}

	// login
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	resp, err = client.Post("http://localhost:8080/login", "application/json", bytes.NewReader(data))
	must(err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Failed to login:", resp.StatusCode, string(body))
		os.Exit(1)
	}

	return client, username
}

func runTestLogout() bool {
	client, _ := registerAndLogin()

	// logout
	req, _ := http.NewRequest("POST", "http://localhost:8080/logout", nil)
	resp, err := client.Do(req)
	must(err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Printf("FAIL: logout: status=%d, body=%s\n", resp.StatusCode, body)
		return false
	}
	fmt.Printf("PASS: logout: status=%d, body=%s\n", resp.StatusCode, body)

	// profile request after logout should fail
	resp2, err := client.Get("http://localhost:8080/profile")
	must(err)
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	if resp2.StatusCode != 401 {
		fmt.Printf("FAIL: profile after logout: status=%d, body=%s\n", resp2.StatusCode, body2)
		return false
	}
	fmt.Printf("PASS: profile after logout: got 401 as expected\n")

	return true
}

func main() {
	passed := 0
	if runTestLogout() {
		passed++
	}

	fmt.Printf("=== SUMMARY: %d/1 passed ===\n", passed)
	if passed != 1 {
		os.Exit(1)
	}
}
