package aws

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client interface {
	S3ListObjects(bucket, fromFile string) *s3.ListObjectsV2Paginator
	S3GetObjectData(ctx context.Context, bucket, filename string) (io.ReadCloser, error)
}
