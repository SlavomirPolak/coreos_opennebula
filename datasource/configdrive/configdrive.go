package main

import (
	"fmt"
	"os"
	"log"
	"io/ioutil"
	b64 "encoding/base64"
	
	"github.com/SlavomirPolak/bashParser/src/bashParser"
	//"github.com/coreos/coreos-cloudinit/datasource"
)

const (
	openstackApiVersion = "latest"
)

type configDrive struct {
	root string
	readFile func(filename string) ([]byte, error)
	variablesMap map[string]string
	err error
	
}

func NewDatasource(root string) *configDrive {
	variablesMap, err := bashParser.UseShlex(root + "context.sh")
	if err != nil {
		log.Printf("Error during parsing script file.\n")
	}
	return &configDrive{root, ioutil.ReadFile, variablesMap, err}
}

func (cd *configDrive) IsAvailable() bool {
	_, err := os.Stat(cd.root)
	return !os.IsNotExist(err)
}

func (cd *configDrive) AvailabilityChanges() bool {
	return true
}

func (cd *configDrive) ConfigRoot() string {
	return cd.root
}

/*preskakujem openstackVersionRoot a openstackRoot
func (cd *configDrive) openstackRoot() string {
	return cd.root
}*/

func (cd *configDrive) FetchMetadata() ([]byte, error) {
	var metadata struct {
		SSH_KEY []byte
	}
	
	// searching for SSH_PUBLIC_KEY or SSH_KEY or PUBLIC_SSH_KEY
	var val string
	if cd.variablesMap["SSH_PUBLIC_KEY"] != "" {
		val = cd.variablesMap["SSH_PUBLIC_KEY"]
	} else if cd.variablesMap["SSH_KEY"] != "" {
		val = cd.variablesMap["SSH_KEY"]
	} else if cd.variablesMap["PUBLIC_SSH_KEY"] != "" {
		val = cd.variablesMap["PUBLIC_SSH_KEY"]
	}
	
	if val != "" {
		/*if cd.variablesMap["USERDATA_ENCODING"] == "base64" {
			var err error
			val, err = decodeBase64(val)
			if err != nil {
				return nil, err
			}
		}*/
		metadata.SSH_KEY = []byte(val)
	} else {
		log.Printf("Variable USER_DATA isnt in script file.\n")
	}
	return metadata.SSH_KEY, cd.err
}

func Type() string {
	return "cloud-drive"
}

func (cd *configDrive) FetchUserdata() ([]byte, error) {
	if cd.variablesMap["USER_DATA"] == "" {
		log.Printf("Variable USER_DATA isnt in script file.\n")
		return nil, cd.err
	}
	userData := cd.variablesMap["USER_DATA"]
	
	if cd.variablesMap["USERDATA_ENCODING"] == "base64" {
		var err error
		userData, err = decodeBase64(userData)
		if err != nil {
			return nil, err
		}
	}
	
	return []byte(userData), cd.err
}

func NewVariablesMap(fileName string) (map[string]string, error) {
	variablesMap, err := bashParser.UseShlex(fileName)
	if err != nil {
		log.Printf("Error during parsing script file.")
		return nil, err
	}
	
	return variablesMap, nil
}

func decodeBase64(text string) (string, error) {
	decodedText, err := b64.StdEncoding.DecodeString(text)
	if err != nil {
		log.Printf("Error during decoding from base64.\n")
	}
	return string(decodedText), err
}

func main() {
	ds := NewDatasource("/home/wolfik/gocode/src/coreos_opennebula/data/")
	userData, err := ds.FetchUserdata()
	data, err := ds.FetchMetadata()
	fmt.Println(string(data), err)
	fmt.Println(string(userData), err)
}