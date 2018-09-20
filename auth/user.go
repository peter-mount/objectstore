package auth

type User struct {
  AccessKey   string  `json:"accessKey" yaml:"accessKey"`
  SecretKey   string  `json:"secretKey" yaml:"secretKey"`
  root        bool
}

// GetUser returns a User for an accessKey
func (s *AuthService) getUser( accessKey string ) *User {
  // Root overrides everything
  if accessKey == s.config.Root.AccessKey {
    return &s.config.Root
  }

  if user, exists := s.config.Users[ accessKey ]; exists {
    return &user
  }

  return nil
}

func (s *User) IsRoot() bool {
  return s != nil && s.root
}

func (s *User) String() string {
  if s == nil {
    return "nil"
  }
  return s.AccessKey
}
