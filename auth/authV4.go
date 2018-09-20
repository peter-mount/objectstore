package auth

import (
  "bytes"
  "encoding/hex"
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/awserror"
  "net/http"
  "sort"
  "strings"
  "time"
  "log"
)

// Signature and API related constants.
const (
	signV4Algorithm   = "AWS4-HMAC-SHA256"
	iso8601DateFormat = "20060102T150405Z"
	yyyymmdd          = "20060102"
)

var ignoredHeaders = map[string]bool{
	"Authorization":  true,
	"Content-Type":   true,
	"Content-Length": true,
	"User-Agent":     true,
}

// getCanonicalHeaders generate a list of request headers for
// signature.
func getCanonicalHeaders( r *rest.Rest ) string {
	var headers []string
	vals := make(map[string][]string)
	for k, vv := range r.Request().Header {
		if _, ok := ignoredHeaders[http.CanonicalHeaderKey(k)]; ok {
			continue // ignored header
		}
		headers = append(headers, strings.ToLower(k))
		vals[strings.ToLower(k)] = vv
	}
	headers = append(headers, "host")
	sort.Strings(headers)

	var buf bytes.Buffer
	// Save all the headers in canonical form <header>:<value> newline
	// separated for each header.
	for _, k := range headers {
		buf.WriteString(k)
		buf.WriteByte(':')
		switch {
		case k == "host":
			buf.WriteString( getHostAddr( r ) )
			fallthrough
		default:
			for idx, v := range vals[k] {
				if idx > 0 {
					buf.WriteByte(',')
				}
				buf.WriteString(v)
			}
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

// getSignedHeaders generate all signed request headers.
// i.e lexically sorted, semicolon-separated list of lowercase
// request header names.
func getSignedHeaders( r *rest.Rest ) string {
	var headers []string
	for k := range r.Request().Header {
		if _, ok := ignoredHeaders[http.CanonicalHeaderKey(k)]; ok {
			continue // Ignored header found continue.
		}
		headers = append(headers, strings.ToLower(k))
	}
	headers = append(headers, "host")
	sort.Strings(headers)
	return strings.Join(headers, ";")
}

// getCanonicalRequest generate a canonical request of style.
//
// canonicalRequest =
//  <HTTPMethod>\n
//  <CanonicalURI>\n
//  <CanonicalQueryString>\n
//  <CanonicalHeaders>\n
//  <SignedHeaders>\n
//  <HashedPayload>
func getCanonicalRequest( r *rest.Rest ) string {
	r.Request().URL.RawQuery = strings.Replace(r.Request().URL.Query().Encode(), "+", "%20", -1)
	canonicalRequest := strings.Join([]string{
		r.Request().Method,
		encodePath(r.Request().URL.Path),
		r.Request().URL.RawQuery,
		getCanonicalHeaders( r ),
		getSignedHeaders( r ),
		getHashedPayload( r ),
	}, "\n")
	return canonicalRequest
}

// getStringToSign a string based on selected query values.
func getStringToSignV4(t time.Time, location, canonicalRequest string) string {
	stringToSign := signV4Algorithm + "\n" + t.Format(iso8601DateFormat) + "\n"
	stringToSign = stringToSign + getScope(location, t) + "\n"
	stringToSign = stringToSign + hex.EncodeToString(sum256([]byte(canonicalRequest)))
	return stringToSign
}

func invalidAuth( authorization string ) *Credential {
  return errorCredential( awserror.InvalidArgument( "Invalid Authorization: %s", authorization ) )
}

// Creates an AWS signature version 4 credential
//https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-auth-using-authorization-header.html
func (s *AuthService) getAWS4CredentialHeader( authorization string, r *rest.Rest ) (*Credential,error) {

  a := strings.SplitN( authorization, " ", 2 )
  if len( a ) != 2 {
    return invalidAuth( authorization ), nil
  }

  // The credential header
  m := make(map[string]string)
  for _,e := range strings.Split( a[1], "," ) {
    v := strings.SplitN( strings.TrimSpace( e ), "=", 2 )
    if len(v) == 2 {
      m[strings.ToLower(v[0])] = v[1]
    }
  }
  if s.config.Debug {
    log.Println( "Authorization header:")
    for k,v := range m {
      log.Printf( "   %20s %s", k, v )
    }
  }

  v, ok := m["credential"]
  if !ok {
    return invalidAuth( authorization ), nil
  }
  a = strings.Split( v, "/" )
  if len(a) != 5 {
    return invalidAuth( authorization ), nil
  }
  date := a[1]
  location := a[2]

  user := s.getUser( a[0] )
  if user == nil {
    return invalidCredential(), nil
  }
  if s.config.Debug {
    log.Println( "User:", user )
  }

	// Time of the key
  //t := time.Now().UTC()
  t, err := time.ParseInLocation( yyyymmdd, date, s.timeLocation )
  if err != nil {
    return nil, err
  }

	// Get canonical request.
	canonicalRequest := getCanonicalRequest( r )

	// Get string to sign from canonical request.
	stringToSign := getStringToSignV4(t, location, canonicalRequest)

	// Get hmac signing key.
	signingKey := getSigningKey( user.SecretKey, location, t )

	// Get credential string.
	//credential := getCredential( user.SecretKey, location, t )

	// Get all signed headers.
	//signedHeaders := getSignedHeaders( r )

	// Calculate signature.
	signature := getSignature( signingKey, stringToSign )

  if s.config.Debug {
    log.Println( m["signature"] == signature, signature )
  }

  if signature != m["signature"] {
    return errorCredential( awserror.AccessDenied() ), nil
  }

  return userCredential( user ), nil
}
