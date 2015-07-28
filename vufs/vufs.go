// Copyright 2009 The go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/mbucc/vufs"
	"log"
)

var addr = flag.String("addr", ":5640", "network address")
var debug = flag.Int("debug", 0, "print debug messages")
var root = flag.String("root", "/", "root filesystem")

func main() {
	flag.Parse()
	fs := new(vufs.VuFs)
	fs.Id = "vufs"
	fs.Root = *root
	fs.Debuglevel = *debug
	fs.Upool = vufs.VuUsers
	fs.Start(fs)

	fmt.Print("vufs starting\n")
	// determined by build tags
	//extraFuncs()
	err := fs.StartNetListener("tcp", *addr)
	if err != nil {
		log.Println(err)
	}
}