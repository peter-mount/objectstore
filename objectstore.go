package objectstore

import (
	"encoding/xml"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/auth"
  "time"
)

type ObjectStore struct {
	authService		*auth.AuthService
  boltService    *bolt.BoltService
  restService    *rest.Server
	timeLocation   *time.Location
}

type Storage struct {
  XMLName     xml.Name     `xml:"ListAllMyBucketsResult"`
  Xmlns       string       `xml:"xmlns,attr"`
  Id          string       `xml:"Owner>ID"`
  DisplayName string       `xml:"Owner>DisplayName"`
  Buckets     []BucketInfo `xml:"Buckets"`
}

type BucketInfo struct {
  Name         string `xml:"Bucket>Name"`
  CreationDate string `xml:"Bucket>CreationDate"`
}

type Content struct {
  Key          string `xml:"Key"`
  LastModified string `xml:"LastModified"`
  ETag         string `xml:"ETag"`
  Size         int    `xml:"Size"`
  StorageClass string `xml:"StorageClass"`
}

type Bucket struct {
  XMLName  xml.Name   `xml:"ListBucketResult"`
  Xmlns    string     `xml:"xmlns,attr"`
  Name     string     `xml:"Name"`
  Prefix   string     `xml:"Prefix"`
  Marker   string     `xml:"Marker"`
  Contents []*Content `xml:"Contents"`
}

const (
	// prefix used for obeject metadata. Prefix is used to improve performance
	// when listing a bucket with a large objects as we can filter by prefix in
	// bbolt but not by suffix
	meta_prefix         = "meta\001"
	// Part suffix, used for multipart upload parts, prefix is the UploadId
	part_suffix					= "\001part"
	// Meta suffix used to represent a multipart uplaod
	partmeta_suffix			= "\001partmeta"
	// The common Access-Control-Allow-Headers header in each response
	allow_headers       = "Access-Control-Allow-Headers"
	allow_headers_list  = "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Amz-User-Agent, X-Amz-Date, x-amz-meta-from, x-amz-meta-to, x-amz-meta-filename, x-amz-meta-private"
	// The block size used when reading MultipartForm
	size_24K = (1 << 20) * 24
)
