package objectstore

import (
  "github.com/peter-mount/golib/kernel/bolt"
  "io"
)

// Reader used to read bytes from an object.
// Initially this will treat the object as one large object but we will support
// reading just what's requried to reduce the memory footprint with large objects
// later.
type ObjectReader struct {
  store        *ObjectStore
  bucketName    string
  obj          *Object
  // partNumber
  partNumber    int
  // Offset of start of block
  offset        int
  // Number of bytes remaining
  remaining     int
  // Current position in array
  pos           int
  // Current block
  data        []byte
}

func (r *ObjectReader) Close() error {
  r.store = nil
  r.obj = nil
  r.pos = -1
  r.data = nil
  r.remaining = -1
  return nil
}

// getReader returns an ObjectReader for the specified object
func (o *Object) getReader( store *ObjectStore, bucketName string ) *ObjectReader {
  return &ObjectReader{
    store,
    bucketName,
    o,
    // Default to the entire object
    -1, 0, o.Length,
    // marker to say no data
    -1, nil,
  }
}

// getPartialReader returns an ObjectReader for the specified object but only for
// the requested range
func (o *Object) getPartialReader( store *ObjectStore, bucketName string, s, e int ) *ObjectReader {
  r := o.getReader( store, bucketName )
  r.offset = s
  l := e - s
  r.remaining = r.remaining - s
  if r.remaining > l {
    r.remaining = l
  }
  return r
}

func (r *ObjectReader) Read( p []byte ) (int, error) {
  if r.remaining <= 0 {
    return 0, io.EOF
  }

  if r.pos < 0 {
    err := r.getNextBlock()
    if err != nil {
      return 0, err
    }
  }

  dl := len(r.data)
  l := len(p)
  i := 0
  for ; i < l && r.remaining > 0 && r.pos < dl; i++ {
    p[i] = r.data[r.pos]
    r.pos++
    r.remaining --;
  }

  if r.pos >= dl {
    r.pos = -1
  }

  if r.remaining <= 0 {
    return i, io.EOF
  }

  return i, nil
}

func (r *ObjectReader) getNextBlock() error {
  return r.store.boltService.View( func( tx *bolt.Tx ) error {
    b := tx.Bucket( r.bucketName )
		if b == nil {
      return io.EOF
		}

    r.partNumber++
    d := r.obj.getPart( b, r.partNumber )
    if d == nil {
      return io.EOF
    }

    if r.data != nil {
      r.offset += len( r.data )
    }

    r.data = d
    r.pos = 0

    return nil
  } )
}
