package fs

import (
	"bytes"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/juju/errors"
)

// S3 is a filesystem that uses Amazon's S3 to persist files and directories
type S3 struct {
	Bucket  string
	Service s3iface.S3API
}

func (s *S3) Name() string {
	return "S3"
}

// ReadFile reads the file named by filename and returns the contents.
func (s *S3) ReadFile(filename string) ([]byte, error) {
	input := s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &filename,
	}

	resp, err := s.Service.GetObject(&input)
	if err != nil {
		if e, ok := err.(awserr.Error); ok && e.Code() == "NoSuchKey" {
			return nil, errors.NotFoundf(filename)
		}
		return nil, errors.Annotatef(err, "while reading from s3: %s/%s", s.Bucket, filename)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Annotatef(err, "while decyphering the body: %v", resp.Body)
	}
	return data, err
}

// WriteFile writes data to a file named by filename.
// If the file does not exist, WriteFile creates it with permissions perm;
// otherwise WriteFile truncates it before writing.
// It ignores the perm field
func (s *S3) WriteFile(filename string, data []byte, perm os.FileMode) error {
	ext := path.Ext(filename)
	mime := mime.TypeByExtension(ext)
	length := int64(len(data))
	input := s3.PutObjectInput{
		Bucket:        &s.Bucket,
		Key:           &filename,
		Body:          bytes.NewReader(data),
		ContentType:   &mime,
		ContentLength: &length,
	}
	_, err := s.Service.PutObject(&input)
	if err != nil {
		return errors.Annotatef(err, "while writing on: %s/%s the data %v", s.Bucket, filename, data)
	}
	return nil
}

// MkdirAll doesn't do anything because S3 doesn't have folders
func (s *S3) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

// Remove removes the named file or directory. If there is an error, it will be of type *PathError.
func (s *S3) Remove(name string) error {
	input := s3.DeleteObjectInput{
		Bucket: &s.Bucket,
		Key:    &name,
	}
	_, err := s.Service.DeleteObject(&input)
	if err != nil {
		if e, ok := err.(awserr.Error); ok && e.Code() == "NoSuchKey" {
			return errors.NewNotFound(err, name)
		}
		return errors.Annotatef(err, "while deleting the file: %s/%s", s.Bucket, name)
	}
	return nil
}

// List returns a list of files in a directory
func (s *S3) List(prefix string) ([]File, error) {
	list := []File{}

	finished := false
	continuation := ""
	for !finished {
		input := s3.ListObjectsV2Input{
			Bucket: &s.Bucket,
			Prefix: &prefix,
		}

		if continuation != "" {
			input.ContinuationToken = &continuation
		}

		resp, err := s.Service.ListObjectsV2(&input)
		if err != nil {
			return nil, errors.Annotatef(err, "prefix %s", prefix)
		}

		for _, file := range resp.Contents {
			list = append(list, File{
				Name: filepath.Base(*file.Key),
				Size: *file.Size,
			})
		}

		if resp.NextContinuationToken != nil {
			continuation = *resp.NextContinuationToken
		} else {
			finished = true
		}
	}
	return list, nil
}
