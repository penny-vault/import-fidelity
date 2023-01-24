// Copyright 2022-2023
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

var (
	ErrBucketNotFound = errors.New("bucket not found")
	ErrSha1Mismatch   = errors.New("SHA 1 hash does not match")
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
		return ErrBucketNotFound
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
		return ErrBucketNotFound
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
		return ErrSha1Mismatch
	}

	log.Info().Str("FileName", fileInfo.Name).Int64("Size", fileInfo.ContentLength).Str("ID", fileInfo.ID).Msg("downloaded file from backblaze")
	return nil
}
