package main

import (
	"fmt"
	"os"
	"log"
	"io/ioutil"
	b64 "encoding/base64"
	
	"github.com/SlavomirPolak/bashParser/src/bashParser"
	"github.com/coreos/coreos-cloudinit/datasource"
)

type configDrive struct {
	root string
	readFile func(filename string) ([]byte, error)
	
}

func NewDatasource(root string) *configDrive {
	return &configDrive{root, ioutil.ReadFile}
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

func (cd *configDrive) FetchMetadata() (metadata datasource.Metadata, err error) {
	log.Printf("Attempting to read SSH_KEY from " + cd.root + "context.sh")
	// searching for SSH_PUBLIC_KEY or SSH_KEY or PUBLIC_SSH_KEY
	val, err := fetchVariableFromShellScript(cd.root + "context.sh", "SSH_PUBLIC_KEY")
	if val == "" {
		val, err = fetchVariableFromShellScript(cd.root + "context.sh", "SSH_KEY")
		if val == "" {
			val, err = fetchVariableFromShellScript(cd.root + "context.sh", "PUBLIC_SSH_KEY")
		}
	}
	if err != nil {
		return
	}
	if val != "" {
		var sshKeyMap map[string] string
		sshKeyMap = make(map[string]string)
		sshKeyMap["SSH_KEY"] = val
		metadata.SSHPublicKeys = sshKeyMap
	}
	return 
}

func (cd *configDrive) FetchUserdata() ([]byte, error) {
	log.Printf("Attempting to read USER_DATA from " + cd.root + "context.sh")
	ret, err := fetchVariableFromShellScript(cd.root + "context.sh", "USER_DATA")
	return []byte(ret), err
}

func Type() string {
	return "cloud-drive"
}

func fetchVariableFromShellScript(fileName string, variableName string) (string, error) {
	variablesMap, err := bashParser.UseShlex(fileName)
	if err != nil {
		return "", err
	}
	ret := variablesMap[variableName]
	
	// checking and decoding base64
	if variableName == "USER_DATA" && variablesMap["USERDATA_ENCODING"] == "base64" {
		var err error
		ret, err = decodeBase64(ret)
		if err != nil {
			return "", err
		}
	}	
	return ret, nil
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
	k, _ := ds.FetchUserdata()
	g, _ := ds.FetchMetadata()
	h := string(k)
	fmt.Println("USER_DATA = " + h + ".\nSSH_KEY = " + g.SSHPublicKeys["SSH_KEY"] + ".\n")
}