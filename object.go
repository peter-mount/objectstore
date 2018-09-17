package objectstore

import (
  "fmt"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
	"io"
  "io/ioutil"
  "net/http"
	"strconv"
  "strings"
)

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

	return s.boltService.Update( func( tx *bolt.Tx ) error {
		b := tx.Bucket( bucketName )
		if b == nil {
      r.Status( 404 )
			return nil
		}

		// Remove any existing object
		obj := &Object{}
		if exists, _ := obj.get( b, objectName ); exists {
			obj.delete( b )
		}

		// Now create our new object
		obj = &Object{
	    objectName,
	    meta,
	    s.timeNow(),
			0,
	    etag( body ),
			nil,
		}

		// Add the sole part
		err := obj.putPart( b, body )
		if err != nil {
      return err
		}

		err = obj.put( b )
		if err != nil {
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
		t := Object{}
		exists, err := t.get( b, objectName )
		if !exists  {
      r.Status( 404 )
      // TODO gofakes returned 500 here
			return nil
		}
		if err != nil {
      r.Status( 500 )
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
	return s.boltService.View( func( tx *bolt.Tx ) error {
		bucketName := r.Var( "BucketName" )
		objectName := r.Var( "ObjectName" )

    b := tx.Bucket( bucketName )
		if b == nil {
      r.Status( 404 )
      return nil
		}

    // Get the metadata
		t := Object{}
		exists, err := t.get( b, objectName )
		if !exists  {
      r.Status( 404 )
      // TODO gofakes returned 500 here
			return nil
		}
		if err != nil {
      r.Status( 500 )
			return err
		}

		// Request is asking for a specific range in the object
		if rng, ok := r.Request().Header["Range"]; ok {
			st, en, err := expandRangeHeader( rng[0] )
			if err != nil {
				return err
			}

			if st<0 || en<0 || st>t.Length || en>=t.Length {
				// Requested range not satisfiable
				r.Status( 416 )
				return nil
			}

			if en==0 {
				en = t.Length-1
				//v = v[s:]
			} else {
				//v = v[s:e+1]
			}

			// 206 Partial Content
			r.Status( 206 ).
				AddHeader( "Content-Range", fmt.Sprintf( "bytes %d-%d/%d", st, en, t.Length ) ).
				AddHeader( "Content-Length", fmt.Sprintf( "%v", en - st + 1 ) ).
				Reader( t.getPartialReader( s, bucketName, st, en ) )
		} else {
			// No range requested so status 200 & return the entire object
			r.Status( 200 ).
				AddHeader( "Content-Length", fmt.Sprintf( "%v", t.Length ) ).
				Reader( t.getReader( s, bucketName ) )
		}

		t.addHeaders( r )

    r.AccessControlAllowOrigin("").
      AddHeader( allow_headers, allow_headers_list ).
			CacheControl( -1 ).
			AddHeader( "Accept-Ranges", "bytes" ).
			AddHeader( "x-amz-id-2", "LriYPLdmOdAiIfgSm/F1YsViT1LW94/xUQxMsF7xiEb1a0wiIOIxl+zbwZ163pt7" ).
  		AddHeader( "x-amz-request-id", "0A49CE4060975EAC" ).
			AddHeader( "Last-Modified", t.LastModified.Format(http.TimeFormat) ).
			Etag( t.ETag ).
			AddHeader( "Server", "AmazonS3" )
			//Value( v )

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

		t := Object{}
		if exists, _ := t.get( b, objectName ); exists {
			t.delete( b )
		}

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
