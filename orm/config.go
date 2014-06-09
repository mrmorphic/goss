package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mrmorphic/goss"
)

func init() {
	fns := []func(goss.ConfigProvider) error{setupDB, setupMetadata}
	goss.RegisterInit(fns)
}

// setupDB creates the database connection pool. This is shared across go-routines for all requests,
// and the pool management is managed automatically by the sql package.
func setupDB(config goss.ConfigProvider) error {
	// Get the properties we expect.
	driverName := config.AsString("goss.database.driverName")
	if driverName == "" {
		return errors.New("goss requires config property goss.database.driverName to be set.")
	}

	dataSourceName := config.AsString("goss.database.dataSourceName")
	if dataSourceName == "" {
		return errors.New("goss requires config property goss.database.dataSourceName to be set.")
	}

	maxIdleConnections := -1 // default is no idle connections
	mi := config.Get("goss.database.maxIdleConnections")
	mif, ok := mi.(float64)
	if ok {
		maxIdleConnections = int(mif)
	} else {
		return errors.New("goss expects config property goss.database.maxIdleConnections to be of type 'int'.")

	}

	// put back in once at go 1.2
	maxOpenConnections := -1 // default is no limit on open connections
	mo := config.Get("goss.database.maxOpenConnections")
	mof, ok := mo.(float64)
	if ok {
		maxOpenConnections = int(mof)

	} else {
		return errors.New("goss expects config property goss.database.maxOpenConnections to be of type 'int'.")
	}

	var e error
	database, e = sql.Open(driverName, dataSourceName)
	if e != nil {
		return e
	}

	fmt.Printf("opened database %s: %s\n", driverName, dataSourceName)

	database.SetMaxIdleConns(maxIdleConnections)
	database.SetMaxOpenConns(maxOpenConnections) // requires go 1.2

	// @todo hack alert, refactor driver-specific things.
	if driverName == "mysql" {
		_, e = database.Query("SET GLOBAL TRANSACTION ISOLATION LEVEL SERIALIZABLE;")
		_, e = database.Query("SET GLOBAL sql_mode = 'ANSI'")
	}
	return nil
}

func setupMetadata(conf goss.ConfigProvider) error {
	metadataSource = conf.AsString("goss.metadata")
	if metadataSource == "" {
		return errors.New("goss requires configuration property goss.metadata is set.")
	}

	dbMetadata = new(DBMetadata)
	e := dbMetadata.RefreshOnDemand(metadataSource)

	fmt.Printf("metadata is %s\n", dbMetadata)
	return e
}
