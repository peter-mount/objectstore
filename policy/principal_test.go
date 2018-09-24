package policy

import (
  "testing"
)

func TestPrincipal_Unmarshal_Principal( t *testing.T ) {

  const src = "{\"Version\": \"2012-10-17\",\"Statement\": [{\"Sid\": \"Test\",\"Effect\": \"Deny\",\"Principal\": {\"AWS\": [\"arn:aws:iam::123456789012:root\",\"arn:aws:iam::123456789012:user/Bob\"]},\"Action\": \"s3:*\",\"Resource\": \"arn:aws:s3:::examplebucket/*\",\"Condition\": {\"StringEquals\": {\"s3:signatureversion\": \"AWS4-HMAC-SHA256\"}}}]}"
  policy := &Policy{}
  testPolicy_Unmarshal( t, policy, src, "\"Principal" )
}

func TestPrincipal_Unmarshal_NotPrincipal( t *testing.T ) {

  const src = "{\"Version\": \"2012-10-17\",\"Statement\": [{\"Sid\": \"Test\",\"Effect\": \"Deny\",\"NotPrincipal\": {\"AWS\": [\"arn:aws:iam::123456789012:root\",\"arn:aws:iam::123456789012:user/Bob\"]},\"Action\": \"s3:*\",\"Resource\": \"arn:aws:s3:::examplebucket/*\",\"Condition\": {\"StringEquals\": {\"s3:signatureversion\": \"AWS4-HMAC-SHA256\"}}}]}"
  policy := &Policy{}
  testPolicy_Unmarshal( t, policy, src, "\"NotPrincipal" )
}
