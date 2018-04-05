package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.White
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: inconsolata.Regular8x16,
		Dot:  point,
	}
	d.DrawString(label)
}

func ips() []string {
	res := make([]string, 0)

	ifaces, err := net.Interfaces()
	if err != nil {
		return res
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return res
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if !ip.IsLoopback() {
				res = append(res, ip.String())
			}
		}
	}

	return res
}

func main() {
	s, err := NewSurface(0)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	for {
		{
			now := time.Now()
			img := image.NewRGBA(image.Rect(0, 0, int(s.FB().Width), int(s.FB().Height)))

			addLabel(img, 50, 50, "hostname: "+hostname)
			addLabel(img, 50, 75, "time: "+now.Format(time.RFC3339))
			addLabel(img, 50, 100, fmt.Sprintf("resolution: %dx%d", int(s.FB().Width), int(s.FB().Height)))

			addLabel(img, 50, 125, "ips:")
			for i, ip := range ips() {
				addLabel(img, 75, 150+i*16, ip)
			}

			s.Set(img.Pix)
			fmt.Printf("draw took %s\n", time.Since(now))
		}
		{
			now := time.Now()
			s.Swap()
			fmt.Printf("swap took %s\n", time.Since(now))
		}
		time.Sleep(10 * time.Second)
	}
}
