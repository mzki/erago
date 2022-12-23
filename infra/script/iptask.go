package script

import (
	"sync"

	"github.com/mzki/erago/util/log"
	lua "github.com/yuin/gopher-lua"
)

const ipTaskQueueMaxLen = 128

// interpreter task runs on task queue on interpreter.
type ipTask = func() error

type ipTaskQueue struct {
	mu    *sync.Mutex
	tasks []ipTask
}

func newIpTaskQueue() *ipTaskQueue {
	return &ipTaskQueue{
		mu:    new(sync.Mutex),
		tasks: make([]ipTask, 0, ipTaskQueueMaxLen),
	}
}

func (queue *ipTaskQueue) TryAppend(task ipTask) (ret bool) {
	queue.mu.Lock()
	if len(queue.tasks) < ipTaskQueueMaxLen {
		queue.tasks = append(queue.tasks, task)
		ret = true
	} else {
		ret = false
	}
	queue.mu.Unlock()
	return
}

func (queue *ipTaskQueue) TryTakeFirst() (ret ipTask, ok bool) {
	queue.mu.Lock()
	if len(queue.tasks) > 0 {
		ret = queue.tasks[0]
		ok = true
		queue.tasks = queue.tasks[1:]
	} else {
		ret = nil
		ok = false
	}
	queue.mu.Unlock()
	return
}

// -- Interpreter API

func (ip *Interpreter) appendTask(task ipTask) {
	if ok := ip.taskQueue.TryAppend(task); !ok {
		log.Debugf("ipTaskQueue overflowed %d tasks, may drop some events", ipTaskQueueMaxLen)
	}
}

func (ip *Interpreter) takeFirstTask() (ipTask, bool) { return ip.taskQueue.TryTakeFirst() }

// create task which report fatal error and return it to terminate whole system.
func createTerminateTask(err error) ipTask {
	return ipTask(func() error {
		log.Infof("Fatal Error in ipTask: %v", err)
		return err
	})
}

func (ip *Interpreter) consumeTaskFuncCall(L *lua.LState, fn lua.LGFunction) int {
	for {
		task, ok := ip.takeFirstTask()
		if !ok {
			break
		}
		err := task()
		if err != nil {
			L.RaiseError(err.Error())
		}
	}
	return fn(L)
}

func (ip *Interpreter) wrapConsumeTaskLG(fn lua.LGFunction) lua.LGFunction {
	return lua.LGFunction(func(L *lua.LState) int {
		return ip.consumeTaskFuncCall(L, fn)
	})
}
