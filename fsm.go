package fsm

import (
	"reflect"
	"sync"
)

type Status = int
type Option = interface{}
type ActionFunc func(opts ...Option) (status Status, e error)

// Action 定义了状态变更的接口
type Action interface {
	// TxBuilder 出现错误时，立即返回最后执行成功的状态和错误。如果所有 ActionFunc 执行失败，则状态不会改变。
	TxBuilder() (status Status, e error)
	// AddTx 添加 ActionFunc
	AddTx(actionFunc ActionFunc, opts ...Option)
}

// FSM 实现了 Action 接口
type FSM struct {
	sync.RWMutex
	Txs     map[string]ActionFunc
	TxsOpts map[string][]Option
	Status  Status
}

// AddTx 添加 ActionFunc
func (fsm *FSM) AddTx(actionFunc ActionFunc, opts ...Option) {
	fsm.Lock()
	defer fsm.Unlock()
	if fsm.Txs == nil {
		fsm.Txs = make(map[string]ActionFunc)
	}
	if fsm.TxsOpts == nil {
		fsm.TxsOpts = make(map[string][]Option)
	}
	// 操作的函数名作为 Txs 的键
	var actionName = reflect.ValueOf(actionFunc).String()
	fsm.Txs[actionName] = actionFunc
	if opts != nil {
		for _, opt := range opts {
			fsm.TxsOpts[actionName] = append(fsm.TxsOpts[actionName], opt)
		}
	}
}

// TxBuilder 出现错误时，立即返回最后执行成功的状态和错误。如果所有 ActionFunc 执行失败，则状态不会改变。
func (fsm *FSM) TxBuilder() (status Status, e error) {
	var statuses []Status
	var n int
	statuses = append(statuses, fsm.Status)
	for action, f := range fsm.Txs {
		n += 1
		localStatus, e := f(action)
		if e != nil {
			return statuses[n-1], e
		}
		statuses = append(statuses, localStatus)
	}
	return statuses[n], e
}
