package auth

import (
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/awserror"
  "strings"
  "log"
)

func (s *AuthService) GetCredential( r *rest.Rest ) (*Credential,error) {

  if s.config.AllowFullAnonymous {
    return anonymousCredential(), nil
  }

  authorization, exists := r.Request().Header["Authorization"]
  if exists {
    if strings.HasPrefix( authorization[0], "AWS4-HMAC-SHA256 " ) {
      c, err := s.getAWS4CredentialHeader( authorization[0], r )
      return c, err
    }

    log.Println( "Unsupported Authorization method:", authorization )
    return errorCredential( awserror.CredentialsNotSupported() ), nil
  }

  // TODO query params
  // https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-query-string-auth.html

  // Anonymous access
  return invalidCredential(), nil
}
