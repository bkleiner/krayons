package drm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"syscall"
	"unsafe"
)

const (
	EventVBlank       = 0x1
	EventPageFlip     = 0x2
	EventCrtcSequence = 0x3
)

var (
	ErrNotEnough = errors.New("drm fd did not return enough data to fit an event")
)

type sysEventHeader struct {
	Typ uint32
	Len uint32
}

func (c *Card) ReadEvent() error {
	// libdrm uses 1024, not sure if thats sane
	buf := make([]byte, 1024)

	n, err := syscall.Read(int(c.File.Fd()), buf)
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrNotEnough
	}

	header := sysEventHeader{}
	if n < int(unsafe.Sizeof(header)) {
		return ErrNotEnough
	}

	r := bytes.NewBuffer(buf[:n])
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return err
	}

	return nil
}
