package broadlink

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Manager struct {
	con *net.UDPConn
	Rm  *Device
}

type RmDevice struct {
	*Device
	HeaterCoolers map[string]*HeaterCooler
}

type Device struct {
	Name string
	Host string
	MAC  string
}

type HeaterCooler struct {
	*Device
	Data struct {
		Active    []byte
		DisActive []byte
	}
}

func Discover(timeout time.Duration) (devs []*Device, err error) {
	devs = make([]*Device, 0)
	mc_addr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 80,
	}

	var udpcon *net.UDPConn
	udpcon, err = net.ListenUDP("udp", nil)
	if err != nil {
		return
	}

	//ip and port
	ip_port := strings.SplitN(GetLocalAddr().String(), ":", 2)
	localip := ip_port[0]
	source_port, _ := strconv.Atoi(ip_port[1])
	address := strings.Split(localip, ".")
	ip_1, _ := strconv.Atoi(address[0])
	ip_2, _ := strconv.Atoi(address[1])
	ip_3, _ := strconv.Atoi(address[2])
	ip_4, _ := strconv.Atoi(address[3])

	//time
	starttime := time.Now()
	_, timezone := starttime.Zone()
	timezone = timezone / 3600
	year := starttime.Year()

	//packet
	packet := [0x30]byte{}
	if timezone < 0 {
		packet[0x08] = byte(0xff + timezone - 1)
		packet[0x09] = 0xff
		packet[0x0a] = 0xff
		packet[0x0b] = 0xff
	} else {
		fmt.Println(timezone)
		packet[0x08] = byte(timezone)
		packet[0x09] = 0
		packet[0x0a] = 0
		packet[0x0b] = 0
	}

	packet[0x0c] = byte(year & 0xff)
	packet[0x0d] = byte(year >> 8)
	packet[0x0e] = byte(starttime.Minute())
	packet[0x0f] = byte(starttime.Hour())
	packet[0x10] = byte(year % 1000)
	packet[0x11] = byte(starttime.Weekday())
	packet[0x12] = byte(starttime.Day())
	packet[0x13] = byte(starttime.Month())
	packet[0x18] = byte(ip_1)
	packet[0x19] = byte(ip_2)
	packet[0x1a] = byte(ip_3)
	packet[0x1b] = byte(ip_4)
	packet[0x1c] = byte(source_port & 0xff)
	packet[0x1d] = byte(source_port >> 8)
	packet[0x26] = 6

	//checksum
	checksum := 0xbeaf
	for _, b := range packet {
		checksum += int(b)
	}
	checksum = checksum & 0xffff
	packet[0x20] = byte(checksum & 0xff)
	packet[0x21] = byte(checksum >> 8)

	fmt.Printf("packet: %v\n", packet)

	//udpcon.SetDeadline(starttime.Add(timeout))
	udpcon.WriteTo(packet[:], mc_addr)

	//read
	udpcon.SetReadDeadline(time.Now().Add(timeout))
	resp := make([]byte, 1024)

	var size int
	var raddr net.Addr
	size, raddr, err = udpcon.ReadFrom(resp)
	if err != nil {
		return
	}

	fmt.Printf("get %d bytes from %s\n", size, raddr.String())
	if size > 0 {
		fmt.Printf("%v", resp)
	} else {
		err = errors.New("can't read anything!")
	}
	return
}

func GetLocalAddr() net.Addr {
	udpcon, err := net.DialUDP("udp",
		nil,
		&net.UDPAddr{
			IP:   net.ParseIP("8.8.8.8"),
			Port: 53,
		})
	if err != nil {
		panic(err)
	}
	defer udpcon.Close()

	return udpcon.LocalAddr()
}
