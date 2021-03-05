// Copyright (c) 2019 Romano, Viacoin developer
//
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/sirupsen/logrus"
<<<<<<< HEAD
	"github.com/devt3000/blockexplorer/blockdata"
	"github.com/devt3000/blockexplorer/cmd/rebuilddb"
	"github.com/devt3000/blockexplorer/mongodb"
=======
	"github.com/romanornr/blockexplorer/blockdata"
	"github.com/romanornr/blockexplorer/cmd/rebuilddb"
	"github.com/romanornr/blockexplorer/mongodb"
>>>>>>> 6077ea7947313fc0ee253740827be35e612bad62
)

var dao = mongodb.MongoDAO{
	"127.0.0.1",
	"viacoin",
}

// This function can be exected with: go run cmd/build.go
// this will build the entire database with blocks, transactions etc
func main() {

	dao.Connect()
	dao.DropDatabase() // delete existing database first

	tip, err := blockdata.GetLatestBlock()
	if err != nil {
		logrus.Fatal("could not get the tip/latest block of the chain")
	}
	rebuilddb.BuildDatabase(tip.Height)
}
