package policy

import (
  "bytes"
  "encoding/json"
  "fmt"
  "github.com/peter-mount/objectstore/utils"
)

// A Principal
type Principal struct {
  principal map[string]PrincipalMap
  negate    bool
}

// A Principal Map
type PrincipalMap []utils.ARN

// ForEach invoke a function for each action
func (a *Principal) ForEach( f func(string,PrincipalMap) error ) error {
  if a != nil {
    for k,v := range a.principal {
      err := f(k,v)
      if err != nil {
        return err
      }
    }
  }

  return nil
}

// IsNegate returns true if this is a NotAction rather than Action block
func (a *Principal) IsNegate() bool {
  return a!=nil && a.negate
}

func (a *Principal) UnmarshalJSON( b []byte ) error {
  a.principal = make( map[string]PrincipalMap )

  // Anonymous "*" -> {"AWS":"*"}
  if bytes.Equal( b, []byte("\"*\"") ) {
    b = []byte( "{\"AWS\":\"*\"}" )
  }

  var m map[string]interface{}

  err := json.Unmarshal( b, &m )
  if err != nil {
    return err
  }

  for k, v := range m {
    pm := &PrincipalMap{}
    err := pm.append( v )
    if err != nil {
      return err
    }
    a.principal[k] = *pm
  }

  return nil
}

func (a *PrincipalMap) append( v interface{} ) error {
  if s, ok := v.(string); ok {
    arn, err := utils.ParseARN( s )
    if err != nil {
      return err
    }
    *a = append( *a, *arn )
  } else if ar, ok := v.([]interface{}); ok {
    for _, e := range ar {
      err := a.append( e )
      if err != nil {
        return err
      }
    }
  } else {
    return fmt.Errorf( "Unsupported entry %v", v )
  }
  return nil
}

func (a *Principal) MarshalJSON() ( []byte, error ) {
  if a == nil {
    return []byte("null"), nil
  }

  // Special case, single Anonymous entry
  if aws, ok := a.principal["AWS"]; ok && len( aws ) == 1 && aws[0].IsAnonymous() {
    return []byte( "\"*\""), nil
  }

  buffer := bytes.NewBufferString("{")

  sep := false
  for k, v := range a.principal {
    if sep {
      buffer.WriteString( "," )
    } else {
      sep = true
    }

    b, err := json.Marshal( k )
    if err != nil {
      return nil, err
    }
    buffer.Write( b )
    buffer.WriteString( ":" )

    b, err = json.Marshal( &v )
    if err != nil {
      return nil, err
    }
    buffer.Write( b )
  }

  buffer.WriteString( "}")
  return buffer.Bytes(), nil
}

func (a *PrincipalMap) MarshalJSON() ( []byte, error ) {
  if len(*a) == 1 {
    return json.Marshal( &(*a)[0] )
  }

  var ar []utils.ARN
  for _, e := range *a {
    ar = append( ar, e )
  }
  return json.Marshal( ar )
}
