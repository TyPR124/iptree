package iptree_test

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"

	"iptree"
)

func TestFind(t *testing.T) {
	//Default root
	tree := iptree.NewDefaultRoot(net.IPv4len, "test")

	//Root of 192.168.0.0/16, IPv4
	tree = iptree.NewRoot(net.IPNet{
		IP:   []byte{192, 168, 0, 0},
		Mask: []byte{255, 255, 0, 0},
	}, "mytree")

	//Find with allowSupernet
	v, err := tree.Find(net.IPNet{
		IP:   []byte{192, 168, 1, 1},
		Mask: []byte{255, 255, 255, 255},
	}, true)

	if v.(string) != "mytree" || err != nil {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Find without allowSupernet (expect error)
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 1, 1},
		Mask: []byte{255, 255, 255, 255},
	}, false)

	if v != nil || err != iptree.ErrNotFound {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Find with bad IP Length (expect error)
	v, err = tree.Find(net.IPNet{
		IP:   []byte{1, 2, 3, 4, 5},
		Mask: []byte{255, 255, 255, 255, 255},
	}, true)

	if v != nil || err != iptree.ErrWrongIPLength {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Insert with bad IP Length (expect error)
	err = tree.Insert(net.IPNet{
		IP:   []byte{1, 2, 3, 4, 5},
		Mask: []byte{255, 255, 255, 255, 255},
	}, "blah")

	if err != iptree.ErrWrongIPLength {
		t.Error(err)
	}

	//Remove with bad IP Length (expect error)
	err = tree.Remove(net.IPNet{
		IP:   []byte{192, 168, 0, 0, 0},
		Mask: []byte{255, 255, 0, 0, 0},
	})

	if err != iptree.ErrWrongIPLength {
		t.Error(err)
	}

	//Insert 192.168.2.0/24
	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 2, 0},
		Mask: []byte{255, 255, 255, 0},
	}, 5)

	if err != nil {
		t.Error(err)
	}

	//Find 192.168.2.1/32, allowSuper
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 2, 1},
		Mask: []byte{255, 255, 255, 255},
	}, true)

	if err != nil || v.(int) != 5 {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Find 192.168.2.0/24, no super
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 2, 0},
		Mask: []byte{255, 255, 255, 0},
	}, false)

	if err != nil || v.(int) != 5 {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Find 192.168.3.0/24, no super (expect error)
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 3, 0},
		Mask: []byte{255, 255, 255, 0},
	}, false)

	if err != iptree.ErrNotFound || v != nil {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Find 192.168.0.0/16, no super
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 0, 0},
		Mask: []byte{255, 255, 0, 0},
	}, false)

	if err != nil || v.(string) != "mytree" {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Insert 192.168.2.0/25
	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 2, 0},
		Mask: []byte{255, 255, 255, 128},
	}, "/25")

	if err != nil {
		t.Error(err)
	}

	//Find 192.168.2.255/32, no super (expect error)
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 2, 255},
		Mask: []byte{255, 255, 255, 255},
	}, false)

	if err != iptree.ErrNotFound || v != nil {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Find 192.168.2.255/32, allow super (expecting to hit 192.168.2.0/24)
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 2, 255},
		Mask: []byte{255, 255, 255, 255},
	}, true)

	if err != nil || v.(int) != 5 {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Find 192.168.2.2/32, allow super (expecting to hit 192.168.2.0/25)
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 2, 2},
		Mask: []byte{255, 255, 255, 255},
	}, true)

	if err != nil || v.(string) != "/25" {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Insert 192.168.3.0/24, 192.168.4.0/24, 192.168.5.0/24, 192.168.6.0/24
	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 3, 0},
		Mask: []byte{255, 255, 255, 0},
	}, "3.0/24")

	if err != nil {
		t.Error(err)
	}

	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 6, 0},
		Mask: []byte{255, 255, 255, 0},
	}, "6.0/24")

	if err != nil {
		t.Error(err)
	}

	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 4, 0},
		Mask: []byte{255, 255, 255, 0},
	}, "4.0/24")

	if err != nil {
		t.Error(err)
	}

	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 5, 0},
		Mask: []byte{255, 255, 255, 0},
	}, "5.0/24")

	if err != nil {
		t.Error(err)
	}

	//Insert 192.168.4.0/23
	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 4, 0},
		Mask: []byte{255, 255, 254, 0},
	}, "4.0/23")

	if err != nil {
		t.Error(err)
	}

	//Insert 192.168.6.0/23
	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 6, 0},
		Mask: []byte{255, 255, 254, 0},
	}, "6.0/23")

	if err != nil {
		t.Error(err)
	}

	//Overwrite 192.168.2.0/24
	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 2, 0},
		Mask: []byte{255, 255, 255, 0},
	}, "2.0/24")

	if err != nil {
		t.Error(err)
	}

	//Find 192.168.2.0/24, no super
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 2, 0},
		Mask: []byte{255, 255, 255, 0},
	}, false)

	if err != nil || v.(string) != "2.0/24" {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Overwrite 192.168.2.0/25
	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 2, 0},
		Mask: []byte{255, 255, 255, 128},
	}, "2.0/25")

	if err != nil {
		t.Error(err)
	}

	//Insert 192.168.0.1/32
	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 0, 1},
		Mask: []byte{255, 255, 255, 255},
	}, "0.1/32")

	if err != nil {
		t.Error(err)
	}

	//Insert 192.168.0.2/32
	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 0, 2},
		Mask: []byte{255, 255, 255, 255},
	}, "0.2/32")

	if err != nil {
		t.Error(err)
	}

	//Find 192.168.2.2/32, allow super (expecting to hit 192.168.2.0/25)
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 2, 2},
		Mask: []byte{255, 255, 255, 255},
	}, true)

	if err != nil || v.(string) != "2.0/25" {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Find 192.168.0.2/32, allow super (expecting to hit 192.168.0.2/32)
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 0, 2},
		Mask: []byte{255, 255, 255, 255},
	}, true)

	if err != nil || v.(string) != "0.2/32" {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Find 192.168.5.7/32, allow super (expecting to hit 192.168.5.0/24)
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 5, 7},
		Mask: []byte{255, 255, 255, 255},
	}, true)

	if err != nil || v.(string) != "5.0/24" {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Find 192.168.6.0/32, allow super (expecting to hit 192.168.6.0/24)
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 6, 0},
		Mask: []byte{255, 255, 255, 255},
	}, true)

	if err != nil || v.(string) != "6.0/24" {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Find 192.168.6.0/23, allow super (expecting to hit 192.168.6.0/23)
	v, err = tree.Find(net.IPNet{
		IP:   []byte{192, 168, 6, 0},
		Mask: []byte{255, 255, 254, 0},
	}, true)

	if err != nil || v.(string) != "6.0/23" {
		t.Errorf("Error: %v, v: %v", err, v)
	}

	//Remove 192.168.0.0/16 (current root, should fail)
	err = tree.Remove(net.IPNet{
		IP:   []byte{192, 168, 0, 0},
		Mask: []byte{255, 255, 0, 0},
	})

	if reflect.TypeOf(err) != reflect.TypeOf(iptree.ErrRemovedRoot{}) {
		t.Error(err)
	}

	newRoots := err.(iptree.ErrRemovedRoot).NewRoots
	if len(newRoots) != 6 {
		t.Error(newRoots)
	}
	//Not using new roots, stick with original root

	//Remove 192.168.5.0/24
	err = tree.Remove(net.IPNet{
		IP:   []byte{192, 168, 5, 0},
		Mask: []byte{255, 255, 255, 0},
	})

	if err != nil {
		t.Error(err)
	}

	//Remove 192.168.5.5/32 (expect error)
	err = tree.Remove(net.IPNet{
		IP:   []byte{192, 168, 5, 5},
		Mask: []byte{255, 255, 255, 255},
	})

	if err != iptree.ErrNotFound {
		t.Error(err)
	}

	//Remove 192.168.4.0/24
	err = tree.Remove(net.IPNet{
		IP:   []byte{192, 168, 4, 0},
		Mask: []byte{255, 255, 255, 0},
	})

	if err != nil {
		t.Error(err)
	}

	//Insert 10.0.0.0/8 (should fail)
	err = tree.Insert(net.IPNet{
		IP:   []byte{10, 0, 0, 0},
		Mask: []byte{255, 0, 0, 0},
	}, "10/8")

	if err != iptree.ErrNotFound {
		t.Error(err)
	}

	//Insert 0.0.0.0/0 (new root)
	err = tree.Insert(net.IPNet{
		IP:   []byte{0, 0, 0, 0},
		Mask: []byte{0, 0, 0, 0},
	}, "default")

	if reflect.TypeOf(err) != reflect.TypeOf(iptree.ErrNewRoot{}) {
		t.Error(err)
	}
	tree = err.(iptree.ErrNewRoot).NewRoot

	//Overwrite 192.168.0.0/16 (previous root)
	err = tree.Insert(net.IPNet{
		IP:   []byte{192, 168, 0, 0},
		Mask: []byte{255, 255, 0, 0},
	}, "0.0/16")

	if err != nil {
		t.Error(err)
	}
	tstring := ""

	err = tree.Traverse(func(ipnet net.IPNet, value interface{}, distance int) error {
		tstring += fmt.Sprintf("%v%v: %v\n", strings.Repeat(" ", distance), ipnet.String(), value)
		return nil
	})
	if err != nil {
		t.Error(err)
	}

	if tstring != "0.0.0.0/0: default\n 192.168.0.0/16: 0.0/16\n  192.168.0.1/32: 0.1/32\n  192.168.0.2/32: 0.2/32\n  192.168.2.0/24: 2.0/24\n   192.168.2.0/25: 2.0/25\n  192.168.3.0/24: 3.0/24\n  192.168.4.0/23: 4.0/23\n  192.168.6.0/23: 6.0/23\n   192.168.6.0/24: 6.0/24\n" {
		t.Error(tstring)
	}

	var sbuf bytes.Buffer
	err = iptree.Serialize(tree, &sbuf, func(v interface{}) ([]byte, error) {
		return []byte(v.(string)), nil
	})
	if err != nil {
		t.Error(err)
	}

	tree, err = iptree.Deserialize(&sbuf, func(b []byte) (interface{}, error) {
		return string(b), nil
	})
	if err != nil {
		t.Error(err)
	}

	tstring2 := ""
	err = tree.Traverse(func(ipnet net.IPNet, value interface{}, distance int) error {
		tstring2 += fmt.Sprintf("%v%v: %v\n", strings.Repeat(" ", distance), ipnet.String(), value)
		return nil
	})

	if err != nil {
		t.Error(err)
	}

	if tstring2 != tstring {
		t.Error(tstring2)
	}

}
