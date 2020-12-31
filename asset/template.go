package asset

import (
	"io/ioutil"

	"github.com/markbates/pkger"
)

func GetTemplateFilePath(name string) string {
	templateHandle, err := pkger.Open("/web/templates/" + name)
	if err != nil {
		panic(err)
	}
	defer templateHandle.Close()

	return templateHandle.Path().Name
}

func ReadTemplateFile(name string) string {
	templateHandle, err := pkger.Open("/web/templates/" + name)
	if err != nil {
		panic(err)
	}
	defer templateHandle.Close()

	bytes, err := ioutil.ReadAll(templateHandle)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}
