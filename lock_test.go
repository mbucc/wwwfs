/*
   Copyright (c) 2015, Mark Bucciarelli <mkbucc@gmail.com>
*/

package vufs

import (
	"os"
	"testing"
)



func TestLock(t *testing.T) {

	rootdir := "./tmpfs"

	initfs(rootdir, "1:adm:adm\n2:mark:mark\n")
	defer os.RemoveAll(rootdir)

	conn := runserver(rootdir, ":5000")
	defer conn.Close()

	if conn == nil {
		t.Fail()
	}
	
}
