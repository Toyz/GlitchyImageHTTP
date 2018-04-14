package filemodes

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Toyz/GlitchyImageHTTP/core"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type CDNMode struct {
	s3Client *s3.S3
}

func (cdn *CDNMode) Setup() {
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

	cdn.s3Client = s3Client
}

func (cdn *CDNMode) Write(data []byte, name string) (string, string) {
	resourceURL := cdn.Path()
	bucket := core.GetEnv("AWS_BUCKET", "")

	folder := fmt.Sprintf("%s/%s/", name[0:2], name[2:4])
	filePath := fmt.Sprintf("%s%s", folder, name)
	object := s3.PutObjectInput{
		Body:        bytes.NewReader(data),
		Bucket:      aws.String(bucket),
		Key:         aws.String(filePath),
		ACL:         aws.String("public-read"),
		ContentType: aws.String(core.GetMimeTypeFromBytes(data)),
	}
	_, err := cdn.s3Client.PutObject(&object)

	if err != nil {
		log.Println(err)
		return "", ""
	}

	return fmt.Sprintf("%s%s", resourceURL, filePath), folder
}

func (cdn *CDNMode) Read(path string) []byte {
	folder := fmt.Sprintf("%s/%s/", path[0:2], path[2:4])
	url := fmt.Sprintf("%s%s%s", cdn.Path(), folder, path)

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return make([]byte, 0)
	}
	defer response.Body.Close()

	buff := new(bytes.Buffer)
	io.Copy(buff, response.Body)

	return buff.Bytes()
}

func (*CDNMode) Path() string {
	return strings.TrimSpace(core.GetEnv("AWS_RESOURCE_URL", ""))
}
