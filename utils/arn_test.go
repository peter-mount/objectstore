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
      t.Errorf( "%d:Expected %s got %s", i, src, s )
    }
  }
}
