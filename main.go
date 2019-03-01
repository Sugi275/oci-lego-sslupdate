package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Sugi275/oci-env-configprovider/envprovider"
	"github.com/Sugi275/oci-lego-sslupdate/loglib"
	"github.com/xenolf/lego/platform/config/env"
)

//UpdateCertificater UpdateCertificater
type UpdateCertificater struct {
	LoadbalancerID          string
	ListenerNames           []string
	CertificateName         string
	PrivateKeyName          string
	PrivateKey              string
	PublicCertificate       string
	ObjectStorageBucketName string
	ObjectStorageNamespace  string
	CompartmentID           string
	Context                 context.Context
}

const (
	envLoadbalancerID          = "OCI_LB_OCID"
	envListenerNames           = "OCI_LISTENERS"
	envObjectStorageBucketName = "OCI_OS_BUCKETNAME"
	envObjectStorageNamespace  = "OCI_OS_NAMESPACE"
)

func main() {
	ceritficateUpdateHandler()
}

func ceritficateUpdateHandler() {
	loglib.InitSugar()
	defer loglib.Sugar.Sync()

	// Let's Encrypt
	certificates, err := getCertificates()

	if err != nil {
		loglib.Sugar.Error(err)
		return
	}

	// updateCertificaterを生成して、パラメータを設定
	updateCertificater := newUpdateCertificater()

	loadbalancerID, ok := os.LookupEnv(envLoadbalancerID)
	if !ok {
		err = fmt.Errorf("can not read envLoadbalancerID from environment variable %s", envLoadbalancerID)
		loglib.Sugar.Error(err)
		return
	}
	updateCertificater.LoadbalancerID = loadbalancerID

	// 環境変数から、カンマ区切りのListenerNameを取得。カンマで文字列を分割して処理をする
	listenerNamesValue, ok := os.LookupEnv(envListenerNames)
	if !ok {
		err = fmt.Errorf("can not read envListenerNames from environment variable %s", envListenerNames)
		loglib.Sugar.Error(err)
		return
	}
	listenerNames := strings.Split(listenerNamesValue, ",")
	for _, ln := range listenerNames {
		updateCertificater.ListenerNames = append(updateCertificater.ListenerNames, ln)
	}

	updateCertificater.PrivateKey = string(certificates.PrivateKey)
	updateCertificater.PublicCertificate = string(certificates.Certificate)

	// Update to SSL Backend
	loglib.Sugar.Infof("Starting updateCertificate. LoadbalancerID:%s ListenerNames:%s",
		updateCertificater.LoadbalancerID,
		updateCertificater.ListenerNames)
	err = updateCertificate(updateCertificater)
	if err != nil {
		loglib.Sugar.Error(err)
		return
	}
	loglib.Sugar.Infof("Successful updateCertificate.")

	// Upload certificate to Object Storage
	bucketName := env.GetOrDefaultString(envObjectStorageBucketName, "lego-cert")
	updateCertificater.ObjectStorageBucketName = bucketName

	namespace, ok := os.LookupEnv(envObjectStorageNamespace)
	if !ok {
		err = fmt.Errorf("can not read namespace from environment variable %s", envObjectStorageNamespace)
		loglib.Sugar.Error(err)
		return
	}
	updateCertificater.ObjectStorageNamespace = namespace

	compartmentID, err := envprovider.GetCompartmentID()
	if err != nil {
		loglib.Sugar.Error(err)
		return
	}
	updateCertificater.CompartmentID = compartmentID

	err = uploadCertificateToObjectStorage(updateCertificater)
	if err != nil {
		loglib.Sugar.Error(err)
		return
	}

	loglib.Sugar.Infof("Successful! Complete update SSL certificate")
}

func newUpdateCertificater() UpdateCertificater {
	// Generate certificate name
	const DateFormat = "20060102-1504"

	return UpdateCertificater{
		CertificateName: "lego-cert-" + time.Now().Format(DateFormat),
		PrivateKeyName:  "lego-privatekey-" + time.Now().Format(DateFormat),
		Context:         context.Background(),
	}
}
