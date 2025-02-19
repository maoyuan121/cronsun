package event

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

const (
	EXIT = "exit"
	WAIT = "wait"
)

var (
	// key 为事件名
	// value 为处理函数数组，签名为 func(interface{})
	Events = make(map[string][]func(interface{}), 2)
)

// 注册事件
// param: name 为事件名
// param: fs 为处理函数数组，签名为 func(interface{})
func On(name string, fs ...func(interface{})) error {
	evs, ok := Events[name]
	if !ok {
		evs = make([]func(interface{}), 0, len(fs))
	}

	for _, f := range fs {
		if f == nil {
			continue
		}

		fp := reflect.ValueOf(f).Pointer()
		for i := 0; i < len(evs); i++ {
			if reflect.ValueOf(evs[i]).Pointer() == fp {
				return fmt.Errorf("func[%v] already exists in event[%s]", fp, name)
			}
		}
		evs = append(evs, f)
	}
	Events[name] = evs
	return nil
}

// 触发事件
// param: name 事件名
// param: arg 参数
func Emit(name string, arg interface{}) {
	evs, ok := Events[name]
	if !ok {
		return
	}

	for _, f := range evs {
		f(arg)
	}
}

// 触发所有事件
func EmitAll(arg interface{}) {
	for _, fs := range Events {
		for _, f := range fs {
			f(arg)
		}
	}
	return
}

// 取消注册事件
// param: name 事件名
// param: f 处理函数
func Off(name string, f func(interface{})) error {
	evs, ok := Events[name]
	if !ok || len(evs) == 0 {
		return fmt.Errorf("envet[%s] doesn't have any funcs", name)
	}

	fp := reflect.ValueOf(f).Pointer()
	for i := 0; i < len(evs); i++ {
		if reflect.ValueOf(evs[i]).Pointer() == fp {
			evs = append(evs[:i], evs[i+1:]...)
			Events[name] = evs
			return nil
		}
	}

	return fmt.Errorf("%v func dones't exist in event[%s]", fp, name)
}

// 取消注册所有事件
func OffAll(name string) error {
	Events[name] = nil
	return nil
}

// 等待信号
// 如果信号参数为空，则会等待常见的终止信号
// SIGINT 2 A 键盘中断（如break键被按下）
// SIGTERM 15 A 终止信号
func Wait(sig ...os.Signal) os.Signal {
	c := make(chan os.Signal, 1)
	if len(sig) == 0 {
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	} else {
		signal.Notify(c, sig...)
	}
	return <-c
}
