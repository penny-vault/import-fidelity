package backblaze

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kothar/go-backblaze"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func Upload(fn, bucketName, dirName string) error {
	b2, err := backblaze.NewB2(backblaze.Credentials{
		KeyID:          viper.GetString("backblaze.application_id"),
		ApplicationKey: viper.GetString("backblaze.application_key"),
	})
	if err != nil {
		log.Error().Str("OriginalError", err.Error()).Str("BucketName", bucketName).Msg("authorize backblaze failed")
		return err
	}

	bucket, err := b2.Bucket(bucketName)
	if err != nil {
		log.Error().Str("OriginalError", err.Error()).Str("BucketName", bucketName).Msg("lookup bucket failed")
		return err
	}
	if bucket == nil {
		log.Error().Str("BucketName", bucketName).Msg("bucket does not exist")
		return errors.New("bucket not found")
	}

	reader, _ := os.Open(fn)
	defer reader.Close()

	var outName string
	if dirName == "." {
		outName = filepath.Base(fn)
	} else {
		outName = fmt.Sprintf("%s/%s", dirName, filepath.Base(fn))
	}

	metadata := make(map[string]string)

	file, err := bucket.UploadFile(outName, metadata, reader)
	if err != nil {
		log.Error().Str("OriginalError", err.Error()).Str("FileName", outName).Str("BucketName", bucketName).Msg("save file to backblaze failed")
		return err
	}

	log.Info().Str("FileName", file.Name).Int64("Size", file.ContentLength).Str("ID", file.ID).Msg("uploaded file to backblaze")
	return nil
}

func Download(fn, bucketName string) error {
	b2, err := backblaze.NewB2(backblaze.Credentials{
		KeyID:          viper.GetString("backblaze.application_id"),
		ApplicationKey: viper.GetString("backblaze.application_key"),
	})
	if err != nil {
		log.Error().Str("OriginalError", err.Error()).Str("BucketName", bucketName).Msg("authorize backblaze failed")
		return err
	}

	bucket, err := b2.Bucket(bucketName)
	if err != nil {
		log.Error().Str("OriginalError", err.Error()).Str("BucketName", bucketName).Msg("lookup bucket failed")
		return err
	}
	if bucket == nil {
		log.Error().Str("BucketName", bucketName).Msg("bucket does not exist")
		return errors.New("bucket not found")
	}

	fileInfo, reader, err := bucket.DownloadFileByName(fn)
	if err != nil {
		log.Error().Err(err).Str("FileName", fn).Str("BucketName", bucketName).Msg("read file from backblaze failed")
		return err
	}
	defer reader.Close()

	file, err := os.Create(fileInfo.Name)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := file

	sha := sha1.New()
	tee := io.MultiWriter(sha, writer)

	_, err = io.Copy(tee, reader)
	if err != nil {
		return err
	}

	// Check SHA
	sha1Hash := hex.EncodeToString(sha.Sum(nil))
	if sha1Hash != fileInfo.ContentSha1 {
		log.Error().Str("Sha1", sha1Hash).Str("ExpectedSha1", fileInfo.ContentSha1).Msg("downloaded data does not match SHA1 hash")
		return errors.New("downloaded data does not match SHA1 hash")
	}

	log.Info().Str("FileName", fileInfo.Name).Int64("Size", fileInfo.ContentLength).Str("ID", fileInfo.ID).Msg("downloaded file from backblaze")
	return nil
}
