package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

func HashApikey(apikey string) string {
	h := sha256.Sum256([]byte(apikey))
	return hex.EncodeToString(h[:])
}

func Sha256Buffer(buf []byte) string {
	h := sha256.Sum256(buf)
	return hex.EncodeToString(h[:])
}

func getEncryptionKey(secret string) []byte {
	h := sha256.Sum256([]byte(secret))
	return h[:]
}

func EncryptApikey(apikey string, secret string) string {
	key := getEncryptionKey(secret)
	iv := make([]byte, 12)
	if _, err := rand.Read(iv); err != nil {
		panic(err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	ciphertext := aesgcm.Seal(nil, iv, []byte(apikey), nil)
	tag := ciphertext[len(ciphertext)-16:]
	encrypted := ciphertext[:len(ciphertext)-16]
	return fmt.Sprintf("%s.%s.%s", base64.StdEncoding.EncodeToString(iv), base64.StdEncoding.EncodeToString(tag), base64.StdEncoding.EncodeToString(encrypted))
}

func DecryptApikey(cipherText string, secret string) (string, error) {
	parts := strings.Split(cipherText, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("apikey 密文格式无效")
	}
	iv, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("apikey 密文格式无效")
	}
	tag, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("apikey 密文格式无效")
	}
	encrypted, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("apikey 密文格式无效")
	}
	key := getEncryptionKey(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ciphertext := append(encrypted, tag...)
	plaintext, err := aesgcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("apikey 解密失败")
	}
	return string(plaintext), nil
}
