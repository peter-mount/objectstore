package auth

import (
  "fmt"
  "github.com/peter-mount/objectstore/utils"
)

// A credential tied to a request
type Credential struct {
  // true if anonymous was enabled
  anon          bool
  // true if no Authorization in the request or it failed
  deny          bool
  // The authenticated user's accessKey
  accessKey     string
  // The users ARN
  arn          *utils.ARN
  // true if this user is root
  root        bool
}

func userCredential( user *User ) *Credential {
  if user == nil {
    return denyCredential()
  }
  return &Credential{
    accessKey: user.AccessKey,
    arn: &user.Arn,
    root: user.root,
  }
}

func anonymousCredential() *Credential {
  return &Credential{anon:true}
}

func denyCredential() *Credential {
  return &Credential{deny:true}
}

func (s *Credential) IsAnonymous() bool {
  return s==nil || s.anon
}

func (s *Credential) IsDeny() bool {
  return s==nil || s.deny
}

func (s *Credential) AccessKey() string {
  return s.accessKey
}

func (s *Credential) Arn() *utils.ARN {
  return s.arn
}

func (s *Credential) IsRoot() bool {
  return s != nil && s.root
}

func (s *Credential) String() string {
  if s == nil {
    return "nil"
  }
  return fmt.Sprintf( "Credential[type=%s,arn=%v,key=%s]", s.Type(), s.arn, s.accessKey)
}

func (s *Credential) Type() string {
  if s == nil {
    return "nil"
  }
  if s.root {
    return "root"
  }
  if s.anon {
    return "anonymous"
  }
  if s.deny {
    return "deny"
  }
  return "invalid"
}
