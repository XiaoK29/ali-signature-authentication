package client

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
)

// hmacSHA256Encrypt 使用HmacSHA256加密
func hmacSHA256Encrypt(appSecret string, stringToSign string) (string, error) {
	appSecretBytes := []byte(appSecret)
	hash := hmac.New(sha256.New, appSecretBytes)
	_, err := hash.Write([]byte(stringToSign))
	if err != nil {
		return "", err
	}

	md5Result := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(md5Result), nil
}

// hmacSHA1Encrypt 使用HmacSHA1加密
func hmacSHA1Encrypt(appSecret string, stringToSign string) (string, error) {
	appSecretBytes := []byte(appSecret)
	hmacSha1 := hmac.New(sha1.New, appSecretBytes)
	_, err := hmacSha1.Write([]byte(stringToSign))
	if err != nil {
		return "", err
	}

	signature := hmacSha1.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature), nil
}
