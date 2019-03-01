package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/xenolf/lego/certcrypto"
	"github.com/xenolf/lego/certificate"
	_ "github.com/xenolf/lego/challenge/dns01"
	"github.com/xenolf/lego/lego"
	"github.com/xenolf/lego/platform/config/env"
	"github.com/xenolf/lego/providers/dns"
	"github.com/xenolf/lego/registration"
)

// MyUser You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

// GetEmail get Email
func (u *MyUser) GetEmail() string {
	return u.Email
}

// GetRegistration get Registration
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey get Key
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func generateMyUser() MyUser {
	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	email := env.GetOrDefaultString("LETSENCRYPT_MY_MAILADDRESS", "default@email.com")

	return MyUser{
		Email: email,
		key:   privateKey,
	}
}

func getDomains() ([]string, error) {
	domainsString, ok := os.LookupEnv("LETSENCRYPT_DOMAINS")

	if !ok {
		err := fmt.Errorf("can not read domains from environment variable %s", "LETSENCRYPT_DOMAINS")
		return nil, err
	}

	domains := strings.Split(domainsString, ",")

	return domains, nil
}

func getCertificates() (*certificate.Resource, error) {
	myUser := generateMyUser()
	config := lego.NewConfig(&myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	// If developping, staging URL is useful.
	// https://acme-staging-v02.api.letsencrypt.org/directory
	config.CADirURL = env.GetOrDefaultString("LETSENCRYPT_CA_URL", "https://acme-v02.api.letsencrypt.org/directory")
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	provider, err := dns.NewDNSChallengeProviderByName("oraclecloud")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	err = client.Challenge.SetDNS01Provider(provider)

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	myUser.Registration = reg

	domains, err := getDomains()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return certificates, nil
}
