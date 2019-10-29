package client

import "testing"

func TestTable_Range(t *testing.T) {
	tableMap := map[AppIdKey]Client{}
	table := NewTable()
	table.LoadOrStore("test1", &Client{AppId: "test_1"})
	table.LoadOrStore("test2", &Client{AppId: "test_2"})
	table.LoadOrStore("test", &Client{AppId: "test"})
	table.LoadOrStore("test1", &Client{AppId: "test_11111111"})
	table.LoadOrStore("test3", &Client{AppId: "test_3"})
	table.Range(func(key AppIdKey, val *Client) bool {
		if key == "test" {
			table.Store(key, &Client{AppId: "test_new"})
			v, _ := table.Load(key)
			tableMap[key] = *v
		} else if key == "test3" {
			//pass
		} else {
			v, _ := table.Load(key)
			tableMap[key] = *v
		}
		return true
	})
	t.Log("table:", table)
	t.Logf("table map: %+v", tableMap)
	if len(tableMap) != 3 {
		t.Fatal()
	}
	if tableMap["test"].AppId != "test_new" {
		t.Fatal()
	}
}
