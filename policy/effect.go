package policy

import (
  "bytes"
  "fmt"
)

// The effect of a Policy.
// If not defined this will default to denied.
// https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_effect.html
type Effect bool

// True if the effect is to allow some action
func (a *Effect) Allowed() bool {
  return a != nil && *a == true
}

// True if the effect is to deny some action
func (a *Effect) Denied() bool {
  return a == nil || *a == false
}

func (a *Effect) String() string {
  if a == nil {
    return "nil"
  }
  if a.Denied() {
    return "Denied"
  }
  return "Allowed"
}

func (a *Effect) UnmarshalJSON( b []byte ) error {
  if bytes.Equal( b, []byte("null") )  || bytes.Equal( b, []byte("\"Deny\"") ) {
    *a = false
  } else if bytes.Equal( b, []byte("\"Allow\"") ) {
    *a = true
  } else {
    return fmt.Errorf( "Invalid Effect: %s", b )
  }
  return nil
}

func (a *Effect) MarshalJSON() ( []byte, error ) {
  if a == nil || !*a {
    return []byte("\"Deny\""), nil
  }
  return []byte("\"Allow\""), nil
}
