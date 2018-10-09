package objectstore

import (
	"encoding/xml"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/auth"
  eventservice "github.com/peter-mount/objectstore/event/service"
  "time"
)

type ObjectStore struct {
	authService		 *auth.AuthService
  boltService    *bolt.BoltService
	eventService	 *eventservice.EventService
  restService    *rest.Server
	timeLocation   *time.Location

	region				*string
	website				*bool
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
	// Suffix used to represent a multipart uplaod
	partmeta_prefix			= "upload\001"
	// Part suffix, used for multipart upload parts, prefix is the UploadId
	partmeta_suffix			= "\002"
	// The block size used when reading MultipartForm
	size_24K = (1 << 20) * 24
)
