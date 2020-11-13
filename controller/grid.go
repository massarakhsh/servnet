package controller

import (
	"fmt"
	"github.com/massarakhsh/lik/likdom"
	"math/rand"
)

type Column struct {
	Name  string
	Title string
	Width string
}

func (it *DataControl) ShowGrid(part string) (string, likdom.Domer) {
	id := fmt.Sprintf("id_%d", 100000+rand.Int31n(900000))
	path := it.BuildPart(part)
	code := likdom.BuildTableClassId("grid", id, "path", path, "redraw=grid_redraw")
	return id, code
}
