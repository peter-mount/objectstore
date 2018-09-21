package condition

import (
  "bytes"
  "encoding/json"
)

type ConditionType map[string]ConditionValueList

func (a *ConditionType) UnmarshalJSON( b []byte ) error {
  m := make( map[string]ConditionValueList )

  err := json.Unmarshal( b, &m )
  if err != nil {
    return err
  }

  *a = m
  return nil
}

func (a *ConditionType) MarshalJSON() ( []byte, error ) {
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
