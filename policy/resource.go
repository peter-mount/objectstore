package policy

import (
  "encoding/json"
  "github.com/peter-mount/objectstore/utils"
)

// A Resource
type Resource struct {
  resources []utils.ARN
  negate      bool
}

func NewResource( resources ...utils.ARN ) *Resource {
  r := &Resource{}
  r.resources = resources
  return r
}

// IsNil returns true if the Resource is empty
func (a *Resource) IsNil() bool {
  return a == nil || len( a.resources ) == 0
}

// Len returns the number of actions
func (a *Resource) Len() int {
  if a==nil {
    return 0
  }
  return len(a.resources)
}

// Get returns the n'th action
func (a *Resource) Get(i int) *utils.ARN {
  if a==nil || i<0 || i>= len(a.resources) {
    return utils.NilARN()
  }
  return &(a.resources[i])
}

// ForEach invoke a function for each action
func (a *Resource) ForEach( f func(int,utils.ARN) error ) error {
  if a != nil {
    for i,e := range a.resources {
      err := f(i,e)
      if err != nil {
        return err
      }
    }
  }

  return nil
}

// IsNegate returns true if this is a NotAction rather than Action block
func (a *Resource) IsNegate() bool {
  return a!=nil && a.negate
}

func (a *Resource) UnmarshalJSON( b []byte ) error {

  if b == nil || len(b) < 3 {
    return nil
  }
  bl := len(b)

  if b[0]=='"' && b[bl-1]=='"' {
    var s utils.ARN
    err := json.Unmarshal( b, &s )
    if err != nil {
      return err
    }
    a.resources = append( a.resources, s )
  } else if b[0]=='[' && b[bl-1]==']' {
    var s []utils.ARN
    err := json.Unmarshal( b, &s )
    if err != nil {
      return err
    }
    for _, e := range s {
      a.resources = append( a.resources, e )
    }
  }
  return nil
}

func (a *Resource) MarshalJSON() ( []byte, error ) {
  // nil or empty then marshal null
  if a.IsNil() {
    return []byte("null"), nil
  }

  // single entry as a string
  if a.Len() == 1 {
    return json.Marshal( &(a.resources[0]) )
  }

  // normal json array
  return json.Marshal( a.resources )
}
