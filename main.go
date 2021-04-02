package main

import (
	"flag"
	"fmt"
	"github.com/karrick/godirwalk"
	"github.com/pkg/xattr"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type counter struct {
	mu    sync.Mutex
	value int
}

func (c *counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

var dirCount counter
var fileCount counter

func logInit() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

type Config struct {
	Dirnames   []string `yaml:"direxcludes"`
	Extensions []string `yaml:"extexcludes"`
}

func checkErr(e error, msg string) {
	if e != nil {
		log.Errorf("%s: %v\n", msg, e)
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
	e := godirwalk.Walk(root, &godirwalk.Options{
		ErrorCallback: func(path string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
		Callback: func(path string, de *godirwalk.Dirent) error {
			if de.IsDir() {
				for _, d := range dirnames {
					if de.Name() == d {
						log.Debugf("excluding from backup: %s", path)
						e := xattr.Set(path, "com.apple.metadata:com_apple_backup_excludeItem", []byte("com.apple.backupd"))
						checkErr(e, fmt.Sprint("could not set attribute on", path))
						if e == nil {
							dirCount.Inc()
						}
					}
				}
			}

			if de.IsRegular() {
				parts := strings.Split(path, string(filepath.Separator))
				for _, p := range parts {
					if checkInSlice(p, dirnames) {
						return godirwalk.SkipThis
					}
				}
				for _, ext := range extensions {
					if strings.HasSuffix(de.Name(), ext) {
						log.Debugf("excluding from backup: %s", path)
						e := xattr.Set(path, "com.apple.metadata:com_apple_backup_excludeItem", []byte("com.apple.backupd"))
						checkErr(e, fmt.Sprint("could not set attribute on", path))
						if e == nil {
							fileCount.Inc()
						}
					}
				}
			}
			return nil
		},
	})
	return e
}

func main() {
	logInit()
	var e error
	var config Config
	cfg := flag.String("config", "excluder.yaml", "yaml file with config")
	verb := flag.Bool("verbose", false, "run in verbose mode")
	flag.Parse()
	root := flag.Arg(0)
	if *verb {
		log.SetLevel(log.DebugLevel)
		log.Debug("running in debug mode as specified")
	}

	log.Debugf("reading config file %s", *cfg)
	f, e := ioutil.ReadFile(*cfg)
	checkErr(yaml.Unmarshal(f, &config), "could not unmarshal")
	log.Debugf("parsed configfile, excluded directories: %s", config.Dirnames)
	log.Debugf("parsed configfile, excluded extensions: %s", config.Extensions)
	log.Debugf("using root dir: %s", root)

	e = walkAll(root, config.Extensions, config.Dirnames)
	checkErr(e, "could not walk")
	log.Debugf("excluded %d directories, %d files", dirCount.value, fileCount.value)
}
