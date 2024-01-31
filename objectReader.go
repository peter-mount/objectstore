package objectstore

import (
	"github.com/peter-mount/go-kernel/v2/bolt"
	"io"
)

// Reader used to read bytes from an object.
// Initially this will treat the object as one large object but we will support
// reading just what's requried to reduce the memory footprint with large objects
// later.
type ObjectReader struct {
	store      *ObjectStore
	bucketName string
	obj        *Object
	// partNumber
	partNumber int
	// Offset of start of block
	offset int
	// Number of bytes remaining
	remaining int
	// Current position in array
	pos int
	// Current block
	data []byte
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
func (o *Object) getReader(store *ObjectStore, bucketName string) *ObjectReader {
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
func (o *Object) getPartialReader(store *ObjectStore, bucketName string, s, e int) *ObjectReader {
	r := o.getReader(store, bucketName)

	r.offset = 0
	l := e - s + 1
	r.remaining = r.remaining - s
	if r.remaining > l {
		r.remaining = l
	}

	// Find the first part containing the requested range
	if s > 0 {
		for i, part := range o.Parts {
			p := r.offset + part.Length
			if s >= p {
				// Start is not in this part so move to next one
				r.offset += part.Length
				r.partNumber = i
			} else {
				// Start is within this part
				r.pos = s - r.offset
				break
			}
		}
	}

	return r
}

func (r *ObjectReader) Read(p []byte) (int, error) {
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
		r.remaining--
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
	return r.store.boltService.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(r.bucketName)
		if b == nil {
			return io.EOF
		}

		r.partNumber = r.partNumber + 1
		d := r.obj.getPart(b, r.partNumber)
		if d == nil {
			return io.EOF
		}

		if r.data != nil {
			r.offset += len(r.data)
		}

		// As the returned value is technically only valid during the lifetime of
		// the transaction we need to make a copy of it otherwise it can end up
		// pointing to different data or can point to invalid memory which will cause a panic
		// see Caveats in the bbolt documentation
		if r.data == nil || len(r.data) != len(d) {
			r.data = make([]byte, len(d))
		}
		copy(r.data, d)

		r.pos = 0

		return nil
	})
}
