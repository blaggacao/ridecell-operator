// +build ignore

package main

import (
	"log"

	"github.com/Ridecell/ridecell-operator/pkg/controller/summon"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(summon.Templates, vfsgen.Options{
		PackageName:  "summon",
		BuildTags:    "release",
		VariableName: "Templates",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
