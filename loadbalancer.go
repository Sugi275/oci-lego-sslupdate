package main

import (
	"fmt"
	"time"

	"github.com/Sugi275/oci-env-configprovider/envprovider"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/loadbalancer"
)

func updateCertificate(updateCertificater UpdateCertificater) error {
	client, err := loadbalancer.NewLoadBalancerClientWithConfigurationProvider(envprovider.GetEnvConfigProvider())
	if err != nil {
		return err
	}

	// Let's Encryptから取得した新しいSSL証明書を使用して、OCI上でCertificateを作成
	workRequestID, err := createNewOCICertificate(updateCertificater, client)
	if err != nil {
		return err
	}

	// Requestの完了を待機
	err = waitWorkRequest(updateCertificater, client, workRequestID)
	if err != nil {
		return err
	}

	// Listenerに新しいCertificateを設定
	workRequestIDs, deleteCertificateNameMap, err := setNewOCICertificate(updateCertificater, client)
	if err != nil {
		return err
	}

	// Requestの完了を待機
	for _, workRequestID := range workRequestIDs {
		err = waitWorkRequest(updateCertificater, client, workRequestID)
		if err != nil {
			return err
		}
	}

	// 古いCertificateを削除
	for _, deleteCertificateName := range deleteCertificateNameMap {
		workRequestID, err = deleteCertificate(updateCertificater, client, deleteCertificateName)
		if err != nil {
			return err
		}

		// Requestの完了を待機
		err := waitWorkRequest(updateCertificater, client, workRequestID)
		if err != nil {
			return err
		}
	}

	return nil
}

func createNewOCICertificate(updateCertificater UpdateCertificater, client loadbalancer.LoadBalancerClient) (workRequestID string, err error) {
	createCertificateDetails := loadbalancer.CreateCertificateDetails{
		CertificateName:   common.String(updateCertificater.CertificateName),
		PrivateKey:        common.String(updateCertificater.PrivateKey),
		PublicCertificate: common.String(updateCertificater.PublicCertificate),
	}

	request := loadbalancer.CreateCertificateRequest{
		CreateCertificateDetails: createCertificateDetails,
		LoadBalancerId:           common.String(updateCertificater.LoadbalancerID),
	}

	createCertificateResponse, err := client.CreateCertificate(updateCertificater.Context, request)

	if err != nil {
		return "", err
	}

	return *createCertificateResponse.OpcWorkRequestId, nil
}

func waitWorkRequest(updateCertificater UpdateCertificater, client loadbalancer.LoadBalancerClient, workRequestID string) error {
	getWorkRequestRequest := loadbalancer.GetWorkRequestRequest{
		WorkRequestId: common.String(workRequestID),
	}

	state := loadbalancer.WorkRequestLifecycleStateAccepted
	for state == loadbalancer.WorkRequestLifecycleStateAccepted || state == loadbalancer.WorkRequestLifecycleStateInProgress {
		response, err := client.GetWorkRequest(updateCertificater.Context, getWorkRequestRequest)
		if err != nil {
			return err
		}
		state = response.LifecycleState
		time.Sleep(5 * time.Second)
	}

	if state == loadbalancer.WorkRequestLifecycleStateFailed {
		return fmt.Errorf("Failed Work Request: workRequestID %s", workRequestID)
	}

	return nil
}

func setNewOCICertificate(updateCertificater UpdateCertificater, client loadbalancer.LoadBalancerClient) (workRequestIDs []string, deleteCertificateNameMap map[string]string, err error) {
	// LoadBalancerのListenerMapを取得する
	getLoadBalancerRequest := loadbalancer.GetLoadBalancerRequest{
		LoadBalancerId: common.String(updateCertificater.LoadbalancerID),
	}

	getLoadBalancerResponse, err := client.GetLoadBalancer(updateCertificater.Context, getLoadBalancerRequest)
	if err != nil {
		return nil, nil, err
	}

	// 更新対象のListenerNameのみ、新しいCertificateをsetする
	// 更新対象のListenerに設定されているCertificateを削除対象として、deleteCertificateIDMapに格納する。Mapを使用しているのは、重複して格納しないため
	deleteCertificateNameMap = map[string]string{}
	for _, listenerName := range updateCertificater.ListenerNames {
		_, exist := getLoadBalancerResponse.LoadBalancer.Listeners[listenerName]
		if !exist {
			return nil, nil, fmt.Errorf("Lister Not Found in OracleCloud: ListenerName %s", listenerName)
		}

		sslConfigurationDetails := loadbalancer.SslConfigurationDetails{
			CertificateName: common.String(updateCertificater.CertificateName),
		}

		updateListenerDetails := loadbalancer.UpdateListenerDetails{
			DefaultBackendSetName: getLoadBalancerResponse.LoadBalancer.Listeners[listenerName].DefaultBackendSetName,
			Port:                  getLoadBalancerResponse.LoadBalancer.Listeners[listenerName].Port,
			Protocol:              getLoadBalancerResponse.LoadBalancer.Listeners[listenerName].Protocol,
			SslConfiguration:      &sslConfigurationDetails,
		}

		updateListenerRequest := loadbalancer.UpdateListenerRequest{
			UpdateListenerDetails: updateListenerDetails,
			LoadBalancerId:        common.String(updateCertificater.LoadbalancerID),
			ListenerName:          common.String(listenerName),
		}

		response, err := client.UpdateListener(updateCertificater.Context, updateListenerRequest)
		if err != nil {
			return nil, nil, err
		}
		workRequestIDs = append(workRequestIDs, *response.OpcWorkRequestId)

		// 更新対象のListenerに設定されているCertificateを削除対象として、deleteCertificateIDMapに格納する
		oldCetrificateName := *getLoadBalancerResponse.LoadBalancer.Listeners[listenerName].SslConfiguration.CertificateName
		_, exist = deleteCertificateNameMap[oldCetrificateName]
		if !exist {
			deleteCertificateNameMap[oldCetrificateName] = oldCetrificateName
		}
	}
	return workRequestIDs, deleteCertificateNameMap, nil
}

func deleteCertificate(updateCertificater UpdateCertificater, client loadbalancer.LoadBalancerClient, deleteCertificateName string) (workRequestID string, err error) {
	deleteCertificateRequest := loadbalancer.DeleteCertificateRequest{
		LoadBalancerId:  common.String(updateCertificater.LoadbalancerID),
		CertificateName: common.String(deleteCertificateName),
	}

	response, err := client.DeleteCertificate(updateCertificater.Context, deleteCertificateRequest)
	if err != nil {
		return "", err
	}

	return *response.OpcWorkRequestId, nil
}
