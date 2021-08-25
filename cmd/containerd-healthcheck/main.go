package main

import (
	"containerdhealthcheck/internal/models"
	"containerdhealthcheck/internal/server"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"

	cleanenv "github.com/ilyakaznacheev/cleanenv"
	"github.com/sirupsen/logrus"
)

var version, commit, date string

func main() {

	var yamlConfig models.YAMLConfig

	// Config
	env := flag.StringP("env", "e", "development", "Application environment")
	addr := flag.StringP("addr", "a", ":9891", "HTTP address for prometheus endpoint")
	configPath := flag.StringP("config", "c", "config.yml", "Path to configuration file")
	// Version
	versionOpt := flag.BoolP("version", "v", false, "Print app version")

	flag.Parse()

	if *versionOpt {
		fmt.Println(version)
		os.Exit(0)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	appBuildInfo := models.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}

	serverConfig := models.ServerConfig{
		Env:  *env,
		Addr: *addr,
	}

	// read configuration from the file and environment variables
	if err := cleanenv.ReadConfig(*configPath, &yamlConfig); err != nil {
		logger.Fatal(err)
	}

	err := cleanenv.ReadConfig("config.yml", &yamlConfig)

	if err != nil {
		logger.Fatal(err)
	}

	server, err := server.NewApp(serverConfig, yamlConfig, appBuildInfo, logger)

	if err != nil {
		logger.Fatal("unexpected error: ", err)
	}

	server.Run()

}
