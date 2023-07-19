// Licensed to Alexandre VILAIN under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Alexandre VILAIN licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package archival

import (
	"net/url"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"go.temporal.io/server/common/archiver/filestore"
	"go.temporal.io/server/common/archiver/gcloud"
	"go.temporal.io/server/common/archiver/s3store"
	"go.temporal.io/server/common/config"
)

// URI returns archival spec URI using the provided provider.
func URI(provider *v1beta1.ArchivalProvider, spec *v1beta1.ArchivalSpec) string {
	scheme := ""
	switch provider.Kind() {
	case v1beta1.FileStoreArchivalProviderKind:
		scheme = filestore.URIScheme
	case v1beta1.S3ArchivalProviderKind:
		scheme = s3store.URIScheme
	case v1beta1.GCSArchivalProviderKind:
		scheme = gcloud.URIScheme
	default:
		return ""
	}

	u := &url.URL{}
	u.Scheme = scheme
	u.Path = spec.Path

	return u.String()
}

func FilestoreArchiverToTemporalFilestoreArchiver(a *v1beta1.FilestoreArchiver) *config.FilestoreArchiver {
	if a == nil {
		return nil
	}

	return &config.FilestoreArchiver{
		DirMode:  a.DirPermissions,
		FileMode: a.FilePermissions,
	}
}

func S3ArchiverToTemporalS3Archiver(a *v1beta1.S3Archiver) *config.S3Archiver {
	if a == nil {
		return nil
	}

	return &config.S3Archiver{
		Region:           a.Region,
		Endpoint:         a.Endpoint,
		S3ForcePathStyle: false, // TODO(alexandrevilain): See implications
	}
}

func GCSArchiverToTemporalGstorageArchiver(a *v1beta1.GCSArchiver) *config.GstorageArchiver {
	if a == nil {
		return nil
	}

	return &config.GstorageArchiver{
		CredentialsPath: a.CredentialsFileMountPath(),
	}
}
