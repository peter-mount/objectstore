package objectstore

import (
	"encoding/xml"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/awserror"
	"strings"
  "time"
)

type CopyObjectResult struct {
  XMLName       xml.Name    `xml:"CopyObjectResult"`
  LastModified  time.Time
  ETag          string
}

// copyObject copies an object
// https://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectCOPY.html
func (s *ObjectStore) copyObject( r *rest.Rest ) error {
	// As we can't have name values in Headers
	// "/{srcBucketName}/{srcObjectName:.{1,}}"
	src := strings.SplitN( r.GetHeader( "X-Amz-Copy-Source" ), "/", 3 )
	// This can be caused by an invalid client request
	if len( src ) != 3 {
		return awserror.InternalError()
	}
	srcBucketName, srcObjectName := src[1], src[2]

  destBucketName := r.Var( "DestBucketName" )
  destObjectName := r.Var( "DestObjectName" )

	return s.boltService.Update( func( tx *bolt.Tx ) error {
		sb, err := s.getBucket( tx, srcBucketName )
    if err != nil {
      return err
    }

    db, err := s.getBucket( tx, destBucketName )
    if err != nil {
      return err
    }

    srcObj := &Object{}
    err = srcObj.get( sb, srcObjectName )
    if err != nil {
      return err
    }

    dstObj := &Object{
      Name: destObjectName,
      // FIXME: This is default of copy directive
      Metadata: srcObj.Metadata,
      LastModified: srcObj.LastModified,
      Length: srcObj.Length,
      ETag: srcObj.ETag,
    }

    for _, part := range srcObj.Parts {
      b := srcObj.getPart( sb, part.PartNumber )
      err = dstObj.putPart( db, b )
      if err != nil {
        dstObj.delete( db )
        return err
      }
    }

    err = dstObj.put( db )
    if err != nil {
      return err
    }

    r.Status( 200 ).
      XML().
      Value( &CopyObjectResult{
        LastModified: dstObj.LastModified,
        ETag: dstObj.ETag,
      })

    return nil
  } )
}
