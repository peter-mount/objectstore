package auth

import (
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "path/filepath"
)

type config struct {
  Auth struct {
    // true to allow full anonymous access, false require authentication
    AllowFullAnonymous  bool              `yaml:"anonymousAccess"`
    // true to disable V2 signatures
    DisableV2           bool              `yaml:"disableV2"`
    // true to disable V4 signatures
    DisableV4           bool              `yaml:"disableV4"`
    // Enable debugging of authentication
    Debug               bool              `yaml:"debug"`
  }
  // Enable debugging
  //Debug               bool              `yaml:"debug"`
  // The root user - this user has full control on this server
  Root                User              `yaml:"rootUser"`
  // The individual users (other than root)
  Users               map[string]User   `yaml:"users"`
}

func (s *AuthService) loadConfig() error {
  filename, _ := filepath.Abs( *s.configFile )

  yml, err := ioutil.ReadFile( filename )
  if err != nil {
    return err
  }

  // root is root
  s.config.Root.root = true
  // Ensure users have the correct access key
  for k, v := range s.config.Users {
    v.AccessKey = k
  }

  return yaml.Unmarshal( yml, &s.config )
}
