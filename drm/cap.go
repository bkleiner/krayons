package drm

import (
	"unsafe"

	"github.com/bkleiner/krayons/drm/ioctl"
)

const (
	CapDumbBuffer = uint64(0x1)
)

var (
	IOCTLGetCap = ioctl.NewCmd(ioctl.Read|ioctl.Write, uint16(unsafe.Sizeof(capability{})), IOCTLBase, 0x0c)
)

type capability struct {
	id  uint64
	val uint64
}

func (c *Card) GetCap(capid uint64) (uint64, error) {
	cap := &capability{
		id: capid,
	}
	if err := c.ioctl(IOCTLGetCap, uintptr(unsafe.Pointer(cap))); err != nil {
		return 0, err
	}
	return cap.val, nil
}
