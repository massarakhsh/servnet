package task

import (
	"github.com/massarakhsh/servnet/ruler"
	"sync"
	"time"
)

//	Дескриптор задачи
type Task struct {
	Self      Tasker
	Name      string //	Наименование задачи
	pause     time.Duration
	isStoping bool //	Признак необходима остановка
	isStoped  bool //	Признак задача остановлена
}

//	Интерфейс задачи
type Tasker interface {
	GetName() string //	Имя задачи
	OnStoping()      //	Начать остановку задачи
	IsStoping() bool //	Проверка. что задача останавливается
	OnStoped()       //	Задача остановлена
	IsStoped() bool  //	Проверка, что задача остановлена
	DoStep()         //	Шаг исполнения
}

var syncList sync.Mutex //	Семафор списка задач
var taskList []Tasker   //	Список интерфейсов задач

//	Инициализация задачи
func (it *Task) Initialize(name string, self Tasker) {
	it.Name = name
	it.Self = self
	syncList.Lock()
	taskList = append(taskList, self)
	syncList.Unlock()
	go it.run()
}

//	Получить имя задачи
func (it *Task) GetName() string {
	return it.Name
}

//	Начать остановку задачи
func (it *Task) OnStoping() {
	it.isStoping = true
}

//	Проверка, что задача останавливается
func (it *Task) IsStoping() bool {
	return it.isStoping || ruler.IsStoping()
}

//	Определить, что задача остановлена
func (it *Task) OnStoped() {
	it.isStoped = true
}

//	Проверка, что задача остановлена
func (it *Task) IsStoped() bool {
	return it.isStoped
}

//	Установка паузы
func (it *Task) SetPause(pause time.Duration) {
	it.pause = pause
}

//	Основной процесс
func (it *Task) run() {
	timeNextRequest := time.Now().Add(time.Second * 0)
	for !it.IsStoping() {
		if !time.Now().Before(timeNextRequest) {
			it.pause = time.Second * 1
			it.Self.DoStep()
			timeNextRequest = time.Now().Add(it.pause)
		}
		if dura := timeNextRequest.Sub(time.Now()); dura <= 0 {
		} else if dura < time.Second {
			time.Sleep(dura)
		} else {
			time.Sleep(time.Second)
		}
	}
	it.OnStoped()
}
