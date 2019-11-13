package com

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type IdKey string

type Server struct {
	Id  IdKey
	Val int
}

type Table struct {
	table sync.Map
}

func NewTable() *Table {
	return &Table{
		table: sync.Map{},
	}
}

func (t *Table) Load(key IdKey) (*Server, bool) {
	val, ok := t.table.Load(key)
	return val.(*Server), ok
}

func (t *Table) Store(key IdKey, val *Server) {
	t.table.Store(key, val)
}

func (t *Table) Delete(key IdKey) {
	t.table.Delete(key)
}

func (t *Table) Range(f func(key IdKey, val *Server) bool) {
	t.table.Range(func(k, v interface{}) bool {
		return f(k.(IdKey), v.(*Server))
	})
}

func (t *Table) LoadOrStore(key IdKey, val *Server) (*Server, bool) {
	res, loaded := t.table.LoadOrStore(key, val)
	return res.(*Server), loaded
}

func TestTable(t *testing.T) {
	table := NewTable()
	table.Store("test1", &Server{
		Id:  "test1_id",
		Val: 0,
	})
	table.Store("test2", &Server{
		Id:  "test2_id",
		Val: 0,
	})
	table.Store("test3", &Server{
		Id:  "test3_id",
		Val: 0,
	})

	go func() {
		for {
			data1, _ := table.Load("test1")
			fmt.Println(data1.Id, data1.Val)
			data1.Val = 2
			table.Store("test1", data1)
			time.Sleep(time.Second * 2)

			data2, _ := table.Load("test2")
			fmt.Println(data2.Id, data2.Val)
			data2.Val = 2
			table.Store("test2", data2)
			time.Sleep(time.Second * 2)

			data3, _ := table.Load("test3")
			fmt.Println(data3.Id, data3.Val)
			data3.Val = 2
			table.Store("test3", data3)
			time.Sleep(time.Second * 2)
		}
	}()

	for {
		count := 0
		table.Range(func(key IdKey, val *Server) bool {
			count++
			fmt.Println("++++++++wait:", key, ":", val.Val, count)
			val.Val--
			table.Store(key, val)
			time.Sleep(time.Second * 1)
			return true
		})
		time.Sleep(time.Second * 1)
	}
}
