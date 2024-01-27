package services

import (
	"cloud.google.com/go/storage"
	"context"
	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v2/log"
	"google.golang.org/api/option"
	"io"
	"mime/multipart"
	"time"
)

func InitFirebase() (*firebase.App, error) {
	opt := option.WithCredentialsFile("config/settlesphere-56478-firebase-adminsdk-pzxjh-55b078fb52.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}
	return app, nil
}

func InitStorageClient() (*storage.Client, error) {
	storageClient, err := storage.NewClient(context.Background(), option.WithCredentialsFile("config/settlesphere-56478-firebase-adminsdk-pzxjh-55b078fb52.json"))
	if err != nil {
		return nil, err
	}
	return storageClient, nil
}

func UploadToFirebase(storageClient *storage.Client, file *multipart.FileHeader) (string, error) {
	ctx := context.Background()
	bucketName := "settlesphere-56478.appspot.com"
	bucket := storageClient.Bucket(bucketName)
	//bucket, err := storageClient.DefaultBucket()

	//objectName := url2.QueryEscape(file.Filename)
	objectName := file.Filename
	wc := bucket.Object(objectName).NewWriter(ctx)
	defer wc.Close()

	fileReader, err := file.Open()
	if err != nil {
		return "", err
	}
	log.Debug(fileReader)
	defer fileReader.Close()
	if _, err := io.Copy(wc, fileReader); err != nil {
		return "", err
	}
	//:= bucket.Object(objectName)
	//attrs, err := object.Attrs(ctx)
	//for key, value := range attrs.Metadata {
	//	fmt.Printf("%s: %s\n", key, value)
	//}
	//url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)
	//url := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s", bucketName, objectName)
	downloadUrl, err := bucket.SignedURL(file.Filename, &storage.SignedURLOptions{
		Expires: time.Now().AddDate(100, 0, 0),
		Method:  "GET",
	})
	//url, err := object.SignedURL(ctx, time.Now().Add(expiration), nil)
	log.Debug("file uploaded successfully")
	log.Debug(downloadUrl)
	return downloadUrl, nil
}
