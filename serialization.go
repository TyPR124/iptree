package iptree

import (
	"encoding/binary"
	"io"
	"net"
)

//smark - serialization markers
//basically, one byte before/between/after every node
const (
	smarkBegin byte = iota
	smarkDownLevel
	smarkUpLevel
	smarkSameLevel
	smarkEnd
)

func serialize(root Root, out io.Writer, serializer ValueSerializer) error {
	//Get length, write as uint16
	iplen := root.GetIPLength()

	if err := binary.Write(out, binary.BigEndian, uint16(iplen)); err != nil {
		return err
	}
	lastd := -1
	if err := root.Traverse(func(ipnet net.IPNet, value interface{}, distance int) error {
		//For all serialization, double check IP and mask lens
		if len(ipnet.IP) != iplen || len(ipnet.Mask) != iplen {
			return ErrWrongIPLength
		}
		//Determine Node type
		var bsmark byte
		if distance == 0 {
			bsmark = smarkBegin
		} else if distance == lastd {
			bsmark = smarkSameLevel
		} else if distance > lastd {
			bsmark = smarkDownLevel
		} else {
			bsmark = smarkUpLevel
		}
		lastd = distance

		//Write node type and IP+Mask
		if err := binary.Write(out, binary.BigEndian, bsmark); err != nil {
			return err
		}
		if err := binary.Write(out, binary.BigEndian, ipnet.IP); err != nil {
			return err
		}
		if err := binary.Write(out, binary.BigEndian, ipnet.Mask); err != nil {
			return err
		}

		//Value
		vbuf, err := serializer(value)
		if err != nil {
			return err
		}
		vlen := uint16(len(vbuf))
		if err := binary.Write(out, binary.BigEndian, vlen); err != nil {
			return err
		}
		return binary.Write(out, binary.BigEndian, vbuf)
	}); err != nil {
		return err
	}
	//End
	return binary.Write(out, binary.BigEndian, smarkEnd)
}

func deserialize(in io.Reader, deserializer ValueDeserializer) (Root, error) {
	//Get IP Len
	var iplen uint16
	if err := binary.Read(in, binary.BigEndian, &iplen); err != nil {
		return nil, err
	}

	var mark byte
	var vlen uint16
	var vbuf, ipbuf, maskbuf []byte
	//recent := make([]*node, 0, 10)
	//ri := 0

	if err := binary.Read(in, binary.BigEndian, &mark); err != nil {
		return nil, err
	}

	var root Root

	for mark != smarkEnd {
		//Read IP and mask
		ipbuf = make([]byte, iplen*2, iplen*2)
		maskbuf = ipbuf[iplen:]
		ipbuf = ipbuf[:iplen]
		if err := binary.Read(in, binary.BigEndian, ipbuf); err != nil {
			return nil, err
		}
		if err := binary.Read(in, binary.BigEndian, maskbuf); err != nil {
			return nil, err
		}

		//Read value length
		if err := binary.Read(in, binary.BigEndian, &vlen); err != nil {
			return nil, err
		}

		//Reset vbuf and read value bytes
		if cap(vbuf) >= int(vlen) {
			vbuf = vbuf[:vlen]
		} else {
			vbuf = make([]byte, vlen, vlen)
		}
		if err := binary.Read(in, binary.BigEndian, vbuf); err != nil {
			return nil, err
		}

		value, err := deserializer(vbuf)
		if err != nil {
			return nil, err
		}

		//Following does not work, not sure why, fallback to Root.Insert()
		//TODO: Optimize deserialization to not rely on insert

		//Determine parent for this node
		// switch mark {
		// case smarkBegin:
		// 	recent = append(recent, newNode)
		// 	ri = 0
		// case smarkDownLevel:
		// 	recent = append(recent, newNode)
		// 	recent[ri].children = append(recent[ri].children, newNode)
		// 	ri++
		// case smarkSameLevel:
		// 	recent[ri-1].children = append(recent[ri].children, newNode)
		// 	recent[ri] = newNode
		// case smarkUpLevel:
		// 	//Ensure ri is at least 2 greater than 0
		// 	if ri < 2 {
		// 		return nil, ErrInvalidData
		// 	}
		// 	recent = recent[:ri]
		// 	ri--
		// 	recent[ri-1].children = append(recent[ri-1].children, newNode)
		// 	recent[ri] = newNode

		if mark == smarkBegin {
			root = makeNode(net.IPNet{
				IP:   net.IP(ipbuf),
				Mask: net.IPMask(maskbuf),
			}, value, nil)
		} else {
			root.Insert(net.IPNet{
				IP:   net.IP(ipbuf),
				Mask: net.IPMask(maskbuf),
			}, value)
		}

		//Read next mark
		if err := binary.Read(in, binary.BigEndian, &mark); err != nil {
			return nil, err
		}
	} //for mark != smarkEnd

	return root, nil
}
