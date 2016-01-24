package common

import (
	"sync"
	"time"
)

type NetTable struct {
	BaseService
	localNode *LocalNode
	waitGroup sync.WaitGroup
	dhtInChan chan string

	lock      sync.RWMutex
	blackList map[string]time.Time
}

func (nt NetTable) Name() string {
	return "net-table"
}

func (nt *NetTable) Init(ln *LocalNode) error {
	nt.localNode = ln
	nt.dhtInChan = make(chan string, 10)
	nt.blackList = make(map[string]time.Time)
	return nil
}

func (nt *NetTable) Run() error {
	for i := 0; i < 10; i++ {
		nt.waitGroup.Add(1)
		go nt.processDHTIn()
	}

	/*for i := 0; i < 10; i++ {
		nt.waitGroup.Add(1)
		go nt.processUnTrusted()
	}*/

	nt.waitGroup.Wait()
	return nil
}

func (nt *NetTable) Stop() {
	defer close(nt.dhtInChan)
	nt.SetStatus(StatusStopping)
}

func (nt *NetTable) GetDHTInChannel() chan<- string {
	return nt.dhtInChan
}

func (nt *NetTable) processDHTIn() {
	defer nt.waitGroup.Done()

	for nt.Status() != StatusStopping {
		select {
		case host, ok := <-nt.dhtInChan:
			if !ok {
				return
			}
			nt.lock.Lock()
			_, ok := nt.blackList[host]
			nt.lock.Unlock()

			if !ok {
				nt.tryConnect(host)
			}
		}
	}
}

func (nt *NetTable) tryConnect(h string) {
	rn, err := TryConnect(h, nt.localNode.NetworkKey())
	if err != nil {
		nt.addToBlackList(h)
		return
	}
	if rn == nil {
		return
	}
}

func (nt *NetTable) addToBlackList(h string) {
	nt.lock.Lock()
	defer nt.lock.Unlock()
	nt.blackList[h] = time.Now()
}
