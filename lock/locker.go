package lock

type Locker interface {
	Lock()
	Unlock()
}
