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

package meta

import "fmt"

// ContainerImage builds a container image reference from a repository and a
// default tag, optionally overriding the tag and/or pinning the image to an
// immutable digest. The result follows the OCI "repository:tag[@digest]" form:
//
//   - repository:defaultTag         when neither override is set
//   - repository:tagOverride        when tagOverride is set (defaultTag ignored)
//   - repository:tag@digest         when digest is set (the tag is kept for
//     readability while the digest guarantees immutability)
//
// The digest is expected to already carry its algorithm prefix (e.g.
// "sha256:abc..."); CRD validation enforces the format on the API fields.
func ContainerImage(repository, defaultTag, tagOverride, digest string) string {
	tag := defaultTag
	if tagOverride != "" {
		tag = tagOverride
	}

	image := fmt.Sprintf("%s:%s", repository, tag)
	if digest != "" {
		image = fmt.Sprintf("%s@%s", image, digest)
	}

	return image
}
