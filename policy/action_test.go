package policy

import (
  "bytes"
  "encoding/json"
  "strings"
  "testing"
)

func test_marshall( t *testing.T, a interface{}, e string ) {
  b, err := json.Marshal( a )
  if err != nil {
    t.Fatal( err )
  }
  test_expectByte( t, b, e )
}

func test_expectByte( t *testing.T, b []byte, e string ) {
  if !bytes.Equal( b, []byte(e) ) {
    t.Errorf( "Expected %s got %s", e, b )
  }
}

func TestAction_Unmarshal_IsNil( t *testing.T ) {
  test := func(a *Action) {
    if !a.IsNil() {
      t.Errorf( "Expected true got %v on %v", a.IsNil(), a )
    }
  }
  test( nil )
  test( &Action{} )
}

// Test we can unmarshal actions formed of a single string
//
// "s3:*"
// "ec2:StartInstances"
// "cloudformation:*"
// "*"
func TestAction_Unmarshal_null( t *testing.T ) {
  for _, src := range []string{
    "null",
    "[]",
    "\"\"",
    "[\"\"]",
    "[\"\",\"\"]",
  } {
    action := Action{}
    err := json.Unmarshal( []byte(src), &action )
    if err != nil {
      t.Fatal( err )
    }
    if !action.IsNil() {
      t.Errorf( "Expected true from %s got %v on %v", src, action.IsNil(), action )
    }
  }
}

// Test we can unmarshal actions formed of a single string
//
// "s3:*"
// "ec2:StartInstances"
// "cloudformation:*"
// "*"
func TestAction_Unmarshal_single( t *testing.T ) {
  for _, expected := range []string{
    "s3:*",
    "ec2:StartInstances",
    "cloudformation:*",
    "*",
  } {
    src := "\"" + expected + "\""

    action := Action{}
    err := json.Unmarshal( []byte(src), &action )
    if err != nil {
      t.Fatal( err )
    }

    if len( action.actions ) != 1 {
      t.Errorf( "Expected 1 action got %d for %s", len( action.actions ), src )
      } else if action.actions[0] != expected {
        t.Errorf( "Expected %s for %s got %v", expected, src, action )
      }
  }
}

// Test we can unmarshal actions formed of a slice of strings string
//
// ["ec2:StartInstances","ec2:StopInstances"]
// ["s3:Get*","s3:List*"]
func TestAction_Unmarshal_slice( t *testing.T ) {
  for _, expected := range [][]string{
    []string{"ec2:StartInstances","ec2:StopInstances"},
    []string{"s3:Get*","s3:List*"},
  } {
    src := "[\"" + strings.Join( expected, "\",\"" ) + "\"]"

    action := Action{}
    err := json.Unmarshal( []byte(src), &action )
    if err != nil {
      t.Fatal( err )
    }

    if len( action.actions ) != len( expected ) {
      t.Errorf( "Expected %d action got %d for %s", len( expected ), len( action.actions ), src )
    } else {
      for i, e := range expected {
        if action.actions[i] != e {
          t.Errorf( "%d: Expected %s got %s", i, e, action.actions[i] )
        }
      }
    }
  }
}

// Test we can marshal values
func TestAction_Marshal( t *testing.T ) {
  var a *Action

  // null Action
  test_marshall( t, a, "null" )

  // empty Action
  a = &Action{}
  test_marshall( t, a, "null" )

  // empty Action
  a = NewAction()
  test_marshall( t, a, "null" )

  // Single element Action
  a = NewAction( "s3:*" )
  test_marshall( t, a, "\"s3:*\"" )

  // multiple element Action
  a = NewAction( "s3:*", "*" )
  test_marshall( t, a, "[\"s3:*\",\"*\"]" )

  a = NewAction( "s3:*", "*", "cloudformation:*" )
  test_marshall( t, a, "[\"s3:*\",\"*\",\"cloudformation:*\"]" )
}
