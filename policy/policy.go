package policy

const (
  // Current policy version
  // https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_version.html
  VERSION_CURRENT = "2012-10-17"
  // Earlier version
  VERSION_OLDER   = "2008-10-17"
  // Statement.Effect
  EFFECT_ALLOW    = "Allow"
  EFFECT_DENY     = "Deny"
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

type Statement struct {
  // Optional SID
  Sid           string      `json:"Sid,omitempty" xml:"Sid,omitempty", yaml:"Sid"`
  // Effect required
  Effect        string      `json:"Effect,omitempty" xml:"Effect,omitempty", yaml:"Effect"`
  // Optional Principal/NotPrincipal
  // FIXME this needs principal_map adding
  Principal     string      `json:"Principal,omitempty" xml:"Principal,omitempty", yaml:"Principal"`
  NotPrincipal  string      `json:"NotPrincipal,omitempty" xml:"NotPrincipal,omitempty", yaml:"NotPrincipal"`
  // Required Action/NotAction
  // FIXME this needs either string or[]string
  Action        Action      `json:"Action,omitempty" xml:"Action,omitempty", yaml:"Action"`
  NotAction     Action      `json:"NotAction,omitempty" xml:"NotAction,omitempty", yaml:"NotAction"`
  // Required resource block
  Resource      Resource    `json:"Resource,omitempty" xml:"Resource,omitempty", yaml:"Resource"`
  NotResource   Resource    `json:"NotResource,omitempty" xml:"NotResource,omitempty", yaml:"NotResource"`
  // Condition optional
  Condition     map[string]map[string]interface{}
}
