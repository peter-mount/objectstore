package policy

import (
  "bytes"
  "encoding/json"
  "fmt"
  "github.com/peter-mount/objectstore/condition"
)

type Statement struct {
  // Optional SID
  Sid           string
  // Effect required
  Effect        Effect
  // Optional Principal/NotPrincipal
  Principal     Principal
  // Required Action/NotAction
  Action        Action
  // Required resource block
  Resource      Resource
  // Condition optional
  Condition     condition.Condition
}

func (a *Statement) UnmarshalJSON( b []byte ) error {

  if b == nil || len(b) < 3 {
    return nil
  }

  // Unmarshal into a map of RawMessage's
  m := make(map[string]json.RawMessage)
  err := json.Unmarshal( b, &m )
  if err != nil {
    return err
  }

  // Now run through the keys & unmarshal into each one
  for k, v := range m {
    switch k {
      case "Sid":
        err = json.Unmarshal( v, &a.Sid)
      case "Effect":
        err = json.Unmarshal( v, &a.Effect)
      case "Principal":
        err = json.Unmarshal( v, &a.Principal )
      case "Action":
        err = json.Unmarshal( v, &a.Action)
      case "NotAction":
        err = json.Unmarshal( v, &a.Action)
        a.Action.negate = true
      case "Resource":
        err = json.Unmarshal( v, &a.Resource)
      case "Condition":
        err = json.Unmarshal( v, &a.Condition)
    }
    if err != nil {
      return fmt.Errorf( "Failed to unmarshal %s: %v", k, err )
    }
  }

  return nil
}

func (a *Statement) MarshalJSON() ( []byte, error ) {
  if a == nil {
    return []byte("null"), nil
  }

  buffer := bytes.NewBufferString("{")

  // util to marshal a single value
  wv := func( v interface{} ) error {
    b, err := json.Marshal( v )
    if err != nil {
      return err
    }
    buffer.Write( b )
    return nil
  }

  // IsNegate interface for Action, Principal & Resource so if the entry is
  // negated then prefix key with "Not"
  type negatable interface {
    IsNegate() bool
  }

  // Write a key, value in the statement map/
  f := false
  w := func(k string, v interface{} ) error {

    // Add "," separator betweek key/value pairs
    if f {
      buffer.WriteString( "," )
    } else {
      f = true
    }

    // Negated operation?
    if neg, ok := v.(negatable); ok && neg.IsNegate() {
      k = "Not" + k
    }
    err := wv(&k)
    if err != nil {
      return err
    }
    buffer.WriteString( ":" )

    return wv(v)
  }

  if err := w("Sid", &a.Sid); err != nil {
    return nil, err
  }

  if err := w("Effect", &a.Effect); err != nil {
    return nil, err
  }

  if err := w("Principal", &a.Principal); err != nil {
    return nil, err
  }

  if err := w("Action", &a.Action); err != nil {
    return nil, err
  }

  if err := w("Resource", &a.Resource); err != nil {
    return nil, err
  }

  if err := w("Condition", &a.Condition); err != nil {
    return nil, err
  }

  buffer.WriteString( "}")
  return buffer.Bytes(), nil
}
