package main

import (
	"fmt"
	"log"

	"github.com/bkleiner/krayons/drm"
	"github.com/bkleiner/krayons/drm/mmap"
)

type Framebuffer struct {
	ID uint32

	Width  uint32
	Height uint32

	Buffer []byte
	Stride uint32

	handle uint32
}

type Modeset struct {
	Card *drm.Card
	Mode *drm.Mode
	Conn uint32
	Crtc uint32
}

func NewModeset(minor uint) (*Modeset, error) {
	card, err := drm.OpenCard(minor)
	if err != nil {
		return nil, err
	}
	m := &Modeset{
		Card: card,
	}
	if err := m.setup(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Modeset) setup() error {
	r, err := m.Card.GetResources()
	if err != nil {
		return err
	}

	for _, c := range r.Connectors {
		conn, err := m.Card.GetConnector(c)
		if err != nil {
			return err
		}

		if conn.Connection != drm.Connected {
			log.Printf("ignoring unused (unconnected) connector %d\n", conn.ID)
			continue
		}

		if len(conn.Modes) == 0 {
			log.Printf("no valid mode for connector %d\n", conn.ID)
			continue
		}

		// TODO: find encoder if EncoderID == 0
		if conn.EncoderID == 0 {
			log.Printf("no valid encoder for connector %d\n", conn.ID)
			continue
		}

		e, err := m.Card.GetEncoder(conn.EncoderID)
		if err != nil {
			return err
		}

		m.Conn = conn.ID
		m.Crtc = e.CrtcID
		m.Mode = &conn.Modes[0]

		return nil
	}

	return fmt.Errorf("no matching mode found")
}

func (m *Modeset) CreateFB() (*Framebuffer, error) {
	fb, err := m.Card.CreateDumb(uint32(m.Mode.Hdisplay), uint32(m.Mode.Vdisplay), 32)
	if err != nil {
		return nil, err
	}

	fbID, err := m.Card.AddFB(
		uint32(m.Mode.Hdisplay),
		uint32(m.Mode.Vdisplay),
		24,
		32,
		fb.Pitch,
		fb.Handle,
	)
	if err != nil {
		return nil, err
	}

	offset, err := m.Card.MapDumb(fb.Handle)
	if err != nil {
		return nil, err
	}

	buf, err := mmap.Map(m.Card.File, int64(offset), int(fb.Size))
	if err != nil {
		return nil, err
	}

	if err := m.Card.SetCrtc(m.Crtc, fbID, 0, 0, &m.Conn, 1, m.Mode); err != nil {
		return nil, err
	}

	return &Framebuffer{
		ID:     fbID,
		Width:  uint32(m.Mode.Hdisplay),
		Height: uint32(m.Mode.Vdisplay),
		Buffer: buf,
		Stride: fb.Pitch,
		handle: fb.Handle,
	}, nil
}

func (m *Modeset) DestroyFB(f *Framebuffer) error {
	if err := mmap.Unmap(f.Buffer); err != nil {
		return err
	}

	if err := m.Card.RmFB(f.ID); err != nil {
		return err
	}

	if err := m.Card.DestroyDumb(f.handle); err != nil {
		return err
	}

	return nil
}

func (m *Modeset) WaitFlip(fb uint32) error {
	if err := m.Card.ModePageFlip(fb, m.Crtc, 0x01); err != nil {
		return err
	}
	if err := m.Card.ReadEvent(); err != nil {
		return err
	}
	return nil
}
