package awserror

import (
	bbolt "github.com/etcd-io/bbolt"
	"github.com/peter-mount/go-kernel/v2/rest"
	"log"
)

func RestErrorWrapper(h rest.RestHandler) rest.RestHandler {
	return func(r *rest.Rest) error {
		err := h(r)
		if err != nil {
			log.Println(err)

			// Map known errors to aws ones
			switch err {
			case bbolt.ErrBucketExists:
				err = BucketAlreadyExists()
			case bbolt.ErrBucketNotFound:
				err = NoSuchBucket()
			}

			log.Println(err)
			// Aws errors send correct response
			if e, ok := err.(*Error); ok {
				e.Send(r)
				return nil
			}
		}
		return err
	}
}
