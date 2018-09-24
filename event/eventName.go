package event

import (
  "github.com/peter-mount/go-glob"
  "github.com/peter-mount/sortfold"
  "strings"
)

func TestEventName( a, b string ) bool {
  if a == "" || b == "" {
    return false
  }

  // Wildcard so match all event names
  if a == "*" {
    return true
  }

  as := strings.Split( string(a), ":" )
  bs := strings.Split( string(b), ":" )
  if len(as) != len(bs) {
    return false
  }

  r := true
  for i := 0; i < len(as) && r; i++ {
    r = arnTest( as[i], bs[i] )
  }
  return r
}

func arnTest( a, b string ) bool {
  // Wildcard so match to true
  if a == "*" {
    return true
  }

  if strings.Contains( a, "*" ) {
    // Contains "*" so glob pattern
    return glob.Glob( a, b )
  }

  // Field is equal
  return sortfold.CompareFold( a, b ) == 0
}
