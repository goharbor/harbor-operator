package storage

import (
	"log"

	minv6 "github.com/minio/minio-go/v6"
)

type Minio interface {
	IsBucketExists(bucket string) (bool, error)
	CreateBucket(bucket string) error
}

type MinioClient struct {
	Client   *minv6.Client
	Location string
}

func GetMinioClient(endpoint, accessKeyID, secretAccessKey, location string, useSSL bool) (*MinioClient, error) {
	client, err := minv6.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)

		return nil, err
	}

	return &MinioClient{
		Client:   client,
		Location: location,
	}, nil
}

func (m MinioClient) IsBucketExists(bucket string) (bool, error) {
	exists, err := m.Client.BucketExists(bucket)
	if err != nil {
		log.Fatalln(err)

		return exists, err
	}

	return exists, nil
}

func (m MinioClient) CreateBucket(bucket string) error {
	err := m.Client.MakeBucket(bucket, m.Location)
	if err != nil {
		log.Fatalln(err)

		return err
	}

	return nil
}
