package asset

import (
	"io/ioutil"

	"github.com/markbates/pkger"
)

func GetTemplateFilePath(name string) string {
	templateHandle, pkgerError := pkger.Open("/web/templates/" + name)
	if pkgerError != nil {
		panic(pkgerError)
	}
	defer templateHandle.Close()

	return templateHandle.Path().Name
}

func ReadTemplateFile(name string) string {
	templateHandle, pkgerError := pkger.Open("/web/templates/" + name)
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
