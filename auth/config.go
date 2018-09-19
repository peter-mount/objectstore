package auth

import (
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "path/filepath"
)

type config struct {
  Root    User              `yaml:"rootUser"`
  Users   map[string]User   `yaml:"users"`
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
