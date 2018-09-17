package objectstore

import (
	"crypto/md5"
	"encoding/hex"
  "fmt"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
	"gopkg.in/mgo.v2/bson"
	"io"
  "io/ioutil"
  "net/http"
	"strconv"
  "strings"
)

// etag calculates the object's etag
func etag( d []byte ) string {
	hash := md5.Sum( d )
	return hex.EncodeToString(hash[:])
}

// CreateObject creates a new S3 object.
func (s *ObjectStore) CreateObject( r *rest.Rest ) error {

	// Delegate to multipart if necessary
	query := r.Request().URL.Query()
	if _, ok := query["uploads"]; ok {
		return s.initiateMultipart( r )
	}
	if _, ok := query["partNumber"]; ok {
		return s.uploadPart( r )
	}
	if _, ok := query["uploadId"]; ok {
		return s.completeMultipart( r )
	}

	bucketName := r.Var( "BucketName" )
  objectName := r.Var( "ObjectName" )

  reader, err := r.BodyReader()
  if err != nil {
    return err
	}

	return s.createObject( r, bucketName, objectName, r.Request().Header, reader )
}

// CreateObjectBrowserUpload creates a new S3 object using a MultipartForm
func (s *ObjectStore) CreateObjectBrowserUpload( r *rest.Rest ) error {

	err := r.Request().ParseMultipartForm( size_24K )
	if err != nil {
		return err
	}

	bucketName := r.Var("BucketName")
	form := r.Request().MultipartForm
	key := form.Value["key"][0]

	fileHeader := form.File["file"][0]
	infile, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer infile.Close()

	return s.createObject( r, bucketName, key, form.Value, infile )
}

func (s *ObjectStore) getBody( headers map[string][]string, reader io.Reader ) ([]byte, error) {
	// Read the raw body
	body, err := ioutil.ReadAll( reader )
	if err != nil {
		return nil, err
	}

	// If it#s a chunked signed stream then we have to dechunk it to get the original object.
	if dl, ok := headers[ "X-Amz-Decoded-Content-Length" ]; ok {
		body, err = s.dechunk( body, dl[0] )
		if err != nil {
			return nil, err
		}
	}

	return body, nil
}

func (s *ObjectStore) createObject( r *rest.Rest, bucketName, objectName string, headers map[string][]string, reader io.Reader ) error {

	body, err := s.getBody( headers, reader )
	if err != nil {
    return err
  }

	// Extract the headers for the meta-data
	meta := make(map[string]string)
	for hk, hv := range headers {
		if strings.Contains(hk, "X-Amz-") || hk == "Content-Type" {
			meta[hk] = hv[0]
		}
	}

	obj := &Object{
    objectName,
    meta,
    s.timeNow(),
    len( body ),
    etag( body ),
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

		// Store the metadata separately from the body
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
			Etag( obj.ETag ).
			AddHeader( "Server", "AmazonS3" )

		return nil
	})
}

func (t *Object) addHeaders( r *rest.Rest ) {
	for mk, mv := range t.Metadata {
		if mk == "Content-Type" {
			r.ContentType( mv )
		} else {
			r.AddHeader(mk, mv)
		}
	}
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
			CacheControl( -1 ).
			AddHeader( "Accept-Ranges", "bytes" ).
  		AddHeader( "x-amz-id-2", "LriYPLdmOdAiIfgSm/F1YsViT1LW94/xUQxMsF7xiEb1a0wiIOIxl+zbwZ163pt7" ).
  		AddHeader( "x-amz-request-id", "0A49CE4060975EAC" )

		t.addHeaders( r )

		r.AddHeader( "Last-Modified", t.LastModified.Format(http.TimeFormat) ).
  		Etag( t.ETag ).
  		AddHeader( "Server", "AmazonS3" ).
			AddHeader( "Content-Length", fmt.Sprintf("%v", t.Length ) )

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

		// Request is asking for a specific range in the object
		if rng, ok := r.Request().Header["Range"]; ok {
			s, e, err := expandRangeHeader( rng[0] )
			if err != nil {
				return err
			}

			if s<0 || e<0 || s>t.Length || e>=t.Length {
				// Requested range not satisfiable
				r.Status( 416 )
				return nil
			}

			if e==0 {
				e = t.Length-1
				v = v[s:]
			} else {
				v = v[s:e+1]
			}

			// 206 Partial Content
			r.Status( 206 ).
				AddHeader( "Content-Range", fmt.Sprintf( "bytes %d-%d/%d", s, e, t.Length ) )
		} else {
			// No range requested so status 200
			r.Status( 200 )
		}

		t.addHeaders( r )

    r.AccessControlAllowOrigin("").
      AddHeader( allow_headers, allow_headers_list ).
			CacheControl( -1 ).
			AddHeader( "Accept-Ranges", "bytes" ).
			AddHeader( "Content-Length", fmt.Sprintf("%v", len(v) ) ).
  		AddHeader( "x-amz-id-2", "LriYPLdmOdAiIfgSm/F1YsViT1LW94/xUQxMsF7xiEb1a0wiIOIxl+zbwZ163pt7" ).
  		AddHeader( "x-amz-request-id", "0A49CE4060975EAC" ).
			AddHeader( "Last-Modified", t.LastModified.Format(http.TimeFormat) ).
			Etag( t.ETag ).
			AddHeader( "Server", "AmazonS3" ).
			Value( v )

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

// range handles converting the range header
// FIXME this needs doing properly
func expandRangeHeader( v string ) (int, int, error) {
	r := strings.Split( v, "=" )
	if len(r) == 2 && r[0] == "bytes" {
		r = strings.Split( r[1], "-" )
		if len(r) == 2 {
			var s, e int
			s, err := atoi( r[0] )
			if err == nil {
				e, err = atoi( r[1] )
				if err == nil {
					return s, e, nil
				}
			}
			return 0, 0, fmt.Errorf( "Invalid Range: \"%v\" %s", v, err )
		}
	}

	return 0, 0, fmt.Errorf( "Invalid Range: \"%v\"", v )
}

func atoi( s string ) (int,error) {
	if s == "" {
		return 0, nil
	}
	i, err := strconv.Atoi( s )
	return i, err
}
