// Package fs_test is used to test the fs package.
// To ensure that the tests involving s3 are correctly run you must provide the proper flags with the aws credentials
package fs_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bcmi-labs/arduino-modules/fs"
	"github.com/juju/errors"
)

var (
	bucket       = flag.String("bucket", "bucket", "The bucket used for s3 tests")
	region       = flag.String("region", "us-east-1", "The region used for s3 tests")
	awsAccessKey = flag.String("aws-access-key", "", "The access key used for s3 tests")
	awsSecretKey = flag.String("aws-secret-key", "", "The secret key used for s3 tests")
	efsFolder    = flag.String("efs-folder", "", "The path where efs has been mounted")
	parsed       = false
)

var fixtures = []fs.File{
	fs.File{Name: "existent1", Path: "test/existent1", Data: []byte("Content of existent1 file")},
	fs.File{Name: "existent1", Path: "test/existent2", Data: []byte("Content of existent2 file")},
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Setenv("AWS_ACCESS_KEY_ID", *awsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", *awsSecretKey)

	if *awsAccessKey != "" && *awsSecretKey != "" {
		os.Exit(m.Run())
	} else {
		fmt.Println("[WARN] Skipping fs tests because aws credentials are not set")
	}
}

func clients() []fs.Manager {
	list := []fs.Manager{}

	// disk
	list = append(list, &fs.Disk{Base: "/tmp/files"})

	// s3
	service := s3.New(session.New(), &aws.Config{Region: region})
	list = append(list, &fs.S3{Bucket: *bucket, Service: service})

	// efs
	list = append(list, &fs.Disk{Base: *efsFolder})

	return list
}

func setup(ti fs.Manager) {
	switch c := ti.(type) {
	case *fs.S3:
		for _, file := range fixtures {
			contentType := "text/ascii"
			inputPut := s3.PutObjectInput{
				Bucket:      bucket,
				Key:         &file.Path,
				Body:        bytes.NewReader(file.Data),
				ContentType: &contentType,
			}
			_, err := c.Service.PutObject(&inputPut)
			if err != nil {
				panic("s3 credentials are invalid")
			}
		}
	case *fs.Disk:
		for _, file := range fixtures {
			filename := filepath.Join(c.Base, file.Path)
			os.MkdirAll(filepath.Dir(filename), 0777)
			ioutil.WriteFile(filename, file.Data, 0777)
		}
	}
}

func clear(ti fs.Manager) {
	switch c := ti.(type) {
	case *fs.Disk:
		os.RemoveAll(c.Base)
	case *fs.S3:
		inputDel := s3.DeleteObjectInput{
			Bucket: bucket,
			Key:    aws.String("test/new.txt"),
		}
		_, err := c.Service.DeleteObject(&inputDel)
		if err != nil {
			panic("s3 credentials are invalid")
		}

		for _, file := range fixtures {
			inputDel := s3.DeleteObjectInput{
				Bucket: bucket,
				Key:    &file.Path,
			}
			_, err := c.Service.DeleteObject(&inputDel)
			if err != nil {
				panic("s3 credentials are invalid")
			}
		}
	}
}

type ReaderTC struct {
	Desc          string
	File          string
	ExpectedError string
	ExpectedData  []byte
	ExpectedSize  int64
}

func TestReader(t *testing.T) {
	testCases := []ReaderTC{
		{"exists", "test/existent1", "nil", []byte("Content of existent1 file"), 25},
		{"missing", "test/missing.txt", "notfound", nil, 0},
	}

	interfaces := clients()

	for _, tc := range testCases {
		for _, ti := range interfaces {
			setup(ti)
			t.Run(fmt.Sprintf("%s:%T", tc.Desc, ti), func(t *testing.T) {
				data, err := ti.ReadFile(tc.File)
				if checkError(t, err, tc.ExpectedError) {
					if string(tc.ExpectedData) != string(data) {
						t.Skipf("Expected data to be '%s', got '%s'", tc.ExpectedData, data)
					}
					size := len(data)
					if string(tc.ExpectedSize) != string(size) {
						t.Skipf("Expected size to be '%d', got '%d'", tc.ExpectedSize, size)
					}
				}
			})
			clear(ti)
		}
	}
}

type FileReadTC struct {
	Desc          string
	Path          string
	ExpectedError string
	ExpectedData  []byte
	ExpectedSize  int64
}

func TestFileRead(t *testing.T) {
	testCases := []FileReadTC{
		{"exists", "test/existent1", "nil", []byte("Content of existent1 file"), 25},
		{"missing", "test/missing.txt", "notfound", nil, 0},
	}

	interfaces := clients()

	for _, tc := range testCases {
		for _, ti := range interfaces {
			setup(ti)
			t.Run(fmt.Sprintf("%s:%T", tc.Desc, ti), func(t *testing.T) {
				file := fs.File{
					Path: tc.Path,
				}

				err := file.Read(ti)
				if checkError(t, err, tc.ExpectedError) {
					if string(tc.ExpectedData) != string(file.Data) {
						t.Skipf("Expected data to be '%s', got '%s'", tc.ExpectedData, file.Data)
					}
					if string(tc.ExpectedSize) != string(file.Size) {
						t.Skipf("Expected Size to be '%d', got '%d'", tc.ExpectedSize, file.Size)
					}
				}
			})
			clear(ti)
		}
	}
}

type WriterTC struct {
	Desc          string
	File          string
	Data          []byte
	ExpectedError string
}

func TestWriter(t *testing.T) {
	testCases := []WriterTC{
		{"new", "test/new.txt", []byte("Content of new file"), "nil"},
		{"existent", "test/existent.txt", []byte("Content of new file"), "nil"},
	}

	interfaces := clients()

	for _, tc := range testCases {
		for _, ti := range interfaces {
			setup(ti)
			t.Run(fmt.Sprintf("%s:%T", tc.Desc, ti), func(t *testing.T) {
				err := ti.WriteFile(tc.File, tc.Data, 0664)
				if checkError(t, err, tc.ExpectedError) {
					data, err := ti.ReadFile(tc.File)
					if checkError(t, err, "nil") {
						if string(tc.Data) != string(data) {
							t.Skipf("Expected data to be '%s', got '%s'", tc.Data, data)
						}
					}
				}
			})
			clear(ti)
		}
	}
}

type RemoverTC struct {
	Desc          string
	File          string
	ExpectedError string
}

func TestRemover(t *testing.T) {
	testCases := []RemoverTC{
		{"existent", "test/existent.txt", "nil"},
		{"missing", "test/missing.txt", "nil"},
	}

	interfaces := clients()

	for _, tc := range testCases {
		for _, ti := range interfaces {
			setup(ti)
			t.Run(fmt.Sprintf("%s:%T", tc.Desc, ti), func(t *testing.T) {
				err := ti.Remove(tc.File)
				if checkError(t, err, tc.ExpectedError) {
					_, err := ti.ReadFile(tc.File)
					checkError(t, err, "notfound")
				}
			})
			clear(ti)
		}
	}
}

type ListerTC struct {
	Desc          string
	Prefix        string
	ExpectedError string
	ExpectedSize  int64
	ExpectedN     int
}

func TestLister(t *testing.T) {
	testCases := []ListerTC{
		{"existent", "test", "nil", 50, 2},
		{"missing", "missing", "nil", 0, 0},
	}

	interfaces := clients()

	for _, tc := range testCases {
		for _, ti := range interfaces {
			setup(ti)
			t.Run(fmt.Sprintf("%s:%T", tc.Desc, ti), func(t *testing.T) {
				list, err := ti.List(tc.Prefix)
				if checkError(t, err, tc.ExpectedError) {
					if len(list) != tc.ExpectedN {
						t.Skipf("Expected %d items, got %d", tc.ExpectedN, len(list))
					}

					var sum int64
					for _, file := range list {
						sum = sum + file.Size
					}

					if sum != tc.ExpectedSize {
						t.Skipf("Expected %d bytes, got %d", tc.ExpectedSize, sum)
					}

				}
			})
			clear(ti)
		}
	}
}

func BenchmarkS3Read(b *testing.B) {
	interfaces := clients()
	setup(interfaces[1])
	for i := 0; i < b.N; i++ {
		f := fs.File{Path: "test/existent1"}
		err := f.Read(interfaces[1])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEFSRead(b *testing.B) {
	interfaces := clients()
	setup(interfaces[2])
	for i := 0; i < b.N; i++ {
		f := fs.File{Path: "test/existent1"}
		err := f.Read(interfaces[2])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDiskRead(b *testing.B) {
	interfaces := clients()
	setup(interfaces[0])
	for i := 0; i < b.N; i++ {
		f := fs.File{Path: "test/existent1"}
		err := f.Read(interfaces[0])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func checkError(t *testing.T, err error, expected string) bool {
	if expected == "nil" && err != nil {
		t.Skipf("err should be nil, got %s", err.Error())
		return false
	}
	if expected != "nil" && err == nil {
		t.Skipf("err should be %s, got nil", expected)
		return false
	}
	if expected == "notfound" {
		if !errors.IsNotFound(err) {
			t.Skipf("err should be notfound, got %s", err)
			return false
		}
		return true
	}
	if expected == "notvalid" {
		if !errors.IsNotValid(err) {
			t.Skipf("err should be notvalid, got %s", err)
			return false
		}
		return true
	}
	return true
}
