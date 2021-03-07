package main

import (
	"log"

	"github.com/matopenKW/tool/modelcreater"
)

const (
	workDir       = "work/"
	modelDir      = "model/"
	repositoryDir = "repository/"
)

func init() {
}

func main() {

	ms, err := modelcreater.New()
	if err != nil {
		panic(err.Error())
	}

	for _, m := range ms {
		log.Println(m)
	}

}
