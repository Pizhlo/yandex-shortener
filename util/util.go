package util

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"net/url"
)

func Shorten(s string) string {
	hash := GetMD5Hash(s)
	encoded := base64.StdEncoding.EncodeToString([]byte(hash))

	return encoded[:7]
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func MakeURL(baseURL, identifier string) (string, error) {
	scheme := "http://"

	parsed, err := url.Parse(scheme + baseURL)
	if err != nil {
		return "", err
	}

	parsed.Path = identifier

	return parsed.String(), nil
}
