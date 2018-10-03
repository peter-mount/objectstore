package objectstore

import (
	"crypto/md5"
  "encoding/hex"
	"encoding/xml"
  "fmt"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/awserror"
  "gopkg.in/mgo.v2/bson"
	"strings"
  "time"
)

type MultipartUpload struct {
	// Final object name for this upload
	ObjectName	string
	// Generated uploadId
	UploadId  	string
	// Index of uploaded parts
	Parts		  	map[string]string
  // Time of when upload was initiated.
  // TODO this is for future use, we'll use this to cleanup incomplete uploads
  Time      	time.Time
	// Metadata
	Meta				map[string]string
}

func (u *MultipartUpload) get( b *bolt.Bucket, uploadId string ) error {
	v := b.Get( partmeta_prefix + uploadId )
	if v == nil {
		return awserror.NoSuchUpload()
	}

	return bson.Unmarshal(v, u)
}

func (u *MultipartUpload) put( b *bolt.Bucket ) error {
  v, err := bson.Marshal( u )
  if err != nil {
    return err
  }

  return b.Put( partmeta_prefix + u.UploadId, v )
}

func (u *MultipartUpload) delete( b *bolt.Bucket ) error {
	for _,n := range u.Parts {
		b.Delete( n )
	}
	return b.Delete( partmeta_prefix + u.UploadId )
}

type InitiateMultipartUploadResult struct {
  XMLName   xml.Name  `xml:"InitiateMultipartUploadResult"`
  Xmlns     string    `xml:"xmlns,attr"`
  Bucket    string    `xml:"Bucket"`
  Key       string    `xml:"Key"`
  UploadId  string    `xml:"UploadId"`
}

type CompleteMultipartUpload struct {
  XMLName   xml.Name              `xml:"CompleteMultipartUpload"`
  Parts     []MultipartUploadPart `xml:"Part"`
}

type CompleteMultipartUploadResult struct {
	XMLName   xml.Name              `xml:"CompleteMultipartUploadResult"`
	Location	string
	Bucket		string
	Key				string
	ETag			string
}

type MultipartUploadPart struct {
  XMLName   xml.Name  `xml:"Part"`
  PartNumber  string  `xml:"PartNumber"`
  ETag        string  `xml:"ETag"`
}

// initiateMultipart initiates a multipart upload
// See https://docs.aws.amazon.com/AmazonS3/latest/API/mpUploadInitiate.html
func (s *ObjectStore) initiateMultipart( r *rest.Rest ) error {
	bucketName := r.Var( "BucketName" )
	objectName := r.Var( "ObjectName" )

	// Generate the uploadId from the bucketName, objectName & start time
	startTime := s.timeNow()
	fullName := fmt.Sprintf( "%s\002%s\002%s", bucketName, objectName, startTime.Format(time.RFC3339) )
	hash := md5.Sum( []byte(fullName) )
	uploadId := hex.EncodeToString(hash[:])

	upload := &MultipartUpload{
		objectName,
		uploadId,
		make( map[string]string),
		startTime,
		make( map[string]string),
	}

	// Extract the headers for the meta-data
	for hk, hv := range r.Request().Header {
		if strings.HasPrefix(hk, "X-Amz-") || hk == "Content-Type" {
			upload.Meta[hk] = hv[0]
		}
	}

  err := s.boltService.Update( func( tx *bolt.Tx ) error {
		b, err := s.getBucket( tx, bucketName )
    if err != nil {
      return err
    }
    return upload.put( b )
  } )
	if err != nil {
		return err
	}

	r.Status( 200 ).
		XML().
		Value( &InitiateMultipartUploadResult{
			Bucket: bucketName,
			Key: objectName,
			UploadId: uploadId,
		} )

	return nil
}

// uploadPart handles the upload of a part
// see https://docs.aws.amazon.com/AmazonS3/latest/API/mpUploadUploadPart.html
func (s *ObjectStore) uploadPart( r *rest.Rest ) error {
	bucketName := r.Var( "BucketName" )
	partNumber := r.Var( "PartNumber" )
	uploadId := r.Var( "UploadId" )

  reader, err := r.BodyReader()
  if err != nil {
    return err
  }

  body, err := s.getBody( r.Request().Header, reader )
  if err != nil {
    return err
  }

  err = s.boltService.Update( func( tx *bolt.Tx ) error {
		b, err := s.getBucket( tx, bucketName )
    if err != nil {
      return err
    }

    // Get the metadata
		upload := MultipartUpload{}
    if err := upload.get( b, uploadId ); err != nil {
			return err
		}

    partKey := partmeta_prefix + uploadId + partmeta_suffix + partNumber
    upload.Parts[partNumber] = partKey

    if err := b.Put( partKey, body ); err != nil {
      return err
    }

    if err := upload.put( b ); err != nil {
      b.Delete( partKey )
      return err
    }
		return nil
  } )
	if err != nil {
		return err
	}

	checksum := etag( body )

	r.Status( 200 ).
		AddHeader( "Connection", "keep-alive" ).
		AddHeader( "Content-MD5", checksum ).
		AddHeader( "Content-Length", "0" ).
		Etag( checksum )

	return nil
}

func (s *ObjectStore) completeMultipart( r *rest.Rest ) error {
  bucketName := r.Var( "BucketName" )
	uploadId := r.Var( "UploadId" )

  // Get payload
  reader, err := r.BodyReader()
  if err != nil {
    return err
  }

  body, err := s.getBody( r.Request().Header, reader )
  if err != nil {
    return err
  }

  req := &CompleteMultipartUpload{}
  err = xml.Unmarshal( body, req )
  if err != nil {
    return err
  }

  return s.boltService.Update( func( tx *bolt.Tx ) error {
		b, err := s.getBucket( tx, bucketName )
    if err != nil {
      return err
    }

    // Get the metadata
    upload := MultipartUpload{}
		if err := upload.get( b, uploadId ); err != nil {
			return err
		}

    obj := &Object{
      upload.ObjectName,
      upload.Meta,
      s.timeNow(),
      0,
      "",
      nil,
    }

    // Delete the upload on exit
    defer upload.delete( b )

    // Now add the parts to the final object
    hash := md5.New()
    for _, p := range req.Parts {
      n := upload.Parts[ p.PartNumber ]

      d := b.Get( n )
      if d == nil {
				return awserror.InvalidPart()
      }
			// TODO should check part's etag matches & if not return InvalidPart

      // Write this part
      err = obj.putPart( b, d )
      if err != nil {
        return err
      }

      // Add it to the hash
      hash.Write( d )
    }

    obj.ETag = hex.EncodeToString(hash.Sum(nil)[:])

    // Save the metadata
    err = obj.put( b )
    if err != nil {
      return err
    }

		s.sendObjectEvent( "ObjectCreated:CompleteMultipartUpload", bucketName, obj )

    r.Status( 200 ).
      XML().
      Value( &CompleteMultipartUploadResult{
        Location: "/" + bucketName + "/" + obj.Name,
        Bucket: bucketName,
        Key: obj.Name,
        ETag: obj.ETag,
      })

    return nil
  } )
}

func (s *ObjectStore) abortMultipart( r *rest.Rest ) error {
  bucketName := r.Var( "BucketName" )
	uploadId := r.Var( "UploadId" )

  return s.boltService.Update( func( tx *bolt.Tx ) error {
		b, err := s.getBucket( tx, bucketName )
    if err != nil {
      return err
    }

    // Get the metadata
    upload := MultipartUpload{}
		if err := upload.get( b, uploadId ); err != nil {
			return err
		}
    upload.delete( b )

    r.Status( 204 )

    return nil
  } )
}
