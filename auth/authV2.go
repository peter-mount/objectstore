package auth

import (
  "bytes"
	"crypto/hmac"
	"crypto/sha1"
  "encoding/base64"
  "fmt"
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/awserror"
  "net/http"
  "net/url"
  "sort"
  "strings"
)

// Signature and API related constants.
const (
	signV2Algorithm = "AWS"
)

// Encode input URL path to URL encoded path.
func encodeURL2Path( req *http.Request ) string {
	return encodePath( req.URL.Path )
}

func stringToSignV2( r *rest.Rest ) string {
	buf := new(bytes.Buffer)
  req := r.Request()

	// Write standard headers.
  buf.WriteString(req.Method + "\n")
  buf.WriteString(req.Header.Get("Content-Md5") + "\n")
  buf.WriteString(req.Header.Get("Content-Type") + "\n")
  buf.WriteString(req.Header.Get("Date") + "\n")

	// Write canonicalized protocol headers if any.
	writeCanonicalizedHeaders(buf, req)

	// Write canonicalized Query resources if any.
	writeCanonicalizedResource(buf, req)
	return buf.String()
}

// writeCanonicalizedHeaders - write canonicalized headers.
func writeCanonicalizedHeaders(buf *bytes.Buffer, req *http.Request) {
	var protoHeaders []string
	vals := make(map[string][]string)
	for k, vv := range req.Header {
		// All the AMZ headers should be lowercase
		lk := strings.ToLower(k)
		if strings.HasPrefix(lk, "x-amz") {
			protoHeaders = append(protoHeaders, lk)
			vals[lk] = vv
		}
	}
	sort.Strings(protoHeaders)
	for _, k := range protoHeaders {
		buf.WriteString(k)
		buf.WriteByte(':')
		for idx, v := range vals[k] {
			if idx > 0 {
				buf.WriteByte(',')
			}
			if strings.Contains(v, "\n") {
				// TODO: "Unfold" long headers that
				// span multiple lines (as allowed by
				// RFC 2616, section 4.2) by replacing
				// the folding white-space (including
				// new-line) by a single space.
				buf.WriteString(v)
			} else {
				buf.WriteString(v)
			}
		}
		buf.WriteByte('\n')
	}
}

// AWS S3 Signature V2 calculation rule is give here:
// http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html#RESTAuthenticationStringToSign

// Whitelist resource list that will be used in query string for signature-V2 calculation.
// The list should be alphabetically sorted
var resourceList = []string{
	"acl",
	"delete",
	"lifecycle",
	"location",
	"logging",
	"notification",
	"partNumber",
	"policy",
	"requestPayment",
	"response-cache-control",
	"response-content-disposition",
	"response-content-encoding",
	"response-content-language",
	"response-content-type",
	"response-expires",
	"torrent",
	"uploadId",
	"uploads",
	"versionId",
	"versioning",
	"versions",
	"website",
}

// From the Amazon docs:
//
// CanonicalizedResource = [ "/" + Bucket ] +
// 	  <HTTP-Request-URI, from the protocol name up to the query string> +
// 	  [ sub-resource, if present. For example "?acl", "?location", "?logging", or "?torrent"];
func writeCanonicalizedResource(buf *bytes.Buffer, req *http.Request) {
	// Save request URL.
	requestURL := req.URL
	// Get encoded URL path.
	buf.WriteString(encodeURL2Path(req))
	if requestURL.RawQuery != "" {
		var n int
		vals, _ := url.ParseQuery(requestURL.RawQuery)
		// Verify if any sub resource queries are present, if yes
		// canonicallize them.
		for _, resource := range resourceList {
			if vv, ok := vals[resource]; ok && len(vv) > 0 {
				n++
				// First element
				switch n {
				case 1:
					buf.WriteByte('?')
				// The rest
				default:
					buf.WriteByte('&')
				}
				buf.WriteString(resource)
				// Request parameters
				if len(vv[0]) > 0 {
					buf.WriteByte('=')
					buf.WriteString(vv[0])
				}
			}
		}
	}
}

// Authorization = "AWS" + " " + AWSAccessKeyId + ":" + Signature;
// Signature = Base64( HMAC-SHA1( YourSecretAccessKeyID, UTF-8-Encoding-Of( StringToSign ) ) );
//
// StringToSign = HTTP-Verb + "\n" +
//  	Content-Md5 + "\n" +
//  	Content-Type + "\n" +
//  	Date + "\n" +
//  	CanonicalizedProtocolHeaders +
//  	CanonicalizedResource;
//
// CanonicalizedResource = [ "/" + Bucket ] +
//  	<HTTP-Request-URI, from the protocol name up to the query string> +
//  	[ subresource, if present. For example "?acl", "?location", "?logging", or "?torrent"];
//
// CanonicalizedProtocolHeaders = <described below>

// SignV2 sign the request before Do() (AWS Signature Version 2).
func signV2( r *rest.Rest, accessKeyID, secretAccessKey string ) (string,error) {

	// Calculate HMAC for secretAccessKey.
	stringToSign := stringToSignV2( r )

	hm := hmac.New(sha1.New, []byte(secretAccessKey))
	hm.Write([]byte(stringToSign))

	// Prepare auth header.
	authHeader := new(bytes.Buffer)
	authHeader.WriteString(fmt.Sprintf("%s %s:", signV2Algorithm, accessKeyID))

	encoder := base64.NewEncoder(base64.StdEncoding, authHeader)
	encoder.Write(hm.Sum(nil))
	encoder.Close()

	// Authorization header.
  return authHeader.String(), nil
}

// Creates an AWS signature version 4 credential
//https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-auth-using-authorization-header.html
func (s *AuthService) getAWS2CredentialHeader( authorization string, r *rest.Rest ) (*Credential,error) {

  a := strings.SplitN( authorization, " ", 2 )
  a = strings.SplitN( a[1], ":", 2 )
  user := s.getUser( a[0] )
  if user == nil {
    return nil, awserror.InvalidAccessKeyId()
  }

  signature, err := signV2( r, user.AccessKey, user.SecretKey )
  if err != nil {
    return nil, err
  }

  if signature != authorization {
    return nil, awserror.AccessDenied()
  }

  return userCredential( user ), nil
}
