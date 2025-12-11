package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aabbtree77/schatzhauser/internal/config"
	"github.com/aabbtree77/schatzhauser/internal/logger"
)

type ChallengeResp struct {
	Challenge  string `json:"challenge"`
	Difficulty uint8  `json:"difficulty"`
	TTLSecs    int64  `json:"ttl_secs"`
	Token      string `json:"token"`
}

type RegisterResp struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h[:])
}

func solvePOW(challenge string, diff uint8) (string, int) {
	raw, _ := base64.RawStdEncoding.DecodeString(challenge)
	for nonce := 0; ; nonce++ {
		h := sha256.New()
		h.Write(raw)
		h.Write([]byte(fmt.Sprintf("%d", nonce)))
		sum := h.Sum(nil)

		var bits uint8
		ok := true
		for _, b := range sum {
			for i := uint(7); i < 8; i-- {
				if bits == diff {
					return fmt.Sprintf("%d", nonce), nonce
				}
				if (b>>i)&1 != 0 {
					ok = false
					break
				}
				bits++
				if bits == diff {
					return fmt.Sprintf("%d", nonce), nonce
				}
			}
			if !ok {
				break
			}
		}
	}
}

func main() {

	//Disable the ip rate limiter for this test

	// Load configuration
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		logger.Error("failed to load config", "err", err)
	}

	if cfg.IPRateLimiter.Register.Enable {
		fmt.Println("Make sure enable is false inside config.toml [ip_rate_limiter.register]")
		os.Exit(1)
	}

	if cfg.AccountPerIPLimiter.Enable {
		fmt.Println("Make sure enable is false inside config.toml [account_per_ip_limiter]")
		os.Exit(1)
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 10 * time.Second}

	// Generate unique username prefix
	seed := time.Now().UnixMilli()
	okCount := 0

	fmt.Println("== Running PoW real-register tests ==")

	for i := 0; i < 5; i++ {
		username := fmt.Sprintf("pow_user_%d_%d", seed, i)
		password := "secret123"

		// 1. Fetch challenge (server tells us if PoW enabled)
		chReq, _ := http.NewRequest("GET", baseURL+"/api/pow/challenge", nil)
		chResp, err := client.Do(chReq)
		if err != nil {
			fmt.Println("❌ Challenge request failed:", err)
			continue
		}

		if chResp.StatusCode == http.StatusNoContent {
			// === PoW disabled ===
			fmt.Println("[PoW disabled] Registering without PoW...")

			ok := doRegister(client, baseURL, username, password, "", "", "")
			if ok {
				okCount++
			}
			continue
		}

		if chResp.StatusCode != 200 {
			fmt.Println("❌ Challenge HTTP error:", chResp.Status)
			continue
		}

		var ch ChallengeResp
		body, _ := io.ReadAll(chResp.Body)
		chResp.Body.Close()
		json.Unmarshal(body, &ch)

		fmt.Printf("[%s] solving difficulty %d...\n", username, ch.Difficulty)

		// 2. Solve PoW
		nonceStr, _ := solvePOW(ch.Challenge, ch.Difficulty)

		// 3. Send real registration request with headers
		ok := doRegister(client, baseURL, username, password, ch.Challenge, nonceStr, ch.Token)
		if ok {
			okCount++
		}
	}

	if okCount == 5 {
		fmt.Println("🎉 All tests passed!")
	} else {
		fmt.Printf("❌ %d of 5 tests passed\n", okCount)
	}
}

func doRegister(client *http.Client, baseURL, username, password, challenge, nonce, token string) bool {
	body := map[string]any{
		"username": username,
		"password": password,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", baseURL+"/api/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	if challenge != "" {
		req.Header.Set("X-PoW-Challenge", challenge)
		req.Header.Set("X-PoW-Nonce", nonce)
		req.Header.Set("X-PoW-Token", token)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ Register error:", err)
		return false
	}

	defer resp.Body.Close()
	resBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 201 {
		fmt.Printf("✅ %s registered OK\n", username)
		return true
	}

	fmt.Printf("❌ %s register FAILED: HTTP %d: %s\n", username, resp.StatusCode, string(resBody))
	return false
}
