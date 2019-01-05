package s3

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/andreimarcu/linx-server/helpers"
	"github.com/andreimarcu/linx-server/torrent"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/zeebo/bencode"
)

type S3Backend struct {
	bucket string
	svc *s3.S3
}

func (b S3Backend) Delete(key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(key),
	}
	_, err := b.svc.DeleteObject(input)
	if err != nil {
		return err
	}
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

func (b S3Backend) Head(key string) (metadata backends.Metadata, err error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(key),
	}
	result, err := b.svc.HeadObject(input)
	if err != nil {
		return
	}

	metadata, err = unmapMetadata(result.Metadata)
	return
}

func (b S3Backend) Get(key string) (metadata backends.Metadata, r io.ReadCloser, err error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(key),
	}
	result, err := b.svc.GetObject(input)
	if err != nil {
		return
	}

	metadata, err = unmapMetadata(result.Metadata)
	r = result.Body
	return
}

func mapMetadata(m backends.Metadata) map[string]*string {
	return map[string]*string{
		"expiry": aws.String(strconv.FormatInt(m.Expiry.Unix(), 10)),
		"delete_key": aws.String(m.DeleteKey),
		"size": aws.String(strconv.FormatInt(m.Size, 10)),
		"mimetype": aws.String(m.Mimetype),
		"sha256sum": aws.String(m.Sha256sum),
	}
}

func unmapMetadata(input map[string]*string) (m backends.Metadata, err error) {
	expiry, err := strconv.ParseInt(*input["expiry"], 10, 64)
	if err != nil {
		return
	}
	m.Expiry = time.Unix(expiry, 0)

	m.Size, err = strconv.ParseInt(*input["size"], 10, 64)
	if err != nil {
		return
	}

	m.DeleteKey = *input["delete_key"]
	m.Mimetype = *input["mimetype"]
	m.Sha256sum = *input["sha256sum"]
	return
}

func (b S3Backend) Put(key string, r io.Reader, expiry time.Time, deleteKey string) (m backends.Metadata, err error) {
	tmpDst, err := ioutil.TempFile("", "linx-server-upload")
	if err != nil {
		return
	}
	defer tmpDst.Close()
	defer os.Remove(tmpDst.Name())

	bytes, err := io.Copy(tmpDst, r)
	if bytes == 0 {
		return m, backends.FileEmptyError
	} else if err != nil {
		return m, err
	}

	m.Expiry = expiry
	m.DeleteKey = deleteKey
	m.Size = bytes
	m.Mimetype, _ = helpers.DetectMime(tmpDst)
	m.Sha256sum, _ = helpers.Sha256sum(tmpDst)
	// XXX: we may not be able to write this to AWS easily
	//m.ArchiveFiles, _ = helpers.ListArchiveFiles(m.Mimetype, m.Size, tmpDst)

	uploader := s3manager.NewUploaderWithClient(b.svc)
	input := &s3manager.UploadInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(key),
		Body: tmpDst,
		Metadata: mapMetadata(m),
	}
	_, err = uploader.Upload(input)
	if err != nil {
		return
	}

	return
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
