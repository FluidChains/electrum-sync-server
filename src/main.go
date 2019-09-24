package main

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ianschenck/envflag"
	"log"
	"net/http"
	"os"
)

const (
	Version = "0.0.4"
)

var dbType = envflag.String("DBTYPE", "postgres", "Database to use, valid options sqlite3, mysql or postgres")
var dbDir = envflag.String("DBDIR", defaultDbDir(), "Sqlite3 - Path where the application should look for the database file.")

var user = envflag.String("DBUSER", "postgres", "postgres/mysql - user name")
var password = envflag.String("DBPASSWORD", "''", "postgres/mysql - password")
var host = envflag.String("DBHOST", "db_labelsync", "postgres/mysql - hostname")
var port = envflag.Int("DBPORT", 5432, "postgres/mysql - port")
var db = envflag.String("DBDATABASE", "postgres", "postgres/mysql - database name")

var listenPort = envflag.String("LISTENPORT", "localhost:8080", "Port where the json api should listen at in host:port format.")

var useTLS = envflag.Bool("useTls", false, "Serve json api conncetions over TLS.")
var certPath = envflag.String("certPath", "cert.pem", "Path to TLS certificate")
var keyPath = envflag.String("keyPath", "key.pem", "Path to Keyfile")

func main() {
	envflag.Parse()
	var sm SyncMaster

	if *dbType == "sqlite3" {
		var opts DbOpts
		opts.DbType = *dbType
		opts.DbPath = *dbDir
		sm = newSyncMaster(opts)
	} else if (*dbType == "postgres" || *dbType == "mysql") {
		var opts DbOpts
		opts.DbType = *dbType
		opts.User = *user
		opts.Password = *password
		opts.Host = *host
		opts.Dbname = *db
                opts.Port = *port
		sm = newSyncMaster(opts)
	} else {
		log.Fatal("Please define which database to use, sqlite3 or postgres")
	}

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	router, err := rest.MakeRouter(
		&rest.Route{"GET", "/labels/since/:nonce/for/:mpk", sm.GetLabels},
		&rest.Route{"POST", "/label", sm.CreateLabel},
		&rest.Route{"POST", "/labels", sm.CreateLabels},
	)

	if err != nil {
		log.Fatal(err)
	}

	api.SetApp(router)

	// Special exception to make Heroku like deployment easy
	if os.Getenv("PORT") != "" {
		*listenPort = fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT"))
	}

	sm.logger.Info("Server started and listening on %s", *listenPort)
	if *useTLS {
		sm.logger.Info("Using SSL with certificate '%s' and keyfile '%s'", *certPath, *keyPath)
		log.Fatal(http.ListenAndServeTLS(*listenPort, *certPath, *keyPath, api.MakeHandler()))
	} else {
		log.Fatal(http.ListenAndServe(*listenPort, api.MakeHandler()))
	}
}
