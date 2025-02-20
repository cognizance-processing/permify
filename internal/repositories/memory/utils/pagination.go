package utils

import (
	"encoding/base64"

	"permify/pkg/database"
)

type (
	// ContinuousToken - Structure for continuous token
	ContinuousToken struct {
		Value string
	}
	// EncodedContinuousToken - Structure for encoded continuous token
	EncodedContinuousToken struct {
		Value string
	}
)

// NewContinuousToken - Creates a new continuous token
func NewContinuousToken(value string) database.ContinuousToken {
	return &ContinuousToken{
		Value: value,
	}
}

// Encode - Encodes the token to a string
func (t ContinuousToken) Encode() database.EncodedContinuousToken {
	return EncodedContinuousToken{
		Value: base64.StdEncoding.EncodeToString([]byte(t.Value)),
	}
}

// Decode decodes the token from a string
func (t EncodedContinuousToken) Decode() (database.ContinuousToken, error) {
	b, err := base64.StdEncoding.DecodeString(t.Value)
	if err != nil {
		return nil, err
	}
	return ContinuousToken{
		Value: string(b),
	}, nil
}

// Decode decodes the token from a string
func (t EncodedContinuousToken) String() string {
	return t.Value
}

type (
	NoopContinuousToken struct {
		Value string
	}
	NoopEncodedContinuousToken struct {
		Value string
	}
)

// NewNoopContinuousToken - Creates a new continuous token
func NewNoopContinuousToken() database.ContinuousToken {
	return &NoopContinuousToken{
		Value: "",
	}
}

// Encode - Encodes the token to a string
func (t NoopContinuousToken) Encode() database.EncodedContinuousToken {
	return NoopEncodedContinuousToken{
		Value: "",
	}
}

// Decode decodes the token from a string
func (t NoopEncodedContinuousToken) Decode() (database.ContinuousToken, error) {
	return NoopContinuousToken{
		Value: "",
	}, nil
}

// Decode decodes the token from a string
func (t NoopEncodedContinuousToken) String() string {
	return ""
}
