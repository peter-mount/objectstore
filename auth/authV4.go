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

// getV4Credential extracts the content of the Authorization header.
//
// Returns m, accessKey, location, service, valid:
//   valid is false for an error, else true
//   m map of the individual components of the authorization string
//   accessKey of the user
//   location usually us-east-1
//   service "s3" but could be different for other services
//
func (s *AuthService) getV4Credential( authorization string ) (map[string]string, string, string, string, bool ) {
  a := strings.SplitN( authorization, " ", 2 )
  if len( a ) == 2 {

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

    if v, ok := m["credential"]; ok {
      // accessKey/date/location/service/"aws4_request"
      a = strings.Split( v, "/" )
      if len(a) == 5 {
          return m, a[0], a[2], a[3], true
      }
    }
  }
  return nil, "", "", "", false
}

// Creates an AWS signature version 4 credential
//https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-auth-using-authorization-header.html
func (s *AuthService) getAWS4CredentialHeader( authorization string, r *rest.Rest ) (*Credential,error) {

  // The X-Amz-Date or Date header
  t, err := getSigningDate( r )
  if err != nil {
    return nil, err
  }

  m, accessKey, location, _, valid := s.getV4Credential( authorization )
  if !valid {
    return invalidAuth( authorization ), nil
  }

  user := s.getUser( accessKey )
  if user == nil {
    return invalidCredential(), nil
  }
  if s.config.Debug {
    log.Println( "User:", user )
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
