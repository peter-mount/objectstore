package objectstore

import (
	"encoding/xml"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
  "time"
)

type ObjectStore struct {
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
  // Actual data
  //Obj          []byte
}

const (
  meta_suffix         = "\001meta"
  allow_headers       = "Access-Control-Allow-Headers"
  allow_headers_list  = "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Amz-User-Agent, X-Amz-Date, x-amz-meta-from, x-amz-meta-to, x-amz-meta-filename, x-amz-meta-private"
)
