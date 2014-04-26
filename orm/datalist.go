package orm

type DataList struct {
	Items []DataObject
}

func NewDataList() *DataList {
	// return &DataList{make([]*DataObject, 10)}
	return &DataList{make([]DataObject, 0, 10)}
}

func (set *DataList) First() DataObject {
	if len(set.Items) < 1 {
		return nil
	}
	return set.Items[0]
}
