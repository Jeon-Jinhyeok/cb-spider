// Cloud Control Manager's Rest Runtime of CB-Spider.
// The CB-Spider is a sub-Framework of the Cloud-Barista Multi-Cloud Project.
// The CB-Spider Mission is to connect all the clouds with a single interface.
//
//      * Cloud-Barista: https://github.com/cloud-barista
//
// by CB-Spider Team, 2022.09.

package commonruntime

import (
	"io/ioutil"

	ccm "github.com/cloud-barista/cb-spider/cloud-control-manager"
	cres "github.com/cloud-barista/cb-spider/cloud-control-manager/cloud-driver/interfaces/resources"
)

//================ AnyCall Handler

func AnyCall(connectionName string, reqInfo cres.AnyCallInfo) (*cres.AnyCallInfo, error) {
	cblog.Info("call AnyCall()")

	// check empty and trim user inputs
	connectionName, err := EmptyCheckAndTrim("connectionName", connectionName)
	if err != nil {
		cblog.Error(err)
		return nil, err
	}

	cldConn, err := ccm.GetCloudConnection(connectionName)
	if err != nil {
		cblog.Error(err)
		return nil, err
	}

	handler, err := cldConn.CreateAnyCallHandler()
	if err != nil {
		cblog.Error(err)
		return nil, err
	}

	//  Call AnyCall
	info, err := handler.AnyCall(reqInfo)
	if err != nil {
		cblog.Error(err)
		return nil, err
	}

	return &info, nil
}

// // for Spiderlet
func SpiderletAnyCall(connectionName string, reqInfo cres.AnyCallInfo) (*cres.AnyCallInfo, error) {
	cblog.Info("call AnyCall()")

	// check empty and trim user inputs
	connectionName, err := EmptyCheckAndTrim("connectionName", connectionName)
	if err != nil {
		cblog.Error(err)
		return nil, err
	}

	// Create a cloud connection using the driver and credential info obtained from the Spider server.
	// spiderlet ==> Spider server
	cldConn, err := ccm.CreateCloudConnection(connectionName)
	if err != nil {
		cblog.Error(err)
		return nil, err
	}

	handler, err := cldConn.CreateAnyCallHandler()
	if err != nil {
		cblog.Error(err)
		return nil, err
	}

	//  Call AnyCall
	info, err := handler.AnyCall(reqInfo)
	if err != nil {
		cblog.Error(err)
		return nil, err
	}

	return &info, nil
}

// loading config file with yaml format
func LoadConfigFileYAML(configFilePath string) (string, error) {
	cblog.Info("call LoadConfigFileYAML()")

	// check empty and trim user inputs
	configFilePath, err := EmptyCheckAndTrim("configFilePath", configFilePath)
	if err != nil {
		cblog.Error(err)
		return "", err
	}

	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		cblog.Error(err)
		return "", err
	}

	return string(configFile), nil
}
