package s3

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

type Client struct {
	Bucket string
	id     string
	secret string
}

func (c *Client) Save(filename string, data []byte) string {
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, _ := session.NewSession(&aws.Config{
		Region:           aws.String("us-west-1"),
		Credentials:      credentials.NewStaticCredentials(c.id, c.secret, ""),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String("https://minio-api.tiny-dev.de"), // replace with your endpoint
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

	/*
		req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String(c.Bucket),
			Key:    aws.String(filename),
		})
	*/
	// urlStr, err := req.Presign(24 * time.Hour) // The link will be valid for 15 minutes.
	urlStr := fmt.Sprintf("https://minio-api.tiny-dev.de/%s/%s", c.Bucket, filename)
	return urlStr
}

func NewClient(bucket string) Client {
	return Client{
		bucket,
		"0de4ab5fbc6cefcc18f5f7e8",
		"1a055d4113ad38cbfa2e58fd9d1a73c63969e3",
	}
}

/*
❗️ IMPORTANT: Before launching MinIO be sure to change the following settings:


Step 1: Go to the settings for minio and minio-api
Step 2: Enable HTTPS
Step 3: Enable Websocket Support


🔐 Dashboard access uses the keys set during deployment. If left blank, MinIO sets the keys by default to minioadmin.


Access Key: 0de4ab5fbc6cefcc18f5f7e8
Secret Key: 1a055d4113ad38cbfa2e58fd9d1a73c63969e3


🌐 After completing the steps above, you can access MinIO at:


Dashboard: https://minio.tiny-dev.de
S3 API Endpoint: https://minio-api.tiny-dev.de
*/
