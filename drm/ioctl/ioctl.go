package ioctl

import (
	"fmt"
	"syscall"
)

const (
	None  = uint8(0x0)
	Write = uint8(0x1)
	Read  = uint8(0x2)
)

func NewCmd(typ uint8, sz uint16, uniq, fn uint8) uint32 {
	var cmd uint32
	if typ > Write|Read {
		panic(fmt.Errorf("invalid ioctl cmd value: %d", typ))
	}

	if sz > 2<<14 {
		panic(fmt.Errorf("invalid ioctl size value: %d", sz))
	}

	cmd = cmd | (uint32(typ) << 30)
	cmd = cmd | (uint32(sz) << 16) // sz has 14bits
	cmd = cmd | (uint32(uniq) << 8)
	cmd = cmd | uint32(fn)

	return cmd
}

func Call(fd, cmd, ptr uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if errno != 0 {
		return errno
	}
	return nil
}
