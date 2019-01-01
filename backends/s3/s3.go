package localfs

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/andreimarcu/linx-server/torrent"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/zeebo/bencode"
)

type S3Backend struct {
	bucket string
	svc *S3
}

func (b S3Backend) Delete(key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(key),
	}
	_, err := b.svc.DeleteObject(input)
	return os.Remove(path.Join(b.bucket, key))
}

func (b S3Backend) Exists(key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(key),
	}
	_, err := b.svc.HeadObject(input)
	return err == nil, err
}

func (b S3Backend) Get(key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(key),
	}
	result, err := b.svc.GetObject(input)
	if err != nil {
		return []byte{}, err
	}
	defer result.Body.Close()

	return ioutil.ReadAll(result.Body)
}

func (b S3Backend) Put(key string, r io.Reader) (int64, error) {
	uploader := s3manager.NewUploaderWithClient(b.svc)
	input := &s3manager.UploadInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(key),
		Body: r,
	}
	result, err := uploader.Upload(input)
	if err != nil {
		return 0, err
	}

	return -1, nil
}

func (b S3Backend) Open(key string) (backends.ReadSeekCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(key),
	}
	result, err := b.svc.GetObject(input)
	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

func (b S3Backend) ServeFile(key string, w http.ResponseWriter, r *http.Request) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(key),
	}
	result, err := b.svc.GetObject(input)
	if err != nil {
		return err
	}
	defer result.Body.Close()

	http.ServeContent(w, r, key, *result.LastModified, result.Body)
	return nil
}

func (b S3Backend) Size(key string) (int64, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(key),
	}
	result, err := b.svc.HeadObject(input)
	if err != nil {
		return 0, err
	}

	return *result.ContentLength, nil
}

func (b S3Backend) GetTorrent(fileName string, url string) (t torrent.Torrent, err error) {
	input := &s3.GetObjectTorrentInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(fileName),
	}
	result, err := b.svc.GetObjectTorrent(input)
	if err != nil {
		return
	}
	defer result.Body.Close()

	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return
	}

	err = bencode.DecodeBytes(data, &t)
	if err != nil {
		return
	}

	t.Info.Name = fileName
	t.UrlList = []string{url}
	return
}

func (b S3Backend) List() ([]string, error) {
	var output []string
	input := &s3.ListObjectsInput{
		bucket: aws.String(b.bucket),
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

func NewS3Backend(bucket string, region string, endpoint string) S3Backend {
	awsConfig := &aws.Config{}
	if region != "" {
		awsConfig.Region = aws.String(region)
	}
	if endpoint != "" {
		awsConfig.Endpoint = aws.String(endpoint)
	}

	sess := session.Must(session.NewSession(awsConfig))
	svc := s3.New(sess)
	return S3Backend{bucket: bucket, svc: svc}
}
