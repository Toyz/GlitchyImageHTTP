package filemodes

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type CDNMode struct{}

func (CDNMode) Write(data []byte, name string) string {
	resourceURL := strings.TrimSpace(core.GetEnv("AWS_RESOURCE_URL", ""))
	bucket := core.GetEnv("AWS_BUCKET", "")
	accessKey := core.GetEnv("AWS_ACCESS_KEY", "")
	secretKey := core.GetEnv("AWS_SECRET_KEY", "")
	region := core.GetEnv("AWS_REGION", "us-east-1")
	endpoint := core.GetEnv("AWS_ENDPOINT", "")

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(endpoint),
		Region:      aws.String(region), // This is counter intuitive, but it will fail with a non-AWS region name.
	}

	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)

	filePath := fmt.Sprintf("%s/%s/%s", name[0:2], name[2:4], name)
	object := s3.PutObjectInput{
		Body:        bytes.NewReader(data),
		Bucket:      aws.String(bucket),
		Key:         aws.String(filePath),
		ACL:         aws.String("public-read"),
		ContentType: aws.String("image/png"),
	}
	_, err := s3Client.PutObject(&object)

	if err != nil {
		log.Panic(err)
		return ""
	}

	return fmt.Sprintf("%s%s", resourceURL, filePath)
}
