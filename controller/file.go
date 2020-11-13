package controller

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likdom"
	"github.com/massarakhsh/servnet/ruler"
	"io/ioutil"
	"os"
	"sync"
)

const dirMain = "./var/file"

type FileControl struct {
	DataControl
	Start        int
	Length       int
	Total        int
	KeySel       string
	UpSync       sync.Mutex
	IsNeedRemenu bool
}

type Filer interface {
	Controller
}

func BuildFile(rule ruler.DataRuler, level int, path []string) Filer {
	it := &FileControl{}
	it.Mode = "all"
	rule.SetControl(level, it)
	it.Execute(rule, path)
	return it
}

func (it *FileControl) ShowMenu(rule ruler.DataRuler) likdom.Domer {
	tbl := likdom.BuildTableClassId("menu", fmt.Sprintf("menu%d", it.Level))
	row := tbl.BuildTr()
	namely := "Корневая папка"
	if name := it.calcKeyName(rule, it.GetLevel()); name != "" {
		namely = name
	}
	it.MenuItemCmd(rule, row, "", namely, "seek")
	if page := rule.GetItPage(); page != nil {
		if it.GetLevel()+1 == len(page.Controls) {
			if nup := len(page.UpList); nup > 0 {
				it.MenuItemText(rule, row, "|")
				it.MenuItemCmd(rule, row, "last", fmt.Sprintf("Последние (%d)", nup), "last")
				it.MenuItemCmd(rule, row, "", "Сохранить", "store")
				it.MenuItemCmd(rule, row, "", "Забыть", "clear")
			}
			it.MenuItemText(rule, row, "|")
			it.MenuItemProc(rule, row, "", "Новая папка", it.BuildProc("create_folder", "folder", ""))
			it.MenuItemCmd(rule, row, "import", "Загрузка", "import")
		}
	}
	it.MenuItemText(rule, row, "|")
	it.MenuTools(rule, row)
	return tbl
}

func (it *FileControl) ShowInfo(rule ruler.DataRuler) likdom.Domer {
	div := likdom.BuildDivClass("grid")
	if it.Mode == "import" {
		div.AppendItem(it.ShowImport(rule))
	} else {
		row := div.BuildTable().BuildTr()
		row.BuildTdClass("top").AppendItem(it.ShowDirectory(rule))
		row.BuildTdClass("top").AppendItem(it.ShowFoto(rule))
	}
	return div
}

func (it *FileControl) Execute(rule ruler.DataRuler, path []string) {
	if cmd := ruler.PopCommand(&path); cmd == "" {
	} else if cmd == "fileinit" {
		it.execGridInit(rule)
	} else if cmd == "filedata" {
		it.execGridData(rule)
	} else if cmd == "select" {
		it.execSelect(rule)
	} else if cmd == "open" {
		it.execOpen(rule)
	} else if cmd == "last" {
		it.execLast(rule)
	} else if cmd == "import" {
		it.execImport(rule)
	} else if cmd == "folder" {
		it.execFolder(rule)
	} else if cmd == "upload" {
		it.execUpload(rule)
	} else if cmd == "store" {
		it.execStore(rule)
	} else if cmd == "clear" {
		it.execClear(rule)
	} else if cmd == "rename" {
		it.execRename(rule)
	} else if cmd == "delete" {
		it.execDelete(rule)
	} else if cmd == "seek" {
		it.Mode = ""
		it.ExecuteController(rule, cmd)
	} else {
		it.ExecuteController(rule, cmd)
	}
}

func (it *FileControl) Marshal(rule ruler.DataRuler) {
	if it.IsNeedRemenu {
		it.IsNeedRemenu = false
		rule.StoreItem(it.ShowMenu(rule))
	}
}

//func (it *FileControl) BuildProc(part string) string {
//	path := it.BuildPart(part)
//	return fmt.Sprintf("%s('%s')", "db_" + part, path)
//}

func (it *FileControl) ShowDirectory(rule ruler.DataRuler) likdom.Domer {
	_, tbl := it.ShowGrid("file")
	if row := tbl.BuildItem("thead").BuildTr(); row != nil {
		row.BuildItem("th").BuildString("№")
		row.BuildItem("th").BuildString("&nbsp")
		row.BuildItem("th").BuildString("&nbsp")
		row.BuildItem("th").BuildString("&nbsp")
		row.BuildItem("th").BuildString("Наименование")
	}
	return tbl
}

func (it *FileControl) ShowFoto(rule ruler.DataRuler) likdom.Domer {
	tbl := likdom.BuildTable("id=PhotoSel")
	path := it.calcPath(rule, it.GetLevel()) + "/" + it.KeySel
	if lik.RegExCompare(path, "(jpg|jpeg|png|gif|tif|tiff)$") {
		tbl.BuildTrTd().BuildString(path)
		if _, err := os.Stat(dirMain + path); err == nil {
			td := tbl.BuildTrTdClass("photo")
			td.BuildUnpairItem("img", fmt.Sprintf("src='%s'", dirMain+path))
		}
	}
	return tbl
}

func (it *FileControl) execGridInit(rule ruler.DataRuler) {
	grid := lik.BuildSet()
	grid.SetItem(true, "serverSide")
	grid.SetItem(true, "processing")
	grid.SetItem(it.execInitLanguage(rule), "language")
	grid.SetItem(false, "searching")
	grid.SetItem(false, "lengthChange")
	grid.SetItem("single", "select/style")
	if it.KeySel != "" {
		grid.SetItem(it.KeySel, "likSelect")
	}
	columns := lik.BuildList()
	columns.AddItemSet("data", "Num", "width=30px")
	columns.AddItemSet("data", "CO", "width=24px")
	columns.AddItemSet("data", "CR", "width=24px")
	columns.AddItemSet("data", "CD", "width=24px")
	columns.AddItemSet("data", "Name", "width=600px")
	grid.SetItem(columns, "columns")
	grid.SetItem(rule.BuildUrl("/front/"+ruler.GetIdLevel(it.Level)+"/griddata"), "ajax")
	rule.SetResponse(grid, "grid")
}

func (it *FileControl) execInitLanguage(rule ruler.DataRuler) lik.Seter {
	data := lik.BuildSet()
	data.SetItem("Поиск", "search")
	data.SetItem("Таблица пуста", "emptyTable")
	data.SetItem("Строки от _START_ до _END_, всего _TOTAL_", "info")
	data.SetItem("Загрузка ...", "fileingRecords")
	data.SetItem("Обработка ...", "processing")
	data.SetItem("Нет строк в таблице", "infoEmpty")
	data.SetItem("В начало", "paginate/first")
	data.SetItem("Назад", "paginate/previos")
	data.SetItem("Вперёд", "paginate/next")
	data.SetItem("В конец", "paginate/last")
	return data
}

func (it *FileControl) execGridData(rule ruler.DataRuler) {
	if parm := rule.GetContext("draw"); parm != "" {
		rule.SetResponse(lik.StrToInt(parm), "draw")
	}
	if parm := rule.GetContext("start"); parm != "" {
		it.Start = lik.StrToInt(parm)
	}
	if parm := rule.GetContext("length"); parm != "" {
		it.Length = lik.StrToInt(parm)
	}
	if it.Length == 0 {
		it.Length = 10
	}
	data := lik.BuildList()
	var pots []*ruler.FilePot
	if it.Mode != "last" {
		pots = it.readDir(rule, it.GetLevel())
	} else if page := rule.GetItPage(); page != nil {
		pots = page.UpList
	}
	if pots != nil {
		it.Total = len(pots)
		rule.SetResponse(it.Total, "recordsTotal")
		rule.SetResponse(it.Total, "recordsFiltered")
		for n := 0; n < it.Length && it.Start+n < it.Total; n++ {
			nr := (it.Start + n)
			pot := pots[nr]
			row := lik.BuildSet("DT_RowId", pot.Name)
			row.SetItem(nr+1, "Num")
			if pot.IsDir {
				img := likdom.BuildUnpairItem("img", "src", "/images/disk.png", "title", "Открыть")
				row.SetItem(it.LinkItemCmd(img.ToString(), "open/"+pot.Name, "cmd").ToString(), "CO")
			} else {
				row.SetItem("", "CO")
			}
			if true {
				img := likdom.BuildUnpairItem("img", "src", "/images/to_edit.gif", "title", "Переименовать")
				proc := it.BuildProc("rename_file", "rename/"+pot.Name, "'"+pot.Name+"'")
				row.SetItem(it.LinkItemProc(img.ToString(), proc, "cmd").ToString(), "CR")
			}
			if true {
				img := likdom.BuildUnpairItem("img", "src", "/images/to_delete.gif", "title", "Удалить")
				proc := it.BuildProc("delete_file", "delete/"+pot.Name, "'"+pot.Name+"'")
				row.SetItem(it.LinkItemProc(img.ToString(), proc, "cmd").ToString(), "CD")
			}
			if pot.IsDir {
				adr := it.LinkItemCmd(pot.Name, "open/"+pot.Name, "cmd")
				row.SetItem(adr.ToString(), "Name")
			} else {
				path := dirMain[1:] + it.calcPath(rule, it.GetLevel()) + "/" + pot.Name
				adr := likdom.BuildItem("a", "href", path, "target=_blank")
				adr.BuildString(pot.Name)
				row.SetItem(adr.ToString(), "Name")
			}
			row.SetItem(pot.Name, "File")
			data.AddItems(row)
		}
	}
	rule.SetResponse(data, "data")
}

func (it *FileControl) ShowImport(rule ruler.DataRuler) likdom.Domer {
	tbl := likdom.BuildTableClass("")
	td := tbl.BuildTrTd()
	url := "/front" + it.BuildUrl(rule, "upload?_mf=1")
	td.BuildItem("form", "class=dropzone", "id=mediaDropzone", "action", url)
	script := "var options = { addRemoveLinks: true };\n"
	script += "var myDropzone = new Dropzone(\"#mediaDropzone\", options);\n"
	td.BuildItem("script").BuildString("jQuery(function(){ " + script + " });")
	return tbl
}

func (it *FileControl) execSelect(rule ruler.DataRuler) {
	it.KeySel = rule.Shift()
	rule.StoreItem(it.ShowFoto(rule))
}

func (it *FileControl) execLast(rule ruler.DataRuler) {
	it.Mode = "last"
	rule.SetNeedRedraw()
}

func (it *FileControl) execOpen(rule ruler.DataRuler) {
	key := rule.Shift()
	if key == "" {
		key = it.KeySel
	}
	if key != "" {
		path := dirMain + it.calcPath(rule, it.GetLevel()) + "/" + key
		if stat, err := os.Stat(path); err == nil {
			if stat.IsDir() {
				it.Mode = "dir_" + key
				BuildFile(rule, it.Level+1, nil)
				rule.SetNeedRedraw()
			}
		}
	}
}

func (it *FileControl) execFolder(rule ruler.DataRuler) {
	name := lik.Transliterate(lik.StringFromXS(rule.Shift()))
	path := dirMain + it.calcPath(rule, it.GetLevel()) + "/" + name
	if os.MkdirAll(path, 0777) == nil {
		it.KeySel = name
	}
	rule.SetNeedRedraw()
}

func (it *FileControl) execRename(rule ruler.DataRuler) {
	if key := rule.Shift(); key != "" {
		if name := lik.Transliterate(lik.StringFromXS(rule.Shift())); name != "" {
			pathold := dirMain + it.calcPath(rule, it.GetLevel()) + "/" + key
			pathnew := dirMain + it.calcPath(rule, it.GetLevel()) + "/" + name
			os.Rename(pathold, pathnew)
			it.KeySel = name
		}
	}
	rule.SetNeedRedraw()
}

func (it *FileControl) execDelete(rule ruler.DataRuler) {
	if key := rule.Shift(); key != "" {
		path := dirMain + it.calcPath(rule, it.GetLevel()) + "/" + key
		os.RemoveAll(path)
	}
	rule.SetNeedRedraw()
}

func (it *FileControl) execUpload(rule ruler.DataRuler) {
	if buffers := rule.GetBuffers(); buffers != nil {
		it.UpSync.Lock()
		if page := rule.GetItPage(); page != nil {
			for name, val := range buffers {
				pot := &ruler.FilePot{Name: lik.Transliterate(name), Data: val}
				page.UpList = append(page.UpList, pot)
			}
		}
		it.UpSync.Unlock()
		it.IsNeedRemenu = true
	}
}

func (it *FileControl) execStore(rule ruler.DataRuler) {
	if page := rule.GetItPage(); page != nil {
		for _, pot := range page.UpList {
			if pot.Data != nil {
				file := lik.Transliterate(pot.Name)
				//if match := lik.RegExParse(pot.Name, "^(.*)\\.(\\w*)$"); match != nil {
				//	pot.Name = match[1]
				//	pot.Key += "." + strings.ToLower(match[2])
				//}
				path := dirMain + it.calcPath(rule, it.GetLevel()) + "/" + file
				_ = ioutil.WriteFile(path, pot.Data, 0666)
			}
		}
		page.UpList = []*ruler.FilePot{}
	}
	it.Mode = ""
	rule.SetNeedRedraw()
}

func (it *FileControl) execClear(rule ruler.DataRuler) {
	it.UpSync.Lock()
	if page := rule.GetItPage(); page != nil {
		page.UpList = []*ruler.FilePot{}
	}
	it.UpSync.Unlock()
	it.Mode = ""
	rule.SetNeedRedraw()
}

func (it *FileControl) execImport(rule ruler.DataRuler) {
	it.Mode = "import"
	rule.SetNeedRedraw()
}

func (it *FileControl) readDir(rule ruler.DataRuler, level int) []*ruler.FilePot {
	path := it.calcPath(rule, level)
	pots := []*ruler.FilePot{}
	if files, err := ioutil.ReadDir(dirMain + path); err == nil {
		for _, file := range files {
			if name := file.Name(); name != "" {
				pot := &ruler.FilePot{IsDir: file.IsDir(), Name: name}
				pots = append(pots, pot)
			}
		}
	}
	for np := len(pots) - 1; np >= 0; np-- {
		for np1 := 0; np1 < np; np1++ {
			pot := pots[np1]
			pot1 := pots[np1+1]
			if !pot.IsDir && pot1.IsDir ||
				pot.IsDir == pot1.IsDir && pot.Name > pot1.Name {
				pots[np1] = pot1
				pots[np1+1] = pot
			}
		}
	}
	return pots
}

func (it *FileControl) calcPath(rule ruler.DataRuler, level int) string {
	path := ""
	page := rule.GetItPage()
	for lev := level - 1; lev >= 0; lev-- {
		mode := page.Controls[lev].GetMode()
		if match := lik.RegExParse(mode, "^dir_(.+)$"); match != nil {
			path = "/" + match[1] + path
		} else {
			break
		}
	}
	return path
}

func (it *FileControl) calcKeyName(rule ruler.DataRuler, level int) string {
	name := ""
	if level > 0 {
		page := rule.GetItPage()
		mode := page.Controls[level-1].GetMode()
		if match := lik.RegExParse(mode, "^dir_(.+)$"); match != nil {
			name = match[1]
		}
	}
	return name
}
