package condition

import (
  "bytes"
  "encoding/json"
)

// Condition
type Condition map[string]ConditionType

func (a *Condition) UnmarshalJSON( b []byte ) error {
  m := make( map[string]ConditionType )

  err := json.Unmarshal( b, &m )
  if err != nil {
    return err
  }

  *a = m
  return nil
}

func (a *Condition) MarshalJSON() ( []byte, error ) {
  var b bytes.Buffer
  b.WriteString( "{")
  for k, v := range *a {
    vb, err := json.Marshal( k )
    if err != nil {
      return nil, err
    }
    b.Write( vb )

    b.WriteString( ":" )

    vb, err = v.MarshalJSON()
    if err != nil {
      return nil, err
    }
    b.Write( vb )
  }
  b.WriteString( "}")
  return b.Bytes(), nil
}
