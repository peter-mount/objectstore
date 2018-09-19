package auth

import (
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/awserror"
)

// A credential tied to a request
type Credential struct {
  anon          bool
  invalid       bool
  accessKey     string
  Error        *awserror.Error
}

func (c *Credential) SendErrorResponse( r *rest.Rest ) *rest.Rest {
  if c.Error == nil {
    c.Error = awserror.InternalError()
  }
  return c.Error.Send( r )
}

func userCredential( user *User ) *Credential {
  if user == nil {
    return invalidCredential()
  }
  return &Credential{
    accessKey: user.AccessKey,
  }
}

func anonymousCredential() *Credential {
  return &Credential{anon:true}
}

func invalidCredential() *Credential {
  return &Credential{invalid:true}
}

func errorCredential( e *awserror.Error ) *Credential {
  return &Credential{invalid:true,Error:e}
}

func (s *Credential) IsAnonymous() bool {
  return s==nil || s.anon
}

func (s *Credential) IsInvalid() bool {
  return s==nil || s.invalid
}

func (s *Credential) AccessKey() string {
  return s.accessKey
}
