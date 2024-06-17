// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package encryption

type EncryptInterface interface {
	// Encrypt a plain text string, returns the encrypted string in hex format or an error
	Encrypt(string) (string, error)
	// Decrypt a hex string, returns the decrypted string or an error
	Decrypt(string) (string, error)
}
