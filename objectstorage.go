package main

import (
	"bytes"
	"io/ioutil"
	"strings"

	"github.com/Sugi275/oci-env-configprovider/envprovider"
	"github.com/Sugi275/oci-lego-sslupdate/loglib"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/objectstorage"
)

func uploadCertificateToObjectStorage(updateCertificater UpdateCertificater) error {
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(envprovider.GetEnvConfigProvider())
	if err != nil {
		return err
	}

	setBucketRequest := objectstorage.GetBucketRequest{
		NamespaceName: common.String(updateCertificater.ObjectStorageNamespace),
		BucketName:    common.String(updateCertificater.ObjectStorageBucketName),
	}

	// Bucketが存在していなければ,Bucketを作成
	_, err = client.GetBucket(updateCertificater.Context, setBucketRequest)
	if err != nil {
		errString := err.Error()
		if strings.Contains(errString, "does not exist in namespace") {
			loglib.Sugar.Infof("BucketName %s is not found. Request create bucket.", updateCertificater.ObjectStorageBucketName)
			err := createBucket(updateCertificater, client)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// 秘密鍵とPublicCertificateファイルをPut
	err = putFile(updateCertificater, client, updateCertificater.PrivateKeyName, updateCertificater.PrivateKey)
	if err != nil {
		return err
	}

	err = putFile(updateCertificater, client, updateCertificater.CertificateName, updateCertificater.PublicCertificate)
	if err != nil {
		return err
	}

	return nil
}

func createBucket(updateCertificater UpdateCertificater, client objectstorage.ObjectStorageClient) error {
	createBucketDetails := objectstorage.CreateBucketDetails{
		Name:             common.String(updateCertificater.ObjectStorageBucketName),
		CompartmentId:    common.String(updateCertificater.CompartmentID),
		PublicAccessType: objectstorage.CreateBucketDetailsPublicAccessTypeNopublicaccess,
	}

	createBucketRequest := objectstorage.CreateBucketRequest{
		NamespaceName:       common.String(updateCertificater.ObjectStorageNamespace),
		CreateBucketDetails: createBucketDetails,
	}

	loglib.Sugar.Infof("Request createBucket. BucketName:%s", updateCertificater.ObjectStorageBucketName)

	_, err := client.CreateBucket(updateCertificater.Context, createBucketRequest)
	if err != nil {
		return err
	}

	loglib.Sugar.Infof("Response createBucket.")

	return nil
}

func putFile(updateCertificater UpdateCertificater, client objectstorage.ObjectStorageClient, objectName string, bodyString string) error {
	buffer := bytes.NewBufferString(bodyString)
	putObjectRequest := objectstorage.PutObjectRequest{
		NamespaceName: common.String(updateCertificater.ObjectStorageNamespace),
		BucketName:    common.String(updateCertificater.ObjectStorageBucketName),
		ObjectName:    common.String(objectName),
		ContentLength: common.Int64(int64(buffer.Len())),
		PutObjectBody: ioutil.NopCloser(buffer),
	}

	loglib.Sugar.Infof("Request PutObject in ObjectStorage. BucketName:%s ObjectName:%s",
		updateCertificater.ObjectStorageBucketName,
		objectName)

	_, err := client.PutObject(updateCertificater.Context, putObjectRequest)
	if err != nil {
		return err
	}

	loglib.Sugar.Infof("Response PutObject.")

	return nil
}
