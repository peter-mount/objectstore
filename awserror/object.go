package awserror

import (
  "net/http"
)

func NoSuchUpload() *Error {
	return &Error{
    Status:   http.StatusNotFound,
    Code:     "NoSuchUpload",
    Message:  "The specified multipart upload does not exist. The upload ID might be invalid, or the multipart upload might have been aborted or completed.",
  }
}

func InvalidPart() *Error {
	return &Error{
    Status:   http.StatusBadRequest,
    Code:     "InvalidPart",
    Message:  "One or more of the specified parts could not be found. The part might not have been uploaded, or the specified entity tag might not have matched the part's entity tag.",
  }
}
