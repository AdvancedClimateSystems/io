package gpio

import (
	"errors"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockSys struct {
	ecErr    error
	ewErr    error
	snbErr   error
	ectlbErr error

	// the Watch tests needs to be able to replace EpollWait completely
	eWaitFn func(int, []syscall.EpollEvent, int) (int, error)
}

func (m *mockSys) EpollCreate1(flag int) (fd int, err error) {
	return 15, m.ecErr
}

func (m *mockSys) EpollWait(epfd int, events []syscall.EpollEvent, msec int) (n int, err error) {
	if m.eWaitFn != nil {
		return m.eWaitFn(epfd, events, msec)
	}
	return 0, m.ewErr
}

func (m *mockSys) SetNonblock(fd int, nonblocking bool) (err error) {
	return m.snbErr
}

func (m *mockSys) EpollCtl(epfd int, op int, fd int, event *syscall.EpollEvent) (err error) {
	return m.ectlbErr
}

func TestNewWatch(t *testing.T) {
	w, err := newWatch(&mockSys{})
	assert.Nil(t, err)
	assert.Equal(t, 15, w.fd)

	w, err = newWatch(&mockSys{ecErr: errors.New("error")})
	assert.Equal(t, "Unable to create epoll FD: error", err.Error())
	assert.Nil(t, w)
}

func TestHandleEvent(t *testing.T) {
	w, _ := newWatch(&mockSys{})

	called := 0
	w.addCallback(1, func() { called++ })
	tests := []struct {
		fd      int
		called  int
		initial bool
	}{
		{1, 0, true},
		{1, 1, false},
		{1, 2, false},
		{2, 2, false},
		{6, 2, false},
	}
	for _, test := range tests {
		if cb, ok := w.callbacks[test.fd]; ok {
			assert.Equal(t, test.initial, cb.initial)
		}
		w.handleEvent(test.fd)
		assert.Equal(t, test.called, called)
	}
}

func TestAddEvent(t *testing.T) {
	w, _ := newWatch(&mockSys{})

	tests := []struct {
		fd          int
		snbErr      error
		ctlErr      error
		expected    error
		expectedLen int
	}{
		{1, nil, nil, nil, 1},
		{1, errors.New("err"), nil, errors.New("err"), 1},
		{1, nil, errors.New("err"), errors.New("err"), 1},
		{1, errors.New("err1"), errors.New("err2"), errors.New("err1"), 1},
		{2, errors.New("err"), nil, errors.New("err"), 1},
		{2, nil, errors.New("err"), errors.New("err"), 1},
		{2, errors.New("err1"), errors.New("err2"), errors.New("err1"), 1},
		{2, nil, nil, nil, 2},
	}

	for _, test := range tests {
		w.sysH = &mockSys{snbErr: test.snbErr, ectlbErr: test.ctlErr}
		err := w.AddEvent(test.fd, func() {})
		assert.Equal(t, test.expected, err)
		assert.Equal(t, test.expectedLen, len(w.callbacks))
		if err == nil {
			assert.Equal(t, true, w.callbacks[test.fd].initial)
		}
	}
}

func TestWatch(t *testing.T) {
	w, _ := newWatch(&mockSys{})

	tests := []struct {
		callbacks   int
		err         error
		expectedMax int
		expectedErr error
	}{
		{
			1, nil, 1, nil,
		},
		{
			0, nil, 1, nil,
		},
		{
			1, syscall.EAGAIN, 1, nil,
		},
		{
			0, errors.New("error"), 1, errors.New("stopping watch loop: error"),
		},
	}

	for _, test := range tests {
		for i := 0; i < test.callbacks; i++ {
			w.addCallback(i, func() {})
		}
		eWaitFn := func(epfd int, events []syscall.EpollEvent, msec int) (int, error) {
			assert.Equal(t, test.expectedMax, len(events))
			// Stopping in the waitfunc ensures the loop only loops once
			w.StopWatch()
			return 1, test.err
		}
		w.sysH = &mockSys{eWaitFn: eWaitFn}
		err := w.Watch()
		assert.Equal(t, test.expectedErr, err)
	}
}
