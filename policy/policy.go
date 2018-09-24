package policy

const (
  // Current policy version
  // https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_version.html
  VERSION_CURRENT = "2012-10-17"
  // Earlier version
  VERSION_OLDER   = "2008-10-17"
)

// An AWS policy
// Grammar: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_grammar.html
type Policy struct {
  // Policy version, see VERSION_CURRENT or VERSION_OLDER
  Version     string      `json:"Version" xml:"Version" yaml:"Version"`
  // optional Policy ID (SQS or SNS use this, S3 doesn't)
  Id          string      `json:"Id,omitempty" xml:"Id,omitempty" yaml:"Id"`
  Statement []Statement   `json:"Statement" xml:"Statement" yaml:"Statement"`
}
