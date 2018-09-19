package objectstore

import (
  "github.com/peter-mount/golib/kernel"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/auth"
  "time"
)

func (s *ObjectStore) Name() string {
  return "objectstore"
}

func (s *ObjectStore) Init( k *kernel.Kernel ) error {

  timeLocation, err := time.LoadLocation("GMT")
	if err != nil {
		return err
	}
  s.timeLocation = timeLocation

  service, err := k.AddService( &bolt.BoltService{} )
  if err != nil {
    return err
  }
  s.boltService = (service).(*bolt.BoltService)

  service, err = k.AddService( &rest.Server{ Port: 80 } )
  if err != nil {
    return err
  }
  s.restService = (service).(*rest.Server)

  service, err = k.AddService( &auth.AuthService{} )
  if err != nil {
    return err
  }
  s.authService = (service).(*auth.AuthService)

  return nil
}

func (s *ObjectStore) PostInit() error {

  r := s.restService

  // todo Add support to rest for this
  //r.Queries( "marker", "prefix" )

  // Note: trailing / required by minio client whilst s3 client doesn't use that
  // List all buckets
  r.Handle("/", s.GetBuckets).Methods("GET")

	r.Handle("/{BucketName}", s.GetBucket).Methods("GET")
	r.Handle("/{BucketName}/", s.GetBucket).Methods("GET")

  r.Handle("/{BucketName}", s.HeadBucket).Methods("HEAD")
  r.Handle("/{BucketName}/", s.HeadBucket).Methods("HEAD")

  r.Handle("/{BucketName}", s.CreateBucket).Methods("PUT")
  r.Handle("/{BucketName}/", s.CreateBucket).Methods("PUT")

  r.Handle("/{BucketName}", s.DeleteBucket).Methods("DELETE")
  r.Handle("/{BucketName}/", s.DeleteBucket).Methods("DELETE")

	// Object operations
  r.Handle("/{BucketName}/", s.CreateObjectBrowserUpload).Methods("POST")
  r.Handle("/{BucketName}/{ObjectName:.{1,}}", s.CreateObject).Methods("PUT")
  r.Handle("/{BucketName}/{ObjectName:.{0,}}", s.CreateObject).Methods("POST")

  r.Handle("/{BucketName}/{ObjectName:.{0,}}", s.HeadObject).Methods("HEAD")
  r.Handle("/{BucketName}/{ObjectName:.{1,}}", s.GetObject).Methods("GET")

  r.Handle("/{BucketName}/{ObjectName:.{1,}}", s.DeleteObject).Methods("DELETE")

  return nil
}

func (s *ObjectStore) timeNow() time.Time {
	return time.Now().In( s.timeLocation )
}

func (s *ObjectStore) timeNowRFC() string {
	return s.timeNow().Format(time.RFC3339)
}
