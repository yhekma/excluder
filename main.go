package main

import (
	"flag"
	"fmt"
	"github.com/pkg/xattr"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Dirnames []string `yaml:"direxcludes"`
	Extensions []string `yaml:"extexcludes"`
}

func checkErr(e error, msg string) {
	if e != nil {
		log.Fatalf("%s: %v\n", msg, e)
	}
}

func checkInSlice(item string, slice []string) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}

func walkAll(root string, extensions, dirnames []string) error {
	e := filepath.Walk(root, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if info.Mode().IsDir() {
			for _, d := range dirnames {
				if info.Name() == d {
					e = xattr.Set(path, "com.apple.metadata:com_apple_backup_excludeItem", []byte("com.apple.backupd"))
					fmt.Println("setting: ", path)
					checkErr(e, fmt.Sprint("could not set attribute on", path))
				}
			}
		}

		if info.Mode().IsRegular() {
			process := true
			parts := strings.Split(path, string(filepath.Separator))
			for _, p := range parts {
				if checkInSlice(p, dirnames) {
					process = false
				}
			}
			for _, ext := range extensions {
				if strings.HasSuffix(info.Name(), ext) && process {
					fmt.Println("setting: ", path)
					e = xattr.Set(path, "com.apple.metadata:com_apple_backup_excludeItem", []byte("com.apple.backupd"))
					checkErr(e, fmt.Sprint("could not set attribute on", path))
				}
			}
		}
		return nil
	})
	return e
}

func main() {
	var e error
	var config Config
	cfg := flag.String("config", "excluder.yaml", "yaml file with config")
	flag.Parse()

	f, e := ioutil.ReadFile(*cfg)
	checkErr(yaml.Unmarshal(f, &config), "could not unmarshal")

	e = walkAll("tests", config.Extensions, config.Dirnames)
	checkErr(e, "could not walk")
}