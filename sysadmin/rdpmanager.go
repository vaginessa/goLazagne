package sysadmin

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/kerbyj/goLazagne/common"
	"log"
	"os"
	"strings"
)

func decryptPassword(encryptedPassword string) string {
	passwordDecrypted := ""
	defer func() {
		// 获取异常信息
		if err := recover(); err != nil {
			//  输出异常信息
			passwordDecrypted = strings.ReplaceAll(encryptedPassword, "\x00", "")
			fmt.Println("passwordDecrypted:", passwordDecrypted)
		}
	}()
	if encryptedPassword == "" {
		return ""
	}

	decoded, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		log.Println("Error decoding password:", err)
		panic(err)
	}
	//fmt.Println("decoded:", decoded)
	decryptedPassword, errUnprotectData := common.Win32CryptUnprotectData(string(decoded), false)
	if errUnprotectData != nil {
		log.Println("Error decrypting password:", errUnprotectData)
		panic(errUnprotectData)
	}
	passwordDecrypted = string(decryptedPassword)
	passwordDecrypted = strings.ReplaceAll(passwordDecrypted, "\x00", "")

	return passwordDecrypted
}

func RDPManagerRun() common.ExtractCredentialsResult {
	settings := []string{
		common.LocalAppData + "\\Microsoft Corporation\\Remote Desktop Connection Manager\\RDCMan.settings",
		common.LocalAppData + "\\Microsoft\\Remote Desktop Connection Manager\\RDCMan.settings",
		//"/Users/peng/PROGRAM/GitHub/goLazagne/test/rdp.xml",
	}

	var pwdFound []common.UrlNamePass

	for _, setting := range settings {
		if _, err := os.Stat(setting); err == nil {

			log.Println(fmt.Sprintf("Setting file found: %s", setting))
			xmlFile, err := os.ReadFile(setting)
			if err != nil {
				fmt.Println("Error reading file:", err)
				return common.EmptyResult
			}

			type FilesToOpen struct {
				FilesToOpen struct {
					Item []string `xml:"item"`
				} `xml:"FilesToOpen"`
			}
			var fileToOpens FilesToOpen
			err = xml.Unmarshal(xmlFile, &fileToOpens)
			if err != nil {
				fmt.Println("Error unmarshalling XML data:", err)
				return common.EmptyResult
			}
			for _, item := range fileToOpens.FilesToOpen.Item {
				if _, err := os.Stat(item); err == nil {
					log.Println(fmt.Sprintf("New setting file found: %s", item))
					if pwdFoundSub, err := ParseXml(item); err == nil {
						pwdFound = append(pwdFound, pwdFoundSub...)
					}
				}
			}
		}
	}

	return common.ExtractCredentialsResult{
		Success: true,
		Data:    pwdFound,
	}
}
func ParseXml(xmlFile string) ([]common.UrlNamePass, error) {
	xmlData, err := os.ReadFile(xmlFile)
	if err != nil {
		return nil, err
	}
	type Server struct {
		Server []struct {
			Host     string `xml:"properties>name"`
			UserName string `xml:"logonCredentials>userName"`
			Password string `xml:"logonCredentials>password"`
		} `xml:"file>server"`
	}
	var servers Server
	err = xml.Unmarshal(xmlData, &servers)
	if err != nil {
		log.Println("Error unmarshalling XML data:", err)
		return nil, err
	}

	var res []common.UrlNamePass
	for _, server := range servers.Server {
		password := decryptPassword(server.Password)

		// print the results
		//fmt.Printf("host: %s, Username: %s, Password: %s\n", server.Host, server.UserName, password)

		res = append(res, common.UrlNamePass{
			Url:      server.Host,
			Username: server.UserName,
			Pass:     password,
		})
	}

	return res, nil
}
