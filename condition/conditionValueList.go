package condition

import (
  "encoding/json"
)

type ConditionValueList []ConditionValue

func (a *ConditionValueList) UnmarshalJSON( b []byte ) error {

  if b == nil {
    return nil
  }

  bl := len(b)
  if b[0]=='"' && b[bl-1]=='"' {
    var s ConditionValue
    err := json.Unmarshal( b, &s )
    if err != nil {
      return err
    }
    *a = append( *a, s )
  } else if b[0]=='[' && b[bl-1]==']' {
    var s []ConditionValue
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

func (a *ConditionValueList) MarshalJSON() ( []byte, error ) {
  // nil or empty then marshal null
  if a == nil || len(*a) == 0 {
    return []byte("null"), nil
  }

  // single entry as a string
  if len(*a) == 1 {
    return json.Marshal( &(*a)[0] )
  }

  // normal json array
  var v []ConditionValue
  for _, e := range *a {
    v = append( v, e )
  }
  if len( v ) == 0 {
    return []byte("null"), nil
  }
  return json.Marshal( &v )
}
