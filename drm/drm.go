package drm

import (
	"fmt"
	"os"

	"github.com/bkleiner/krayons/drm/ioctl"
)

const (
	IOCTLBase = 'd'

	driPath = "/dev/dri"
)

type Card struct {
	File *os.File
}

func OpenCard(num uint) (*Card, error) {
	path := fmt.Sprintf("%s/card%d", driPath, num)
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	return &Card{f}, nil
}

func (c *Card) Close() error {
	return c.File.Close()
}

func (c *Card) ioctl(cmd uint32, ptr uintptr) error {
	return ioctl.Call(uintptr(c.File.Fd()), uintptr(cmd), ptr)
}
