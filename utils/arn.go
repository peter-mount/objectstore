package utils

import (
  "encoding/json"
  "fmt"
  "github.com/peter-mount/go-glob"
  "github.com/peter-mount/sortfold"
  "strings"
)

// A representation of an ARN.
// https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
// https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html#identifiers-arns
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

func NewS3ARN( t, region, resource string ) *ARN {
  return NewARN( t, "aws", "s3", region, "", resource )
}

func NilARN() *ARN {
  return &ARN{"","","","","",""}
}

func AnonymousARN() *ARN {
  return &ARN{"","","","","*",""}
}

func (a *ARN) String() string {
  if a == nil {
    return ""
  }

  if a.IsAnonymous() {
    return "*"
  }

  if a.IsUserId() {
    return a.Account
  }

  return fmt.Sprintf( "%s:%s:%s:%s:%s:%s", a.Type, a.Partition, a.Service, a.Region, a.Account, a.Resource)
}

// IsNil returns true if this ARN is nil or every field is ""
func (a *ARN) IsNil() bool {
  return a == nil || (
    a.Type == "" &&
    a.Partition == "" &&
    a.Service == "" &&
    a.Region == "" &&
    a.Account == "" &&
    a.Resource == "" )
}

// IsAnonymous returns true if this ARN is for the Anonymous user.
// i.e. it's value is "*" but in this struct it's got Account == "*" but all other fields set to ""
func (a *ARN) IsAnonymous() bool {
  return a != nil &&
    a.Type == "" &&
    a.Partition == "" &&
    a.Service == "" &&
    a.Region == "" &&
    a.Account == "*" &&
    a.Resource == ""
}

// IsUserId returns true if this ARN is for a specific Account.
// Like IsAnonymous() where Account is "*", this is true if all other fields are ""
// but Account is not "" and not "*" (which is Anonymous)
func (a *ARN) IsUserId() bool {
  return a != nil &&
    a.Type == "" &&
    a.Partition == "" &&
    a.Service == "" &&
    a.Region == "" &&
    a.Account != "" &&
    a.Account != "*" &&
    a.Resource == ""
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
  // empty
  if src == "" {
    src = ":::::"
  }

  // Anonymous shorthand
  if src == "*" {
    src = "::::*:"
  }

  // Just the userId
  if !strings.Contains( src, ":" ) {
    src = "::::" + src + ":"
  }

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

// Matches tests an ARN against this one.
// Note: This ARN may contain globs (*). If the other ARN contains an * then that should fail
func (a *ARN) Matches( b *ARN ) bool {
  // If a is nil then b must be nil
  if a.IsNil() {
    return b.IsNil()
  }
  // b nil then false as a is not nil
  if b.IsNil() {
    return false
  }
  // If either is Anonymous then both must be to match
  if a.IsAnonymous() || b.IsAnonymous() {
    return a.IsAnonymous() && b.IsAnonymous()
  }

  return arnTest( a.Type, b.Type ) &&
    arnTest( a.Partition, b.Partition  ) &&
    arnTest( a.Service, b.Service ) &&
    arnTest( a.Region, b.Region ) &&
    arnTest( a.Account, b.Account ) &&
    arnTest( a.Resource, b.Resource )
}

func arnTest( a, b string ) bool {
  // Wildcard so match to true
  if a == "*" {
    return true
  }

  if strings.Contains( a, "*" ) {
    // Contains "*" so glob pattern
    return glob.Glob( a, b )
  }

  // Field is equal
  return sortfold.CompareFold( a, b ) == 0
}

func (a *ARN) UnmarshalYAML( f func(interface{}) error ) error {
  var s string
  err := f( &s )
  if err == nil {
    err = a.Parse( s )
  }
  return err
}

func (a *ARN) MarshalYAML() (interface{},error) {
  return a.String, nil
}
