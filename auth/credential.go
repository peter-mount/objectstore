package auth

// A credential tied to a request
type Credential struct {
  anon          bool
  deny          bool
  accessKey     string
}

func userCredential( user *User ) *Credential {
  if user == nil {
    return denyCredential()
  }
  return &Credential{
    accessKey: user.AccessKey,
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
