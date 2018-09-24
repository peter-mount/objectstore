package event

import (
  "github.com/peter-mount/objectstore/utils"
  "time"
)

// Records is the wrapper around events.
type Records struct {
  Records          []*Event             `json:Records`
}

type Event struct {
  // "2.0"
  Version             string            `json:"eventVersion"`
  // "aws:s3"
  Source              string            `json:"eventSource"`
  // "us-east-1"
  Region              string            `json:awsRegion`
  // The time, in ISO-8601 format, for example, 1970-01-01T00:00:00.000Z,
  // when S3 finished processing the request
  Time                time.Time         `json:"eventTime"`
  // The event name, e.g. "ObjectCreated:Put"
  Name                string            `json:"eventName"`
  // Amazon-customer-ID-of-the-user-who-caused-the-event
  Identity            Identity          `json:"userIdentity"`
  RequestParameters   map[string]string `json:"requestParameters"`
  // Response parameters, normally x-amz-request-id & x-amz-id-2
  ResponseElements    map[string]string `json:"responseElements"`
  S3                 *S3                `json:"s3,omitifempty"`
}

type Identity struct {
  PrincipalId       string            `json:"principalId"`
}

type S3 struct {
  // "1.0"
  Version           string            `json:"s3SchemaVersion"`
  ConfigId          string            `json:"configurationId"`
  Bucket            S3Bucket          `json:"bucket"`
  Object           *S3Object          `json:"object,omitifempty"`
}

type S3Bucket struct {
  Name              string            `json:"Bucket"`
  Identity          Identity          `json:"ownerIdentity"`
  Arn              *utils.ARN         `json:"arn,omitifempty"`
}

type S3Object struct {
  Key             string            `json:"key"`
  Size            int               `json:"size"`
  ETag            string            `json:"eTag"`
  VersionId      *string            `json:"versionId"`
  Sequencer       string            `json:"sequencer"`
}
