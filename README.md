# csqlite

Self-contained cgo SQLite bindings for Go's `database/sql` — the `c` is the
C/cgo layer. This is a fork of [go-sqlite3](https://github.com/mattn/go-sqlite3)
(`package csqlite`), vendored in-tree so the Hanzo stack depends on **zero**
external go-sqlite3 module.

It carries the SQLite amalgamation (`sqlite3-binding.c/.h`), the cgo driver
(`sqlite3.go` + callback trampolines), and the optional-feature files verbatim
from upstream, rebranded to `package csqlite`. It registers the `sqlite3`
`database/sql` driver name from this repo.

## Layering

```
hanzoai/csqlite   ← forked cgo SQLite bindings (this repo)  — the mattn replacement
      ↑
hanzoai/sqlite    ← high-level wrapper: CEK/SQLCipher API, connector, dual cgo/pure-Go
      ↑
cloud · o11y · ai · base · commerce  ← consumers
```

The `hanzoai/sqlite` wrapper imports this package on its cgo path and
`modernc.org/sqlite` on its pure-Go (`!cgo`) path. Consumers import
`hanzoai/sqlite`, never this repo directly.

## Encryption (SQLCipher)

The default cgo build compiles the vendored amalgamation (plain SQLite, no
codec). At-rest AES-256 encryption is obtained by linking the **system
libsqlcipher** via the `libsqlite3` build tag — identical to upstream:

```sh
CGO_ENABLED=1 \
CGO_CFLAGS="-DSQLITE_HAS_CODEC -DSQLITE_USE_URI=1 -I/usr/include/sqlcipher" \
CGO_LDFLAGS="-lsqlcipher" \
go build -tags "libsqlite3 sqlite_fts5" ./...
```

## License

MIT (see `LICENSE`), forked from go-sqlite3 © 2014 Yasuhiro Matsumoto. The
embedded SQLite amalgamation is public domain. See `NOTICE`.
