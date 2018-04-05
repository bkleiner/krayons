package mmap

import (
	"os"
	"syscall"
)

func Map(f *os.File, offset int64, size int) ([]byte, error) {
	return syscall.Mmap(int(f.Fd()), offset, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
}

func Unmap(b []byte) error {
	return syscall.Munmap(b)
}
