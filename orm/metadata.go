package orm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// This file handles things to do with the metadata that generated a database we're connecting with.
// The ORM needs this to make sense of the tables and relationships etc, and allow the user to reference
// the classes as in PHP.

// DBMetadata represents the structure of a SilverStripe database. We create one instance for each database
// we connect to. This is common across all connections.
type DBMetadata struct {
	// Identifier
	DBIdentifier string

	// Raw classes read in from metadata file.
	Classes []*ClassInfo

	// Map of class name to ClassInfo object.
	ClassMap map[string]*ClassInfo
}

type DBField struct {
	Name   string
	SSType string
}

// This holds metadata about each class. We read this in from the file system.
type ClassInfo struct {
	ClassName   string
	HasTable    bool
	Versioned   bool
	TableName   string
	Ancestors   []string
	Descendents []string
	//	Fields []*DBField
	//	SuperClasses []*ClassInfo
	//	SubClasses []*ClassInfo

	// We'll pre-calculate the FROM clause for queries where this class is the base class. This makes query
	// construction very fast and almost trivial.
	defaultFrom string

	// Likewise we'll precalculate a part of the where clause that selects ClassName being the base class or
	// any of its descendents.
	defaultWhere string
}

var Databases map[string]*DBMetadata

func init() {
	Databases = make(map[string]*DBMetadata)
}

// Refresh this metadata object from the metadata source file provided. We only actually do that under two circumstances:
// - we have not initialised yet
// - the metadata file has a modification time that is more recent than when we last refreshed.
func (dbm *DBMetadata) RefreshOnDemand(metadataSource string) error {
	// fmt.Printf("DBMetadata::RefreshOnDemand called with %s\n", metadataSource)

	// @todo check if file has changed since we last read it. Return if it hasn't.

	file, err := os.Open(metadataSource)
	if err != nil {
		return err
	}

	// Read the json file. This is marshalled directly into dbm.
	decoder := json.NewDecoder(bufio.NewReader(file))
	err = decoder.Decode(&dbm)
	if err != nil && err != io.EOF {
		return err
	}

	// for c, def := range dbm.Classes {
	// 	fmt.Printf("Class define for class %s is %s\n", c, def)
	// }

	dbm.precache()

	fmt.Printf("After loading metadata, dbm now looks like %s\n", dbm)
	return nil
}

// After loading metadata from the JSON file, we then pre-calculate some cached info that makes the ORM faster. As much as possible we want to bypass looking up classes
// by name, instead creating an object graph that can be traversed in SQL generation.
func (dbm *DBMetadata) precache() {
	// First create a map of classes by name.
	dbm.ClassMap = make(map[string]*ClassInfo)
	for _, c := range dbm.Classes {
		dbm.ClassMap[c.ClassName] = c
	}
	// fmt.Printf("before precache, class map is %s\n", dbm.ClassMap)
	for _, c := range dbm.Classes {
		c.precacheDefaultFromWhere(dbm)
	}
}

// Calculate the defaultFrom property, which is going to be a join clause
func (ci *ClassInfo) precacheDefaultFromWhere(dbm *DBMetadata) {
	fromClause := ""
	whereClause := ""
	first := true
	lastTable := ""
	rootTable := ""

	// fmt.Printf("precacheDefaultFromWhere: class info is %s\n", ci)

	// fmt.Printf("... processing ancestors\n")
	// Join all ancestors leading up to this class, including the class itself, but only where a class has a table.
	for _, c := range ci.Ancestors {
		a := dbm.GetClass(c)

		if rootTable == "" {
			rootTable = a.TableName
		}

		if a.HasTable {
			if !first {
				fromClause += " inner join "
			}
			fromClause += "\"" + a.TableName + "\" \"" + a.TableName + "\""
			if !first {
				fromClause += " on \"" + lastTable + "\".\"ID\"=\"" + a.TableName + "\".\"ID\""
			}
			first = false
			lastTable = a.TableName
		}
	}

	// Join all ancestors leading up to this class, including the class itself, but only where a class has a table.
	// fmt.Printf("... processing descendents\n")
	whereClause = "(\"" + rootTable + "\".\"ClassName\"='" + ci.ClassName + "'"
	for _, c := range ci.Descendents {
		d := dbm.GetClass(c)
		// fmt.Printf("...looking up descendent class %s\n", c)
		if d == nil {
			fmt.Printf("class %s could not be found\n", c)
		}
		if d.HasTable {
			fromClause += " left join "
			fromClause += "\"" + d.TableName + "\" \"" + d.TableName + "\""
			fromClause += " on \"" + lastTable + "\".\"ID\"=\"" + d.TableName + "\".\"ID\""
		}
		whereClause += " or \"" + rootTable + "\".\"ClassName\"='" + d.ClassName + "'"

	}
	whereClause += ")"

	//	fmt.Printf("precache: calculated defaultFrom for class %s: %s\n\n\n", ci.ClassName, fromClause)
	ci.defaultFrom = fromClause
	//	fmt.Printf("precache: calculated defaultWhere for class %s: %s\n\n\n", ci.ClassName, whereClause)
	ci.defaultWhere = whereClause
}

func (dbm *DBMetadata) IsHierarchical(className string) bool {
	// @todo implement
	if className == "SiteTree" || className == "ProjectPage" || className == "EntryPage" || className == "ProjectHolder" || className == "SpecialProjectPage" || className == "HomePage" || className == "FeedbackPage" || className == "MessagesPage" || className == "Page" {
		return true
	}
	return false
}

func (dbm *DBMetadata) IsVersioned(className string) bool {
	// @todo implement
	c := dbm.GetClass(className)
	return c.Versioned
}

// Return a ClassInfo and all it's defined properties given a class name.
func (dbm *DBMetadata) GetClass(className string) *ClassInfo {
	// fmt.Printf("GetClass: getting %s\n", className)
	//	fmt.Printf("GetClass: map is %s\n", dbm.ClassMap)
	return dbm.ClassMap[className]
}

func (dbm *DBMetadata) IsSubclass(className string, inclusive bool) (bool, error) {
	// @todo implement
	return false, nil
}

func (dbm *DBMetadata) TablesForClass(className string) ([]string, error) {
	var t = []string{"SiteTree"}
	return t, nil
}
