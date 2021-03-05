// Copyright (c) 2019 Romano, Viacoin developer
//
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package client

import (
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/spf13/viper"
)

var instance *rpcclient.Client

// Using the singleton design pattern to check if an instance already exist
// if not, only then create a new one
func GetInstance() *rpcclient.Client {
	if instance != nil {
		return instance
	}

	var err error
	connCfg := loadConfig()
	instance, err = rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
		instance.Shutdown()
	}
	return instance
}

// get the current path: client/
var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// config path returns a string which should look
// like: /home/username/go/src/github.com/devt3000/projectname/client/config
func getConfigPath() string {
	path := strings.Split(basepath, "client")
	configPath := path[:len(path)-1][0] + "config"
	return configPath
}

func GetViperConfig() error {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(getConfigPath())
	viper.SetConfigName("app")

	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("No configuration file loaded !\n%s", err)
	}
	return err
}

// load config file from config/app.yml with viper
// the config file should contain the correct RPC details
func loadConfig() *rpcclient.ConnConfig {

	GetViperConfig()

	connCfg := &rpcclient.ConnConfig{
		Host:         viper.GetString("rpc.ip") + ":" + viper.GetString("rpc.port"), //127.0.0.1:8332
		User:         viper.GetString("rpc.username"),
		Pass:         viper.GetString("rpc.password"),
		HTTPPostMode: true, // Viacoin core only supports HTTP POST mode
		DisableTLS:   true, // Viacoin core does not provide TLS by default
	}

	return connCfg
}
