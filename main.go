package main

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"

	"github.com/imdario/mergo"
	log "github.com/qiniu/log.v1"
	"github.com/qiniu/version"
	fileserver "qiniu.com/poc/fileserver/service"
	config "qiniupkg.com/x/config.v7"
)

type Config struct {
	MaxProcs   int    `json:"max_procs"`
	BindHost   string `json:"bind_host"`
	DebugLevel int    `json:"debug_level"`
	FilePath   string `json:"file_path"`
}

const defaultConfig = `
{
    "max_procs": 0,
    "bind_host": ":80",
	"debug_level": 0
}
`
const appName = "file-server"

var conf Config

func init() {
	log.Println("version:", version.Version())
	// load default config
	// var conf Config
	config.LoadString(&conf, defaultConfig)
	config.Init("f", "qiniu", appName+".conf")

	var fileConf Config
	if e := config.Load(&fileConf); e != nil {
		log.Fatal("config.Load failed:", e)
	}
	mergo.MergeWithOverwrite(&conf, fileConf)
	buf, _ := json.MarshalIndent(conf, "", "    ")
	log.Printf("loaded conf \n%s", string(buf))

	if _, err := os.Stat(conf.FilePath); err != nil {
		os.MkdirAll(conf.FilePath, 0777)
	}
}

func main() {
	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	mux := http.NewServeMux()

	service, err := fileserver.NewFileserverService(conf.FilePath)
	if err != nil {
		log.Fatal("failed to create file server service instance:", err)
	}

	mux.HandleFunc("/upload", service.PostUpload)
	mux.Handle("/", http.FileServer(http.Dir(conf.FilePath)))

	log.Infof("Starting %s..., listen on %s", appName, conf.BindHost)
	log.Fatal("http.ListenAndServe:", http.ListenAndServe(conf.BindHost, mux))
	log.Info(appName + " stopped, process exit")

}
