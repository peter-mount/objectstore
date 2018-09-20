package utils

import (
  "encoding/json"
  "fmt"
  "strings"
)

// A representation of an ARN
type ARN struct {
  // "arn" for AWS but allows us to have our own
  Type      string
  // Partition, "AWS" for standard aws
  Partition string
  // Service, e.g. "s3", "iam" etc
  Service   string
  // Region (or location) e.g. "us-east-1" or "" for some
  Region    string
  // Account ID
  Account   string
  // Specific for a source
  Resource  string
}

func NewARN( t, partition, service, region, account, resource string ) *ARN {
  return &ARN{t, partition, service, region, account, resource}
}

func (a *ARN) String() string {
  return fmt.Sprintf( "%s:%s:%s:%s:%s:%s", a.Type, a.Partition, a.Service, a.Region, a.Account, a.Resource)
}

func (a *ARN) Equal( b *ARN ) bool {
  if a==nil {
    return b==nil
  }

  return b != nil &&
    a.Type == b.Type &&
    a.Partition == b.Partition &&
    a.Service == b.Service &&
    a.Region == b.Region &&
    a.Account == b.Account &&
    a.Resource == b.Resource
}

func (a *ARN) UnmarshalJSON( b []byte ) error {
  if b == nil {
    return nil
  }

  bl := len( b )
  if b[0]!='"' || b[bl-1]!='"' {
    return fmt.Errorf( "Invalid ARN %s", b )
  }

  return a.Parse( string(b[1:bl-1]) )
}

func ParseARN( src string ) (*ARN, error) {
  a := &ARN{}
  err := a.Parse( src )
  if err != nil {
    return nil, err
  }
  return a, nil
}

func ( a *ARN) Parse( src string ) error {
  s := strings.SplitN( src, ":", 6 )
  if len(s) != 6 {
    return fmt.Errorf( "Invalid ARN %s", src )
  }

  a.Type = s[0]
  a.Partition = s[1]
  a.Service = s[2]
  a.Region = s[3]
  a.Account = s[4]
  a.Resource = s[5]
  return nil
}

func (a *ARN) MarshalJSON() ( []byte, error ) {
  if a == nil {
    return []byte("null"), nil
  }

  return json.Marshal( a.String() )
}