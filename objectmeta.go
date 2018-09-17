package objectstore

import (
	"crypto/md5"
	"encoding/hex"
  "github.com/peter-mount/golib/kernel/bolt"
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
}

// etag calculates the object's etag
func etag( d []byte ) string {
	hash := md5.Sum( d )
	return hex.EncodeToString(hash[:])
}

// Put an object's metadata and it's content
func (o *Object) put( b *bolt.Bucket, body []byte ) error {
	o.Metadata["Last-Modified"] = o.LastModified.Format("Mon, 2 Jan 2006 15:04:05 MST")

	metadata, err := bson.Marshal( o )
	if err != nil {
		return err
	}

	err = b.Put( o.Name + meta_suffix, metadata )
	if err != nil {
		return err
	}

	if body != nil {
		err = b.Put( o.Name, body )
		if err != nil {
			b.Delete( o.Name + meta_suffix )
			return err
		}
	}

	return nil
}

// get retrieves an object's metadata
func (o *Object) get( b *bolt.Bucket, objectName string ) (bool,error) {
	v := b.Get( objectName + meta_suffix )
	if v == nil {
		return false, nil
	}
	e, err := o.getBytes( v )
	return e, err
}

// Unmarshal the metadata - usually used with a Cursor
func (o *Object) getBytes( v []byte ) (bool,error) {
	err := bson.Unmarshal(v, o)
	return err == nil, err
}

// getObject returns the entire object a a byte slice
func (o *Object) getObject( b *bolt.Bucket ) []byte {
	return b.Get( o.Name )
}

// delete Deletes an object and it's metadata
func (o *Object) delete( b *bolt.Bucket ) error {
	b.Delete( o.Name + meta_suffix )
	b.Delete( o.Name )
	return nil
}
