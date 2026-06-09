// +build ignore

package main

import (
	"log"

	//	"git.brobridge.com/pilotwave/pilotwave/web"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

// Assets contains project assets.
var Assets http.FileSystem = http.Dir("../../../web/dist")

func main() {
	err := vfsgen.Generate(Assets, vfsgen.Options{
		PackageName:  "static",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
