package crypto

import "encoding/base64"

// Phase 2 stores upstream keys behind an abstraction. AES-GCM encryption lands
// with production hardening; base64 avoids accidental plain display meanwhile.
func Protect(secret string) string {
	return base64.StdEncoding.EncodeToString([]byte(secret))
}

func Unprotect(protected string) string {
	raw, err := base64.StdEncoding.DecodeString(protected)
	if err != nil {
		return protected
	}
	return string(raw)
}
