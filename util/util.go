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
	if len(encoded) > 6 {
		return encoded[:7]
	} else {
		return encoded
	}
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func MakeURL(baseURL, identifier string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	parsed.Path = identifier

	return parsed.String(), nil
}
