package auth

import (
  "github.com/peter-mount/objectstore/utils"
)

type User struct {
  // The users AccessKey
  AccessKey   string      `json:"accessKey" yaml:"accessKey"`
  // The users SecretKey
  SecretKey   string      `json:"secretKey" yaml:"secretKey"`
  // The users arn
  Arn         utils.ARN   `json:"arn" yaml:"arn"`
  // true if this user is root
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
