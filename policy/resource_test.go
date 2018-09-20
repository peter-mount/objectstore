package policy

import (
  "encoding/json"
  "strings"
  "testing"
)

func TestResource_Unmarshal_IsNil( t *testing.T ) {
  test := func(a *Resource) {
    if !a.IsNil() {
      t.Errorf( "Expected true got %v on %v", a.IsNil(), a )
    }
  }
  test( nil )
  test( &Resource{} )
}

// Test we can unmarshal actions formed of a single string
//
// "s3:*"
// "ec2:StartInstances"
// "cloudformation:*"
// "*"
func TestResource_Unmarshal_null( t *testing.T ) {
  for _, src := range []string{
    "null",
    "[]",
    "\"\"",
    "[\"\"]",
    "[\"\",\"\"]",
  } {
    action := Resource{}
    err := json.Unmarshal( []byte(src), &action )
    if err != nil {
      t.Fatal( err )
    }
    if !action.IsNil() {
      t.Errorf( "Expected true from %s got %v on %s", src, action.IsNil(), action )
    }
  }
}

// Test we can unmarshal actions formed of a single string
//
// "s3:*"
// "ec2:StartInstances"
// "cloudformation:*"
// "*"
func TestResource_Unmarshal_single( t *testing.T ) {
  for _, expected := range []string{
    "s3:*",
    "ec2:StartInstances",
    "cloudformation:*",
    "*",
  } {
    src := "\"" + expected + "\""

    action := Resource{}
    err := json.Unmarshal( []byte(src), &action )
    if err != nil {
      t.Fatal( err )
    }

    if len( action ) != 1 {
      t.Errorf( "Expected 1 action got %d for %s", len( action ), src )
      } else if action[0] != expected {
        t.Errorf( "Expected %s for %s got %v", expected, src, action )
      }
  }
}

// Test we can unmarshal actions formed of a slice of strings string
//
// ["ec2:StartInstances","ec2:StopInstances"]
// ["s3:Get*","s3:List*"]
func TestResource_Unmarshal_slice( t *testing.T ) {
  for _, expected := range [][]string{
    []string{"ec2:StartInstances","ec2:StopInstances"},
    []string{"s3:Get*","s3:List*"},
  } {
    src := "[\"" + strings.Join( expected, "\",\"" ) + "\"]"

    action := Resource{}
    err := json.Unmarshal( []byte(src), &action )
    if err != nil {
      t.Fatal( err )
    }

    if len( action ) != len( expected ) {
      t.Errorf( "Expected %d action got %d for %s", len( expected ), len( action ), src )
    } else {
      for i, e := range expected {
        if action[i] != e {
          t.Errorf( "%d: Expected %s got %s", i, e, action[i] )
        }
      }
    }
  }
}

// Test we can marshal values
func TestResource_Marshal( t *testing.T ) {
  var a *Resource

  // null Resource
  test_marshall( t, a, "null" )

  // empty Resource
  a = &Resource{}
  test_marshall( t, a, "null" )

  // Single element Resource
  a = &Resource{ "s3:*" }
  test_marshall( t, a, "\"s3:*\"" )

  // multiple element Resource
  a = &Resource{ "s3:*", "*" }
  test_marshall( t, a, "[\"s3:*\",\"*\"]" )

  a = &Resource{ "s3:*", "*", "cloudformation:*" }
  test_marshall( t, a, "[\"s3:*\",\"*\",\"cloudformation:*\"]" )
}
