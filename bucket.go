package objectstore

import (
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
  "strings"
  "time"
)

// GetBuckets returns a list of all Buckets
func (s *ObjectStore) GetBuckets( r *rest.Rest ) error {
  buckets := []BucketInfo{}

  now := s.timeNowRFC()

  err := s.boltService.View( func( tx *bolt.Tx ) error {
    return tx.ForEach( func( name string, _ *bolt.Bucket ) error {
      buckets = append( buckets, BucketInfo{name,now} )
      return nil
    })
  } )

  if err != nil {
    return err
  }

  r.Status( 200 ).
    XML().
    AccessControlAllowOrigin("").
    AddHeader( allow_headers, allow_headers_list ).
    Value( &Storage{
      Xmlns:       "http://s3.amazonaws.com/doc/2006-03-01/",
      Id:          "fe7272ea58be830e56fe1663b10fafef",
      DisplayName: "Area51ObjectStore",
      Buckets:     buckets,
    } )
  return nil
}

// CreateBucket creates a new S3 bucket in the BoltDB storage.
func (s *ObjectStore) CreateBucket( r *rest.Rest ) error {
	bucketName := r.Var("BucketName")

  exists := false
  err := s.boltService.Update( func ( tx *bolt.Tx ) error {
    b := tx.Bucket( bucketName )
    if b != nil {
      exists = true
      return nil;
    }

    _, err := tx.CreateBucket( bucketName )
    return err
  })

  if err != nil {
    return err
  }

  if exists {
    r.Status( 409 )
  } else {
    r.Status( 200 ).
    AccessControlAllowOrigin("").
    AddHeader( allow_headers, allow_headers_list ).
    AddHeader( "Host", r.Request().Host ).
    AddHeader( "Location", "/" + bucketName )
  }

	return nil
}

// DeleteBucket deletes a S3 bucket in the BoltDB storage.
func (s *ObjectStore) DeleteBucket( r *rest.Rest ) error {
  bucketName := r.Var("BucketName")

  err := s.boltService.Update( func ( tx *bolt.Tx ) error {
    return tx.DeleteBucket( bucketName )
  })

  if err != nil {
    r.Status( 404 )
  } else {
    r.Status( 200 ).
    AddHeader( "x-amz-id-2", "LriYPLdmOdAiIfgSm/F1YsViT1LW94/xUQxMsF7xiEb1a0wiIOIxl+zbwZ163pt7" ).
    AddHeader( "x-amz-request-id", "0A49CE4060975EAC" ).
    AddHeader( "Server", "AmazonS3" )
  }

	return nil
}

// HeadBucket checks whether a bucket exists.
func (s *ObjectStore) HeadBucket( r *rest.Rest ) error {
	bucketName := r.Var("BucketName")

	s.boltService.View( func( tx *bolt.Tx ) error {
		b := tx.Bucket( bucketName )
		if b == nil {
      r.Status( 404 )
		} else {
      r.Status( 200 ).
      AddHeader( "x-amz-id-2", "LriYPLdmOdAiIfgSm/F1YsViT1LW94/xUQxMsF7xiEb1a0wiIOIxl+zbwZ163pt7" ).
      AddHeader( "x-amz-request-id", "0A49CE4060975EAC" ).
      AddHeader( "Server", "AmazonS3" )
    }

		return nil
	} )

  r.AccessControlAllowOrigin("").
    AddHeader( allow_headers, allow_headers_list )

  return nil
}

// GetBucket lists the contents of a bucket.
func (s *ObjectStore) GetBucket( r *rest.Rest ) error {
	bucketName := r.Var("BucketName")
  prefix := r.Request().URL.Query().Get("prefix")

  bucketc := &Bucket{
    Xmlns:    "http://s3.amazonaws.com/doc/2006-03-01/",
    Name:     bucketName,
    Prefix:   prefix,
    Marker:   "",
    Contents: []*Content{},
  }

	return s.boltService.View( func( tx *bolt.Tx ) error {
		// Assume bucket exists and has keys
		b := tx.Bucket( bucketName )
		if b == nil {
      r.Status( 404 )
			return nil
		}

    // prefix with our meta_prefix prefixed to it
    pre := meta_prefix + prefix

		c := b.Cursor()
		for k, v := c.Seek( meta_prefix ); k != "" && strings.HasPrefix( k, pre ); k, v = c.Next() {
      t := Object{}
      if _, err := t.getBytes( v ); err != nil {
        return err
      }

      bucketc.Contents = append( bucketc.Contents, &Content{
        Key:          t.Name,
        LastModified: t.LastModified.Format(time.RFC3339),
        ETag:         t.ETag,
        Size:         t.Length,
        StorageClass: "STANDARD",
      })
		}

    r.Status( 200 ).
      XML().
      AccessControlAllowOrigin("").
      AddHeader( allow_headers, allow_headers_list ).
      AddHeader( "Host", r.Request().Host ).
      AddHeader( "Location", "/" + bucketName ).
      Value( bucketc )

    return nil
	})
}
