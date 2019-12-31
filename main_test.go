package main

import (
	"os"
	"reflect"
	"testing"
)

func TestListFiles(t *testing.T) {
	fileList := []string{"a", "b", "c"}
	err := os.Mkdir("test", 0777)
	if err != nil {
		t.Errorf(err.Error())
	}
	for _, f := range fileList {
		_, err := os.Create("test/" + f)
		if err != nil {
			t.Errorf(err.Error())
		}
	}

	files := ListFiles("test")
	if !reflect.DeepEqual(fileList, files) {
		t.Errorf("Expected %#v, got %#v", fileList, files)
	}

	err = os.RemoveAll("test")
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestHasItem(t *testing.T) {
	a := []string{"a", "b", "c"}
	patterns := []struct {
		item   string
		items  []string
		result bool
	}{
		{
			item:   "a",
			items:  a,
			result: true,
		},
	}

	for i, p := range patterns {
		actual := HasItem(p.items, p.item)
		if p.result != actual {
			t.Errorf("%d => %t", i, actual)
		}
	}
}

func TestLoadWorkspaces(t *testing.T) {
	workspaces, err := LoadWorkspaces()
	if err != nil {
		t.Errorf(err.Error())
	} else {
		for _, w := range workspaces {
			err := os.RemoveAll(w.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
		}
	}
}

func TestDownloadImage(t *testing.T) {
	DownloadImage("https://blog.golang.org/lib/godoc/images/go-logo-blue.svg", "test.svg")
	if _, err := os.Stat("test.svg"); os.IsNotExist(err) {
		t.Errorf(err.Error())
	} else {
		err := os.Remove("test.svg")
		if err != nil {
			t.Errorf(err.Error())
		}
	}
}
