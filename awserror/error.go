package awserror

import (
	"encoding/xml"
  "fmt"
  "github.com/peter-mount/golib/rest"
  "net/http"
)

type Error struct {
  XMLName     xml.Name  `xml:"Error"`
  Status      int       `xml:"-"`
  Code        string    `xml:"Code"`
  Message     string    `xml:"Message"`
  Resource    string    `xml:"Resource"`
  RequestId   string    `xml:"RequestId"`
}

// Error enables us to use our error's as a go error
func (e *Error) Error() string {
	return fmt.Sprintf( "%d: %s %s", e.Status, e.Code, e.Message )
}

func (e *Error) Send( r *rest.Rest ) *rest.Rest {
  if e == nil {
    e = AccessDenied()
  }

  if e.Status == 0 {
    e.Status = http.StatusBadRequest
  }

  r.Status( e.Status ).
    XML().
    Value( e )

  return r
}

func AccessDenied() *Error {
  return &Error{
    Status:   http.StatusForbidden,
    Code:     "AccessDenied",
    Message:  "AccessDenied",
  }
}

func AllAccessDisabled() *Error {
  return &Error{
    Status:   http.StatusForbidden,
    Code:     "",
    Message:  "All access to this resource has been disabled.",
  }
}

func CredentialsNotSupported() *Error {
  return &Error{
    Status:   http.StatusForbidden,
    Code:     "CredentialsNotSupported",
    Message:  "This request does not support credentials.",
  }
}

func InvalidAccessKeyId() *Error {
  return &Error{
    Status:   http.StatusForbidden,
    Code:     "InvalidAccessKeyId",
    Message:  "The AWS access key ID you provided does not exist in our records.",
  }
}

func InvalidArgument( f string, a ...interface{}) *Error {
  if f == "" {
    f = "Invalid Argument"
  }
  return &Error{
    Status:   http.StatusBadRequest,
    Code:     "InvalidArgument",
    Message:  fmt.Sprintf( f, a... ),
  }
}

func InternalError() *Error {
  return &Error{
    Status:   http.StatusInternalServerError,
    Code:     "InternalError",
    Message:  "We encountered an internal error. Please try again.",
  }
}
