package policy

import (
  "encoding/json"
  "strings"
  "testing"
)

func TestEffect_Allowed_Denied( t *testing.T ) {
  // Test a is allowed if eb true, denied if false
  test := func(a *Effect, eb bool) {
    if a.Allowed() != eb {
      t.Errorf( "Expected Allowed %v got %v on %s", eb, a.Allowed(), a )
    }
    if a.Denied() == eb {
      t.Errorf( "Expected Denied %v got %v on %s", !eb, a.Denied(), a )
    }
  }

  test( nil, false )

  var a Effect
  test( &a, false )

  a = true
  test( &a, true )
}

// Test we can unmarshal actions formed of a single string
//
// "s3:*"
// "ec2:StartInstances"
// "cloudformation:*"
// "*"
func TestEffect_Unmarshal( t *testing.T ) {
  for i, src := range []string{
    "\"Allow\"",
    "null",
    "\"Deny\"",
    "\"Wibble\"",
  } {
    var action Effect
    err := json.Unmarshal( []byte(src), &action )
    if err != nil {
      // Invalid Effect is valid
      if !strings.HasPrefix( err.Error(), "Invalid Effect:" ) {
        t.Fatal( err )
      }
    }

    allowed := i == 0
    if action.Allowed() != allowed {
      t.Errorf( "Expected Allowed %v from %s got %v on %v", allowed, src, action.Allowed(), action )
    }
    if action.Denied() == allowed {
      t.Errorf( "Expected Denied %v from %s got %v on %v", !allowed, src, action.Denied(), action )
    }
  }
}

func TestEffect_Marshal( t *testing.T ) {
  const deny = "\"Deny\""
  const allow = "\"Allow\""

  var a Effect

  test := func( e string, eb bool ) {
    b, err := json.Marshal( &a )
    if err != nil {
      t.Fatal( err )
    }
    test_expectByte( t, b, e )

    if a.Allowed() != eb {
      t.Errorf( "Expected Allowed %v from %v got %v", eb, a, a.Allowed() )
    }
    if a.Denied() == eb {
      t.Errorf( "Expected Denied %v from %v got %v", !eb, a, a.Denied() )
    }
  }

  test( deny, false )

  a = false
  test( deny, false )

  a = true
  test( allow, true )
}
