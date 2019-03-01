package envprovider

import (
	"crypto/rsa"
	b64 "encoding/base64"
	"fmt"
	"os"

	"github.com/oracle/oci-go-sdk/common"
)

const (
	envPrivKeyEncoded    = "OCI_PRIVKEY_BASE64"
	envPrivKeyPassphrase = "OCI_PRIVKEY_PASS"
	envTenancyID         = "OCI_TENANCY_OCID"
	envUserID            = "OCI_USER_OCID"
	envPubKeyFingerPrint = "OCI_PUBKEY_FINGERPRINT"
	envRegion            = "OCI_REGION"
)

//GetEnvConfigProvider 環境変数を使用したConfigrationProviderを取得する
func GetEnvConfigProvider() common.ConfigurationProvider {
	return envConfigProvider{}
}

type envConfigProvider struct {
}

func (p envConfigProvider) PrivateRSAKey() (key *rsa.PrivateKey, err error) {
	var privateKeyEncoded string
	var privateKeyPassphrase string
	var ok bool

	if privateKeyEncoded, ok = os.LookupEnv(envPrivKeyEncoded); !ok {
		err = fmt.Errorf("can not read PrivateKeyEncoded from environment variable %s", envPrivKeyEncoded)
		return nil, err
	}

	if privateKeyPassphrase, ok = os.LookupEnv(envPrivKeyPassphrase); !ok {
		// err = fmt.Errorf("can not read PrivateKeyPassphrase from environment variable %s", envKeyPassphrase)
		// err = fmt.Errorf("and set kuuhaku to privateKeyPassphrase")
		privateKeyPassphrase = ""
	}

	privateKeyDecoded, _ := b64.StdEncoding.DecodeString(privateKeyEncoded)

	key, err = common.PrivateKeyFromBytes(privateKeyDecoded, &privateKeyPassphrase)
	return key, nil
}

func (p envConfigProvider) KeyID() (keyID string, err error) {
	ocid, err := p.TenancyOCID()
	if err != nil {
		return "", err
	}

	userocid, err := p.UserOCID()
	if err != nil {
		return "", err
	}

	fingerprint, err := p.KeyFingerprint()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", ocid, userocid, fingerprint), nil
}

func (p envConfigProvider) TenancyOCID() (value string, err error) {
	var ok bool
	if value, ok = os.LookupEnv(envTenancyID); !ok {
		err = fmt.Errorf("can not read Tenancy from environment variable %s", envTenancyID)
		return "", err
	}

	return value, nil
}

func (p envConfigProvider) UserOCID() (value string, err error) {
	var ok bool
	if value, ok = os.LookupEnv(envUserID); !ok {
		err = fmt.Errorf("can not read user id from environment variable %s", envUserID)
		return "", err
	}

	return value, nil
}

func (p envConfigProvider) KeyFingerprint() (value string, err error) {
	var ok bool
	if value, ok = os.LookupEnv(envPubKeyFingerPrint); !ok {
		err = fmt.Errorf("can not read fingerprint from environment variable %s", envPubKeyFingerPrint)
		return "", err
	}

	return value, nil
}

func (p envConfigProvider) Region() (value string, err error) {
	var ok bool
	if value, ok = os.LookupEnv(envRegion); !ok {
		err = fmt.Errorf("can not read region from environment variable %s", envRegion)
		return "", err
	}

	return value, nil
}
