package policy

import (
  "encoding/json"
  "github.com/peter-mount/objectstore/utils"
)

// A Resource
type Resource []utils.ARN

// IsNil returns true if the Resource is empty
func (a *Resource) IsNil() bool {
  return a == nil || len( *a ) == 0
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
    *a = append( *a, s )
  } else if b[0]=='[' && b[bl-1]==']' {
    var s []utils.ARN
    err := json.Unmarshal( b, &s )
    if err != nil {
      return err
    }
    for _, e := range s {
      *a = append( *a, e )
    }
  }
  return nil
}

func (a *Resource) MarshalJSON() ( []byte, error ) {
  // nil or empty then marshal null
  if a == nil || len(*a) == 0 {
    return []byte("null"), nil
  }

  // single entry as a string
  if len(*a) == 1 {
    return json.Marshal( &(*a)[0] )
  }

  // normal json array
  var v []utils.ARN
  for _, e := range *a {
    v = append( v, e )
  }
  return json.Marshal( v )
}
