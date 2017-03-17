package broadlink

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"net"
)

type Device interface{}

type BaseDevice struct {
	con   *net.UDPConn
	count uint16

	Host string
	MAC  []byte

	key []byte
	iv  []byte
	id  []byte
}

func newBaseDevice(con *net.UDPConn, host string, mac []byte) (dev *BaseDevice) {
	dev = &BaseDevice{
		Host: host,
		MAC:  mac,

		con: con,
		key: []byte{0x09, 0x76, 0x28, 0x34, 0x3f, 0xe9, 0x9e, 0x23, 0x76, 0x5c, 0x15, 0x13, 0xac, 0xcf, 0x8b, 0x02},
		iv:  []byte{0x56, 0x2e, 0x17, 0x99, 0x6d, 0x09, 0x3d, 0x28, 0xdd, 0xb3, 0xba, 0x69, 0x5a, 0x2e, 0x6f, 0x58},
		id:  []byte{0, 0, 0, 0},
	}

	return
}

func (bd *BaseDevice) newDevice(devtype uint16) (dev Device) {
	fmt.Printf("devtype:%x host:%s mac:%x\n", devtype, bd.Host, bd.MAC)

	switch devtype {
	// RM Mini
	case 0x2737:
		dev = newRM(bd)
	default:
	}

	return
}

func (bd *BaseDevice) Auth() error {
	payload := make([]byte, 0x50)
	payload[0x04] = 0x31
	payload[0x05] = 0x31
	payload[0x06] = 0x31
	payload[0x07] = 0x31
	payload[0x08] = 0x31
	payload[0x09] = 0x31
	payload[0x0a] = 0x31
	payload[0x0b] = 0x31
	payload[0x0c] = 0x31
	payload[0x0d] = 0x31
	payload[0x0e] = 0x31
	payload[0x0f] = 0x31
	payload[0x10] = 0x31
	payload[0x11] = 0x31
	payload[0x12] = 0x31
	payload[0x1e] = 0x01
	payload[0x2d] = 0x01
	payload[0x30] = 'T'
	payload[0x31] = 'e'
	payload[0x32] = 's'
	payload[0x33] = 't'
	payload[0x34] = ' '
	payload[0x35] = ' '
	payload[0x36] = '1'

	response, err := bd.SendPacket(0x65, payload)
	if err != nil {
		return err
	}

	//decode
	enc_payload := response[0x38:]
	//aes = AES.new(bytes(self.key), AES.MODE_CBC, bytes(self.iv))
	//payload = aes.decrypt(bytes(enc_payload))
	block, aes_err := aes.NewCipher(bd.key)
	if aes_err != nil {
		return aes_err
	}
	mode := cipher.NewCBCDecrypter(block, bd.iv)
	mode.CryptBlocks(payload, enc_payload)

	bd.id = payload[0x00:0x04]
	bd.key = payload[0x04:0x14]

	return nil
}

func (bd *BaseDevice) SendPacket(command byte, payload []byte) (resp []byte, err error) {
	resp = make([]byte, 1024)
	bd.count = (bd.count + 1) & 0xffff

	packet := make([]byte, 0x38)

	packet[0x00] = 0x5a
	packet[0x01] = 0xa5
	packet[0x02] = 0xaa
	packet[0x03] = 0x55
	packet[0x04] = 0x5a
	packet[0x05] = 0xa5
	packet[0x06] = 0xaa
	packet[0x07] = 0x55
	packet[0x24] = 0x2a
	packet[0x25] = 0x27
	packet[0x26] = command
	packet[0x28] = byte(bd.count & 0xff)
	packet[0x29] = byte(bd.count >> 8)
	packet[0x2a] = bd.MAC[0]
	packet[0x2b] = bd.MAC[1]
	packet[0x2c] = bd.MAC[2]
	packet[0x2d] = bd.MAC[3]
	packet[0x2e] = bd.MAC[4]
	packet[0x2f] = bd.MAC[5]
	packet[0x30] = bd.id[0]
	packet[0x31] = bd.id[1]
	packet[0x32] = bd.id[2]
	packet[0x33] = bd.id[3]

	//checksum
	checksum := getCheckSum(packet)
	packet[0x34] = byte(checksum & 0xff)
	packet[0x35] = byte(checksum >> 8)

	//aes = AES.new(bytes(self.key), AES.MODE_CBC, bytes(self.iv))
	//payload = aes.encrypt(bytes(payload))

	block, aes_err := aes.NewCipher(bd.key)
	if aes_err != nil {
		err = aes_err
		return
	}
	ciphertext := make([]byte, 16)
	mode := cipher.NewCBCEncrypter(block, bd.iv)
	mode.CryptBlocks(ciphertext, payload)
	//payload := fmt.Sprintf("%X", ciphertext)
	packet = append(packet, ciphertext...)

	checksum = getCheckSum(packet)
	packet[0x20] = byte(checksum & 0xff)
	packet[0x21] = byte(checksum >> 8)

	rudp, addr_err := net.ResolveUDPAddr("udp", bd.Host)
	if addr_err != nil {
		err = addr_err
		return
	}

	_, werr := bd.con.WriteToUDP(packet, rudp)
	if werr != nil {
		err = werr
		return
	}

	size, raddr, rerr := bd.con.ReadFrom(resp)
	if rerr != nil {
		err = rerr
		return
	}
	fmt.Printf("get %d bytes from %s\n", size, raddr.String())

	return
}

func getCheckSum(packet []byte) (checksum uint16) {
	checksum = 0xbeaf
	for _, b := range packet {
		checksum += uint16(b)
		checksum = checksum & 0xffff
	}
	return
}
