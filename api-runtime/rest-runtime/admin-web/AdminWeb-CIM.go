// Cloud Info Manager's Rest Runtime of CB-Spider.
// The CB-Spider is a sub-Framework of the Cloud-Barista Multi-Cloud Project.
// The CB-Spider Mission is to connect all the clouds with a single interface.
//
//      * Cloud-Barista: https://github.com/cloud-barista
//
// by CB-Spider Team, 2020.06.

package adminweb

import (
	"fmt"
	"strconv"

	cr "github.com/cloud-barista/cb-spider/api-runtime/common-runtime"
	ccim "github.com/cloud-barista/cb-spider/cloud-info-manager/connection-config-info-manager"
	cim "github.com/cloud-barista/cb-spider/cloud-info-manager/credential-info-manager"
	dim "github.com/cloud-barista/cb-spider/cloud-info-manager/driver-info-manager"
	rim "github.com/cloud-barista/cb-spider/cloud-info-manager/region-info-manager"

	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// number, Provider Name, Driver File, Driver Name, checkbox
func makeDriverTRList_html(bgcolor string, height string, fontSize string, infoList []*dim.CloudDriverInfo) string {
	if bgcolor == "" {
		bgcolor = "#FFFFFF"
	}
	if height == "" {
		height = "30"
	}
	if fontSize == "" {
		fontSize = "2"
	}

	// make base TR frame for info list
	strTR := fmt.Sprintf(`
                <tr bgcolor="%s" align="center" height="%s">
                    <td>
                            <font size=%s>$$NUM$$</font>
                    </td>
                    <td>
                            <font size=%s>$$S1$$</font>
                    </td>
                    <td>
                            <font size=%s>$$S2$$</font>
                    </td>
                    <td>
                            <font size=%s>$$S3$$</font>
                    </td>
                    <td>
                        <input type="checkbox" name="check_box" value=$$S3$$>
                    </td>
                </tr>
       		`, bgcolor, height, fontSize, fontSize, fontSize, fontSize)

	strData := ""
	// set data and make TR list
	for i, one := range infoList {
		str := strings.ReplaceAll(strTR, "$$NUM$$", strconv.Itoa(i+1))
		str = strings.ReplaceAll(str, "$$S1$$", one.ProviderName)
		str = strings.ReplaceAll(str, "$$S2$$", one.DriverLibFileName)
		str = strings.ReplaceAll(str, "$$S3$$", one.DriverName)
		strData += str
	}

	return strData
}

// make the string of javascript function
func makeOnchangeDriverProviderFunc_js() string {
	strFunc := `
              function onchangeProvider(source) {
                var providerName = source.value
                document.getElementById('2').value= providerName.toLowerCase() + "-driver-v1.0.so";
                document.getElementById('3').value= providerName.toLowerCase() + "-driver-01";
              }
        `

	return strFunc
}

// make the string of javascript function
func makeCheckBoxToggleFunc_js() string {

	strFunc := `
              function toggle(source) {
                var checkboxes = document.getElementsByName('check_box');
                for (var i = 0; i < checkboxes.length; i++) {
                  checkboxes[i].checked = source.checked;
                }
              }
        `

	return strFunc
}

// make the string of javascript function
func makePostDriverFunc_js() string {

	// curl -X POST http://$RESTSERVER:1024/spider/driver -H 'Content-Type: application/json'  -d '{"DriverName":"aws-driver01","ProviderName":"AWS", "DriverLibFileName":"aws-driver-v1.0.so"}'

	strFunc := `
                function postDriver() {
                        var textboxes = document.getElementsByName('text_box');
			sendJson = '{ "ProviderName" : "$$PROVIDER$$", "DriverLibFileName" : "$$$DRVFILE$$", "DriverName" : "$$NAME$$" }'
                        for (var i = 0; i < textboxes.length; i++) { // @todo make parallel executions
                                switch (textboxes[i].id) {
                                        case "1":
                                                sendJson = sendJson.replace("$$PROVIDER$$", textboxes[i].value);
                                                break;
                                        case "2":
                                                sendJson = sendJson.replace("$$$DRVFILE$$", textboxes[i].value);
                                                break;
                                        case "3":
                                                sendJson = sendJson.replace("$$NAME$$", textboxes[i].value);
                                                break;
                                        default:
                                                break;
                                }
                        }
                        var xhr = new XMLHttpRequest();
                        xhr.open("POST", "$$SPIDER_SERVER$$/spider/driver", false);
                        xhr.setRequestHeader('Content-Type', 'application/json');
                        //xhr.send(JSON.stringify({ "DriverName": driverName, "ProviderName": providerName, "DriverLibFileName": driverLibFileName}));
			//xhr.send(JSON.stringify(sendJson));

			// client logging
			parent.frames["log_frame"].Log("curl -sX POST " + "$$SPIDER_SERVER$$/spider/driver -H 'Content-Type: application/json' -d '" + sendJson + "'");

			xhr.send(sendJson);

			// client logging
			parent.frames["log_frame"].Log("   => " + xhr.response);

                        //setTimeout(function(){ // when async call
                                location.reload();
                        //}, 400);

                }
        `
	strFunc = strings.ReplaceAll(strFunc, "$$SPIDER_SERVER$$", "http://"+cr.ServiceIPorName+cr.ServicePort) // cr.ServicePort = ":1024"
	return strFunc
}

// make the string of javascript function
func makeDeleteDriverFunc_js() string {
	// curl -X DELETE http://$RESTSERVER:1024/spider/driver/gcp-driver01 -H 'Content-Type: application/json'

	strFunc := `
                function deleteDriver() {
                        var checkboxes = document.getElementsByName('check_box');
                        for (var i = 0; i < checkboxes.length; i++) { // @todo make parallel executions
                                if (checkboxes[i].checked) {
                                        var xhr = new XMLHttpRequest();
                                        xhr.open("DELETE", "$$SPIDER_SERVER$$/spider/driver/" + checkboxes[i].value, false);
                                        xhr.setRequestHeader('Content-Type', 'application/json');

                                        // client logging
                                        parent.frames["log_frame"].Log("curl -sX DELETE " + "$$SPIDER_SERVER$$/spider/driver/" + checkboxes[i].value + " -H 'Content-Type: application/json'" );

                                        xhr.send(null);

                                        // client logging
                                        parent.frames["log_frame"].Log("   => " + xhr.response);
                                }
                        }
			location.reload();
                }
        `
	strFunc = strings.ReplaceAll(strFunc, "$$SPIDER_SERVER$$", "http://"+cr.ServiceIPorName+cr.ServicePort) // cr.ServicePort = ":1024"
	return strFunc
}

// ================ Driver Info Management
// create driver page
func Driver(c echo.Context) error {
	cblog.Info("call Driver()")

	// make page header
	htmlStr := ` 
		<html>
		<head>
		    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
		    <script type="text/javascript">
		`
	// (1) make Javascript Function
	htmlStr += makeOnchangeDriverProviderFunc_js()
	htmlStr += makeCheckBoxToggleFunc_js()
	htmlStr += makePostDriverFunc_js()
	htmlStr += makeDeleteDriverFunc_js()

	htmlStr += `
		    </script>
		</head>

		<body>
		    <table border="0" bordercolordark="#F8F8FF" cellpadding="0" cellspacing="1" bgcolor="#FFFFFF"  style="font-size:small;">      
		`

	// (2) make Table Action TR
	// colspan, f5_href, delete_href, fontSize
	htmlStr += makeActionTR_html("5", "driver", "deleteDriver()", "2")

	// (3) make Table Header TR

	nameWidthList := []NameWidth{
		{"Provider Name", "200"},
		{"Driver Library Name", "300"},
		{"Driver Name", "200"},
	}
	htmlStr += makeTitleTRList_html("#DDDDDD", "2", nameWidthList, true)

	// (4) make TR list with info list
	// (4-1) get info list @todo if empty list

	// client logging
	htmlStr += genLoggingGETResURL("driver")

	resBody, err := getResourceList_JsonByte("driver")
	if err != nil {
		cblog.Error(err)
		// client logging
		htmlStr += genLoggingGETResURL(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// client logging
	htmlStr += genLoggingResult(string(resBody[:len(resBody)-1]))

	var info struct {
		ResultList []*dim.CloudDriverInfo `json:"driver"`
	}
	json.Unmarshal(resBody, &info)

	// (4-2) make TR list with info list
	htmlStr += makeDriverTRList_html("", "", "", info.ResultList)

	// (5) make input field and add
	// attach text box for add
	nameList := cloudosList()
	htmlStr += `
			<tr bgcolor="#FFFFFF" align="center" height="30">
			    <td bgcolor="#FFEFBA">
                                    <font size=2>&nbsp;create:&nbsp;</font>
			    </td>
			    <td>
				<!-- <input style="font-size:12px;text-align:center;" type="text" name="text_box" id="1" value="AWS"> -->
		`
	// Select format of CloudOS  name=text_box, id=1
	htmlStr += makeSelect_html("onchangeProvider", nameList, "1")

	htmlStr += `
			    </td>
			    <td>
				<input style="font-size:12px;text-align:center;" type="text" name="text_box" id="2" value="aws-driver-v1.0.so">
			    </td>
			    <td>
				<input style="font-size:12px;text-align:center;" type="text" name="text_box" id="3" value="aws-driver-01">
			    </td>
			    <td>
				<a href="javascript:postDriver()">
				    <font size=3><b>+</b></font>
				</a>
			    </td>
			</tr>
		`
	// make page tail
	htmlStr += `
                    </table>
		    <hr>
                </body>
                </html>
        `

	//fmt.Println(htmlStr)
	return c.HTML(http.StatusOK, htmlStr)
}

func cloudosList() []string {
	resBody, err := getResourceList_JsonByte("cloudos")
	if err != nil {
		cblog.Error(err)
	}
	var info struct {
		ResultList []string `json:"cloudos"`
	}
	json.Unmarshal(resBody, &info)

	return info.ResultList
}

func genLoggingGETResURL(rsType string) string {
	/* return example
	   <script type="text/javascript">
	           parent.frames["log_frame"].Log("curl -sX GET http://localhost:1024/spider/driver -H 'Content-Type: application/json' ");
	   </script>
	*/

	url := "http://" + "localhost" + cr.ServerPort + "/spider/" + rsType + " -H 'Content-Type: application/json' "
	htmlStr := `
                <script type="text/javascript">
                `
	htmlStr += `    parent.frames["log_frame"].Log("curl -sX GET ` + url + `");`
	htmlStr += `
                </script>
                `
	return htmlStr
}

// make the string of javascript function
func makeOnchangeCredentialProviderFunc_js() string {
	strFunc := `
              function onchangeProvider(source) {
                var providerName = source.value
		// for credential info
		switch(providerName) {
		  case "AWS":
			credentialInfo = '[{"Key":"ClientId", "Value":"XXXXXX"}, {"Key":"ClientSecret", "Value":"XXXXXX"}]'
		    break;
		  case "AZURE":
			credentialInfo = '[{"Key":"ClientId", "Value":"XXXX-XXXX"}, {"Key":"ClientSecret", "Value":"xxxx-xxxx"}, {"Key":"TenantId", "Value":"xxxx-xxxx"}, {"Key":"SubscriptionId", "Value":"xxxx-xxxx"}]'
		    break;
		  case "GCP":
			credentialInfo = '[{"Key":"PrivateKey", "Value":"-----BEGIN PRIVATE KEY-----\nXXXX\n-----END PRIVATE KEY-----\n"},{"Key":"ProjectID", "Value":"powerkimhub"}, {"Key":"ClientEmail", "Value":"xxxx@xxxx.iam.gserviceaccount.com"}]'
		    break;
		  case "ALIBABA":
			credentialInfo = '[{"Key":"ClientId", "Value":"XXXXXX"}, {"Key":"ClientSecret", "Value":"XXXXXX"}]'
		    break;
		  case "TENCENT":
			credentialInfo = '[{"Key":"ClientId", "Value":"XXXXXX"}, {"Key":"ClientSecret", "Value":"XXXXXX"}]'
		    break;
		  case "IBM":
			credentialInfo = '[{"Key":"ApiKey", "Value":"XXXXXX"}]'
		    break;
		  case "OPENSTACK":
			credentialInfo = '[{"Key":"IdentityEndpoint", "Value":"http://123.456.789.123:5000/v3"}, {"Key":"Username", "Value":"etri"}, {"Key":"Password", "Value":"xxxx"}, {"Key":"DomainName", "Value":"default"}, {"Key":"ProjectID", "Value":"xxxx"}]'
		    break;

		  case "NCPVPC":
			credentialInfo = '[{"Key":"ClientId", "Value":"XXXXXXXXXXXXXXXXXXX"}, {"Key":"ClientSecret", "Value":"XXXXXXXXXXXXXXXXXXXXXXXXXXX"}]'
		    break;
		  case "NCP":
			credentialInfo = '[{"Key":"ClientId", "Value":"XXXXXXXXXXXXXXXXXXX"}, {"Key":"ClientSecret", "Value":"XXXXXXXXXXXXXXXXXXXXXXXXXXX"}]'
		    break;
		  case "NHNCLOUD":
			credentialInfo = '[{"Key":"IdentityEndpoint", "Value":"https://api-identity-infrastructure.nhncloudservice.com"}, {"Key":"Username", "Value":"XXXXX@XXXXXXXXXXXXXXXX"}, {"Key":"Password", "Value":"XXXXXXXXXXXXXXXXXX"}, {"Key":"DomainName", "Value":"default"}, {"Key":"TenantId", "Value":"XXXXXXXXXXXXXXXXX"}]'
		    break;
		case "KTCLOUD":
			credentialInfo = '[{"Key":"ClientId", "Value":"XXXXXXXXXXXXXXXXXXX"}, {"Key":"ClientSecret", "Value":"XXXXXXXXXXXXXXXXXXXXXXXXXXX"}]'
		    break;
		case "KTCLOUDVPC":
			credentialInfo = ' [{"Key":"IdentityEndpoint", "Value":"https://api.ucloudbiz.olleh.com/d1/identity/v3/"}, {"Key":"Username", "Value":"~~~@~~~.com"}, {"Key":"Password", "Value":"XXXXXXXXXX"}, {"Key":"DomainName", "Value":"default"}, {"Key":"ProjectID", "Value":"XXXXXXXXXX"}]'
		    break;

		  case "MOCK":
			credentialInfo = '[{"Key":"MockName", "Value":"mock_name00"}]'
		    break;
		  case "CLOUDTWIN":
			credentialInfo = '[{"Key":"IdentityEndpoint", "Value":"http://123.456.789.123:8192"}, {"Key":"DomainName", "Value":"cloud-1"}, {"Key":"MockName", "Value":"mock_name01"}]'
		    break;
		  default:
			credentialInfo = '[{"Key":"ClientId", "Value":"XXXXXX"}, {"Key":"ClientSecret", "Value":"XXXXXX"}]'
		}
                document.getElementById('2').value= credentialInfo

		// for credential name
                document.getElementById('3').value= providerName.toLowerCase() + "-credential-01";
              }
        `
	return strFunc
}

// number, Provider Name, Credential Info, Credential Name, checkbox
func makeCredentialTRList_html(bgcolor string, height string, fontSize string, infoList []*cim.CredentialInfo) string {
	if bgcolor == "" {
		bgcolor = "#FFFFFF"
	}
	if height == "" {
		height = "30"
	}
	if fontSize == "" {
		fontSize = "2"
	}

	// make base TR frame for info list
	strTR := fmt.Sprintf(`
                <tr bgcolor="%s" align="center" height="%s">
                    <td>
                            <font size=%s>$$NUM$$</font>
                    </td>
                    <td>
                            <font size=%s>$$S1$$</font>
                    </td>
                    <td>
                            <font size=%s>$$S2$$</font>
                    </td>
                    <td>
                            <font size=%s>$$S3$$</font>
                    </td>
                    <td>
                        <input type="checkbox" name="check_box" value=$$S3$$>
                    </td>
                </tr>
                `, bgcolor, height, fontSize, fontSize, fontSize, fontSize)

	strData := ""
	// set data and make TR list
	for i, one := range infoList {
		str := strings.ReplaceAll(strTR, "$$NUM$$", strconv.Itoa(i+1))
		str = strings.ReplaceAll(str, "$$S1$$", one.ProviderName)
		strKeyList := ""
		for _, kv := range one.KeyValueInfoList {
			strKeyList += kv.Key + ":" + kv.Value + ", "
		}
		strKeyList = strings.TrimSuffix(strKeyList, ", ")
		str = strings.ReplaceAll(str, "$$S2$$", strKeyList)
		str = strings.ReplaceAll(str, "$$S3$$", one.CredentialName)
		strData += str
	}

	return strData
}

// make the string of javascript function
func makePostCredentialFunc_js() string {

	// curl -X POST http://$RESTSERVER:1024/spider/credential -H 'Content-Type: application/json' '{"CredentialName":"aws-credential-01","ProviderName":"AWS", "KeyValueInfoList": [{"Key":"ClientId", "Value":"XXXXXX"}, {"Key":"ClientSecret", "Value":"XXXXXX"}]}'

	strFunc := `
                function postCredential() {
                        var textboxes = document.getElementsByName('text_box');
			sendJson = '{ "ProviderName" : "$$PROVIDER$$", "KeyValueInfoList" : $$CREDENTIALINFO$$, "CredentialName" : "$$NAME$$" }'

                        for (var i = 0; i < textboxes.length; i++) { // @todo make parallel executions
                                switch (textboxes[i].id) {
                                        case "1":
                                                sendJson = sendJson.replace("$$PROVIDER$$", textboxes[i].value);
                                                break;
                                        case "2":
                                                sendJson = sendJson.replace("$$CREDENTIALINFO$$", textboxes[i].value);
                                                break;
                                        case "3":
                                                sendJson = sendJson.replace("$$NAME$$", textboxes[i].value);
                                                break;
                                        default:
                                                break;
                                }
                        }
                        var xhr = new XMLHttpRequest();
                        xhr.open("POST", "$$SPIDER_SERVER$$/spider/credential", false);
                        xhr.setRequestHeader('Content-Type', 'application/json');
                        //xhr.send(JSON.stringify({ "CredentialName": credentialName, "ProviderName": providerName, "KeyValueInfoList": credentialInfo}));
                        //xhr.send(JSON.stringify(sendJson));

			// client logging
			parent.frames["log_frame"].Log("curl -sX POST " + "$$SPIDER_SERVER$$/spider/credential -H 'Content-Type: application/json' -d '" + sendJson + "'");

                        xhr.send(sendJson);

			// client logging
			parent.frames["log_frame"].Log("   => " + xhr.response);

                        // setTimeout(function(){ // when async call
                                location.reload();
                        // }, 400);

                }
        `
	strFunc = strings.ReplaceAll(strFunc, "$$SPIDER_SERVER$$", "http://"+cr.ServiceIPorName+cr.ServicePort) // cr.ServicePort = ":1024"
	return strFunc
}

// make the string of javascript function
func makeDeleteCredentialFunc_js() string {
	// curl -X DELETE http://$RESTSERVER:1024/spider/credential/aws-credential-01 -H 'Content-Type: application/json'

	strFunc := `
                function deleteCredential() {
                        var checkboxes = document.getElementsByName('check_box');
                        for (var i = 0; i < checkboxes.length; i++) { // @todo make parallel executions
                                if (checkboxes[i].checked) {
                                        var xhr = new XMLHttpRequest();
                                        xhr.open("DELETE", "$$SPIDER_SERVER$$/spider/credential/" + checkboxes[i].value, false);
                                        xhr.setRequestHeader('Content-Type', 'application/json');

                                        // client logging
                                        parent.frames["log_frame"].Log("curl -sX DELETE " + "$$SPIDER_SERVER$$/spider/credential/" + checkboxes[i].value + " -H 'Content-Type: application/json'" );

                                        xhr.send(null);

                                        // client logging
                                        parent.frames["log_frame"].Log("   => " + xhr.response);
                                }
                        }
			location.reload();
                }
        `
	strFunc = strings.ReplaceAll(strFunc, "$$SPIDER_SERVER$$", "http://"+cr.ServiceIPorName+cr.ServicePort) // cr.ServicePort = ":1024"
	return strFunc
}

// ================ Credential Info Management
// create credential page
func Credential(c echo.Context) error {
	cblog.Info("call Credential()")

	// make page header
	htmlStr := `
                <html>
                <head>
                    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
                    <script type="text/javascript">
                `
	// (1) make Javascript Function
	htmlStr += makeOnchangeCredentialProviderFunc_js()
	htmlStr += makeCheckBoxToggleFunc_js()
	htmlStr += makePostCredentialFunc_js()
	htmlStr += makeDeleteCredentialFunc_js()

	htmlStr += `
                    </script>
                </head>

                <body>
                    <table border="0" bordercolordark="#F8F8FF" cellpadding="0" cellspacing="1" bgcolor="#FFFFFF"  style="font-size:small;">
                `

	// (2) make Table Action TR
	// colspan, f5_href, delete_href, fontSize
	htmlStr += makeActionTR_html("5", "credential", "deleteCredential()", "2")

	// (3) make Table Header TR
	nameWidthList := []NameWidth{
		{"Provider Name", "200"},
		{"Credential Info", "300"},
		{"Credential Name", "200"},
	}
	htmlStr += makeTitleTRList_html("#DDDDDD", "2", nameWidthList, true)

	// (4) make TR list with info list
	// (4-1) get info list @todo if empty list

	// client logging
	htmlStr += genLoggingGETResURL("credential")

	resBody, err := getResourceList_JsonByte("credential")
	if err != nil {
		cblog.Error(err)
		// client logging
		htmlStr += genLoggingGETResURL(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// client logging
	htmlStr += genLoggingResult(string(resBody[:len(resBody)-1]))

	var info struct {
		ResultList []*cim.CredentialInfo `json:"credential"`
	}
	json.Unmarshal(resBody, &info)

	// (4-2) make TR list with info list
	htmlStr += makeCredentialTRList_html("", "", "", info.ResultList)

	// (5) make input field and add
	// attach text box for add
	nameList := cloudosList()
	htmlStr += `
                        <tr bgcolor="#FFFFFF" align="center" height="30">
                            <td bgcolor="#FFEFBA">
                                    <font size=2>&nbsp;create:&nbsp;</font>
                            </td>
                            <td>
				<!-- <input style="font-size:12px;text-align:center;" type="text" name="text_box" id="1" value="AWS"> -->
		`
	// Select format of CloudOS  name=text_box, id=1
	htmlStr += makeSelect_html("onchangeProvider", nameList, "1")

	htmlStr += `	
                            </td>
                            <td>
                                <textarea style="font-size:12px;text-align:center;" name="text_box" id="2" cols=50>[{"Key":"ClientId", "Value":"XXXXXX"}, {"Key":"ClientSecret", "Value":"XXXXXX"}]</textarea>
                            </td>
                            <td>
                                <input style="font-size:12px;text-align:center;" type="text" name="text_box" id="3" value="aws-credential-01">
                            </td>
                            <td>
                                <a href="javascript:postCredential()">
                                    <font size=3><b>+</b></font>
                                </a>
                            </td>
                        </tr>
                `
	// make page tail
	htmlStr += `
                    </table>
		    <hr>
                </body>
                </html>
        `

	//fmt.Println(htmlStr)
	return c.HTML(http.StatusOK, htmlStr)
}

// make the string of javascript function
func makeOnchangeRegionProviderFunc_js() string {
	strFunc := `
              function onchangeProvider(source) {
                var providerName = source.value
        // for region info
        switch(providerName) {
          case "AWS":
            regionInfo = '[{"Key":"Region", "Value":"us-east-2"}, {"Key":"Zone", "Value":"us-east-2a"}]'
            region = '(ohio)us-east-2a'
            break;
          case "AZURE":
            regionInfo = '[{"Key":"Region", "Value":"northeurope"}, {"Key":"Zone", "Value":"1"}]'
            region = 'northeurope'
            break;
          case "GCP":
            regionInfo = '[{"Key":"Region", "Value":"us-central1"},{"Key":"Zone", "Value":"us-central1-a"}]'
            region = 'us-central1-a'
            break;
          case "ALIBABA":
            regionInfo = '[{"Key":"Region", "Value":"ap-northeast-1"}, {"Key":"Zone", "Value":"ap-northeast-1a"}]'
            region = 'ap-northeast-1a'
            break;
          case "TENCENT":
            regionInfo = '[{"Key":"Region", "Value":"ap-beijing"}, {"Key":"Zone", "Value":"ap-beijing-3"}]'
            region = 'ap-beijing-3'
            break;
          case "IBM":
            regionInfo = '[{"Key":"Region", "Value":"us-south"}, {"Key":"Zone", "Value":"us-south-1"}]'
            region = 'us-south-1'
            break;
          case "OPENSTACK":
            regionInfo = '[{"Key":"Region", "Value":"RegionOne"}]'
            region = 'RegionOne'
            break;

          case "NCPVPC":
            regionInfo = '[{"Key":"Region", "Value":"KR"}, {"Key":"Zone", "Value":"KR-1"}]'
            region = 'KR-1'
            break;
          case "NCP":
            regionInfo = '[{"Key":"region", "Value":"KR"}]'
            region = 'KR'
            break;
          case "NHNCLOUD":
            regionInfo = '[{"Key":"Region", "Value":"KR1"}]'
            region = 'KR1'
            break;
		case "KTCLOUD":
            regionInfo = '[{"Key":"Region", "Value":"KOR-Seoul"}, {"Key":"Zone", "Value":"95e2f517-d64a-4866-8585-5177c256f7c7"}]'
            region = 'KOR-Seoul-M'
            break;
		case "KTCLOUDVPC":
            regionInfo = '[{"Key":"Region", "Value":"KR1"}, {"Key":"Zone", "Value":"DX-M1"}]'
            region = 'KR1-DX-M1'
            break;

          case "MOCK":
            regionInfo = '[{"Key":"Region", "Value":"default"}]'
            region = 'default'
            break;
          case "CLOUDTWIN":
            regionInfo = '[{"Key":"Region", "Value":"default"}]'
            region = 'default'
            break;
          default:
            regionInfo = '[{"Key":"Region", "Value":"default"}]'
            region = 'default'
        }
                document.getElementById('2').value= regionInfo

        // for region-zone name
                document.getElementById('3').value= providerName.toLowerCase() + "-" + region;
              }
        `
	return strFunc
}

// number, Provider Name, Region Info, Region Name, checkbox
func makeRegionTRList_html(bgcolor string, height string, fontSize string, infoList []*rim.RegionInfo) string {
	if bgcolor == "" {
		bgcolor = "#FFFFFF"
	}
	if height == "" {
		height = "30"
	}
	if fontSize == "" {
		fontSize = "2"
	}

	// make base TR frame for info list
	strTR := fmt.Sprintf(`
                <tr bgcolor="%s" align="center" height="%s">
                    <td>
                            <font size=%s>$$NUM$$</font>
                    </td>
                    <td>
                            <font size=%s>$$S1$$</font>
                    </td>
                    <td>
                            <font size=%s>$$S2$$</font>
                    </td>
                    <td>
                            <font size=%s>$$S3$$</font>
                    </td>
                    <td>
                        <input type="checkbox" name="check_box" value=$$S3$$>
                    </td>
                </tr>
                `, bgcolor, height, fontSize, fontSize, fontSize, fontSize)

	strData := ""
	// set data and make TR list
	for i, one := range infoList {
		str := strings.ReplaceAll(strTR, "$$NUM$$", strconv.Itoa(i+1))
		str = strings.ReplaceAll(str, "$$S1$$", one.ProviderName)
		strKeyList := ""
		for _, kv := range one.KeyValueInfoList {
			strKeyList += kv.Key + ":" + kv.Value + ", "
		}
		str = strings.ReplaceAll(str, "$$S2$$", strKeyList)
		str = strings.ReplaceAll(str, "$$S3$$", one.RegionName)
		strData += str
	}

	return strData
}

// make the string of javascript function
func makePostRegionFunc_js() string {

	// curl -X POST http://$RESTSERVER:1024/spider/region -H 'Content-Type: application/json'
	//      -d '{"RegionName":"aws-(ohio)us-east-2","ProviderName":"AWS", "KeyValueInfoList":
	//.       '[{"Key":"Region", "Value":"us-east-2"}, {"Key":"Zone", "Value":"us-east-2a"}]'}'

	strFunc := `
                function postRegion() {
                        var textboxes = document.getElementsByName('text_box');
            sendJson = '{ "ProviderName" : "$$PROVIDER$$", "KeyValueInfoList" : $$REGIONINFO$$, "RegionName" : "$$NAME$$" }'

                        for (var i = 0; i < textboxes.length; i++) { // @todo make parallel executions
                                switch (textboxes[i].id) {
                                        case "1":
                                                sendJson = sendJson.replace("$$PROVIDER$$", textboxes[i].value);
                                                break;
                                        case "2":
                                                sendJson = sendJson.replace("$$REGIONINFO$$", textboxes[i].value);
                                                break;
                                        case "3":
                                                sendJson = sendJson.replace("$$NAME$$", textboxes[i].value);
                                                break;
                                        default:
                                                break;
                                }
                        }
                        var xhr = new XMLHttpRequest();
                        xhr.open("POST", "$$SPIDER_SERVER$$/spider/region", false);
                        xhr.setRequestHeader('Content-Type', 'application/json');
                        //xhr.send(JSON.stringify({ "RegionName": regionName, "ProviderName": providerName, "KeyValueInfoList": regionInfo}));
                        //xhr.send(JSON.stringify(sendJson));

			// client logging
			parent.frames["log_frame"].Log("curl -sX POST " + "$$SPIDER_SERVER$$/spider/region -H 'Content-Type: application/json' -d '" + sendJson + "'");

                        xhr.send(sendJson);

			// client logging
			parent.frames["log_frame"].Log("   => " + xhr.response);

                        // setTimeout(function(){ // when async call
                                location.reload();
                        // }, 400);

                }
        `
	strFunc = strings.ReplaceAll(strFunc, "$$SPIDER_SERVER$$", "http://"+cr.ServiceIPorName+cr.ServicePort) // cr.ServicePort = ":1024"
	return strFunc
}

// make the string of javascript function
func makeDeleteRegionFunc_js() string {
	// curl -X DELETE http://$RESTSERVER:1024/spider/region/aws-(ohio)us-east-2 -H 'Content-Type: application/json'

	strFunc := `
                function deleteRegion() {
                        var checkboxes = document.getElementsByName('check_box');
                        for (var i = 0; i < checkboxes.length; i++) { // @todo make parallel executions
                                if (checkboxes[i].checked) {
                                        var xhr = new XMLHttpRequest();
                                        xhr.open("DELETE", "$$SPIDER_SERVER$$/spider/region/" + checkboxes[i].value, false);
                                        xhr.setRequestHeader('Content-Type', 'application/json');

                                        // client logging
                                        parent.frames["log_frame"].Log("curl -sX DELETE " + "$$SPIDER_SERVER$$/spider/region/" + checkboxes[i].value + " -H 'Content-Type: application/json'" );

                                        xhr.send(null);

                                        // client logging
                                        parent.frames["log_frame"].Log("   => " + xhr.response);
                                }
                        }
			location.reload();
                }
        `
	strFunc = strings.ReplaceAll(strFunc, "$$SPIDER_SERVER$$", "http://"+cr.ServiceIPorName+cr.ServicePort) // cr.ServicePort = ":1024"
	return strFunc
}

// ================ Region Info Management
// create region page
func Region(c echo.Context) error {
	cblog.Info("call Region()")

	// make page header
	htmlStr := `
                <html>
                <head>
                    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
                    <script type="text/javascript">
                `
	// (1) make Javascript Function
	htmlStr += makeOnchangeRegionProviderFunc_js()
	htmlStr += makeCheckBoxToggleFunc_js()
	htmlStr += makePostRegionFunc_js()
	htmlStr += makeDeleteRegionFunc_js()

	htmlStr += `
                    </script>
                </head>

                <body>
                    <table border="0" bordercolordark="#F8F8FF" cellpadding="0" cellspacing="1" bgcolor="#FFFFFF"  style="font-size:small;">
                `

	// (2) make Table Action TR
	// colspan, f5_href, delete_href, fontSize
	htmlStr += makeActionTR_html("5", "region", "deleteRegion()", "2")

	// (3) make Table Header TR
	nameWidthList := []NameWidth{
		{"Provider Name", "200"},
		{"Region Info", "300"},
		{"Region Name", "200"},
	}
	htmlStr += makeTitleTRList_html("#DDDDDD", "2", nameWidthList, true)

	// (4) make TR list with info list
	// (4-1) get info list @todo if empty list

	// client logging
	htmlStr += genLoggingGETResURL("region")

	resBody, err := getResourceList_JsonByte("region")
	if err != nil {
		cblog.Error(err)
		// client logging
		htmlStr += genLoggingGETResURL(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// client logging
	htmlStr += genLoggingResult(string(resBody[:len(resBody)-1]))

	var info struct {
		ResultList []*rim.RegionInfo `json:"region"`
	}
	json.Unmarshal(resBody, &info)

	// (4-2) make TR list with info list
	htmlStr += makeRegionTRList_html("", "", "", info.ResultList)

	// (5) make input field and add
	// attach text box for add
	nameList := cloudosList()
	htmlStr += `
                        <tr bgcolor="#FFFFFF" align="center" height="30">
                            <td bgcolor="#FFEFBA">
                                    <font size=2>&nbsp;create:&nbsp;</font>
                            </td>
                            <td>
                <!-- <input style="font-size:12px;text-align:center;" type="text" name="text_box" id="1" value="AWS"> -->
        `
	// Select format of CloudOS  name=text_box, id=1
	htmlStr += makeSelect_html("onchangeProvider", nameList, "1")

	htmlStr += `    
                            </td>
                            <td>
                                <textarea style="font-size:12px;text-align:center;" name="text_box" id="2" cols=50>[{"Key":"Region", "Value":"us-east-2"}, {"Key":"Zone", "Value":"us-east-2a"}]</textarea>
                            </td>
                            <td>
                                <input style="font-size:12px;text-align:center;" type="text" name="text_box" id="3" value="aws-(ohio)us-east-2">
                            </td>
                            <td>
                                <a href="javascript:postRegion()">
                                    <font size=3><b>+</b></font>
                                </a>
                            </td>
                        </tr>
                `
	// make page tail
	htmlStr += `
                    </table>
		    <hr>
                </body>
                </html>
        `

	//fmt.Println(htmlStr)
	return c.HTML(http.StatusOK, htmlStr)
}

func getProviderName(connConfig string) (string, error) {
	resBody, err := getResource_JsonByte("connectionconfig", connConfig)
	if err != nil {
		cblog.Error(err)
		return "", err
	}
	var configInfo ccim.ConnectionConfigInfo
	json.Unmarshal(resBody, &configInfo)

	return configInfo.ProviderName, nil
}

func getRegionName(connConfig string) (string, error) {
	resBody, err := getResource_JsonByte("connectionconfig", connConfig)
	if err != nil {
		cblog.Error(err)
		return "", err
	}
	var configInfo ccim.ConnectionConfigInfo
	json.Unmarshal(resBody, &configInfo)

	return configInfo.RegionName, nil
}

func getRegionZone(regionName string) (string, string, error) {
	// Region Name List
	resBody, err := getResource_JsonByte("region", regionName)
	if err != nil {
		cblog.Error(err)
		return "", "", err
	}
	var regionInfo rim.RegionInfo
	json.Unmarshal(resBody, &regionInfo)

	region := ""
	zone := ""
	// get the region & zone
	for _, one := range regionInfo.KeyValueInfoList {
		if one.Key == "Region" || one.Key == "region" {
			region = one.Value
		}
		if one.Key == "location" {
			region = one.Value
		}
		if one.Key == "Zone" || one.Key == "zone" {
			zone = one.Value
		}

	}
	return region, zone, nil
}

func makeDriverNameHiddenTRList_html(infoList []*dim.CloudDriverInfo) string {

	// make base Label frame for info list
	strTR := `<label name="driverName-$$CSP$$" hidden>$$DRIVERNAME$$</label>`

	strData := ""
	// set data and make TR list
	for _, one := range infoList {
		str := strings.ReplaceAll(strTR, "$$CSP$$", one.ProviderName)
		str = strings.ReplaceAll(str, "$$DRIVERNAME$$", one.DriverName)
		strData += str
	}

	return strData
}

func makeCredentialNameHiddenTRList_html(infoList []*cim.CredentialInfo) string {

	// make base Label frame for info list
	strTR := `<label name="credentialName-$$CSP$$" hidden>$$CREDENTIALNAME$$</label>`

	strData := ""
	// set data and make TR list
	for _, one := range infoList {
		str := strings.ReplaceAll(strTR, "$$CSP$$", one.ProviderName)
		str = strings.ReplaceAll(str, "$$CREDENTIALNAME$$", one.CredentialName)
		strData += str
	}

	return strData
}

func makeRegionNameHiddenTRList_html(infoList []*rim.RegionInfo) string {

	// make base Label frame for info list
	strTR := `<label name="regionName-$$CSP$$" hidden>$$REGIONNAME$$</label>`

	strData := ""
	// set data and make TR list
	for _, one := range infoList {
		str := strings.ReplaceAll(strTR, "$$CSP$$", one.ProviderName)
		str = strings.ReplaceAll(str, "$$REGIONNAME$$", one.RegionName)
		strData += str
	}

	return strData
}

// ================ Spider Info
func SpiderInfo(c echo.Context) error {
	cblog.Info("call SpiderInfo()")

	htmlStr := `
        <html>
            <head>
                <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
                <title>CB-Spider Information</title>
                <style>
                    body {
                        font-family: Arial, sans-serif;
                        font-size: 14px;
                        margin: 20px;
                        background-color: #f5f5f5;
                    }
                    .container {
                        background-color: #ffffff;
                        padding: 20px;
                        border-radius: 8px;
                        box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
                    }
                    .header {
                        font-size: 20px;
                        font-weight: bold;
                        margin-bottom: 20px;
                        color: #333333;
                    }
                    table {
                        width: 100%;
                        border-collapse: collapse;
                        margin-bottom: 20px;
                    }
                    th, td {
                        border: 1px solid #dddddd;
                        text-align: center;
                        padding: 8px;
                    }
                    th {
                        background-color: #f2f2f2;
                        font-weight: bold;
                    }
                    .info-section {
                        margin-top: 40px;
                    }
                    .info-title {
                        font-size: 18px;
                        font-weight: bold;
                        margin-bottom: 10px;
                    }
                    .api-docs a {
                        color: #007bff;
                        text-decoration: none;
                    }
                    .api-docs a:hover {
                        text-decoration: underline;
                    }
                </style>
            </head>
            <body>
                <div class="container">
                    <div class="header">CB-Spider Information</div>

                    <div class="info-section">
                        <div class="info-title">Server Information</div>
                        <table>
                            <tr>
                                <th>Server Start Time</th>
                                <th>Server Version</th>
                                <th>API Version</th>
                            </tr>
                            <tr>
                                <td>$$STARTTIME$$</td>
                                <td>CB-Spider v0.9.0 (Cinnamon)</td>
                                <td>REST API v0.9.0 (Cinnamon)</td>
                            </tr>
                        </table>
                    </div>

                    <div class="info-section">
                        <div class="info-title">API Endpoint Information</div>
                        <table>
                            <tr>
                                <th>API Endpoint</th>
                                <th>API Documentation</th>
                            </tr>
                            <tr>
                                <td>$$APIENDPOINT$$</td>
                                <td class="api-docs">
                                    <a href="https://github.com/cloud-barista/cb-spider/wiki/CB-Spider-User-Interface" target="_blank">
                                        CB-Spider User Interface Documentation
                                    </a>
                                </td>
                            </tr>
                        </table>
                    </div>

                    <div class="info-section">
                        <div class="info-title">Additional Information</div>
                        <table>
                            <tr>
                                <th>Project Repository</th>
                                <th>Support</th>
                            </tr>
                            <tr>
                                <td class="api-docs">
                                    <a href="https://github.com/cloud-barista/cb-spider" target="_blank">
                                        GitHub: cloud-barista/cb-spider
                                    </a>
                                </td>
                                <td>
                                    For support, please visit the 
                                    <a href="https://github.com/cloud-barista/cb-spider/issues" target="_blank">issue tracker</a>.
                                </td>
                            </tr>
                        </table>
                    </div>
                </div>
            </body>
        </html>
    `

	htmlStr = strings.ReplaceAll(htmlStr, "$$STARTTIME$$", cr.StartTime)
	htmlStr = strings.ReplaceAll(htmlStr, "$$APIENDPOINT$$", "http://"+cr.ServiceIPorName+cr.ServicePort+"/spider")

	return c.HTML(http.StatusOK, htmlStr)
}
