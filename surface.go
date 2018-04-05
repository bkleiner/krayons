package main

import (
	"image"
	"image/color"
	"unsafe"
)

type Surface struct {
	m           *Modeset
	frontbuffer int
	fbs         []*Framebuffer
}

func NewSurface(minor uint) (*Surface, error) {
	m, err := NewModeset(minor)
	if err != nil {
		return nil, err
	}

	fbs := make([]*Framebuffer, 2)

	fbs[0], err = m.CreateFB()
	if err != nil {
		return nil, err
	}

	fbs[1], err = m.CreateFB()
	if err != nil {
		return nil, err
	}

	return &Surface{
		m:           m,
		frontbuffer: 1,
		fbs:         fbs,
	}, nil
}

func (s *Surface) Close() error {
	if err := s.m.DestroyFB(s.fbs[0]); err != nil {
		return err
	}
	if err := s.m.DestroyFB(s.fbs[1]); err != nil {
		return err
	}
	return nil
}

func (s *Surface) FB() *Framebuffer {
	return s.fbs[s.frontbuffer^1]
}

func (s *Surface) Clear(col color.Color) {
	s.Set(make([]byte, len(s.FB().Buffer)))
}

func (s *Surface) Set(buf []byte) {
	fb := s.FB()
	copy(fb.Buffer, buf)
}

func (s *Surface) Rect(rect image.Rectangle, col color.Color) {
	r, g, b, _ := col.RGBA()
	val := (r << 16) | (g << 8) | b
	fb := s.FB()

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			offset := int(fb.Stride)*y + x*4
			*(*uint32)(unsafe.Pointer(&fb.Buffer[offset])) = val
		}
	}
}

func (s *Surface) Swap() error {
	if err := s.m.WaitFlip(s.fbs[s.frontbuffer^1].ID); err != nil {
		return err
	}
	s.frontbuffer = s.frontbuffer ^ 1
	return nil
}
