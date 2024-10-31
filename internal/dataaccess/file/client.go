package file

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"GoLoad/internal/configs"

	"github.com/minio/minio-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client interface {
	Write(ctx context.Context, filePath string) (io.WriteCloser, error)
	Read(ctx context.Context, filePath string) (io.ReadCloser, error)
}

func NewClient(downloadConfig configs.Download) (Client, error) {
	switch downloadConfig.Mode {
	case configs.DownloadModeLocal:
		return NewLocalClient(downloadConfig)
	case configs.DownloadModeS3:
		return NewS3Client(downloadConfig)
	default:
		return nil, fmt.Errorf("unsupported download mode: %s", downloadConfig.Mode)
	}
}

type bufferedFileReader struct {
	file           *os.File
	bufferedReader io.Reader
}

func newBufferedFileReader(
	file *os.File,
) io.ReadCloser {
	return &bufferedFileReader{
		file:           file,
		bufferedReader: bufio.NewReader(file),
	}
}
func (b bufferedFileReader) Close() error {
	return b.file.Close()
}
func (b bufferedFileReader) Read(p []byte) (int, error) {
	return b.bufferedReader.Read(p)
}

type LocalClient struct {
	downloadDirectory string
}

func NewLocalClient(downloadConfig configs.Download) (Client, error) {
	if err := os.MkdirAll(downloadConfig.DownloadDirectory, os.ModeDir); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("failed to create download directory: %w", err)
		}
	}
	return &LocalClient{
		downloadDirectory: downloadConfig.DownloadDirectory,
	}, nil
}
func (l LocalClient) Read(ctx context.Context, filePath string) (io.ReadCloser, error) {
	absolutePath := path.Join(l.downloadDirectory, filePath)
	file, err := os.Open(absolutePath)
	if err != nil {
		log.Printf("failed to open file")
		return nil, status.Error(codes.Internal, "failed to open file")
	}
	return newBufferedFileReader(file), nil
}
func (l *LocalClient) Write(ctx context.Context, filePath string) (io.WriteCloser, error) {
	absolutePath := path.Join(l.downloadDirectory, filePath)
	file, err := os.Create(absolutePath)
	if err != nil {
		log.Printf("failed to open file")
		return nil, status.Error(codes.Internal, "failed to open file")
	}
	return file, nil
}

type s3ClientReadWriteCloser struct {
	writtenData []byte
	isClosed    bool
}

func newS3ClientReadWriteCloser(
	ctx context.Context,
	minioClient *minio.Client,
	bucketName,
	objectName string,
) io.ReadWriteCloser {
	readWriteCloser := &s3ClientReadWriteCloser{
		writtenData: make([]byte, 0),
		isClosed:    false,
	}
	go func() {
		if _, err := minioClient.PutObjectWithContext(
			ctx, bucketName, objectName, readWriteCloser, -1, minio.PutObjectOptions{},
		); err != nil {
			log.Printf("failed to put object")
		}
	}()
	return readWriteCloser
}
func (s *s3ClientReadWriteCloser) Close() error {
	s.isClosed = true
	return nil
}
func (s *s3ClientReadWriteCloser) Read(p []byte) (int, error) {
	if len(s.writtenData) > 0 {
		writtenLength := copy(p, s.writtenData)
		s.writtenData = s.writtenData[writtenLength:]
		return writtenLength, nil
	}
	if s.isClosed {
		return 0, io.EOF
	}
	return 0, nil
}
func (s *s3ClientReadWriteCloser) Write(p []byte) (int, error) {
	s.writtenData = append(s.writtenData, p...)
	return len(p), nil
}

type S3Client struct {
	minioClient *minio.Client
	bucket      string
}

func NewS3Client(downloadConfig configs.Download) (Client, error) {
	minioClient, err := minio.New(downloadConfig.Address, downloadConfig.Username, downloadConfig.Password, false)
	if err != nil {
		log.Printf("failed to create minio client")
		return nil, err
	}
	return &S3Client{
		minioClient: minioClient,
		bucket:      downloadConfig.Bucket,
	}, nil
}
func (s S3Client) Read(ctx context.Context, filePath string) (io.ReadCloser, error) {
	object, err := s.minioClient.GetObjectWithContext(ctx, s.bucket, filePath, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("failed to get s3 object")
		return nil, status.Error(codes.Internal, "failed to get s3 object")
	}
	return object, nil
}
func (s S3Client) Write(ctx context.Context, filePath string) (io.WriteCloser, error) {
	return newS3ClientReadWriteCloser(ctx, s.minioClient, s.bucket, filePath), nil
}
