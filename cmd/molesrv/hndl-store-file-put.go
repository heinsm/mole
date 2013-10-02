package main

import (
	"bytes"
	"github.com/calmh/mole/conf"
	"github.com/calmh/mole/ini"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

var obfuscateKeys = []string{
	"key",
	"password",
	"IPSec_secret",
	"Xauth_password",
}

func init() {
	addHandler(handler{
		pattern: "/store/",
		method:  "PUT",
		fn:      putFile,
		auth:    true,
		ro:      false,
	})
}

func putFile(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		defer listCacheLock.Unlock()
		listCacheLock.Lock()
		listCache = nil
	}()

	iniFile := path.Join(storeDir, "data", req.URL.Path[7:])
	// Read pushed data
	data, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}

	// Verify the configuration
	_, err = conf.Load(bytes.NewBuffer(data))
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}

	// Get the raw INI
	inf := ini.Parse(bytes.NewBuffer(data))

	// Obfuscate
	shouldSaveKeys := false
	for _, section := range inf.Sections() {
		for _, option := range inf.Options(section) {
			for i := range obfuscateKeys {
				if option == obfuscateKeys[i] {
					val := inf.Get(section, option)
					if oval := obfuscate(val); oval != val {
						inf.Set(section, option, oval)
						shouldSaveKeys = true
					}
					break
				}
			}
		}
	}
	if shouldSaveKeys {
		saveKeys()
	}

	// Save
	outf, err := os.Create(iniFile)
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	inf.Write(outf)
	outf.Close()
}