/*
   Copyright (c) 2015, Mark Bucciarelli <mkbucc@gmail.com>
*/

package vufs_test

import (
	"github.com/mbucc/vufs"
	"io/ioutil"

	"net"
	"os"
	"testing"
)

func setup_create_test(t *testing.T, fid uint32, rootdir, uid string) (*vufs.VuFs, net.Conn) {

	fs := vufs.New(rootdir)
	err := fs.Start("tcp", vufs.DEFAULTPORT)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	c, err := net.Dial("tcp", vufs.DEFAULTPORT)
	if err != nil {
		t.Errorf("connection failed: %v", err)
		return nil, nil
	}

	tx := &vufs.Fcall{
		Type:    vufs.Tversion,
		Tag:     vufs.NOTAG,
		Msize:   131072,
		Version: vufs.VERSION9P}
	err = vufs.WriteFcall(c, tx)
	if err != nil {
		t.Errorf("connection write failed: %v", err)
		return nil, nil
	}

	rx, err := vufs.ReadFcall(c)
	if err != nil {
		t.Errorf("connection read failed: %v", err)
		return nil, nil
	}
	if rx.Type != vufs.Rversion {
		t.Errorf("bad message type, expected %d got %d", vufs.Rversion, rx.Type)
		return nil, nil
	}
	if rx.Version != vufs.VERSION9P {
		t.Errorf("bad version response, expected '%s' got '%s'", vufs.VERSION9P, rx.Version)
		return nil, nil
	}

	tx = &vufs.Fcall{
		Type:  vufs.Tattach,
		Fid:   fid,
		Tag:   1,
		Afid:  vufs.NOFID,
		Uname: uid,
		Aname: "/"}
	err = vufs.WriteFcall(c, tx)
	if err != nil {
		t.Fatalf("Tattach write failed: %v", err)
	}

	rx, err = vufs.ReadFcall(c)
	if err != nil {
		t.Errorf("Rattach read failed: %v", err)
	}
	if rx.Type == vufs.Rerror {
		t.Fatalf("Tattach returned error: '%s'", rx.Ename)
	}
	if rx.Type != vufs.Rattach {
		t.Errorf("bad message type, expected %d got %d", vufs.Rattach, rx.Type)
	}
	return fs, c

}

// Can adm create a subdirectory off root?   (Yes.)
func TestCreate(t *testing.T) {

	rootdir, err := ioutil.TempDir("", "testcreate")
	if err != nil {
		t.Fatalf("TempDir failed: %v", err)
	}
	defer os.RemoveAll(rootdir)

	uid := "mark"
	fid := uint32(1)
	fs, c := setup_create_test(t, fid, rootdir, uid)
	if fs == nil || c == nil {
		return
	}
	defer fs.Stop()
	defer c.Close()

	//fs.Chatty(true)

	tx := new(vufs.Fcall)
	tx.Type = vufs.Tcreate
	tx.Fid = fid
	tx.Tag = 1
	tx.Name = "testcreate.txt"
	tx.Mode = 0
	tx.Perm = vufs.Perm(0644)

	err = vufs.WriteFcall(c, tx)
	if err != nil {
		t.Fatalf("Tcreate write failed: %v", err)
	}

	rx, err := vufs.ReadFcall(c)
	if err != nil {
		t.Errorf("Rcreate read failed: %v", err)
	}
	if rx.Type == vufs.Rerror {
		t.Fatalf("attach returned error: '%s'", rx.Ename)
	}
	if rx.Type != vufs.Rcreate {
		t.Errorf("bad message type, expected %d got %d", vufs.Rcreate, rx.Type)
	}

	// Tag must be the same
	if rx.Tag != tx.Tag {
		t.Errorf("wrong tag, expected %d got %d", tx.Tag, rx.Tag)
	}

	// Qid should be loaded
	if rx.Qid.Path == 0 {
		t.Error("Qid.Path was zero")
	}

	// Stat newly created file.
	tx = &vufs.Fcall{Type: vufs.Tstat, Fid: fid, Tag: 1}
	err = vufs.WriteFcall(c, tx)
	if err != nil {
		t.Fatalf("Tstat write failed: %v", err)
	}
	rx, err = vufs.ReadFcall(c)
	if err != nil {
		t.Errorf("Rstat read failed: %v", err)
	}
	dir, err := vufs.UnmarshalDir(rx.Stat)
	if err != nil {
		t.Fatalf("UnmarshalDir failed: %v", rx.Ename)
	}

	// User of file should be same as user passed in
	if dir.Uid != uid {
		t.Errorf("wrong user, expected '%s' got '%s'", uid, dir.Uid)
	}

	if dir.Name != "testcreate.txt" {
		t.Errorf("wrong Name, expected '%s', got '%s'", "testcreate.txt", dir.Name)
	}

	if dir.Length != 0 {
		t.Errorf("newly created empty file should have length 0")
	}

	// Group of file is from directory group.
	if dir.Gid != vufs.DEFAULT_USER {
		t.Errorf("wrong group, expected '%s' got '%s'", vufs.DEFAULT_USER, dir.Uid)
	}
}

// Create takes owner from request and group from parent directory.
// Root directory mode = 550 means no files in entire tree can be created.
// 700
// 570
// 557