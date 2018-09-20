package policy

import (
  "encoding/json"
  "github.com/peter-mount/objectstore/utils"
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
  for _, src := range []string{ "null", "[]" } {
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
    "arn:aws:iam::123456789012:root",
    "arn:aws:iam::123456789012:user/Bob",
    "arn:aws:iam::123456789012:user/division_abc/subdivision_xyz/Bob",
    "arn:aws:iam::123456789012:group/Developers",
  } {
    src := "\"" + expected + "\""

    arn, err := utils.ParseARN(expected)
    if err != nil {
      t.Error( err )
    } else {
      r := Resource{}
      err = json.Unmarshal( []byte(src), &r )
      if err != nil {
        t.Fatal( err )
      }

      if len( r ) != 1 {
        t.Errorf( "Expected 1 action got %d for %s", len( r ), src )
      } else if !arn.Equal( &r[0] ) {
        t.Errorf( "Expected %s for %s got %v", expected, src, r )
      }
    }
  }
}

// Test we can unmarshal actions formed of a slice of strings string
//
func TestResource_Unmarshal_slice( t *testing.T ) {
  for _, expected := range [][]string{
    []string{
      "arn:aws:iam::123456789012:root",
      "arn:aws:iam::123456789012:user/Bob",
    },
    []string{
      "arn:aws:iam::123456789012:user/division_abc/subdivision_xyz/Bob",
      "arn:aws:iam::123456789012:group/Developers",
    },
  } {
    src := "[\"" + strings.Join( expected, "\",\"" ) + "\"]"

    r := Resource{}
    err := json.Unmarshal( []byte(src), &r )
    if err != nil {
      t.Fatal( err )
    }

    if len( r ) != len( expected ) {
      t.Errorf( "Expected %d action got %d for %s", len( expected ), len( r ), src )
    } else {
      for i, e := range expected {
        arn, err := utils.ParseARN(e)
        if err != nil {
          t.Fatal( err )
        }

        if !arn.Equal( &r[i] ) {
          t.Errorf( "%d: Expected %s got %s", i, e, r[i] )
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
  arn0, err := utils.ParseARN( "arn:aws:iam::123456789012:root" )
  if err != nil {
    t.Fatal( err )
  }
  a = &Resource{ *arn0 }
  test_marshall( t, a, "\"arn:aws:iam::123456789012:root\"" )

  // multiple element Resource
  arn1, err := utils.ParseARN( "arn:aws:iam::123456789012:user/Bob" )
  if err != nil {
    t.Fatal( err )
  }
  arn2, err := utils.ParseARN( "arn:aws:iam::123456789012:user/division_abc/subdivision_xyz/Bob" )
  if err != nil {
    t.Fatal( err )
  }
  a = &Resource{ *arn1, *arn2 }
  test_marshall( t, a, "[\"arn:aws:iam::123456789012:user/Bob\",\"arn:aws:iam::123456789012:user/division_abc/subdivision_xyz/Bob\"]" )
}
