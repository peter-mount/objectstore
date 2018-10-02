package objectstore

import (
	"encoding/xml"
  "github.com/peter-mount/golib/kernel/bolt"
  "github.com/peter-mount/golib/rest"
)

type AccessControlPolicy struct {
  XMLName       xml.Name    `xml:"AccessControlPolicy"`
  Id            string      `xml:"Owner>ID"`
  DisplayName   string      `xml:"Owner>DisplayName"`
  Grant       []Grant       `xml:"AccessControlList"`
}

type Grant struct {
  XMLName     xml.Name    `xml:"Grant"`
  Id          string      `xml:"Grantee>ID"`
  DisplayName string      `xml:"Grantee>DisplayName"`
  Permission  string      `xml:"Permission"`
}

// getAcl is a stub until we support this call.
// It's required as some clients issue this request and currently we respond
// with the entire object which breaks the client.
// https://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectGETacl.html
func (s *ObjectStore) getObjectAcl( r *rest.Rest ) error {
    bucketName := r.Var( "BucketName" )
    objectName := r.Var( "ObjectName" )

    obj := &Object{}

  	err := s.boltService.Update( func( tx *bolt.Tx ) error {
      b, err := s.getBucket( tx, bucketName )
      if err != nil {
        return err
      }

  		return obj.get( b, objectName )
  	})
    if err != nil {
      return err
    }

    // Dummy data
    acl := &AccessControlPolicy{
      Id: "75aa57f09aa0c8caeab4f8c24e99d10f8e7faeebf76c078efc7c6caea54ba06a",
      DisplayName: "mtd@amazon.com",
      Grant: []Grant{
        Grant{
          Id: "75aa57f09aa0c8caeab4f8c24e99d10f8e7faeebf76c078efc7c6caea54ba06a",
          DisplayName: "mtd@amazon.com",
          Permission: "FULL_CONTROL",
        },
      },
    }

    r.Status( 200 ).
      AddHeader( "x-amz-id-2", "LriYPLdmOdAiIfgSm/F1YsViT1LW94/xUQxMsF7xiEb1a0wiIOIxl+zbwZ163pt7" ).
      AddHeader( "x-amz-request-id", "0A49CE4060975EAC" ).
      AddHeader( "Content-Length", "0" ).
      XML().
      Value( acl )

    return nil
  }
