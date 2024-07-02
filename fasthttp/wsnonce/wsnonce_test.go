package wsnonce

import (
    "testing"
//    "os"
    "fmt"
    "crypto/rand"
    "encoding/base64"
)

func TestWsNonce(t *testing.T) {

	// generate a  random nonce
    randomBytes := make([]byte, 16)
    _, err := rand.Read(randomBytes)
    if err != nil {t.Error("cannot create random bytes!")}

	encodedStr := base64.StdEncoding.EncodeToString(randomBytes)
    fmt.Printf("encoded nonce: %s %d\n", encodedStr, len(encodedStr))

 	acceptStr, err := GenNonce([]byte(encodedStr))

	if err != nil {t.Errorf("cannot gen accept: %v", err)}

    fmt.Printf("accept Str: %s %d\n", acceptStr, len(acceptStr))

}
