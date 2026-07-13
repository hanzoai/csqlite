// Copyright (C) 2025 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
// Copyright (C) 2025 Jakob Borg <jakob@kastelo.net>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build sqlite_dbstat
// +build sqlite_dbstat

package csqlite

/*
#cgo CFLAGS: -DSQLITE_ENABLE_DBSTAT_VTAB
*/
import "C"
