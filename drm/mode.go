package drm

import (
	"bytes"
	"unsafe"

	"github.com/bkleiner/krayons/drm/ioctl"
)

const (
	NameLen = 32

	Connected         = 1
	Disconnected      = 2
	UnknownConnection = 3
)

type sysName [NameLen]uint8

func (n sysName) String() string {
	return string(n[:bytes.IndexByte(n[:], 0)])
}

type sysResources struct {
	fbIdPtr              uintptr
	crtcIdPtr            uintptr
	connectorIdPtr       uintptr
	encoderIdPtr         uintptr
	CountFbs             uint32
	CountCrtcs           uint32
	CountConnectors      uint32
	CountEncoders        uint32
	MinWidth, MaxWidth   uint32
	MinHeight, MaxHeight uint32
}

type Resources struct {
	sysResources

	Fbs        []uint32
	Crtcs      []uint32
	Connectors []uint32
	Encoders   []uint32
}

type Mode struct {
	Clock                                         uint32
	Hdisplay, HsyncStart, HsyncEnd, Htotal, Hskew uint16
	Vdisplay, VsyncStart, VsyncEnd, Vtotal, Vscan uint16

	Vrefresh uint32

	Flags uint32
	Type  uint32
	Name  sysName
}

type sysGetConnector struct {
	encodersPtr   uintptr
	modesPtr      uintptr
	propsPtr      uintptr
	propValuesPtr uintptr

	countModes    uint32
	countProps    uint32
	countEncoders uint32

	EncoderID uint32
	ID        uint32
	Type      uint32
	TypeID    uint32

	Connection        uint32
	WidthMM, HeightMM uint32 // HxW in millimeters
	Subpixel          uint32
}

type Connector struct {
	sysGetConnector

	Modes []Mode

	Props      []uint32
	PropValues []uint64

	Encoders []uint32
}

type Encoder struct {
	ID   uint32
	Type uint32

	CrtcID uint32

	PossibleCrtcs  uint32
	PossibleClones uint32
}

type FB struct {
	Height, Width uint32
	BPP           uint32
	Flags         uint32

	Handle uint32
	Pitch  uint32
	Size   uint64
}

type sysFBCmd struct {
	fbID          uint32
	width, height uint32
	pitch         uint32
	bpp           uint32
	depth         uint32

	/* driver specific handle */
	handle uint32
}

type sysMapDumb struct {
	handle uint32 // Handle for the object being mapped
	pad    uint32

	// Fake offset to use for subsequent mmap call
	// This is a fixed-size type for 32/64 compatibility.
	offset uint64
}

type sysDestroyDumb struct {
	handle uint32
}

type sysRmFB struct {
	handle uint32
}

type sysCrtc struct {
	setConnectorsPtr uintptr
	countConnectors  uint32

	id   uint32
	fbID uint32 // Id of framebuffer

	x, y uint32 // Position on the frameuffer

	gammaSize uint32
	modeValid uint32
	mode      Mode
}

type sysPageFlip struct {
	crtcID   uint32
	fbID     uint32
	flags    uint32
	reserved uint32
	user     uint64
}

type Crtc struct {
	ID       uint32
	BufferID uint32 // FB id to connect to 0 = disconnect

	X, Y          uint32 // Position on the framebuffer
	Width, Height uint32
	ModeValid     int
	Mode          Mode

	GammaSize int // Number of gamma stops
}

var (
	// DRM_IOWR(0xA0, struct drm_mode_card_res)
	IOCTLModeResources = ioctl.NewCmd(ioctl.Read|ioctl.Write, uint16(unsafe.Sizeof(sysResources{})), IOCTLBase, 0xA0)

	// DRM_IOWR(0xA1, struct drm_mode_crtc)
	IOCTLModeGetCrtc = ioctl.NewCmd(ioctl.Read|ioctl.Write, uint16(unsafe.Sizeof(sysCrtc{})), IOCTLBase, 0xA1)

	// DRM_IOWR(0xA2, struct drm_mode_crtc)
	IOCTLModeSetCrtc = ioctl.NewCmd(ioctl.Read|ioctl.Write, uint16(unsafe.Sizeof(sysCrtc{})), IOCTLBase, 0xA2)

	// DRM_IOWR(0xA6, struct drm_mode_get_encoder)
	IOCTLModeGetEncoder = ioctl.NewCmd(ioctl.Read|ioctl.Write, uint16(unsafe.Sizeof(Encoder{})), IOCTLBase, 0xA6)

	// DRM_IOWR(0xA7, struct drm_mode_get_connector)
	IOCTLModeGetConnector = ioctl.NewCmd(ioctl.Read|ioctl.Write, uint16(unsafe.Sizeof(sysGetConnector{})), IOCTLBase, 0xA7)

	// DRM_IOWR(0xAE, struct drm_mode_fb_cmd)
	IOCTLModeAddFB = ioctl.NewCmd(ioctl.Read|ioctl.Write, uint16(unsafe.Sizeof(sysFBCmd{})), IOCTLBase, 0xAE)

	// DRM_IOWR(0xAF, unsigned int)
	IOCTLModeRmFB = ioctl.NewCmd(ioctl.Read|ioctl.Write, uint16(unsafe.Sizeof(uint32(0))), IOCTLBase, 0xAF)

	// DRM_IOWR(0xB2, struct drm_mode_create_dumb)
	IOCTLModeCreateDumb = ioctl.NewCmd(ioctl.Read|ioctl.Write, uint16(unsafe.Sizeof(FB{})), IOCTLBase, 0xB2)

	// DRM_IOWR(0xB3, struct drm_mode_map_dumb)
	IOCTLModeMapDumb = ioctl.NewCmd(ioctl.Read|ioctl.Write, uint16(unsafe.Sizeof(sysMapDumb{})), IOCTLBase, 0xB3)

	// DRM_IOWR(0xB4, struct drm_mode_destroy_dumb)
	IOCTLModeDestroyDumb = ioctl.NewCmd(ioctl.Read|ioctl.Write, uint16(unsafe.Sizeof(sysDestroyDumb{})), IOCTLBase, 0xB4)

	// DRM_IOWR(0xB0, struct drm_mode_crtc_page_flip)
	IOCTLModeCrtcPageFlip = ioctl.NewCmd(ioctl.Read|ioctl.Write, uint16(unsafe.Sizeof(sysPageFlip{})), IOCTLBase, 0xB0)
)

func (c *Card) GetResources() (*Resources, error) {
	res := &Resources{}
	err := c.ioctl(IOCTLModeResources, uintptr(unsafe.Pointer(&res.sysResources)))
	if err != nil {
		return nil, err
	}

	if res.CountFbs > 0 {
		res.Fbs = make([]uint32, res.CountFbs)
		res.fbIdPtr = uintptr(unsafe.Pointer(&res.Fbs[0]))
	}
	if res.CountCrtcs > 0 {
		res.Crtcs = make([]uint32, res.CountCrtcs)
		res.crtcIdPtr = uintptr(unsafe.Pointer(&res.Crtcs[0]))
	}
	if res.CountEncoders > 0 {
		res.Encoders = make([]uint32, res.CountEncoders)
		res.encoderIdPtr = uintptr(unsafe.Pointer(&res.Encoders[0]))
	}
	if res.CountConnectors > 0 {
		res.Connectors = make([]uint32, res.CountConnectors)
		res.connectorIdPtr = uintptr(unsafe.Pointer(&res.Connectors[0]))
	}

	err = c.ioctl(IOCTLModeResources, uintptr(unsafe.Pointer(&res.sysResources)))
	if err != nil {
		return nil, err
	}
	//TODO: check for changes between ioctl (hotplugging, etc)
	return res, nil
}

func (c *Card) GetConnector(id uint32) (*Connector, error) {
	res := &Connector{}
	res.ID = id
	err := c.ioctl(IOCTLModeGetConnector, uintptr(unsafe.Pointer(&res.sysGetConnector)))
	if err != nil {
		return nil, err
	}

	if res.countProps > 0 {
		res.Props = make([]uint32, res.countProps)
		res.propsPtr = uintptr(unsafe.Pointer(&res.Props[0]))

		res.PropValues = make([]uint64, res.countProps)
		res.propValuesPtr = uintptr(unsafe.Pointer(&res.PropValues))
	}

	if res.countEncoders > 0 {
		res.Encoders = make([]uint32, res.countEncoders)
		res.encodersPtr = uintptr(unsafe.Pointer(&res.Encoders[0]))
	}

	if res.countModes == 0 {
		res.countModes = 1
	}

	res.Modes = make([]Mode, res.countModes)
	res.modesPtr = uintptr(unsafe.Pointer(&res.Modes[0]))

	err = c.ioctl(IOCTLModeGetConnector, uintptr(unsafe.Pointer(&res.sysGetConnector)))
	if err != nil {
		return nil, err
	}
	//TODO: check for changes between ioctl (hotplugging, etc)
	return res, nil
}

func (c *Card) GetEncoder(id uint32) (*Encoder, error) {
	res := &Encoder{
		ID: id,
	}

	err := c.ioctl(IOCTLModeGetEncoder, uintptr(unsafe.Pointer(res)))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Card) CreateDumb(width, height uint32, bpp uint32) (*FB, error) {
	res := &FB{
		Width:  width,
		Height: height,
		BPP:    bpp,
	}
	err := c.ioctl(IOCTLModeCreateDumb, uintptr(unsafe.Pointer(res)))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Card) AddFB(width, height uint32, depth, bpp uint32, pitch, boHandle uint32) (uint32, error) {
	f := &sysFBCmd{}
	f.width = width
	f.height = height
	f.pitch = pitch
	f.bpp = bpp
	f.depth = depth
	f.handle = boHandle
	err := c.ioctl(IOCTLModeAddFB, uintptr(unsafe.Pointer(f)))
	if err != nil {
		return 0, err
	}
	return f.fbID, nil
}

func (c *Card) RmFB(bufferid uint32) error {
	return c.ioctl(IOCTLModeRmFB, uintptr(unsafe.Pointer(&sysRmFB{bufferid})))
}

func (c *Card) MapDumb(boHandle uint32) (uint64, error) {
	mreq := &sysMapDumb{}
	mreq.handle = boHandle
	err := c.ioctl(IOCTLModeMapDumb, uintptr(unsafe.Pointer(mreq)))
	if err != nil {
		return 0, err
	}
	return mreq.offset, nil
}

func (c *Card) DestroyDumb(handle uint32) error {
	return c.ioctl(IOCTLModeDestroyDumb, uintptr(unsafe.Pointer(&sysDestroyDumb{handle})))
}

func (c *Card) GetCrtc(id uint32) (*Crtc, error) {
	crtc := &sysCrtc{}
	crtc.id = id
	err := c.ioctl(IOCTLModeGetCrtc, uintptr(unsafe.Pointer(crtc)))
	if err != nil {
		return nil, err
	}
	ret := &Crtc{
		ID:        crtc.id,
		X:         crtc.x,
		Y:         crtc.y,
		ModeValid: int(crtc.modeValid),
		BufferID:  crtc.fbID,
		GammaSize: int(crtc.gammaSize),
	}

	ret.Mode = crtc.mode
	ret.Width = uint32(crtc.mode.Hdisplay)
	ret.Height = uint32(crtc.mode.Vdisplay)
	return ret, nil
}

func (c *Card) SetCrtc(crtcid, fbID, x, y uint32, connectors *uint32, count int, mode *Mode) error {
	crtc := &sysCrtc{}
	crtc.x = x
	crtc.y = y
	crtc.id = crtcid
	crtc.fbID = fbID
	if connectors != nil {
		crtc.setConnectorsPtr = uintptr(unsafe.Pointer(connectors))
	}
	crtc.countConnectors = uint32(count)
	if mode != nil {
		crtc.mode = *mode
		crtc.modeValid = 1
	}
	return c.ioctl(IOCTLModeSetCrtc, uintptr(unsafe.Pointer(crtc)))
}

func (c *Card) ModePageFlip(fbID, crtcID, flags uint32) error {
	flip := &sysPageFlip{
		fbID:   fbID,
		crtcID: crtcID,
		flags:  flags,
	}
	return c.ioctl(IOCTLModeCrtcPageFlip, uintptr(unsafe.Pointer(flip)))
}
