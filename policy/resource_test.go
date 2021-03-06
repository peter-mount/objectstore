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
    resource := Resource{}
    err := json.Unmarshal( []byte(src), &resource )
    if err != nil {
      t.Fatal( err )
    }
    if !resource.IsNil() {
      t.Errorf( "Expected true from %s got %v on %v", src, resource.IsNil(), resource )
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

      if r.Len() != 1 {
        t.Errorf( "Expected 1 action got %d for %s", r.Len(), src )
      } else if !arn.Equal( r.Get(0) ) {
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

    if r.Len() != len( expected ) {
      t.Errorf( "Expected %d action got %d for %s", len( expected ), r.Len(), src )
    } else {
      for i, e := range expected {
        arn, err := utils.ParseARN(e)
        if err != nil {
          t.Fatal( err )
        }

        if !arn.Equal( r.Get(i) ) {
          t.Errorf( "%d: Expected %s got %s", i, e, r.Get(i) )
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
  a = NewResource()
  test_marshall( t, a, "null" )

  // Single element Resource
  arn0, err := utils.ParseARN( "arn:aws:iam::123456789012:root" )
  if err != nil {
    t.Fatal( err )
  }
  a = NewResource(*arn0)
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
  a = NewResource( *arn1, *arn2 )
  test_marshall( t, a, "[\"arn:aws:iam::123456789012:user/Bob\",\"arn:aws:iam::123456789012:user/division_abc/subdivision_xyz/Bob\"]" )
}

func TestResource_Unmarshal_Resource( t *testing.T ) {

  const src = "{\"Version\": \"2012-10-17\",\"Statement\": [{\"Sid\": \"Test\",\"Effect\": \"Deny\",\"Principal\": {\"AWS\": [\"arn:aws:iam::123456789012:root\",\"arn:aws:iam::123456789012:user/Bob\"]},\"Action\": \"s3:*\",\"Resource\": \"arn:aws:s3:::examplebucket/*\",\"Condition\": {\"StringEquals\": {\"s3:signatureversion\": \"AWS4-HMAC-SHA256\"}}}]}"
  policy := &Policy{}

  testPolicy_Unmarshal( t, policy, src, "\"Resource\"" )
}

func TestResource_Unmarshal_NotResource( t *testing.T ) {

  const src = "{\"Version\": \"2012-10-17\",\"Statement\": [{\"Sid\": \"Test\",\"Effect\": \"Deny\",\"Principal\": {\"AWS\": [\"arn:aws:iam::123456789012:root\",\"arn:aws:iam::123456789012:user/Bob\"]},\"Action\": \"s3:*\",\"NotResource\": \"arn:aws:s3:::examplebucket/*\",\"Condition\": {\"StringEquals\": {\"s3:signatureversion\": \"AWS4-HMAC-SHA256\"}}}]}"
  policy := &Policy{}

  testPolicy_Unmarshal( t, policy, src, "\"NotResource\"" )
}
