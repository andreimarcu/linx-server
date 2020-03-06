package s3

import (
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/andreimarcu/linx-server/helpers"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Backend struct {
	bucket string
	svc    *s3.S3
}

func (b S3Backend) Delete(key string) error {
	_, err := b.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}

func (b S3Backend) Exists(key string) (bool, error) {
	_, err := b.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	return err == nil, err
}

func (b S3Backend) Head(key string) (metadata backends.Metadata, err error) {
	var result *s3.HeadObjectOutput
	result, err = b.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey || aerr.Code() == "NotFound" {
				err = backends.NotFoundErr
			}
		}
		return
	}

	metadata, err = unmapMetadata(result.Metadata)
	return
}

func (b S3Backend) Get(key string) (metadata backends.Metadata, r io.ReadCloser, err error) {
	var result *s3.GetObjectOutput
	result, err = b.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey || aerr.Code() == "NotFound" {
				err = backends.NotFoundErr
			}
		}
		return
	}

	metadata, err = unmapMetadata(result.Metadata)
	r = result.Body
	return
}

func mapMetadata(m backends.Metadata) map[string]*string {
	return map[string]*string{
		"Expiry":    aws.String(strconv.FormatInt(m.Expiry.Unix(), 10)),
		"Deletekey": aws.String(m.DeleteKey),
		"Size":      aws.String(strconv.FormatInt(m.Size, 10)),
		"Mimetype":  aws.String(m.Mimetype),
		"Sha256sum": aws.String(m.Sha256sum),
		"AccessKey": aws.String(m.AccessKey),
	}
}

func unmapMetadata(input map[string]*string) (m backends.Metadata, err error) {
	expiry, err := strconv.ParseInt(aws.StringValue(input["Expiry"]), 10, 64)
	if err != nil {
		return m, err
	}
	m.Expiry = time.Unix(expiry, 0)

	m.Size, err = strconv.ParseInt(aws.StringValue(input["Size"]), 10, 64)
	if err != nil {
		return
	}

	m.DeleteKey = aws.StringValue(input["Deletekey"])
	if m.DeleteKey == "" {
		m.DeleteKey = aws.StringValue(input["Delete_key"])
	}

	m.Mimetype = aws.StringValue(input["Mimetype"])
	m.Sha256sum = aws.StringValue(input["Sha256sum"])

	if key, ok := input["AccessKey"]; ok {
		m.AccessKey = aws.StringValue(key)
	}

	return
}

func (b S3Backend) Put(key string, r io.Reader, expiry time.Time, deleteKey, accessKey string) (m backends.Metadata, err error) {
	tmpDst, err := ioutil.TempFile("", "linx-server-upload")
	if err != nil {
		return m, err
	}
	defer tmpDst.Close()
	defer os.Remove(tmpDst.Name())

	bytes, err := io.Copy(tmpDst, r)
	if bytes == 0 {
		return m, backends.FileEmptyError
	} else if err != nil {
		return m, err
	}

	_, err = tmpDst.Seek(0, 0)
	if err != nil {
		return m, err
	}

	m, err = helpers.GenerateMetadata(tmpDst)
	if err != nil {
		return
	}
	m.Expiry = expiry
	m.DeleteKey = deleteKey
	m.AccessKey = accessKey
	// XXX: we may not be able to write this to AWS easily
	//m.ArchiveFiles, _ = helpers.ListArchiveFiles(m.Mimetype, m.Size, tmpDst)

	_, err = tmpDst.Seek(0, 0)
	if err != nil {
		return m, err
	}

	uploader := s3manager.NewUploaderWithClient(b.svc)
	input := &s3manager.UploadInput{
		Bucket:   aws.String(b.bucket),
		Key:      aws.String(key),
		Body:     tmpDst,
		Metadata: mapMetadata(m),
	}
	_, err = uploader.Upload(input)
	if err != nil {
		return
	}

	return
}

func (b S3Backend) PutMetadata(key string, m backends.Metadata) (err error) {
	_, err = b.svc.CopyObject(&s3.CopyObjectInput{
		Bucket:            aws.String(b.bucket),
		Key:               aws.String(key),
		CopySource:        aws.String("/" + b.bucket + "/" + key),
		Metadata:          mapMetadata(m),
		MetadataDirective: aws.String("REPLACE"),
	})
	if err != nil {
		return
	}

	return
}

func (b S3Backend) Size(key string) (int64, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	}
	result, err := b.svc.HeadObject(input)
	if err != nil {
		return 0, err
	}

	return *result.ContentLength, nil
}

func (b S3Backend) List() ([]string, error) {
	var output []string
	input := &s3.ListObjectsInput{
		Bucket: aws.String(b.bucket),
	}

	results, err := b.svc.ListObjects(input)
	if err != nil {
		return nil, err
	}

	for _, object := range results.Contents {
		output = append(output, *object.Key)
	}

	return output, nil
}

func NewS3Backend(bucket string, region string, endpoint string, forcePathStyle bool) S3Backend {
	awsConfig := &aws.Config{}
	if region != "" {
		awsConfig.Region = aws.String(region)
	}
	if endpoint != "" {
		awsConfig.Endpoint = aws.String(endpoint)
	}
	if forcePathStyle == true {
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	sess := session.Must(session.NewSession(awsConfig))
	svc := s3.New(sess)
	return S3Backend{bucket: bucket, svc: svc}
}
