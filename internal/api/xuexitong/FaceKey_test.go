package xuexitong

import (
	"encoding/pem"
	"strings"
	"testing"
)

func TestFaceSecretsDecrypt(t *testing.T) {
	key := DefaultFaceKeyPEM()
	if key == "" {
		t.Fatal("embedded face key decrypted to empty string")
	}
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		t.Fatal("embedded face key is not valid PEM after decryption")
	}
	if !strings.Contains(key, "PRIVATE KEY") {
		t.Fatal("embedded face key does not look like a private key")
	}
	if got := FaceUploadAppID(); got == "" {
		t.Fatal("upload appid decrypted to empty string")
	}
	if got := FaceUploadSecret(); got == "" {
		t.Fatal("upload secret decrypted to empty string")
	}
}
