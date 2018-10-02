package auth

import (
  "bytes"
  "encoding/hex"
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/awserror"
  "log"
  "strings"
  "time"
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
func getCanonicalHeaders( r *rest.Rest, m map[string]string ) string {
	// Save all the headers in canonical form <header>:<value> newline
	// separated for each header.
  var buf bytes.Buffer
	for _, k := range strings.Split( m["signedheaders"], ";" ) {
		buf.WriteString(k)
		buf.WriteByte(':')
		switch {
  		case k == "host":
  			buf.WriteString( getHostAddr( r ) )
  			fallthrough
  		default:
        buf.WriteString( m[k] )
  			buf.WriteByte( '\n' )
		}
	}
	return buf.String()
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
func getCanonicalRequest( r *rest.Rest, m map[string]string ) string {
	r.Request().URL.RawQuery = strings.Replace(r.Request().URL.Query().Encode(), "+", "%20", -1)
	return strings.Join([]string{
		r.Request().Method,
		encodePath(r.Request().URL.Path),
		r.Request().URL.RawQuery,
		getCanonicalHeaders( r, m ),
		m["signedheaders"],
		getHashedPayload( r ),
	}, "\n")
}

// getStringToSign a string based on selected query values.
func getStringToSignV4(t time.Time, location, canonicalRequest string) string {
	stringToSign := signV4Algorithm + "\n" + t.Format(iso8601DateFormat) + "\n"
	stringToSign = stringToSign + getScope(location, t) + "\n"
	stringToSign = stringToSign + hex.EncodeToString(sum256([]byte(canonicalRequest)))
	return stringToSign
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
func (s *AuthService) getV4Credential( authorization string, r *rest.Rest ) (map[string]string, string, string, string, bool ) {
  a := strings.SplitN( authorization, " ", 2 )
  if len( a ) == 2 {

    m := make(map[string]string)

    for k,v := range r.Request().Header {
      m[strings.ToLower(k)] = strings.Join( v, "," )
    }

    // The credential header
    for _,e := range strings.Split( a[1], "," ) {
      v := strings.SplitN( strings.TrimSpace( e ), "=", 2 )
      if len(v) == 2 {
        m[strings.ToLower(v[0])] = v[1]
      }
    }

    if s.config.Auth.Debug {
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

  m, accessKey, location, _, valid := s.getV4Credential( authorization, r )
  if !valid {
    return nil, awserror.InvalidArgument( "Invalid Authorization: %s", authorization )
  }

  user := s.getUser( accessKey )
  if user == nil {
    return nil, awserror.InvalidAccessKeyId()
  }
  if s.config.Auth.Debug {
    log.Println( "User:", user.AccessKey, user.SecretKey, user.Arn )
  }

  // Get canonical request.
  canonicalRequest := getCanonicalRequest( r, m )
  if s.config.Auth.Debug {
    log.Printf( "canonicalRequest\n%s", canonicalRequest)
  }

  // Get string to sign from canonical request.
  stringToSign := getStringToSignV4( t, location, canonicalRequest )

  // Get hmac signing key.
  signingKey := getSigningKey( user.SecretKey, location, t )

  // Calculate signature.
  signature := getSignature( signingKey, stringToSign )

  if s.config.Auth.Debug {
    log.Println( m["signature"] == signature, signature )
  }

  if signature != m["signature"] {
    return nil, awserror.AccessDenied()
  }

  return userCredential( user ), nil
}
