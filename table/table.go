package table

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likdom"
	"math/rand"
)

type Table struct {
	serverSide bool
	pageLength int
	columns    lik.Lister
}

type Tabler interface {
}

func New(opts ...interface{}) *Table {
	options := lik.BuildSet(opts...)
	it := &Table{}
	it.serverSide = options.GetBool("server")
	it.pageLength = options.GetInt("page")
	return it
}

func (it *Table) Initialize(path string) likdom.Domer {
	div := likdom.BuildDivClass("grid")
	id := fmt.Sprintf("id_%d", 100000+rand.Int31n(900000))
	code := likdom.BuildTableClassId("grid", id, "path", path, "redraw=grid_redraw")
	div.AppendItem(code)
	return div
}

func (it *Table) Show() lik.Seter {
	grid := lik.BuildSet()
	grid.SetItem(it.serverSide, "serverSide")
	grid.SetItem(it.serverSide, "processing")
	grid.SetItem(it.showLanguage(), "language")
	//grid.SetItem("400px", "scrollY")
	if it.pageLength > 0 {
		grid.SetItem(it.pageLength, "pageLength")
		grid.SetItem(true, "paging")
		//grid.SetItem(true, "scrollCollapse")
	} else {
		grid.SetItem(false, "paging")
	}
	grid.SetItem(false, "searching")
	grid.SetItem(false, "lengthChange")
	grid.SetItem("single", "select/style")
	//if it.IdSel > 0 {
	//	grid.SetItem(it.IdSel, "likSelect")
	//}
	columns := lik.BuildList()
	if it.columns != nil {
		for nc := 0; nc < it.columns.Count(); nc++ {
			if col := it.columns.GetSet(nc); col != nil {
				columns.AddItems(col)
			}
		}
	}
	grid.SetItem(columns, "columns")
	return grid
}

func (it *Table) AddColumn(opts ...interface{}) {
	if it.columns == nil {
		it.columns = lik.BuildList()
	}
	it.columns.AddItemSet(opts...)
}

func (it *Table) showLanguage() lik.Seter {
	data := lik.BuildSet()
	data.SetItem("Поиск", "search")
	data.SetItem("Таблица пуста", "emptyTable")
	data.SetItem("Строки от _START_ до _END_, всего _TOTAL_", "info")
	data.SetItem("Загрузка ...", "loadingRecords")
	data.SetItem("Обработка ...", "processing")
	data.SetItem("Нет строк в таблице", "infoEmpty")
	data.SetItem("В начало", "paginate/first")
	data.SetItem("Назад", "paginate/previos")
	data.SetItem("Вперёд", "paginate/next")
	data.SetItem("В конец", "paginate/last")
	return data
}
