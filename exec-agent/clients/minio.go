package clients

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type MinioClient struct {
	client     *minio.Client
	bucketName string
}

func NewMinioClient(endpoint, accessKeyID, secretAccessKey, bucketName string, useSSL bool) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Minio client: %v", err)
	}

	// Test connection and ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %v", err)
		}
		logrus.WithField("bucket", bucketName).Info("Created Minio bucket")
	}

	logrus.WithFields(logrus.Fields{
		"endpoint":   endpoint,
		"bucket":     bucketName,
		"use_ssl":    useSSL,
	}).Info("Minio client initialized")

	return &MinioClient{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (m *MinioClient) DownloadFile(ctx context.Context, objectName, destPath string) error {
	logrus.WithFields(logrus.Fields{
		"object": objectName,
		"dest":   destPath,
	}).Debug("Downloading file from Minio")

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Get object from Minio
	object, err := m.client.GetObject(ctx, m.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object from Minio: %v", err)
	}
	defer object.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destFile.Close()

	// Copy object content to file
	_, err = io.Copy(destFile, object)
	if err != nil {
		return fmt.Errorf("failed to copy object content: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"object": objectName,
		"dest":   destPath,
	}).Info("File downloaded successfully")

	return nil
}

func (m *MinioClient) UploadFile(ctx context.Context, filePath, objectName string) error {
	logrus.WithFields(logrus.Fields{
		"file":   filePath,
		"object": objectName,
	}).Debug("Uploading file to Minio")

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %v", err)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Upload file
	_, err = m.client.PutObject(ctx, m.bucketName, objectName, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("failed to upload object to Minio: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"file":   filePath,
		"object": objectName,
		"size":   fileInfo.Size(),
	}).Info("File uploaded successfully")

	return nil
}

func (m *MinioClient) UploadDirectory(ctx context.Context, dirPath, prefix string) ([]string, error) {
	var uploadedObjects []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Calculate relative path for object name
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		// Convert to forward slashes for object name
		objectName := filepath.ToSlash(filepath.Join(prefix, relPath))

		if err := m.UploadFile(ctx, path, objectName); err != nil {
			return fmt.Errorf("failed to upload %s: %v", path, err)
		}

		uploadedObjects = append(uploadedObjects, objectName)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %v", err)
	}

	return uploadedObjects, nil
}

func (m *MinioClient) DownloadDirectory(ctx context.Context, prefix, destDir string) error {
	logrus.WithFields(logrus.Fields{
		"prefix":   prefix,
		"dest_dir": destDir,
	}).Debug("Downloading directory from Minio")

	// List objects with prefix
	objectCh := m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return fmt.Errorf("failed to list objects: %v", object.Err)
		}

		// Skip if it's a directory marker
		if strings.HasSuffix(object.Key, "/") {
			continue
		}

		// Calculate destination path
		relPath := strings.TrimPrefix(object.Key, prefix)
		relPath = strings.TrimPrefix(relPath, "/")
		destPath := filepath.Join(destDir, relPath)

		if err := m.DownloadFile(ctx, object.Key, destPath); err != nil {
			return fmt.Errorf("failed to download %s: %v", object.Key, err)
		}
	}

	return nil
}

func (m *MinioClient) GenerateObjectPath(executionID, filename string) string {
	return fmt.Sprintf("executions/%s/%s", executionID, filename)
}

func (m *MinioClient) GenerateDirectoryPath(executionID, dirname string) string {
	return fmt.Sprintf("executions/%s/%s", executionID, dirname)
}

func (m *MinioClient) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	_, err := m.client.StatObject(ctx, m.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *MinioClient) DeleteObject(ctx context.Context, objectName string) error {
	return m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
}

func (m *MinioClient) DeleteObjects(ctx context.Context, objectNames []string) error {
	objectsCh := make(chan minio.ObjectInfo, len(objectNames))

	go func() {
		defer close(objectsCh)
		for _, name := range objectNames {
			objectsCh <- minio.ObjectInfo{Key: name}
		}
	}()

	errorCh := m.client.RemoveObjects(ctx, m.bucketName, objectsCh, minio.RemoveObjectsOptions{})

	for err := range errorCh {
		if err.Err != nil {
			return fmt.Errorf("failed to delete object %s: %v", err.ObjectName, err.Err)
		}
	}

	return nil
}