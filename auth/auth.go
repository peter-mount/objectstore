package auth

import (
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/awserror"
  "strings"
  "log"
)
const (
  AUTH_KEY = "RequestCredential"
)

// AuthenticatorDecorator rest decorator to try to resolve the requests authentication
func (s *AuthService) AuthenticatorDecorator( h rest.RestHandler ) rest.RestHandler {
  return func ( r *rest.Rest ) error {
    cred, err := s.GetCredential( r )
    if err != nil {
      return err
    }

    if s.config.Auth.Debug {
      log.Println( cred )
    }

    r.SetAttribute( AUTH_KEY, cred )
    return h(r)
  }
}

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

    return nil, awserror.CredentialsNotSupported()
  }

  // TODO query params
  // https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-query-string-auth.html

  // Deny access
  return denyCredential(), nil
}
