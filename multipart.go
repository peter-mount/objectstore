package objectstore

import (
	"crypto/md5"
//  "encoding/base64"
  "encoding/hex"
	"encoding/xml"
  "fmt"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
  "gopkg.in/mgo.v2/bson"
  "time"
	"log"
)

type MultipartUpload struct {
	UploadId  string
	// Index of uploaded parts
	Parts		  map[string]string
  // Time of when upload was initiated.
  // TODO this is for future use, we'll use this to cleanup incomplete uploads
  Time      time.Time
}

func (u *MultipartUpload) get( b *bolt.Bucket, uploadId string ) (bool,error) {
	v := b.Get( uploadId + partmeta_suffix )
	if v == nil {
		return false, nil
	}

	err := bson.Unmarshal(v, u)
	return err == nil, err
}

func (u *MultipartUpload) put( b *bolt.Bucket ) error {
  v, err := bson.Marshal( u )
  if err != nil {
    return err
  }

  return b.Put( u.UploadId + partmeta_suffix, v )
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

type MultipartUploadPart struct {
  XMLName   xml.Name  `xml:"Part"`
  PartNumber  string    `xml:"PartNumber"`
  ETag        string    `xml:"ETag"`
}

// initiateMultipart initiates a multipart upload
// See https://docs.aws.amazon.com/AmazonS3/latest/API/mpUploadInitiate.html
func (s *ObjectStore) initiateMultipart( r *rest.Rest ) error {
  return s.boltService.Update( func( tx *bolt.Tx ) error {
    bucketName := r.Var( "BucketName" )
    objectName := r.Var( "ObjectName" )

    b := tx.Bucket( bucketName )
    if b == nil {
      r.Status( 404 )
      return nil
    }

    // Generate the uploadId from the bucketName, objectName & start time
    startTime := s.timeNow()
    fullName := fmt.Sprintf( "%s\002%s\002%s", bucketName, objectName, startTime.Format(time.RFC3339) )
    hash := md5.Sum( []byte(fullName) )
    uploadId := hex.EncodeToString(hash[:])

    log.Println( "initiateMultipart", bucketName, objectName, uploadId )

    upload := &MultipartUpload{ uploadId, make( map[string]string), startTime }
    err := upload.put( b )
    if err != nil {
      return err
    }

    r.Status( 200 ).
      AddHeader( "Access-Control-Allow-Origin", "*" ).
      AddHeader( "x-amz-id-2", "LriYPLdmOdAiIfgSm/F1YsViT1LW94/xUQxMsF7xiEb1a0wiIOIxl+zbwZ163pt7" ).
      AddHeader( "x-amz-request-id", "0A49CE4060975EAC" ).
      AddHeader( "Server", "AmazonS3" ).
      XML().
      Value( &InitiateMultipartUploadResult{
        Bucket: bucketName,
        Key: objectName,
        UploadId: uploadId,
      } )

    return nil
  } )
}

// uploadPart handles the upload of a part
// see https://docs.aws.amazon.com/AmazonS3/latest/API/mpUploadUploadPart.html
func (s *ObjectStore) uploadPart( r *rest.Rest ) error {
  reader, err := r.BodyReader()
  if err != nil {
    return err
  }

  body, err := s.getBody( r.Request().Header, reader )
  if err != nil {
    return err
  }

  return s.boltService.Update( func( tx *bolt.Tx ) error {
    bucketName := r.Var( "BucketName" )
    //objectName := r.Var( "ObjectName" )
  	query := r.Request().URL.Query()
    partNumber := query["partNumber"][0]
    uploadId := query["uploadId"][0]

    b := tx.Bucket( bucketName )
    if b == nil {
      r.Status( 404 )
      return nil
    }

    // Get the metadata
    upload := MultipartUpload{}
    if ok, err := upload.get( b, uploadId ); !ok || err != nil {
      log.Println( "Not found", uploadId )
      r.Status( 404 )
			return nil
		}

    partKey := uploadId + part_suffix + partNumber
    upload.Parts[partNumber] = partKey

    log.Println( "Body", len(body) )
    if err := b.Put( partKey, body ); err != nil {
      return err
    }

    log.Println( "UploadPart", partNumber, partKey)

    if err := upload.put( b ); err != nil {
      b.Delete( partKey )
      return err
    }

    hash := md5.Sum( body )
    checksum := hex.EncodeToString(hash[:])
    //checksum := base64.StdEncoding.EncodeToString(body)
    log.Println( checksum )
    log.Println( len(checksum) )

    r.Status( 200 ).
      AddHeader( "Access-Control-Allow-Origin", "*" ).
      AddHeader( "x-amz-id-2", "LriYPLdmOdAiIfgSm/F1YsViT1LW94/xUQxMsF7xiEb1a0wiIOIxl+zbwZ163pt7" ).
      AddHeader( "x-amz-request-id", "0A49CE4060975EAC" ).
      AddHeader( "Connection", "keep-alive" ).
      AddHeader( "Server", "AmazonS3" ).
      AddHeader( "Content-MD5", checksum ).
      AddHeader( "Content-Length", "0" ).
      Etag( checksum )

    return nil
  } )
}

func (s *ObjectStore) completeMultipart( r *rest.Rest ) error {
  log.Println( "completeMultipart" )
  reader, err := r.BodyReader()
  if err != nil {
    return err
  }

  body, err := s.getBody( r.Request().Header, reader )
  if err != nil {
    return err
  }

  log.Printf( "%v", string(body[:]) )
/*
  req := &CompleteMultipartUpload{}
  err := r.Body( req)
  if err != nil {
    return err
  }
  log.Printf( "%v", req )
  */

  return nil
}