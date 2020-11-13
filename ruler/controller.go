package ruler

import (
	"fmt"
	"github.com/massarakhsh/lik/likdom"
)

type DataControl struct {
	Level int
	Mode  string
}

type ControlExecuter interface {
	Run(rule DataPager)
}

type ControlMarshaler interface {
	Run(rule DataPager)
}

type Controller interface {
	SetLevel(lev int)
	GetLevel() int
	GetMode() string
	ShowMenu(rule DataRuler) likdom.Domer
	ShowInfo(rule DataRuler) likdom.Domer
	Marshal(rule DataRuler)
	Execute(rule DataRuler, path []string)
}

func (it *DataControl) SetLevel(lev int) {
	it.Level = lev
}

func (it *DataControl) GetLevel() int {
	return it.Level
}

func (it *DataControl) GetMode() string {
	return it.Mode
}

func PopCommand(path *[]string) string {
	if path == nil || len(*path) == 0 {
		return ""
	} else {
		cmd := (*path)[0]
		*path = (*path)[1:]
		return cmd
	}
}

func GetIdLevel(lev int) string {
	return fmt.Sprintf("c%d", lev)
}
