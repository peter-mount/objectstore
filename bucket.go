package objectstore

import (
	"github.com/peter-mount/go-kernel/v2/bolt"
	"github.com/peter-mount/go-kernel/v2/rest"
	"github.com/peter-mount/objectstore/awserror"
	"strings"
	"time"
)

// getBucket returns a bucket or an error if the bucket is not found
func (s *ObjectStore) getBucket(tx *bolt.Tx, bucketName string) (*bolt.Bucket, error) {
	b := tx.Bucket(bucketName)
	if b == nil {
		return nil, awserror.NoSuchBucket()
	}
	return b, nil
}

// GetBuckets returns a list of all Buckets
func (s *ObjectStore) GetBuckets(r *rest.Rest) error {

	buckets := []BucketInfo{}

	now := s.timeNowRFC()

	err := s.boltService.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name string, _ *bolt.Bucket) error {
			buckets = append(buckets, BucketInfo{name, now})
			return nil
		})
	})

	if err != nil {
		return err
	}

	r.Status(200).
		XML().
		Value(&Storage{
			Xmlns:       "http://s3.amazonaws.com/doc/2006-03-01/",
			Id:          "fe7272ea58be830e56fe1663b10fafef",
			DisplayName: "Area51ObjectStore",
			Buckets:     buckets,
		})
	return nil
}

// CreateBucket creates a new S3 bucket in the BoltDB storage.
func (s *ObjectStore) CreateBucket(r *rest.Rest) error {
	bucketName := r.Var("BucketName")

	err := s.boltService.Update(func(tx *bolt.Tx) error {
		// TODO if bbolt.ErrBucketExists then check ownership & return
		// BucketAlreadyOwnedByYou if caller is the owner
		_, err := tx.CreateBucket(bucketName)
		return err
	})

	if err != nil {
		return err
	}

	r.Status(200).
		AddHeader("Host", r.Request().Host).
		AddHeader("Location", "/"+bucketName)

	return nil
}

// DeleteBucket deletes a S3 bucket in the BoltDB storage.
func (s *ObjectStore) DeleteBucket(r *rest.Rest) error {
	bucketName := r.Var("BucketName")

	err := s.boltService.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(bucketName)
	})

	if err != nil {
		return err
	}

	r.Status(200)

	return nil
}

// HeadBucket checks whether a bucket exists.
func (s *ObjectStore) HeadBucket(r *rest.Rest) error {
	bucketName := r.Var("BucketName")

	err := s.boltService.View(func(tx *bolt.Tx) error {
		_, err := s.getBucket(tx, bucketName)
		return err
	})

	if err != nil {
		return err
	}

	r.Status(200)

	return nil
}

// GetBucket lists the contents of a bucket.
func (s *ObjectStore) GetBucket(r *rest.Rest) error {
	bucketName := r.Var("BucketName")
	prefix := r.Request().URL.Query().Get("prefix")

	bucketc := &Bucket{
		Xmlns:    "http://s3.amazonaws.com/doc/2006-03-01/",
		Name:     bucketName,
		Prefix:   prefix,
		Marker:   "",
		Contents: []*Content{},
	}

	err := s.boltService.View(func(tx *bolt.Tx) error {
		b, err := s.getBucket(tx, bucketName)
		if err != nil {
			return err
		}

		// prefix with our meta_prefix prefixed to it
		pre := meta_prefix + prefix

		c := b.Cursor()
		for k, v := c.Seek(pre); k != "" && strings.HasPrefix(k, pre); k, v = c.Next() {
			t := Object{}
			if err := t.getBytes(v); err != nil {
				return err
			}

			bucketc.Contents = append(bucketc.Contents, &Content{
				Key:          t.Name,
				LastModified: t.LastModified.Format(time.RFC3339),
				ETag:         t.ETag,
				Size:         t.Length,
				StorageClass: "STANDARD",
			})
		}

		return nil
	})
	if err != nil {
		return err
	}

	r.Status(200).
		XML().
		AddHeader("Host", r.Request().Host).
		AddHeader("Location", "/"+bucketName).
		Value(bucketc)

	return nil
}
