package objectstore

import (
	"crypto/md5"
	"encoding/hex"
  "fmt"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
	"gopkg.in/mgo.v2/bson"
  "io/ioutil"
  "net/http"
  "strings"
)

// CreateObject creates a new S3 object.
func (s *ObjectStore) CreateObject( r *rest.Rest ) error {
	bucketName := r.Var( "BucketName" )
  objectName := r.Var( "ObjectName" )

  reader, err := r.BodyReader()
  if err != nil {
    return err
	}

	body, err := ioutil.ReadAll( reader )
	if err != nil {
    return err
	}

  hash := md5.Sum( body )

	meta := make(map[string]string)
	for hk, hv := range r.Request().Header {
		if strings.Contains(hk, "X-Amz-") {
			meta[hk] = hv[0]
		}
	}

	obj := &Object{
    objectName,
    meta,
    s.timeNow(),
    len( body ),
    "\"" + hex.EncodeToString(hash[:]) + "\"",
  }

  meta["Last-Modified"] = obj.LastModified.Format("Mon, 2 Jan 2006 15:04:05 MST")

  metadata, err := bson.Marshal(obj)
  if err != nil {
    return err
  }

	return s.boltService.Update( func( tx *bolt.Tx ) error {
		b := tx.Bucket( bucketName )
		if b == nil {
      r.Status( 404 )
			return nil
		}

		err = b.Put( objectName + meta_suffix, metadata )
		if err != nil {
      return err
		}

		err = b.Put( objectName, body )
		if err != nil {
      b.Delete( objectName + meta_suffix )
      return err
		}

		r.Status( 200 ).
      AddHeader( "Access-Control-Allow-Origin", "*" ).
  		AddHeader( "x-amz-id-2", "LriYPLdmOdAiIfgSm/F1YsViT1LW94/xUQxMsF7xiEb1a0wiIOIxl+zbwZ163pt7" ).
  		AddHeader( "x-amz-request-id", "0A49CE4060975EAC" ).
  		AddHeader( "ETag", obj.ETag ).
  		AddHeader( "Server", "AmazonS3" ).
      Value( "" )

		return nil
	})
}

// HeadObject retrieves only meta information of an object and not the whole.
func (s *ObjectStore) HeadObject( r *rest.Rest ) error {
  bucketName := r.Var( "BucketName" )
  objectName := r.Var( "ObjectName" )

	return s.boltService.View( func( tx *bolt.Tx ) error {
		b := tx.Bucket( bucketName )
		if b == nil {
      r.Status( 404 )
      return nil
		}

    // Get the metadata
		v := b.Get( objectName + meta_suffix )
		if v == nil {
      r.Status( 404 )
      // TODO gofakes returned 500 here
			return nil
		}

		t := Object{}
		err := bson.Unmarshal(v, &t)
		if err != nil {
      return err
		}

    r.Status( 200 ).
      AccessControlAllowOrigin("").
      AddHeader( allow_headers, allow_headers_list ).
  		AddHeader( "x-amz-id-2", "LriYPLdmOdAiIfgSm/F1YsViT1LW94/xUQxMsF7xiEb1a0wiIOIxl+zbwZ163pt7" ).
  		AddHeader( "x-amz-request-id", "0A49CE4060975EAC" )

		for mk, mv := range t.Metadata {
			r.AddHeader(mk, mv)
		}

		r.AddHeader( "Last-Modified", t.LastModified.Format(http.TimeFormat) ).
  		AddHeader( "ETag", t.ETag ).
  		AddHeader( "Server", "AmazonS3" ).
      Value( "" )

		return nil
	})
}

// GetObject retrievs a bucket object.
func (s *ObjectStore) GetObject( r *rest.Rest ) error {
	bucketName := r.Var( "BucketName" )
  objectName := r.Var( "ObjectName" )

	return s.boltService.View( func( tx *bolt.Tx ) error {
    b := tx.Bucket( bucketName )
		if b == nil {
      r.Status( 404 )
      return nil
		}

    // Get the metadata
		v := b.Get( objectName + meta_suffix )
		if v == nil {
      r.Status( 404 )
      // TODO gofakes returned 500 here
			return nil
		}

		t := Object{}
		err := bson.Unmarshal(v, &t)
		if err != nil {
      return err
		}

    // Get the actual object
    v = b.Get( objectName )
		if v == nil {
      r.Status( 404 )
      // TODO gofakes returned 500 here
			return nil
		}

    r.Status( 200 ).
      AccessControlAllowOrigin("").
      AddHeader( allow_headers, allow_headers_list ).
  		AddHeader( "x-amz-id-2", "LriYPLdmOdAiIfgSm/F1YsViT1LW94/xUQxMsF7xiEb1a0wiIOIxl+zbwZ163pt7" ).
  		AddHeader( "x-amz-request-id", "0A49CE4060975EAC" )

		for mk, mv := range t.Metadata {
			r.AddHeader(mk, mv)
		}

		r.AddHeader( "Last-Modified", t.LastModified.Format(http.TimeFormat) ).
  		AddHeader( "ETag", t.ETag ).
  		AddHeader( "Server", "AmazonS3" ).
  		AddHeader( "Content-Length", fmt.Sprintf("%v", t.Length ) ).
  		AddHeader( "Connection", "close" ).
      Writer().
      Write( v )

		return nil
	})
}

// DeleteObject deletes a S3 object from the bucket.
func (s *ObjectStore) DeleteObject( r *rest.Rest ) error {
  bucketName := r.Var( "BucketName" )
  objectName := r.Var( "ObjectName" )

	return s.boltService.Update( func( tx *bolt.Tx ) error {
    b := tx.Bucket( bucketName )
		if b == nil {
      r.Status( 404 )
      return nil
		}

		b.Delete( objectName + meta_suffix )
    b.Delete( objectName )

    r.Status( 204 ).
      AccessControlAllowOrigin("").
      AddHeader( allow_headers, allow_headers_list ).
  		AddHeader( "x-amz-id-2", "LriYPLdmOdAiIfgSm/F1YsViT1LW94/xUQxMsF7xiEb1a0wiIOIxl+zbwZ163pt7" ).
  		AddHeader( "x-amz-request-id", "0A49CE4060975EAC" ).
  		AddHeader( "Content-Length", "0" ).
  		AddHeader( "Connection", "close" )

		return nil
	})
}
