package auth

import (
	"crypto/hmac"
	"crypto/sha256"
  "encoding/hex"
  "github.com/peter-mount/golib/rest"
	"regexp"
  "strings"
  "unicode/utf8"
)

// unsignedPayload - value to be set to X-Amz-Content-Sha256 header when
const (
  unsignedPayload   = "UNSIGNED-PAYLOAD"
  iso8601DateFormat = "20060102T150405Z"
	yyyymmdd          = "20060102"
)

// if object matches reserved string, no need to encode them
var reservedObjectNames = regexp.MustCompile("^[a-zA-Z0-9-_.~/]+$")

func getSigningDate( r *rest.Rest ) string {
  v, ok := r.Request().Header["X-Amz-Date"]
  if !ok {
    v, ok = r.Request().Header["Date"]
  }
	return v[0]
}

// getSigningKey hmac seed to calculate final signature.
func getSigningKey(secret, loc string, t string) []byte {
	date := sumHMAC([]byte("AWS4"+secret), []byte(t[0:8]))
	location := sumHMAC(date, []byte(loc))
	service := sumHMAC(location, []byte("s3"))
	signingKey := sumHMAC(service, []byte("aws4_request"))
	return signingKey
}

// getSignature final signature in hexadecimal form.
func getSignature(signingKey []byte, stringToSign string) string {
	return hex.EncodeToString(sumHMAC(signingKey, []byte(stringToSign)))
}

// getScope generate a string of a specific date, an AWS region, and a
// service.
func getScope(location string, t string) string {
	scope := strings.Join([]string{
		t[0:8],
		location,
		"s3",
		"aws4_request",
	}, "/")
	return scope
}

// getHashedPayload get the hexadecimal value of the SHA256 hash of
// the request payload.
func getHashedPayload( r *rest.Rest ) string {
	hashedPayload := r.Request().Header.Get("X-Amz-Content-Sha256")
	if hashedPayload == "" {
		// Presign does not have a payload, use S3 recommended value.
		hashedPayload = unsignedPayload
	}
	return hashedPayload
}

// GetCredential generate a credential string.
func getCredential(accessKeyID, location string, t string) string {
	scope := getScope(location, t)
	return accessKeyID + "/" + scope
}

// sum256 calculate sha256 sum for an input byte array.
func sum256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// sumHMAC calculate hmac between two input byte array.
func sumHMAC(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}

// getHostAddr returns host header if available, otherwise returns host from URL
func getHostAddr( r *rest.Rest ) string {
  req := r.Request()
	if req.Host != "" {
		return req.Host
	}
	return req.URL.Host
}

func encodePath(pathName string) string {
	if reservedObjectNames.MatchString(pathName) {
		return pathName
	}
	var encodedPathname string
	for _, s := range pathName {
		if 'A' <= s && s <= 'Z' || 'a' <= s && s <= 'z' || '0' <= s && s <= '9' { // §2.3 Unreserved characters (mark)
			encodedPathname = encodedPathname + string(s)
			continue
		}
		switch s {
		case '-', '_', '.', '~', '/': // §2.3 Unreserved characters (mark)
			encodedPathname = encodedPathname + string(s)
			continue
		default:
			len := utf8.RuneLen(s)
			if len < 0 {
				// if utf8 cannot convert return the same string as is
				return pathName
			}
			u := make([]byte, len)
			utf8.EncodeRune(u, s)
			for _, r := range u {
				hex := hex.EncodeToString([]byte{r})
				encodedPathname = encodedPathname + "%" + strings.ToUpper(hex)
			}
		}
	}
	return encodedPathname
}
