package awserror

import (
  "net/http"
)

func BucketAlreadyExists() *Error {
	return &Error{
    Status:   http.StatusConflict,
    Code:     "BucketAlreadyExists",
    Message:  "The requested bucket name is not available. The bucket namespace is shared by all users of the system. Please select a different name and try again.",
  }
}

func NoSuchBucket() *Error {
	return &Error{
    Status:   http.StatusNotFound,
    Code:     "NoSuchBucket",
    Message:  "The specified bucket does not exist.",
  }
}
