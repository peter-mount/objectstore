package utils

import (
  "encoding/json"
  "testing"
)

func TestARN_Unmarshal( t *testing.T ) {
  for i, src := range []string{
    "arn:aws:iam::123456789012:root",
    "arn:aws:iam::123456789012:user/Bob",
    "arn:aws:iam::123456789012:user/division_abc/subdivision_xyz/Bob",
    "arn:aws:iam::123456789012:group/Developers",
    "arn:aws:iam::123456789012:group/division_abc/subdivision_xyz/product_A/Developers",
    "arn:aws:iam::123456789012:role/S3Access",
    "arn:aws:iam::123456789012:policy/ManageCredentialsPermissions",
    "arn:aws:iam::123456789012:instance-profile/Webserver",
    "arn:aws:sts::123456789012:federated-user/Bob",
    "arn:aws:sts::123456789012:assumed-role/Accounting-Role/Mary",
    "arn:aws:iam::123456789012:mfa/Bob",
    "arn:aws:iam::123456789012:server-certificate/ProdServerCert",
    "arn:aws:iam::123456789012:server-certificate/division_abc/subdivision_xyz/ProdServerCert",
    "arn:aws:iam::123456789012:saml-provider/ADFSProvider",
    "arn:aws:iam::123456789012:oidc-provider/GoogleProvider",
    // Anonymous
    "*",
    // User Id only
    "123456789012",
  } {
    a := &ARN{}
    err := json.Unmarshal( []byte("\"" + src + "\""), a )
    if err != nil {
      t.Fatal( err )
    }

    s := a.String()
    if src != s {
      t.Errorf( "%d:Expected %s got %v", i, src, s )
    }
  }
}

func TestARN_IsAnonymous( t *testing.T ) {
  a := &ARN{}
  a.Parse( "*" )
  if !a.IsAnonymous() {
    t.Errorf( "Expected IsAnonymous got %v for %s", a.IsAnonymous(), a )
  }
  if a.IsUserId() {
    t.Errorf( "Expected !IsUserId got %v for %s", a.IsUserId(), a )
  }
  if a.IsNil() {
    t.Errorf( "Expected !IsNil got %v for %s", a.IsNil(), a )
  }
}

func TestARN_IsUserId( t *testing.T ) {
  a := &ARN{}
  a.Parse( "12345shdfgfhsdfr" )
  if a.IsAnonymous() {
    t.Errorf( "Expected !IsAnonymous got %v for %s", a.IsAnonymous(), a )
  }
  if !a.IsUserId() {
    t.Errorf( "Expected IsUserId got %v for %s", a.IsUserId(), a )
  }
  if a.IsNil() {
    t.Errorf( "Expected !IsNil got %v for %s", a.IsNil(), a )
  }
}

func TestARN_IsNil( t *testing.T ) {
  for i := 0; i < 2; i++ {
    var a *ARN

    // i==0 then null, i==1 then instance with ""
    // for both IsNil() should return true
    if i == 1 {
      a = &ARN{}
      a.Parse( "" )
    }
    
    if a.IsAnonymous() {
      t.Errorf( "%d:Expected !IsAnonymous got %v for %s", i, a.IsAnonymous(), a )
    }
    if a.IsUserId() {
      t.Errorf( "%d:Expected !IsUserId got %v for %s", i, a.IsUserId(), a )
    }
    if !a.IsNil() {
      t.Errorf( "%d:Expected IsNil got %v for %s", i, a.IsNil(), a )
    }
  }
}
