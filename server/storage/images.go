package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"strings"
	"time"

	"opentools/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const MaxArticleImage = 10 << 20

func New(cfg config.Config) *s3.Client {
	if cfg.R2AccessKeyID == "" || cfg.R2SecretAccessKey == "" || cfg.R2Bucket == "" || cfg.R2Endpoint == "" {
		return nil
	}
	return s3.New(s3.Options{
		Region: "auto", BaseEndpoint: aws.String(cfg.R2Endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.R2AccessKeyID, cfg.R2SecretAccessKey, ""),
		UsePathStyle: true,
	})
}

var ErrImage = errors.New("请选择 10 MB 以内的 PNG、JPEG、GIF 或 WebP 图片")

func PutArticleImage(ctx context.Context, client *s3.Client, bucket, public string, data []byte) (string, error) {
	if client == nil || public == "" {
		return "", errors.New("图片存储未配置")
	}
	if len(data) == 0 || len(data) > MaxArticleImage {
		return "", ErrImage
	}
	contentType := http.DetectContentType(data)
	extension := ""
	switch contentType {
	case "image/png":
		extension = "png"
	case "image/jpeg":
		extension = "jpg"
	case "image/gif":
		extension = "gif"
	case "image/webp":
		extension = "webp"
	default:
		return "", ErrImage
	}
	key := "articles/" + rand.Text() + "." + extension
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key), Body: bytes.NewReader(data),
		ContentType: aws.String(contentType), CacheControl: aws.String("public, max-age=86400"),
	})
	if err != nil {
		return "", err
	}
	return strings.TrimRight(public, "/") + "/" + key, nil
}
