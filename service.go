package objectstore

import (
  "flag"
  "github.com/peter-mount/golib/kernel"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/auth"
  "github.com/peter-mount/objectstore/awserror"
  eventservice "github.com/peter-mount/objectstore/event/service"
  "os"
  "time"
)

func (s *ObjectStore) Name() string {
  return "objectstore"
}

func (s *ObjectStore) Init( k *kernel.Kernel ) error {

  s.region = flag.String( "region", "", "Region")

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

  service, err = k.AddService( &rest.Server{} )
  if err != nil {
    return err
  }
  s.restService = (service).(*rest.Server)

  service, err = k.AddService( &auth.AuthService{} )
  if err != nil {
    return err
  }
  s.authService = (service).(*auth.AuthService)

  service, err = k.AddService( &eventservice.EventService{} )
  if err != nil {
    return err
  }
  s.eventService = (service).(*eventservice.EventService)

  return nil
}

func (s *ObjectStore) PostInit() error {

  if *s.region == "" {
    *s.region = os.Getenv( "REGION" )
  }
  if *s.region == "" {
    *s.region = "us-east-1"
  }

  // todo Add support to rest for this
  //r.Queries( "marker", "prefix" )

  // Note: trailing / required by minio client whilst s3 client doesn't use that
  // List all buckets
  builder := s.restService.RestBuilder().
    // Common decorators, applied to every endpoint
    Decorate( s.authService.AuthenticatorDecorator ).
    Decorate( awserror.RestErrorWrapper ).
    Decorate( (&rest.AddHeadersDecorator{
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Headers": "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Amz-User-Agent, X-Amz-Date, x-amz-meta-from, x-amz-meta-to, x-amz-meta-filename, x-amz-meta-private",
      "X-Clacks-Overhead": "GNU Terry Pratchett",
    }).Decorator )

  // Bucket operations
  builder.
    // List all buckets
    Method( "GET" ).
    Path( "/" ).
    Handler( s.GetBuckets ).
    Build().
    // GetBucket
    Method( "GET" ).
    Path( "/{BucketName}", "/{BucketName}/" ).
    Handler( s.GetBucket ).
    Build().
    // Check existence of bucket
    Method( "HEAD" ).
    Path( "/{BucketName}", "/{BucketName}/" ).
    Handler( s.HeadBucket ).
    Build().
    // Create bucket
    Method( "PUT" ).
    Path( "/{BucketName}", "/{BucketName}/" ).
    Handler( s.CreateBucket ).
    Build().
    // Delete Bucket
    Method( "DELETE" ).
    Path( "/{BucketName}", "/{BucketName}/" ).
    Handler( s.DeleteBucket ).
    Build()

  // Browser upload
    builder.
    Method( "POST" ).
    Path( "/{BucketName}/" ).
    Handler( s.CreateObjectBrowserUpload ).
    Build()

  // Multipart Uploads
  builder.
    // initiateMultipart
    Method( "POST" ).
    Path( "/{BucketName}/{ObjectName:.{1,}}" ).
    Queries( "uploads", "" ).
    Handler( s.initiateMultipart ).
    Build().
    // uploadPart
    Method( "PUT" ).
    Path( "/{BucketName}/{ObjectName:.{1,}}" ).
    Queries(
      "partNumber", "{PartNumber}",
      "uploadId", "{UploadId}",
    ).
    Handler( s.uploadPart ).
    Build().
    // completeMultipart
    Method( "POST" ).
    Path( "/{BucketName}/{ObjectName:.{1,}}" ).
    Queries( "uploadId", "{UploadId}").
    Handler( s.completeMultipart ).
    Build().
    // abortMultipart
    Method( "DELETE" ).
    Path( "/{BucketName}/{ObjectName:.{1,}}" ).
    Queries( "uploadId", "{UploadId}").
    Handler( s.abortMultipart ).
    Build()

    // Object upload - non multipart
  builder.
    Method( "PUT" ).
    Path( "/{BucketName}/{ObjectName:.{1,}}" ).
    Handler( s.uploadObject ).
    Build().
    // Post new object
    Method( "POST" ).
    Path( "/{BucketName}/{ObjectName:.{0,}}" ).
    Handler( s.uploadObject ).
    Build().
    // Post new object - Browser upload
    Method( "POST" ).
    Path( "/{BucketName}/" ).
    Handler( s.CreateObjectBrowserUpload ).
    Build()

  // Check object exists
  builder.
    Method( "HEAD" ).
    Path( "/{BucketName}/{ObjectName:.{0,}}" ).
    Handler( s.HeadObject ).
    Build()

  // Get object
  builder.
    Method( "GET" ).
    Path( "/{BucketName}/{ObjectName:.{1,}}" ).
    Handler( s.GetObject ).
    Build()

  // Delete object
  builder.
    Method( "DELETE" ).
    Path( "/{BucketName}/{ObjectName:.{1,}}" ).
    Handler( s.DeleteObject ).
    Build()

  return nil
}

func (s *ObjectStore) timeNow() time.Time {
	return time.Now().In( s.timeLocation )
}

func (s *ObjectStore) timeNowRFC() string {
	return s.timeNow().Format(time.RFC3339)
}
