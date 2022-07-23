package provider

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func setup(t *testing.T) (string, string) {
	td := createTempDir(t)
	cwd, _ := os.Getwd()
	fmt.Printf("temp dir: %s\n", td)
	os.Chdir(td)
	return td, cwd
}

func tearDown(t *testing.T, td, cwd string) {
	os.RemoveAll(td)
	os.Chdir(cwd)
}

func createTempDir(t *testing.T) string {
	tmp, err := ioutil.TempDir("", "tf")
	if err != nil {
		t.Fatal(err)
	}
	return tmp
}
