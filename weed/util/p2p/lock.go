package p2p

import (
	"sync/atomic"
)

type MicroLock struct {
	lock    chan bool
	lockNum int32
}

func NewMicroLock() MicroLock {
	return MicroLock{
		lock:    make(chan bool, 1),
		lockNum: 0,
	}
}
func (lock *MicroLock) Lock() {
	lock.lock <- true
	atomic.AddInt32(&lock.lockNum, 1)
}
func (lock *MicroLock) UnLock() {
	if lock.lockNum == 0 {
		return
	}
	<-lock.lock
	atomic.AddInt32(&lock.lockNum, -1)
}
func (lock *MicroLock) IsLock() bool {
	return lock.lockNum > 0
}
func (lock *MicroLock) Safe(handle func()) {
	lock.Lock()
	defer lock.UnLock()
	handle()
}
