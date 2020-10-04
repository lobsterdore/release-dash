package service

import (
	"io/ioutil"

	"github.com/markbates/pkger"
)

func ReadTemplateFile(name string) string {
	templateHandle, pkgerError := pkger.Open("/templates/" + name)
	if pkgerError != nil {
		panic(pkgerError)
	}
	defer templateHandle.Close()

	bytes, readError := ioutil.ReadAll(templateHandle)
	if readError != nil {
		panic(readError)
	}

	return string(bytes)
}
