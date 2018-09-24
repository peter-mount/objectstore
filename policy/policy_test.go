package policy

import (
  "encoding/json"
  "strings"
  "testing"
)

// Tests unmarshal then marshal to see if some string is present.
// Usually this is something like \"Action\" or \"NotAction" - note the quotes!
// t Test environment
// src Source JSON
// expected Expected string, "" to just unmarshal src
func testPolicy_Unmarshal( t *testing.T, v interface{}, src, expected string ) {

  err := json.Unmarshal( []byte(src), v )
  if err != nil {
    t.Fatal( err )
  }

  if expected != "" {
    b, err := json.MarshalIndent( v, "", "  " )
    if err != nil {
      t.Fatal( err )
    }

    bs := string(b)
    if !strings.Contains( bs, expected ) {
      t.Errorf( "Expected %s in unmarshalled JSON:\n%s", expected, bs )
    }
  }
}

func TestPolicyUnmarshal_1( t *testing.T ) {
  const src = "{\"Version\": \"2012-10-17\",\"Statement\": [{\"Sid\": \"Test\",\"Effect\": \"Deny\",\"Principal\": \"*\",\"Action\": \"s3:*\",\"Resource\": \"arn:aws:s3:::examplebucket/*\",\"Condition\": {\"StringEquals\": {\"s3:signatureversion\": \"AWS4-HMAC-SHA256\"}}}]}"

  policy := &Policy{}
  testPolicy_Unmarshal( t, policy, src, "" )

  if len( policy.Statement ) != 1 {
    t.Errorf( "Expected 1 statement got %d", len( policy.Statement ) )
  } else {
    for i, stmt := range policy.Statement {

      if stmt.Action.Len() != 1 {
        t.Errorf( "%d: Expected 1 action got %d", i, stmt.Action.Len() )
      } else  if stmt.Action.Len() == 1 && stmt.Action.Get(0) != "s3:*" {
        t.Errorf( "%d: Expected s3:* action got %v", i, stmt.Action )
      }

    }
  }
}

func TestPolicyUnmarshal_2( t *testing.T ) {
  const src = "{\"Version\": \"2012-10-17\",\"Statement\": [{\"Sid\": \"Test\",\"Effect\": \"Deny\",\"Principal\": {\"AWS\": \"*\"},\"Action\": \"s3:*\",\"Resource\": \"arn:aws:s3:::examplebucket/*\",\"Condition\": {\"StringEquals\": {\"s3:signatureversion\": \"AWS4-HMAC-SHA256\"}}}]}"

  policy := &Policy{}
  testPolicy_Unmarshal( t, policy, src, "" )

  if len( policy.Statement ) != 1 {
    t.Errorf( "Expected 1 statement got %d", len( policy.Statement ) )
  } else {
    for i, stmt := range policy.Statement {

      if stmt.Action.Len() != 1 {
        t.Errorf( "%d: Expected 1 action got %d", i, stmt.Action.Len() )
      } else  if stmt.Action.Len() == 1 && stmt.Action.Get(0) != "s3:*" {
        t.Errorf( "%d: Expected s3:* action got %v", i, stmt.Action )
      }

    }
  }
}

func TestPolicyUnmarshal_3( t *testing.T ) {
  const src = "{\"Version\": \"2012-10-17\",\"Statement\": [{\"Sid\": \"Test\",\"Effect\": \"Deny\",\"Principal\": {\"AWS\": \"arn:aws:iam::123456789012:root\"},\"Action\": \"s3:*\",\"Resource\": \"arn:aws:s3:::examplebucket/*\",\"Condition\": {\"StringEquals\": {\"s3:signatureversion\": [\"AWS4-HMAC-SHA256\",\"ANOTHER\"]}}}]}"

  policy := &Policy{}
  testPolicy_Unmarshal( t, policy, src, "" )

  if len( policy.Statement ) != 1 {
    t.Errorf( "Expected 1 statement got %d", len( policy.Statement ) )
  } else {
    for i, stmt := range policy.Statement {

      if stmt.Action.Len() != 1 {
        t.Errorf( "%d: Expected 1 action got %d", i, stmt.Action.Len() )
      } else  if stmt.Action.Len() == 1 && stmt.Action.Get(0) != "s3:*" {
        t.Errorf( "%d: Expected s3:* action got %v", i, stmt.Action )
      }

    }
  }
}

func TestPolicyUnmarshal_4( t *testing.T ) {

  const src = "{\"Version\": \"2012-10-17\",\"Statement\": [{\"Sid\": \"Test\",\"Effect\": \"Deny\",\"Principal\": {\"AWS\": [\"arn:aws:iam::123456789012:root\",\"arn:aws:iam::123456789012:user/Bob\"]},\"Action\": \"s3:*\",\"Resource\": \"arn:aws:s3:::examplebucket/*\",\"Condition\": {\"StringEquals\": {\"s3:signatureversion\": \"AWS4-HMAC-SHA256\"}}}]}"

  policy := &Policy{}
  testPolicy_Unmarshal( t, policy, src, "" )

  if len( policy.Statement ) != 1 {
    t.Errorf( "Expected 1 statement got %d", len( policy.Statement ) )
  } else {
    for i, stmt := range policy.Statement {

      if stmt.Action.Len() != 1 {
        t.Errorf( "%d: Expected 1 action got %d", i, stmt.Action.Len() )
      } else  if stmt.Action.Len() == 1 && stmt.Action.Get(0) != "s3:*" {
        t.Errorf( "%d: Expected s3:* action got %v", i, stmt.Action )
      }

    }
  }
}
