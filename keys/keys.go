package keys

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"

	"code.google.com/p/gopass"
)

func ReadPEM(content []byte, password bool) (*pem.Block, error) {
	var keyDERBlock *pem.Block
	for {
		keyDERBlock, content = pem.Decode(content)

		if keyDERBlock == nil {
			return nil, errors.New("crypto/tls: failed to parse key PEM data")
		}
		if keyDERBlock.Type == "PRIVATE KEY" || strings.HasSuffix(keyDERBlock.Type, " PRIVATE KEY") {
			break
		}
	}
	var derBytes []byte
	if strings.Contains(keyDERBlock.Headers["Proc-Type"], "ENCRYPTED") && password {
		password, _ := gopass.GetPass("This key is encrypted, please enter the passphrase:")
		derBytes, _ = x509.DecryptPEMBlock(keyDERBlock, []byte(password))
	} else {
		derBytes = keyDERBlock.Bytes
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(derBytes)

	if err != nil {
		return nil, err
	}
	return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}, nil
}
