package policy

import (
  "encoding/json"
)

// A Resource
type Resource []string

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
    var s string
    err := json.Unmarshal( b, &s )
    if err != nil {
      return err
    }
    *a = append( *a, s )
  } else if b[0]=='[' && b[bl-1]==']' {
    var s []string
    err := json.Unmarshal( b, &s )
    if err != nil {
      return err
    }
    for _, e := range s {
      if e != "" {
        *a = append( *a, e )
      }
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
    s := (*a)[0]
    if s == "" {
      return []byte("null"), nil
    }
    return json.Marshal( s )
  }

  // normal json array
  var v []string
  for _, e := range *a {
    if e != "" {
      v = append( v, e )
    }
  }
  if len( v ) == 0 {
    return []byte("null"), nil
  }
  return json.Marshal( v )
}
