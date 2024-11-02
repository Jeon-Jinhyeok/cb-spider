// Proof of Concepts for the Cloud-Barista Multi-Cloud Project.
//      * Cloud-Barista: https://github.com/cloud-barista
//
// NCP VM Spec Handler
//
// by ETRI, 2020.09.

package resources

import (
	"errors"
	"fmt"
	"strconv"
	
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	cblog "github.com/cloud-barista/cb-log"
	call "github.com/cloud-barista/cb-spider/cloud-control-manager/cloud-driver/call-log"
	idrv "github.com/cloud-barista/cb-spider/cloud-control-manager/cloud-driver/interfaces"
	irs "github.com/cloud-barista/cb-spider/cloud-control-manager/cloud-driver/interfaces/resources"
)

type NcpVMSpecHandler struct {
	CredentialInfo 		idrv.CredentialInfo
	RegionInfo     		idrv.RegionInfo
	VMClient         	*server.APIClient
}

func init() {
	// cblog is a global variable.
	cblogger = cblog.GetLogger("NCP VMSpecHandler")
}

func (vmSpecHandler *NcpVMSpecHandler) ListVMSpec() ([]*irs.VMSpecInfo, error) {
	cblogger.Info("NCP Classic Cloud Driver: called ListVMSpec()!")

	InitLog()
	callLogInfo := GetCallLogScheme(vmSpecHandler.RegionInfo.Zone, call.VMIMAGE, "ListVMSpec()", "ListVMSpec()")

	ncpRegion := vmSpecHandler.RegionInfo.Region
	cblogger.Infof("Region : [%s]", ncpRegion)

	vmHandler := NcpVMHandler{
		RegionInfo:     	vmSpecHandler.RegionInfo,
		VMClient:         	vmSpecHandler.VMClient,
	}
	regionNo, err := vmHandler.getRegionNo(vmSpecHandler.RegionInfo.Region)
	if err != nil {
		newErr := fmt.Errorf("Failed to Get the NCP Region No of the Region Code: [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return nil, newErr
	}
	zoneNo, err := vmHandler.getZoneNo(vmSpecHandler.RegionInfo.Region, vmSpecHandler.RegionInfo.Zone)
	if err != nil {
		newErr := fmt.Errorf("Failed to Get NCP Zone No of the Zone Code : [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return nil, newErr
	}

	imageHandler := NcpImageHandler{
		CredentialInfo: 	vmSpecHandler.CredentialInfo,
		RegionInfo:     	vmSpecHandler.RegionInfo,  //CAUTION!!
		VMClient:         	vmSpecHandler.VMClient,
	}
	imageListResult, err := imageHandler.ListImage()
	if err != nil {
		newErr := fmt.Errorf("Failed to Get the NCP Image list!! : [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return nil, newErr
	} else {
		cblogger.Info("Image list Count : ", len(imageListResult))
		// spew.Dump(imageListResult)
	}

	// Note : var vmProductList []*server.Product  //NCP Product(Spec) Info.
	var vmSpecInfoList []*irs.VMSpecInfo //Cloud-Barista Spec Info.

	for _, image := range imageListResult {
		cblogger.Infof("# 기준 NCP Image ID(ImageProductCode) : [%s]", image.IId.SystemId)

		vmSpecReq := server.GetServerProductListRequest{			
			RegionNo: 				regionNo,
			ZoneNo: 				zoneNo,
			ServerImageProductCode: ncloud.String(image.IId.SystemId),  // ***** Caution : ImageProductCode is mandatory. *****			
			// GenerationCode: 		ncloud.String("G2"),  				// # Caution!! : Generations are divided only in the Korean Region.
		}
		result, err := vmSpecHandler.VMClient.V2Api.GetServerProductList(&vmSpecReq)
		if err != nil {
			cblogger.Error(*result.ReturnMessage)
			cblogger.Error(fmt.Sprintf("Failed to Get VMSpec list from NCP : [%v]", err))
			return nil, err
		} else {
			cblogger.Info("Succeeded in Getting VMSpec list!!")
			cblogger.Infof("기준 NCP Image ID(ImageProductCode) : [%s]", image.IId.SystemId)
			cblogger.Infof("에 대해 조회된 VMSpec 정보 수 : [%d]", len(result.ProductList))
		}		

		for _, product := range result.ProductList {
			vmSpecInfo := MappingVMSpecInfo(ncpRegion, image.IId.SystemId, *product)
			vmSpecInfoList = append(vmSpecInfoList, &vmSpecInfo)
		}
	}
	cblogger.Infof("# 총 VMSpec 수 : [%d]", len(vmSpecInfoList))
	return vmSpecInfoList, err
}

func (vmSpecHandler *NcpVMSpecHandler) GetVMSpec(Name string) (irs.VMSpecInfo, error) {
	cblogger.Info("NCP Classic Cloud Driver: called GetVMSpec()!")

	InitLog()
	callLogInfo := GetCallLogScheme(vmSpecHandler.RegionInfo.Zone, call.VMSPEC, Name, "GetVMSpec()")

	ncpRegion := vmSpecHandler.RegionInfo.Region
	cblogger.Infof("Region : [%s] / SpecName : [%s]", ncpRegion, Name)

	vmHandler := NcpVMHandler{
		RegionInfo:     	vmSpecHandler.RegionInfo,
		VMClient:         	vmSpecHandler.VMClient,
	}
	regionNo, err := vmHandler.getRegionNo(vmSpecHandler.RegionInfo.Region)
	if err != nil {
		newErr := fmt.Errorf("Failed to Get the NCP Region No of the Region Code: [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return irs.VMSpecInfo{}, newErr
	}
	zoneNo, err := vmHandler.getZoneNo(vmSpecHandler.RegionInfo.Region, vmSpecHandler.RegionInfo.Zone)
	if err != nil {
		newErr := fmt.Errorf("Failed to Get NCP Zone No of the Zone Code : [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return irs.VMSpecInfo{}, newErr
	}

	imgHandler := NcpImageHandler{
		CredentialInfo: 	vmSpecHandler.CredentialInfo,
		RegionInfo:     	vmSpecHandler.RegionInfo,
		VMClient:         	vmSpecHandler.VMClient,
	}
	cblogger.Infof("imgHandler.RegionInfo.Zone : [%s]", imgHandler.RegionInfo.Zone)  //Need to Check the value!!

	imgListResult, err := imgHandler.ListImage()
	if err != nil {
		cblogger.Infof("Failed to Find Image list!! : ", err)
		return irs.VMSpecInfo{}, errors.New("Failed to Find Image list!!")
	} else {
		cblogger.Info("Succeeded in Getting Image list!!")
		// cblogger.Info(imgListResult)
		cblogger.Infof("Image list Count : [%d]", len(imgListResult))
		// spew.Dump(imgListResult)
	}

	for _, image := range imgListResult {
		cblogger.Infof("# 기준 NCP Image ID(ImageProductCode) : [%s]", image.IId.SystemId)

		specReq := server.GetServerProductListRequest{
			RegionNo:    			regionNo,
			ZoneNo: 				zoneNo,
			ProductCode: 			&Name,			
			ServerImageProductCode: ncloud.String(image.IId.SystemId),  // ***** Caution : ImageProductCode is mandatory. *****
		}
		callLogStart := call.Start()
		result, err := vmSpecHandler.VMClient.V2Api.GetServerProductList(&specReq)
		if err != nil {
			cblogger.Error(*result.ReturnMessage)
			newErr := fmt.Errorf("Failed to Find VMSpec list from NCP : [%v]", err)
			cblogger.Error(newErr.Error())
			LoggingError(callLogInfo, newErr)
			return irs.VMSpecInfo{}, newErr
		}
		LoggingInfo(callLogInfo, callLogStart)
		// spew.Dump(result)

		if len(result.ProductList) > 0 {
			specInfo := MappingVMSpecInfo(ncpRegion, image.IId.SystemId, *result.ProductList[0])
			return specInfo, nil
		}
	}
	return irs.VMSpecInfo{}, errors.New("Not found : VMSpec Name'" + Name + "' Not found!!")
}

func (vmSpecHandler *NcpVMSpecHandler) ListOrgVMSpec() (string, error) {
	cblogger.Info("NCP Classic Cloud Driver: called ListOrgVMSpec()!")

	InitLog()
	callLogInfo := GetCallLogScheme(vmSpecHandler.RegionInfo.Zone, call.VMIMAGE, "ListOrgVMSpec()", "ListOrgVMSpec()")

	ncpRegion := vmSpecHandler.RegionInfo.Region
	cblogger.Infof("Region : [%s]", ncpRegion)

	vmHandler := NcpVMHandler{
		RegionInfo:     	vmSpecHandler.RegionInfo,
		VMClient:         	vmSpecHandler.VMClient,
	}
	regionNo, err := vmHandler.getRegionNo(vmSpecHandler.RegionInfo.Region)
	if err != nil {
		newErr := fmt.Errorf("Failed to Get the NCP Region No of the Region Code: [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return "", newErr
	}
	zoneNo, err := vmHandler.getZoneNo(vmSpecHandler.RegionInfo.Region, vmSpecHandler.RegionInfo.Zone)
	if err != nil {
		newErr := fmt.Errorf("Failed to Get NCP Zone No of the Zone Code : [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return "", newErr
	}

	imageHandler := NcpImageHandler{
		CredentialInfo: 	vmSpecHandler.CredentialInfo,
		RegionInfo:     	vmSpecHandler.RegionInfo,  //CAUTION!!
		VMClient:         	vmSpecHandler.VMClient,
	}
	imageListResult, err := imageHandler.ListImage()
	if err != nil {
		newErr := fmt.Errorf("Failed to Get the NCP Image list!! : [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return "", newErr
	} else {
		cblogger.Info("Image list Count : ", len(imageListResult))
		// spew.Dump(imageListResult)
	}

	var productList []*server.Product  //NCP Product(Spec) Info.
	for _, image := range imageListResult {
		cblogger.Infof("# 기준 NCP Image ID(ImageProductCode) : [%s]", image.IId.SystemId)

		vmSpecReq := server.GetServerProductListRequest{			
			RegionNo: 				regionNo,
			ZoneNo: 				zoneNo,
			ServerImageProductCode: ncloud.String(image.IId.SystemId),  // ***** Caution : ImageProductCode is mandatory. *****			
			// GenerationCode: 		ncloud.String("G2"),  				// # Caution!! : Generations are divided only in the Korean Region.
		}
		result, err := vmSpecHandler.VMClient.V2Api.GetServerProductList(&vmSpecReq)
		if err != nil {
			cblogger.Error(*result.ReturnMessage)
			cblogger.Error(fmt.Sprintf("Failed to Get VMSpec list from NCP : [%v]", err))
			return "", err
		} else {
			cblogger.Info("Succeeded in Getting VMSpec list!!")
			cblogger.Infof("기준 NCP Image ID(ImageProductCode) : [%s]", image.IId.SystemId)
			cblogger.Infof("에 대해 조회된 VMSpec 정보 수 : [%d]", len(result.ProductList))
		}

		for _, product := range result.ProductList {
			productList = append(productList, product)
		}
	}
	cblogger.Infof("# 총 VMSpec 수 : [%d]", len(productList))

	jsonString, jsonErr := ConvertJsonString(productList)
	if jsonErr != nil {
		cblogger.Error(jsonErr)
		return "", jsonErr
	}
	return jsonString, jsonErr
}

func (vmSpecHandler *NcpVMSpecHandler) GetOrgVMSpec(Name string) (string, error) {
	cblogger.Info("NCP Classic Cloud Driver: called GetOrgVMSpec()!")

	InitLog()
	callLogInfo := GetCallLogScheme(vmSpecHandler.RegionInfo.Zone, call.VMSPEC, Name, "GetOrgVMSpec()")

	ncpRegion := vmSpecHandler.RegionInfo.Region
	cblogger.Infof("Region : [%s] / SpecName : [%s]", ncpRegion, Name)

	vmHandler := NcpVMHandler{
		RegionInfo:     	vmSpecHandler.RegionInfo,
		VMClient:         	vmSpecHandler.VMClient,
	}
	regionNo, err := vmHandler.getRegionNo(vmSpecHandler.RegionInfo.Region)
	if err != nil {
		newErr := fmt.Errorf("Failed to Get the NCP Region No of the Region Code: [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return "", newErr
	}
	zoneNo, err := vmHandler.getZoneNo(vmSpecHandler.RegionInfo.Region, vmSpecHandler.RegionInfo.Zone)
	if err != nil {
		newErr := fmt.Errorf("Failed to Get NCP Zone No of the Zone Code : [%v]", err)
		cblogger.Error(newErr.Error())
		LoggingError(callLogInfo, newErr)
		return "", newErr
	}

	imgHandler := NcpImageHandler{
		CredentialInfo: 	vmSpecHandler.CredentialInfo,
		RegionInfo:     	vmSpecHandler.RegionInfo,
		VMClient:         	vmSpecHandler.VMClient,
	}
	cblogger.Infof("imgHandler.RegionInfo.Zone : [%s]", imgHandler.RegionInfo.Zone)  //Need to Check the value!!

	imgListResult, err := imgHandler.ListImage()
	if err != nil {
		cblogger.Infof("Failed to Find Image list!! : ", err)
		return "", errors.New("Failed to Find Image list!!")
	} else {
		cblogger.Info("Succeeded in Getting Image list!!")
		// cblogger.Info(imgListResult)
		cblogger.Infof("Image list Count : [%d]", len(imgListResult))
		// spew.Dump(imgListResult)
	}

	for _, image := range imgListResult {
		cblogger.Infof("# 기준 NCP Image ID(ImageProductCode) : [%s]", image.IId.SystemId)

		specReq := server.GetServerProductListRequest{
			RegionNo:    			regionNo,
			ZoneNo: 				zoneNo,
			ProductCode: 			&Name,			
			ServerImageProductCode: ncloud.String(image.IId.SystemId),  // ***** Caution : ImageProductCode is mandatory. *****
		}
		callLogStart := call.Start()
		result, err := vmSpecHandler.VMClient.V2Api.GetServerProductList(&specReq)
		if err != nil {
			cblogger.Error(*result.ReturnMessage)
			newErr := fmt.Errorf("Failed to Find VMSpec list from NCP : [%v]", err)
			cblogger.Error(newErr.Error())
			LoggingError(callLogInfo, newErr)
			return "", newErr
		}
		LoggingInfo(callLogInfo, callLogStart)
		// spew.Dump(result)

		if len(result.ProductList) > 0 {
			jsonString, jsonErr := ConvertJsonString(*result.ProductList[0])
			if jsonErr != nil {
				cblogger.Error(jsonErr)
				return "", jsonErr
			}
			return jsonString, jsonErr
		}
	}
	
	return "", nil
}

func MappingVMSpecInfo(Region string, ImageId string, NcpVMSpec server.Product) irs.VMSpecInfo {
	//	server ProductList type : []*Product
	cblogger.Infof("*** Mapping VMSpecInfo : Region: [%s] / SpecName: [%s]", Region, *NcpVMSpec.ProductCode)
	// spew.Dump(vmSpec)

	// vmSpec에 리전 정보는 없기 때문에 받은 리전 정보로 기입
	// NOTE 주의 : vmSpec.ProductCode -> specName 으로
	vmSpecInfo := irs.VMSpecInfo{
		Region: Region,
		//Name:   *NcpVMSpec.ProductName,
		Name:   *NcpVMSpec.ProductCode,
		// int32 to string 변환 : String(), int64 to string 변환 : strconv.Itoa()
		VCpu: irs.VCpuInfo{Count: String(*NcpVMSpec.CpuCount), Clock: "N/A"},

		// server.Product에 GPU 정보는 없음.
		Gpu: []irs.GpuInfo{{Count: "N/A", Model: "N/A"}},

		KeyValueList: []irs.KeyValue{
			// {Key: "ProductName", Value: *vmSpec.ProductName}, //This is same to 'ProductDescription'.
			{Key: "ProductType", Value: *NcpVMSpec.ProductType.CodeName},
			{Key: "InfraResourceType", Value: *NcpVMSpec.InfraResourceType.CodeName},
			// {Key: "PlatformType", Value: *NcpVMSpec.PlatformType.CodeName}, //This makes "invalid memory address or nil pointer dereference" error
			{Key: "BaseBlockStorageSize(GB)", Value: strconv.FormatFloat(float64(*NcpVMSpec.BaseBlockStorageSize)/(1024*1024*1024), 'f', 0, 64)},
			{Key: "DiskType", Value: *NcpVMSpec.DiskType.CodeName},
			{Key: "ProductDescription", Value: *NcpVMSpec.ProductDescription},
			{Key: "SupportingImageSystemId", Value: ImageId},
			{Key: "NCPGenerationCode", Value: *NcpVMSpec.GenerationCode},
			{Key: "NCP Region", Value: Region},
		},
	}

	// vmSpecInfo.Mem = strconv.FormatFloat(float64(*vmSpec.MemorySize)*1024, 'f', 0, 64) // GB->MB로 변환
	vmSpecInfo.Mem = strconv.FormatFloat(float64(*NcpVMSpec.MemorySize)/(1024*1024), 'f', 0, 64)

	return vmSpecInfo
}
