package gpio

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testValues struct {
	readVal  []byte
	mockErr  error
	prevPath string
}

type mockReaderWriter struct {
	// the testvalues pointer is needed to store/record the values. Storing
	// them into mockReaderWriter cant wortk, because the functions dont have
	// pointerrecievers, so the values would not actually be set as expected.
	v *testValues
}

// readFromBase reads data from a file into b.
func (m mockReaderWriter) readFromBase(b []byte, pathFromBase string) (int, error) {
	if m.v.mockErr != nil {
		return 0, m.v.mockErr
	}

	n := 0
	for i, v := range m.v.readVal {
		if i < len(b) {
			b[i] = v
			n++
		}
	}
	m.v.prevPath = pathFromBase
	return n, nil
}

// readFromBase writeFromBase writes data to a file.
func (m mockReaderWriter) writeFromBase(b []byte, pathFromBase string) error {
	m.v.prevPath = pathFromBase
	m.v.readVal = b
	if m.v.mockErr != nil {
		return m.v.mockErr
	}
	return nil
}

func TestPinImplements(t *testing.T) {
	assert.Implements(t, (*GPIO)(nil), new(Pin))
}

func TestNewPin(t *testing.T) {
	p := NewPin(1, "gpio1", new(watch))
	assert.Equal(t, []byte("1"), p.kernelIDByte)
	assert.Equal(t, 1, p.KernelID)
	assert.Equal(t, p.pinBase, "gpio1")
}

func TestDirection(t *testing.T) {
	p := NewPin(1, "gpio1", new(watch))

	tests := []struct {
		val      string
		expected Direction
		err      error
	}{
		{"out", OutDirection, nil},
		{"in\n", InDirection, nil},
		{"not-a-valid-value", OutDirection, errors.New("not a known direction: 'not'")},
		{"", OutDirection, errors.New("not enough bytes to read")},
	}
	for _, test := range tests {
		mrw := mockReaderWriter{&testValues{readVal: []byte(test.val)}}
		p.rwHelper = mrw
		dir, err := p.Direction()
		assert.Equal(t, test.err, err)
		assert.Equal(t, test.expected, dir)
		assert.Equal(t, "gpio1/direction", mrw.v.prevPath)
	}

	p.rwHelper = mockReaderWriter{&testValues{mockErr: errors.New("error")}}
	dir, err := p.Direction()
	assert.Equal(t, "error", err.Error())
	assert.Equal(t, OutDirection, dir)
}

func TestSetDirection(t *testing.T) {
	p := NewPin(1, "gpio1", new(watch))
	mrw := mockReaderWriter{&testValues{}}
	p.rwHelper = mrw

	err := p.SetDirection(InDirection)
	assert.Nil(t, err)
	assert.Equal(t, "gpio1/direction", mrw.v.prevPath)
}

func TestValue(t *testing.T) {
	p := NewPin(1, "gpio1", new(watch))

	tests := []struct {
		val      string
		expected int
		err      error
	}{
		{"1", 1, nil},
		{"0", 0, nil},
		{"not-a-valid-value", 0, errors.New("not a known value: 'n'")},
		{"", 0, errors.New("expected 1 byte, got 0")},
	}
	for _, test := range tests {
		mrw := mockReaderWriter{&testValues{readVal: []byte(test.val)}}
		p.rwHelper = mrw
		val, err := p.Value()
		assert.Equal(t, test.err, err)
		assert.Equal(t, test.expected, val)
		assert.Equal(t, "gpio1/value", mrw.v.prevPath)
	}

	p.rwHelper = mockReaderWriter{&testValues{mockErr: errors.New("error")}}
	val, err := p.Value()
	assert.Equal(t, "error", err.Error())
	assert.Equal(t, 0, val)
}

func TestSetHigh(t *testing.T) {
	p := NewPin(1, "gpio1", new(watch))
	mrw := mockReaderWriter{&testValues{}}
	p.rwHelper = mrw

	err := p.SetHigh()
	assert.Nil(t, err)
	assert.Equal(t, "gpio1/value", mrw.v.prevPath)
	assert.Equal(t, []byte("1"), mrw.v.readVal)
}

func TestSetLow(t *testing.T) {
	mrw := mockReaderWriter{&testValues{}}
	p := &Pin{
		KernelID:     1,
		kernelIDByte: []byte("1"),
		pinBase:      "gpio1",
		rwHelper:     mrw,
	}

	err := p.SetLow()
	assert.Nil(t, err)
	assert.Equal(t, "gpio1/value", mrw.v.prevPath)
	assert.Equal(t, []byte("0"), mrw.v.readVal)
}

func TestActiveLow(t *testing.T) {
	p := NewPin(1, "gpio1", new(watch))

	tests := []struct {
		val      string
		expected bool
		err      error
	}{
		{"1", true, nil},
		{"0", false, nil},
		{"not-a-valid-value", false, errors.New("not a known value: 'n'")},
		{"", false, errors.New("expected 1 byte, got 0")},
	}
	for _, test := range tests {
		mrw := mockReaderWriter{&testValues{readVal: []byte(test.val)}}
		p.rwHelper = mrw
		val, err := p.ActiveLow()
		assert.Equal(t, test.err, err)
		assert.Equal(t, test.expected, val)
		assert.Equal(t, "gpio1/active_low", mrw.v.prevPath)
	}

	p.rwHelper = mockReaderWriter{&testValues{mockErr: errors.New("error")}}
	val, err := p.ActiveLow()
	assert.Equal(t, "error", err.Error())
	assert.Equal(t, false, val)
}

func TestSetActiveLow(t *testing.T) {
	p := NewPin(1, "gpio1", new(watch))
	mrw := mockReaderWriter{&testValues{}}
	p.rwHelper = mrw

	err := p.SetActiveLow(true)
	assert.Nil(t, err)
	assert.Equal(t, "gpio1/active_low", mrw.v.prevPath)
	assert.Equal(t, []byte("1"), mrw.v.readVal)

	err = p.SetActiveLow(false)
	assert.Nil(t, err)
	assert.Equal(t, "gpio1/active_low", mrw.v.prevPath)
	assert.Equal(t, []byte("0"), mrw.v.readVal)
}

func TestEdge(t *testing.T) {
	p := NewPin(1, "gpio1", new(watch))

	tests := []struct {
		val      string
		expected Edge
		err      error
	}{
		{"rising\n", RisingEdge, nil},
		{"falling\n", FallingEdge, nil},
		{"none\n", NoneEdge, nil},
		{"both\n", BothEdge, nil},
		{"not-a-valid-value", NoneEdge, errors.New("not a known value: 'not-a-va'")},
		{"", NoneEdge, errors.New("not enough bytes to read")},
	}
	for _, test := range tests {
		mrw := mockReaderWriter{&testValues{readVal: []byte(test.val)}}
		p.rwHelper = mrw
		val, err := p.Edge()
		assert.Equal(t, test.err, err)
		assert.Equal(t, test.expected, val)
		assert.Equal(t, "gpio1/edge", mrw.v.prevPath)
	}

	p.rwHelper = mockReaderWriter{&testValues{mockErr: errors.New("error")}}
	val, err := p.Edge()
	assert.Equal(t, "error", err.Error())
	assert.Equal(t, NoneEdge, val)
}

func TestSetEdge(t *testing.T) {
	// TODO: Find a way to mock opening files
}

func TestExport(t *testing.T) {
	p := NewPin(1, "gpio1", new(watch))
	mrw := mockReaderWriter{&testValues{}}
	p.rwHelper = mrw

	tests := []struct {
		err error
		// mockErr is the error the mock is going to return
		mockErr error
	}{
		{nil, nil},
		{nil, fmt.Errorf("write %v/export: device or resource busy", basePath)},
		{errors.New("error"), errors.New("error")},
	}
	for _, test := range tests {
		mrw := mockReaderWriter{&testValues{mockErr: test.mockErr}}
		p.rwHelper = mrw

		err := p.Export()
		assert.Equal(t, test.err, err)
	}
}

func TestUnexport(t *testing.T) {
	p := NewPin(1, "gpio1", new(watch))
	mrw := mockReaderWriter{&testValues{}}
	p.rwHelper = mrw

	tests := []struct {
		err error
		// mockErr is the error the mock is going to return
		mockErr error
	}{
		{nil, nil},
		{errors.New("error"), errors.New("error")},
	}
	for _, test := range tests {
		mrw := mockReaderWriter{&testValues{mockErr: test.mockErr}}
		p.rwHelper = mrw

		err := p.Unexport()
		assert.Equal(t, test.err, err)
	}

}
