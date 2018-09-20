package auth

import (
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/awserror"
  "strings"
  "log"
)

func (s *AuthService) GetCredential( r *rest.Rest ) (*Credential,error) {

  if s.config.Auth.AllowFullAnonymous {
    return anonymousCredential(), nil
  }

  if s.config.Auth.Debug {
    log.Println( "Request Headers:")
    for k,v := range r.Request().Header {
      log.Printf( "   %20s %s", k, v )
    }
  }

  authorization, exists := r.Request().Header["Authorization"]
  if exists {
    if strings.HasPrefix( authorization[0], signV4Algorithm ) && !s.config.Auth.DisableV4 {
      c, err := s.getAWS4CredentialHeader( authorization[0], r )
      return c, err
    }

    if strings.HasPrefix( authorization[0], signV2Algorithm ) && !s.config.Auth.DisableV2 {
      c, err := s.getAWS2CredentialHeader( authorization[0], r )
      return c, err
    }

    if s.config.Auth.Debug {
      log.Println( "Unsupported Authorization method:", authorization )
    }

    return errorCredential( awserror.CredentialsNotSupported() ), nil
  }

  // TODO query params
  // https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-query-string-auth.html

  // Deny access
  return invalidCredential(), nil
}
