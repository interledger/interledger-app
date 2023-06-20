package client

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gitlab.com/fynbos/backend/aws"
)

type client struct {
	s3Client *s3.Client
}

func New(ctx context.Context) (aws.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws SDK configuration, %v", err)
	}

	cl := s3.NewFromConfig(cfg)

	return &client{s3Client: cl}, nil
}

var _ aws.Client = client{}

func (c client) S3ListObjects(bucket string) *s3.ListObjectsV2Paginator {
	return s3.NewListObjectsV2Paginator(c.s3Client, &s3.ListObjectsV2Input{
		Bucket: &bucket,
	})
}

func (c client) S3GetObjectData(ctx context.Context, bucket, filename string) (io.ReadCloser, error) {
	goo, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &filename,
	})
	if err != nil {
		return nil, err
	}

	return goo.Body, nil
}
