package policy

import (
  "encoding/json"
)

// One or more actions to invoke.
//
// In JSON form this is either null, a single string or an array of strings
// as defined in https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_grammar.html
//
// One difference here, if during json marshal:
// * If the action is nil,  empty or contains just "" entries then null is used.
// * If the action is a single non "" entry then that entry as a string is used.
// * Otherwise a json array is used.
//
// During unmarshalling we do the same and strip out any "" entries.
//
type Action struct {
  actions []string
  negate    bool
}

func NewAction( s ...string ) *Action {
  return &Action{ actions: s }
}

func NewNotAction( s ...string ) *Action {
  return &Action{ actions: s, negate: true }
}

// IsNil returns true if the action is empty
func (a *Action) IsNil() bool {
  return a == nil || len( a.actions ) == 0
}

// Len returns the number of actions
func (a *Action) Len() int {
  if a==nil {
    return 0
  }
  return len(a.actions)
}

// Get returns the n'th action
func (a *Action) Get(i int) string {
  if a==nil || i<0 || i>= len(a.actions) {
    return ""
  }
  return a.actions[i]
}

// ForEach invoke a function for each action
func (a *Action) ForEach( f func(int,string) error ) error {
  if a != nil {
    for i,e := range a.actions {
      err := f(i,e)
      if err != nil {
        return err
      }
    }
  }

  return nil
}

// IsNegate returns true if this is a NotAction rather than Action block
func (a *Action) IsNegate() bool {
  return a!=nil && a.negate
}

func (a *Action) UnmarshalJSON( b []byte ) error {

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
    a.actions = append( a.actions, s )
  } else if b[0]=='[' && b[bl-1]==']' {
    var s []string
    err := json.Unmarshal( b, &s )
    if err != nil {
      return err
    }
    for _, e := range s {
      if e != "" {
        a.actions = append( a.actions, e )
      }
    }
  }
  return nil
}

func (a *Action) MarshalJSON() ( []byte, error ) {
  // nil or empty then marshal null
  if a == nil || len(a.actions) == 0 {
    return []byte("null"), nil
  }

  // single entry as a string
  if len(a.actions) == 1 {
    s := (a.actions)[0]
    if s == "" {
      return []byte("null"), nil
    }
    return json.Marshal( s )
  }

  // normal json array
  var v []string
  for _, e := range a.actions {
    if e != "" {
      v = append( v, e )
    }
  }
  if len( v ) == 0 {
    return []byte("null"), nil
  }
  return json.Marshal( v )
}
