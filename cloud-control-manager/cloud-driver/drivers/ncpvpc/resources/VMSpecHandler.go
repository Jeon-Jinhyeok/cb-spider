// Proof of Concepts for the Cloud-Barista Multi-Cloud Project.
//      * Cloud-Barista: https://github.com/cloud-barista
//
// NCP VPC VMSpec Handler
//
// by ETRI, 2020.12.
// Updated by ETRI, 2025.02.

package resources

import (
	"errors"
	"fmt"
	"strings"
	"strconv"
	"regexp"
	// "github.com/davecgh/go-spew/spew"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	cblog "github.com/cloud-barista/cb-log"
	call "github.com/cloud-barista/cb-spider/cloud-control-manager/cloud-driver/call-log"
	idrv "github.com/cloud-barista/cb-spider/cloud-control-manager/cloud-driver/interfaces"
	irs "github.com/cloud-barista/cb-spider/cloud-control-manager/cloud-driver/interfaces/resources"
)

type NcpVpcVMSpecHandler struct {
	CredentialInfo idrv.CredentialInfo
	RegionInfo     idrv.RegionInfo
	VMClient       *vserver.APIClient
}

func init() {
	// cblog is a global variable.
	cblogger = cblog.GetLogger("NCP VPC VMSpecHandler")
}

func (vmSpecHandler *NcpVpcVMSpecHandler) ListVMSpec() ([]*irs.VMSpecInfo, error) {
	cblogger.Info("NCP VPC Cloud driver: called ListVMSpec()!")
	InitLog()
	callLogInfo := GetCallLogScheme(vmSpecHandler.RegionInfo.Zone, call.VMSPEC, "ListVMSpec()", "ListVMSpec()")

	imageHandler := NcpVpcImageHandler{
		RegionInfo:     vmSpecHandler.RegionInfo,  //CAUTION!!
		VMClient:       vmSpecHandler.VMClient,
	}
	imgListResult, err := imageHandler.ListImage()
	if err != nil {
		newErr := fmt.Errorf("Failed to Get the NCP Image list!! : [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return nil, newErr
	}

	vmSpecMap := make(map[string]*irs.VMSpecInfo)
	for _, image := range imgListResult {
		cblogger.Infof("\n### Lookup by NCP VPC Image ID(ImageProductCode) : [%s]", image.IId.SystemId)
		vmSpecReq := vserver.GetServerSpecListRequest{
			RegionCode:		&vmSpecHandler.RegionInfo.Region,
			ZoneCode:		&vmSpecHandler.RegionInfo.Zone,
			ServerImageNo: 	ncloud.String(image.IId.SystemId), // ***** Caution : Mandatory. *****
		}
		callLogStart := call.Start()
		result, err := vmSpecHandler.VMClient.V2Api.GetServerSpecList(&vmSpecReq)
		if err != nil { 
			rtnErr := logAndReturnError(callLogInfo, "Failed to Get VMSpec list from NCP VPC : ", err)
			return nil, rtnErr
		}
		LoggingInfo(callLogInfo, callLogStart)

		if len(result.ServerSpecList) < 1 {
			rtnErr := logAndReturnError(callLogInfo, "# VMSpec info corresponding to the Image ID does Not Exist!!", "")
			return nil, rtnErr
		} else {
			for _, vmSpec := range result.ServerSpecList {
				vmSpecInfo := vmSpecHandler.mappingVMSpecInfo(image.IId.SystemId, vmSpec)
				if existingSpec, exists := vmSpecMap[vmSpecInfo.Name]; exists {
					// If the VMSpec already exists, add the image ID to the corresponding images list in KeyValueList
					for i, kv := range existingSpec.KeyValueList {
						if kv.Key == "CorrespondingImageIds" {
							existingSpec.KeyValueList[i].Value += "," + image.IId.SystemId
							break
						}
					}
				} else {
					vmSpecInfo.KeyValueList = append(vmSpecInfo.KeyValueList, irs.KeyValue{Key: "CorrespondingImageIds", Value: image.IId.SystemId})
					vmSpecMap[vmSpecInfo.Name] = &vmSpecInfo
				}
			}
		}
	}

	var vmSpecInfoList []*irs.VMSpecInfo
	for _, specInfo := range vmSpecMap {
		vmSpecInfoList = append(vmSpecInfoList, specInfo)
	}
	return vmSpecInfoList, nil
}

func (vmSpecHandler *NcpVpcVMSpecHandler) GetVMSpec(specName string) (irs.VMSpecInfo, error) {
	cblogger.Info("NCP VPC Cloud driver: called GetVMSpec()!")
	InitLog()
	callLogInfo := GetCallLogScheme(vmSpecHandler.RegionInfo.Zone, call.VMSPEC, specName, "GetVMSpec()")

	if strings.EqualFold(specName, "") {
		newErr := fmt.Errorf("Invalid specName!!")
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return irs.VMSpecInfo{}, newErr
	}

	// Note!!) Use ListVMSpec() to include 'CorrespondingImageIds' parameter.
	specListResult, err := vmSpecHandler.ListVMSpec()
	if err != nil {
		newErr := fmt.Errorf("Failed to Get the VMSpec info list!! : [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return irs.VMSpecInfo{}, newErr
	}

	for _, spec := range specListResult {
		if strings.EqualFold(spec.Name, specName) {
			return *spec, nil
		}
	}
	return irs.VMSpecInfo{}, errors.New("Failed to find the VMSpec info : '" + specName)
}

func (vmSpecHandler *NcpVpcVMSpecHandler) ListOrgVMSpec() (string, error) {
	cblogger.Info("NCP VPC Cloud driver: called ListOrgVMSpec()!")
	InitLog()
	callLogInfo := GetCallLogScheme(vmSpecHandler.RegionInfo.Zone, call.VMSPEC, "ListOrgVMSpec()", "ListOrgVMSpec()")

	vmSpecList, err := vmSpecHandler.getNcpVpcVMSpecList()
	if err != nil {
		newErr := fmt.Errorf("Failed to Get the VMSpec info list!! : [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return "", newErr
	}

	jsonString, cvtErr := ConvertJsonString(vmSpecList)
	if cvtErr != nil {
		newErr := fmt.Errorf("Failed to Convert JSON to String : [%v]", cvtErr)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return "", newErr
	}
	return jsonString, nil
}

func (vmSpecHandler *NcpVpcVMSpecHandler) GetOrgVMSpec(specName string) (string, error) {
	cblogger.Info("NCP VPC Cloud driver: called GetOrgVMSpec()!")
	InitLog()
	callLogInfo := GetCallLogScheme(vmSpecHandler.RegionInfo.Zone, call.VMSPEC, specName, "GetOrgVMSpec()")

	if strings.EqualFold(specName, "") {
		newErr := fmt.Errorf("Invalid specName!!")
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return "", newErr
	}

	vmSpec, err := vmSpecHandler.getNcpVpcVMSpec(specName)
	if err != nil {
		newErr := fmt.Errorf("Failed to Get VMSpec from NCP VPC : ", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return "", newErr
	}

	jsonString, cvtErr := ConvertJsonString(vmSpec)
	if cvtErr != nil {
		rtnErr := logAndReturnError(callLogInfo, "Failed to Convert JSON to String : ", cvtErr)
		return "", rtnErr
	}
	return jsonString, nil
}

// func (vmSpecHandler *NcpVpcVMSpecHandler) mappingVMSpecInfo(ImageId string, NcpVMSpec *vserver.Product) irs.VMSpecInfo {
func (vmSpecHandler *NcpVpcVMSpecHandler) mappingVMSpecInfo(ImageId string, vmSpec *vserver.ServerSpec) irs.VMSpecInfo {
	cblogger.Info("NCP VPC Cloud driver: called mappingVMSpecInfo()!")
	// spew.Dump(vmSpec)

	if strings.EqualFold(ImageId, "") {
		newErr := fmt.Errorf("Invalid ImageId!!")
		cblogger.Error(newErr.Error())
		return irs.VMSpecInfo{}
	}

	var memSize string
	if vmSpec.MemorySize != nil {
		memSize = strconv.FormatFloat(float64(*vmSpec.MemorySize), 'f', 2, 64) 
	}

	// Define a regex to match the number before "GB" in "Disk <number>GB"
	// *vmSpec.ServerSpecDescription : Ex) "Tesla T4 GPU 2EA, GPUMemory 32GB, vCPU 32EA, Memory 160GB, Disk 50GB"
	re := regexp.MustCompile(`Disk (\d+)GB`)
	matches := re.FindStringSubmatch(*vmSpec.ServerSpecDescription) // Find the match
	
	var diskSize string
	if len(matches) > 1 {
		diskSize = matches[1] // Extract only the numeric part
	}
	if strings.EqualFold(diskSize, "") {
		diskSize = "-1"
	}

	var gpuCount string
	if vmSpec.GpuCount != nil {
		gpuCount = strconv.Itoa(int(*vmSpec.GpuCount))
	}
	if strings.EqualFold(gpuCount, "") {
		gpuCount = "-1"
	}

	vmSpecInfo := irs.VMSpecInfo{
		Region: vmSpecHandler.RegionInfo.Region,
		Name:   *vmSpec.ServerSpecCode,
		
		// int32 to string 변환 : String(), int64 to string 변환 : strconv.Itoa()
		VCpu: irs.VCpuInfo{Count: String(*vmSpec.CpuCount), Clock: "-1"},
		Mem: memSize,
		Disk: diskSize,
		Gpu: []irs.GpuInfo{{Count: gpuCount, Mfr: "NA", Model: "NA", Mem: "-1"}},

		KeyValueList: []irs.KeyValue{
			{Key: "SpecDescription", Value: *vmSpec.ServerSpecDescription},
			{Key: "Generation", Value: *vmSpec.GenerationCode},
			{Key: "HypervisorType", Value: *vmSpec.HypervisorType.CodeName},
			{Key: "CpuArchitectureType", Value: *vmSpec.CpuArchitectureType.CodeName},
			{Key: "BlockStorageMaxCount", Value: String(*vmSpec.BlockStorageMaxCount)},
		},
	}
	// vmSpecInfo.Mem = strconv.FormatFloat(float64(*vmSpec.MemorySize)*1024, 'f', 0, 64) // GB->MB로 변환
	vmSpecInfo.Mem = strconv.FormatFloat(float64(*vmSpec.MemorySize)/(1024*1024), 'f', 0, 64)
	return vmSpecInfo
}

func (vmSpecHandler *NcpVpcVMSpecHandler) getNcpVpcVMSpecList() ([]*vserver.ServerSpec, error) {
	cblogger.Info("NCP VPC Cloud driver: called getNcpVpcVMSpecList()!")
	InitLog()
	callLogInfo := GetCallLogScheme(vmSpecHandler.RegionInfo.Zone, call.VMSPEC, "getNcpVpcVMSpecList()", "getNcpVpcVMSpecList()")

	imgHandler := NcpVpcImageHandler{
		RegionInfo:     vmSpecHandler.RegionInfo,
		VMClient: 		vmSpecHandler.VMClient,
	}
	imgListResult, err := imgHandler.ListImage()
	if err != nil {
		rtnErr := logAndReturnError(callLogInfo, "Failed to Get Image Info list :  : ", err)
		return nil, rtnErr
	} else {
		cblogger.Infof("Image list Count of the Region : [%d]", len(imgListResult))
	}

	var vmSpecList []*vserver.ServerSpec
	for _, image := range imgListResult {
		// cblogger.Infof("\n### Lookup by NCP VPC Image ID(ImageProductCode) : [%s]", image.IId.SystemId)

		specReq := vserver.GetServerSpecListRequest{
			RegionCode:		&vmSpecHandler.RegionInfo.Region,  //CAUTION!!
			ZoneCode:		&vmSpecHandler.RegionInfo.Zone,
			ServerImageNo: 	ncloud.String(image.IId.SystemId), // ***** Caution : ServerImageNo is mandatory. *****
		}
		callLogStart := call.Start()
		result, err := vmSpecHandler.VMClient.V2Api.GetServerSpecList(&specReq)
		if err != nil { 
			rtnErr := logAndReturnError(callLogInfo, "Failed to Get VMSpec list from NCP VPC : ", err)
			return nil, rtnErr
		}
		LoggingInfo(callLogInfo, callLogStart)

		// spew.Dump(result)
		if len(result.ServerSpecList) < 1 {
			rtnErr := logAndReturnError(callLogInfo, "# VMSpec info corresponding to the Image ID does Not Exist!!", "")
			return nil, rtnErr
		} else {
			vmSpecList = append(vmSpecList, result.ServerSpecList...)
		}
	}
	return vmSpecList, nil
}

func (vmSpecHandler *NcpVpcVMSpecHandler) getNcpVpcVMSpec(specName string) (*vserver.ServerSpec, error) {
	cblogger.Info("NCP VPC Cloud driver: called getNcpVpcVMSpec()!")
	InitLog()
	callLogInfo := GetCallLogScheme(vmSpecHandler.RegionInfo.Zone, call.VMSPEC, specName, "getNcpVpcVMSpec()")

	if strings.EqualFold(specName, "") {
		newErr := fmt.Errorf("Invalid specName!!")
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return nil, newErr
	}

	specReq := vserver.GetServerSpecListRequest{
		RegionCode:  			&vmSpecHandler.RegionInfo.Region,
		ZoneCode:  				&vmSpecHandler.RegionInfo.Zone,
		ServerSpecCodeList: 	[]*string{ncloud.String(specName),},
	}	
	callLogStart := call.Start()
	result, err := vmSpecHandler.VMClient.V2Api.GetServerSpecList(&specReq)
	if err != nil {
		if err != nil {
			rtnErr := logAndReturnError(callLogInfo, "Failed to Get VMSpec list from NCP VPC : ", err)
			return nil, rtnErr
		}
	}
	LoggingInfo(callLogInfo, callLogStart)

	// spew.Dump(result)
	if len(result.ServerSpecList) < 1 {
		rtnErr := logAndReturnError(callLogInfo, "The VMSpec Name does Not Exist!!", "")
		return nil, rtnErr
	} else {
		return result.ServerSpecList[0], nil
	}
}

func (vmSpecHandler *NcpVpcVMSpecHandler) getNcpVpcServerProductCode(specName string) (string, error) {
	cblogger.Info("NCP VPC Cloud driver: called getNcpVpcServerProductCode()!")
	InitLog()
	callLogInfo := GetCallLogScheme(vmSpecHandler.RegionInfo.Zone, call.VMSPEC, specName, "getNcpVpcServerProductCode()")

	if strings.EqualFold(specName, "") {
		newErr := fmt.Errorf("Invalid specName!!")
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return "", newErr
	}

	vmSpec, err := vmSpecHandler.getNcpVpcVMSpec(specName)
	if err != nil {
		newErr := fmt.Errorf("Failed to Get VMSpec from NCP VPC : ", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return "", newErr
	}

	// spew.Dump(result)
	if len(*vmSpec.ServerProductCode) < 1 {
		newErr := fmt.Errorf("Failed to Get ServerProductCode from NCP VPC : ", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return "", newErr
	} else {
		return *vmSpec.ServerProductCode, nil
	}
}
