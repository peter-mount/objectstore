# objectstore

A prototype object store written in go using the S3 protocol using the bbolt object database as the backend.

It is loosely based on [gofakes3](https://github.com/johannesboyne/gofakes3) by johannesboyne but converted to use both of my kernel & rest frameworks, making it not just standalone but embedable within other projects. It also stores objects differently to return the correct object sizes when listing them and to reduce the memory footprint when listing large objects.

## Supported features

* Create bucket
* Delete bucket
* List buckets
* Create object
* List objects
* Retrieve object
* Delete object
* Docker container

## Supported clients
These are only partially tested so far:
* aws command line
* minio client (mc) - deleting buckets and objects does not work as it uses an odd API

## ToDo
* Add authentication
* Add ACL control
* Add Event notification using RabbitMQ (similar to how [Minio](https://github.com/minio/minio) does)

## Similar notable projects
- https://github.com/johannesboyne/gofakes3 also written in Go & where I started this from
- https://github.com/minio/minio **not similar but powerfull ;-)**
- https://github.com/andrewgaul/s3proxy by @andrewgaul
