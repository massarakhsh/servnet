package base

import (
	"github.com/massarakhsh/lik"
	"sync"
)

var register_data lik.Seter
var register_sync sync.Mutex

func InitRegister() {
	register_data = lik.BuildSet()
}

func SetRegister(path string, val interface{}) {

}
