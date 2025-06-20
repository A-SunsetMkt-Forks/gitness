// Copyright 2023 Harness, Inc.
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

package enum

type PublicKeyScheme string

const (
	PublicKeySchemeSSH PublicKeyScheme = "ssh"
	PublicKeySchemePGP PublicKeyScheme = "pgp"
)

var publicKeySchemes = sortEnum([]PublicKeyScheme{
	PublicKeySchemeSSH, PublicKeySchemePGP,
})

func (PublicKeyScheme) Enum() []interface{} { return toInterfaceSlice(publicKeySchemes) }
func (s PublicKeyScheme) Sanitize() (PublicKeyScheme, bool) {
	return Sanitize(s, GetAllPublicKeySchemes)
}
func GetAllPublicKeySchemes() ([]PublicKeyScheme, PublicKeyScheme) {
	return publicKeySchemes, PublicKeySchemeSSH
}

// PublicKeyUsage represents usage type of public key.
type PublicKeyUsage string

// PublicKeyUsage enumeration.
const (
	PublicKeyUsageAuth PublicKeyUsage = "auth"
	PublicKeyUsageSign PublicKeyUsage = "sign"
)

var publicKeyTypes = sortEnum([]PublicKeyUsage{
	PublicKeyUsageAuth, PublicKeyUsageSign,
})

func (PublicKeyUsage) Enum() []interface{} { return toInterfaceSlice(publicKeyTypes) }
func (s PublicKeyUsage) Sanitize() (PublicKeyUsage, bool) {
	return Sanitize(s, GetAllPublicKeyUsages)
}
func GetAllPublicKeyUsages() ([]PublicKeyUsage, PublicKeyUsage) {
	return publicKeyTypes, PublicKeyUsageAuth
}

// PublicKeySort is used to specify sorting of public keys.
type PublicKeySort string

// PublicKeySort enumeration.
const (
	PublicKeySortCreated    PublicKeySort = "created"
	PublicKeySortIdentifier PublicKeySort = "identifier"
)

var publicKeySorts = sortEnum([]PublicKeySort{
	PublicKeySortCreated,
	PublicKeySortIdentifier,
})

func (PublicKeySort) Enum() []interface{}               { return toInterfaceSlice(publicKeySorts) }
func (s PublicKeySort) Sanitize() (PublicKeySort, bool) { return Sanitize(s, GetAllPublicKeySorts) }
func GetAllPublicKeySorts() ([]PublicKeySort, PublicKeySort) {
	return publicKeySorts, PublicKeySortCreated
}

// RevocationReason is the reason why a public key has been revoked.
type RevocationReason string

// RevocationReason enumeration.
const (
	RevocationReasonUnknown     RevocationReason = "unknown"
	RevocationReasonSuperseded  RevocationReason = "superseded"
	RevocationReasonRetired     RevocationReason = "retired"
	RevocationReasonCompromised RevocationReason = "compromised"
)

var revocationReasons = sortEnum([]RevocationReason{
	RevocationReasonUnknown,
	RevocationReasonSuperseded,
	RevocationReasonRetired,
	RevocationReasonCompromised,
})

func (RevocationReason) Enum() []interface{} { return toInterfaceSlice(revocationReasons) }
func (s RevocationReason) Sanitize() (RevocationReason, bool) {
	return Sanitize(s, GetAllRevocationReasons)
}
func GetAllRevocationReasons() ([]RevocationReason, RevocationReason) {
	return revocationReasons, ""
}
