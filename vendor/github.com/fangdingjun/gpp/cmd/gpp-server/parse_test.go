package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	//"net/url"
	"os"
	"testing"
)

func TestParser(t *testing.T) {

	var _cfg CFG
	buf, err := ioutil.ReadFile("config.json")
	if err != nil {
		t.Fatalf("open config.json failed: %s", err.Error())
	}

	err = json.Unmarshal(buf, &_cfg)
	if err != nil {
		t.Errorf("parser json error: %s", err)
	}

	fmt.Printf("%+v\n", _cfg)
}
