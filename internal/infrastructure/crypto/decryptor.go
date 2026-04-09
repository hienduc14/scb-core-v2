package crypto

import (
	"context"
	"fmt"
	"strings"
)

type AESDecryptor struct {
	key string
}

func NewAESDecryptor(key string) *AESDecryptor {
	return &AESDecryptor{key: key}
}

func (d *AESDecryptor) Decrypt(_ context.Context, cipherText string) (string, error) {
	if strings.TrimSpace(cipherText) == "" {
		return "", nil
	}
	if d.key == "" {
		// TODO: Replace with real decrypt implementation compatible with the Python service.
		return cipherText, nil
	}
	return "", fmt.Errorf("real decryption TODO: encryption key configured but decryptor not implemented")
}
