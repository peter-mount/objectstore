package condition

import (
  "encoding/json"
)

type ConditionValue struct {
  t int
  s string
  i int
  f float64
  b bool
}

const (
  VAL_NIL = iota
  VAL_STRING
  VAL_INT
  VAL_FLOAT
  VAL_BOOL
)

func (a *ConditionValue) UnmarshalJSON( b []byte ) error {

  if b == nil || len(b) < 3 {
    a.t = VAL_NIL
    return nil
  }

  var s string
  err := json.Unmarshal( b, &s )
  if err != nil {
    return err
  }

  // All values have the string so we can use it to unmarshal
  a.s = s

  // TODO process here, for now just string
  a.t = VAL_STRING

  return nil
}

func (a *ConditionValue) MarshalJSON() ( []byte, error ) {
  // nil or empty then marshal null
  if a == nil || a.t == VAL_NIL {
    return []byte("null"), nil
  }

  // Marshal the string value
  return json.Marshal( &(a.s) )
}
