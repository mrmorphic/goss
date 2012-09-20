package goss

/**
 * This file handles things to do with the metadata that generated a database we're connecting with.
 * The ORM needs this to make sense of the tables and relationships etc, and allow the user to reference
 * the classes as in PHP.
 */

/**
 * Represents a database. We hold one such record for each database we connect to. This is common across
 * all connections.
 */
type DBMetadata struct {
	DBIdentifier string
	Classes []ClassInfo
}

type DBField struct {
	Name string
	SSType string
}

type ClassInfo struct {
	ClassName string
	HasTable bool					// physical table mapping
	Fields []*DBField
	SuperClasses []*ClassInfo
	SubClasses []*ClassInfo
}

var Databases map[string] *DBMetadata

func init() {
	Databases = make(map[string] *DBMetadata)
}

func (dbm *DBMetadata) IsHierarchical(className string) (bool, error) {
	// @todo implement
	if className == "SiteTree" || className == "ProjectPage" || className == "EntryPage" || className == "ProjectHolder" || className == "SpecialProjectPage" {
		return true, nil
	}
	return false, nil
}

func (dbm *DBMetadata) IsVersioned(className string) (bool, error) {
	// @todo implement
	return false, nil
}

/**
 * Return a class and all it's defined properties.
 */
func (dbm *DBMetadata) GetClass(className string) (* ClassInfo, error) {
	return nil, nil
}

func (dbm *DBMetadata) IsSubclass(className string, inclusive bool) (bool, error) {
	// @todo implement
	return false, nil
}