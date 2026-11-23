package api

import (
	"bytes"
	"context"
	"log"
	"net/url"
	"strings"
	"time"

	"opentools/config"
	"opentools/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var r2Client *s3.Client
var r2Bucket string
var r2Public string

func Configure(cfg config.Config) {
	sessionKey = cfg.SessionSecret
	configureMail(cfg)
	r2Bucket = cfg.R2Bucket
	r2Public = strings.TrimRight(cfg.R2PublicBase, "/")
	r2Client = storage.New(cfg)
}

func iconBase() string {
	if r2Public == "" {
		return "/ico?host="
	}
	return r2Public + "/ico/"
}

func iconURL(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return ""
	}
	if r2Public == "" {
		return "/ico?host=" + url.QueryEscape(host)
	}
	return r2Public + "/ico/" + url.PathEscape(host)
}

func putIcon(host string, body []byte, kind string) error {
	if r2Client == nil || r2Public == "" || len(body) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	key := "ico/" + host
	_, err := r2Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(r2Bucket),
		Key:          aws.String(key),
		Body:         bytes.NewReader(body),
		ContentType:  aws.String(kind),
		CacheControl: aws.String("public, max-age=604800"),
	})
	if err != nil {
		log.Printf("r2 put %s: %v", host, err)
	}
	return err
}
