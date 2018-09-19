package auth

import (
  "flag"
  "github.com/peter-mount/golib/kernel"
  "time"
)

// AuthService provides support for AWS authentication server side.
// Although this is primarily for objectstore & S3 this could be reused for other
// services.
type AuthService struct {
  configFile   *string
  config        config
	timeLocation   *time.Location
}

func (s *AuthService) Name() string {
  return "AuthService"
}

func (s *AuthService) Init( k *kernel.Kernel ) error {
  s.configFile = flag.String( "config", "", "The config file to use" )

  timeLocation, err := time.LoadLocation("GMT")
	if err != nil {
		return err
	}
  s.timeLocation = timeLocation

  return nil
}

func (s *AuthService) PostInit() error {
  if *s.configFile != "" {
    err := s.loadConfig()
    if err != nil {
      return nil
    }
  }
  return nil
}
