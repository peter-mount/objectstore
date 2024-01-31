package objectstore

import (
	//"github.com/peter-mount/go-kernel/v2/rest"
	"github.com/peter-mount/objectstore/event"
	"github.com/peter-mount/objectstore/utils"
)

func (s *ObjectStore) sendObjectEvent(eventName, bucketName string, obj *Object) {

	s.eventService.Notify(&event.Event{
		Source: "aws:s3",
		Region: *s.region,
		Time:   obj.LastModified,
		Name:   eventName,
		Identity: event.Identity{
			PrincipalId: "AIDAJDPLRKLG7UEXAMPLE",
		},
		S3: &event.S3{
			Bucket: event.S3Bucket{
				Name: bucketName,
				Identity: event.Identity{
					PrincipalId: "A3NL1KOZZKExample",
				},
				Arn: utils.NewS3ARN("aws", "", bucketName),
			},
			Object: &event.S3Object{
				Key:  obj.Name,
				Size: obj.Length,
				ETag: obj.ETag,
			},
		},
	})

}
