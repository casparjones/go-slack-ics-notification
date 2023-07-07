package s3

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"os"
)

type Client struct {
	Bucket   string
	id       string
	secret   string
	endpoint string
}

func (c *Client) Save(filename string, data []byte) string {
	// Initialize a session that the SDK will use to load
	sess, _ := session.NewSession(&aws.Config{
		Region:           aws.String("us-west-1"),
		Credentials:      credentials.NewStaticCredentials(c.id, c.secret, ""),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String(c.endpoint),
	})

	svc := s3.New(sess)

	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(data),
	})

	if err != nil {
		log.Fatalf("failed to upload file, %v", err)
	}

	urlStr := fmt.Sprintf("%s/%s/%s", c.endpoint, c.Bucket, filename)
	return urlStr
}

func NewClient(bucket string) Client {
	return Client{
		bucket,
		os.Getenv("S3_ID"),
		os.Getenv("S3_SECRET"),
		os.Getenv("S3_ENDPOINT"),
	}
}
