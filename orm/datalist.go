package orm

// DataList represents a set of DataObjects. DataList uses a DataQuery to retrieve values, and does so
// lazily. It's implementation of Evaluater means that Sort, Filter etc can be chained and appended to the
// underlying DataQuery. Only when GetItems is invoked does data actually get fetched.
type DataList interface {
	DataQuery

	Append(interface{})

	Items() ([]interface{}, error)
}

type DataListStruct struct {
	// query is the underlying DataQuery that we manipulate.
	query DataQuery

	fetched bool

	// once we've  executed the query, we'll store the resulting items here.
	items []interface{}
}

func NewDataList(query DataQuery) DataList {
	return &DataListStruct{items: make([]interface{}, 0, 10), query: query}
}

func (set *DataListStruct) Append(do interface{}) {
	set.fetched = true
	set.items = append(set.items, do)
}

func (set *DataListStruct) Sort(clause string, rest ...string) DataQuery {
	set.query = set.query.Sort(clause, rest...)
	return set
}

func (set *DataListStruct) Where(clause interface{}) DataQuery {
	set.query = set.query.Where(clause)
	return set
}

func (set *DataListStruct) Limit(offset int, length int) DataQuery {
	set.query = set.query.Limit(offset, length)
	return set
}

func (set *DataListStruct) Run() (interface{}, error) {
	res, e := set.query.Run()

	if e != nil {
		return nil, e
	}

	// @todo ensure the return type is OK for us.

	set.items = res.([]interface{})

	return set.items, nil
}

func (set *DataListStruct) Items() ([]interface{}, error) {
	if !set.fetched {
		_, e := set.Run()
		if e != nil {
			return nil, e
		}
	}
	return set.items, nil
}

// @todo perform on-demand fetch
func (set *DataListStruct) First() interface{} {
	if !set.fetched {
		_, e := set.Run()
		if e != nil {
			panic(e.Error())
		}
	}

	if len(set.items) < 1 {
		return nil
	}
	return set.items[0]
}
