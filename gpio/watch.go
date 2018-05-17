// +build linux

package gpio

import (
	"fmt"
	"os"
	"sync"
	"syscall"
)

type watchCallback struct {
	initial  bool
	callback func()
}

// Watcher watches files for events and executes a callback when an event occurs.
type Watcher interface {
	Watch() error
	StopWatch()
	AddEvent(fpnt int, callback func()) error
	AddFile(file *os.File)
	Close() error
}

// watch is the implementation of Watcher. It is used to watch gpio files for
// changes and handle events assiociated with those files.
type watch struct {
	// sysH is used to make all the syscalls, this way all syscalls can be
	// mocked during tests.
	sysH syscaller
	// File descriptor for epoll
	fd        int
	callbacks map[int]*watchCallback

	// Keep a reference to the files, otherwise it might get garbage collected,
	// which causes epoll not recieveing any events.
	files []*os.File
	run   bool
	m     sync.RWMutex
}

// NewWatcher Creates a new Watcher.
func NewWatcher() (Watcher, error) {
	return newWatch(new(syscallHelper))
}

// newWatch creates a new watch with the given syscall helper. This function is
// useful for testing, because it allows the creation of Watchers with a
// custom sycall handler
func newWatch(sysH syscaller) (*watch, error) {
	// Open an epoll file descriptor.
	// EpollCreate1 does the same thing as epoll create, except with a 0 as
	// argument the obsolete size argument is dropped.
	epollFD, err := sysH.EpollCreate1(0)
	if err != nil {
		return nil, fmt.Errorf("Unable to create epoll FD: %v", err.Error())
	}

	w := &watch{
		sysH:      sysH,
		fd:        epollFD,
		callbacks: make(map[int]*watchCallback),
		m:         sync.RWMutex{},
	}
	return w, nil
}

// Watch handles incoming epoll events.
func (w *watch) Watch() error {
	w.run = true
	// maxEvents is the maximum of events handled at once.
	w.m.RLock()
	maxEvents := len(w.callbacks)
	w.m.RUnlock()

	if maxEvents == 0 {
		// At least one event must be handeld.
		maxEvents = 1
	}
	events := make([]syscall.EpollEvent, maxEvents)
	for w.run {
		// The last argument is the timeout, the timeout specifies how long the call will block,
		// Setting a timeout of -1 will make it block indefinitely.
		numEvents, err := w.sysH.EpollWait(w.fd, events, -1)
		if err != nil {
			if err == syscall.EAGAIN {
				continue
			}
			return fmt.Errorf("stopping watch loop: %v", err)
		}
		for i := 0; i < numEvents; i++ {
			go w.handleEvent(int(events[i].Fd))
		}
	}
	return nil
}

// StopWatch stops the watcher after next event.
func (w *watch) StopWatch() {
	w.run = false
}

// handleEvent runs the callback funcion if one has been registered for a file.
// The first event is ignored, because this is the event fired when this file
// is first added .
func (w *watch) handleEvent(fd int) {
	w.m.Lock()
	wcb, exists := w.callbacks[fd]
	w.m.Unlock()

	if exists {
		if !wcb.initial {
			wcb.callback()
		}
		w.m.Lock()
		wcb.initial = false
		w.m.Unlock()
	}
}

func (w *watch) addCallback(fpntr int, callback func()) {
	w.m.Lock()
	w.callbacks[fpntr] = &watchCallback{
		true,
		callback,
	}
	w.m.Unlock()
}

func (w *watch) AddFile(file *os.File) {
	w.m.Lock()
	w.files = append(w.files, file)
	w.m.Unlock()
}

func (w *watch) AddEvent(fpntr int, callback func()) error {
	var event syscall.EpollEvent
	event.Events = syscall.EPOLLIN | (syscall.EPOLLET & 0xffffffff)
	event.Fd = int32(fpntr)

	// An application that employs the EPOLLET flag should use nonblocking
	// file descriptors to avoid having a blocking read or write starve a
	// task that is handling multiple file descriptors.
	//  - http://man7.org/linux/man-pages/man7/epoll.7.html
	if err := w.sysH.SetNonblock(fpntr, true); err != nil {
		return err
	}

	if err := w.sysH.EpollCtl(w.fd, syscall.EPOLL_CTL_ADD, fpntr, &event); err != nil {
		return err
	}
	w.addCallback(fpntr, callback)
	return nil
}

func (w *watch) Close() error {
	return syscall.Close(w.fd)
}

// syscaller is an interface specifing the syscall-based functions watcher needs
// to watch. This way the syscalls can be replaced with a mock during tests.
type syscaller interface {
	EpollCreate1(flag int) (fd int, err error)
	EpollWait(epfd int, events []syscall.EpollEvent, msec int) (n int, err error)
	SetNonblock(fd int, nonblocking bool) (err error)
	EpollCtl(epfd int, op int, fd int, event *syscall.EpollEvent) (err error)
}

// syscallHelper implements the syscaller interface and should be used to make
// syscalls in watcher.
type syscallHelper struct{}

func (syscallHelper) EpollCreate1(flag int) (fd int, err error) {
	return syscall.EpollCreate1(flag)
}

func (syscallHelper) EpollWait(epfd int, events []syscall.EpollEvent, msec int) (n int, err error) {
	return syscall.EpollWait(epfd, events, msec)
}

func (syscallHelper) SetNonblock(fd int, nonblocking bool) (err error) {
	return syscall.SetNonblock(fd, nonblocking)
}

func (syscallHelper) EpollCtl(epfd int, op int, fd int, event *syscall.EpollEvent) (err error) {
	return syscall.EpollCtl(epfd, op, fd, event)
}
