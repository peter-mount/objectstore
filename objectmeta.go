package objectstore

import (
	"crypto/md5"
	"encoding/hex"
  "fmt"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/objectstore/awserror"
	"gopkg.in/mgo.v2/bson"
  "time"
)

// The metadata for each stored object
type Object struct {
  // The true object name
  Name          string
  // Metadata
  Metadata      map[string]string
  // When last modified
  LastModified  time.Time
  // Length
  Length        int
  // ETag
  ETag          string
  // The parts
  Parts       []ObjectPart
}

type ObjectPart struct {
  // The part number
  PartNumber    int
  // The start position in the object
  Start         int
  // The length of this part
  Length        int
}

// etag calculates the object's etag
func etag( d []byte ) string {
	hash := md5.Sum( d )
	return hex.EncodeToString(hash[:])
}

// Put an object's metadata and it's content
func (o *Object) put( b *bolt.Bucket ) error {
	o.Metadata["Last-Modified"] = o.LastModified.Format("Mon, 2 Jan 2006 15:04:05 MST")

	metadata, err := bson.Marshal( o )
	if err != nil {
		return err
	}

	return b.Put( meta_prefix + o.Name, metadata )
}

func (o *Object) putPart( b *bolt.Bucket, body []byte ) error {
  partNo := len( o.Parts )

  o.Parts = append( o.Parts, ObjectPart{ partNo, o.Length, len(body) } )

  l := 0
  for _, p := range o.Parts {
    l += p.Length
  }
  o.Length = l

  return b.Put( fmt.Sprintf( "%s\003%d", o.Name, partNo ), body )
}

// get retrieves an object's metadata
func (o *Object) get( b *bolt.Bucket, objectName string ) error {
	v := b.Get( meta_prefix + objectName )
	if v == nil {
		return awserror.NoSuchKey()
	}
	return o.getBytes( v )
}

// Unmarshal the metadata - usually used with a Cursor
func (o *Object) getBytes( v []byte ) error {
	return bson.Unmarshal(v, o)
}

// getObject returns the entire object a a byte slice
func (o *Object) getPart( b *bolt.Bucket, partNumber int ) []byte {
  return b.Get( fmt.Sprintf( "%s\003%d", o.Name, partNumber ) )
}

// delete Deletes an object and it's metadata
func (o *Object) delete( b *bolt.Bucket ) error {
	b.Delete( meta_prefix + o.Name )
  for _, p := range o.Parts {
    b.Delete( fmt.Sprintf( "%s\003%d", o.Name, p.PartNumber ) )
  }
  o.Parts = []ObjectPart{}
	return nil
}
