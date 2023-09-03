package dir

import (
	"encoding/binary"
	"errors"
	"io/fs"
	"time"
)

var (
	ErrHdrLenNoEnough   = errors.New("cache file header length no enough")
	ErrEntryHdrNoEnough = errors.New("cache entry header length no enough")
	ErrDataNoEnough     = errors.New("cache data no enough")
	ErrFileTagUnmatched = errors.New("cache file tag unmatched")
)

func readBytes(b []byte, n int) ([]byte, []byte, error) {
	if len(b) < n {
		return nil, nil, ErrDataNoEnough
	}
	ret := make([]byte, n)
	copy(ret, b)
	return ret, b[n:], nil
}

// -----------------------------------------------------------------------------------------

const (
	cacheFileTag = 0x17936825
)

type cacheHdr struct {
	tag   uint32
	count uint32
}

const (
	cacheHdrLen = 8
)

func (p *cacheHdr) read(b []byte) ([]byte, error) {
	if len(b) < cacheHdrLen {
		return nil, ErrHdrLenNoEnough
	}
	p.tag = binary.LittleEndian.Uint32(b)
	if p.tag != cacheFileTag {
		return nil, ErrFileTagUnmatched
	}
	p.count = binary.LittleEndian.Uint32(b[4:])
	return b[8:], nil
}

func WriteCacheHdr(b []byte, entries int) []byte {
	binary.LittleEndian.PutUint32(b, cacheFileTag)
	binary.LittleEndian.PutUint32(b[4:], uint32(entries))
	return b[8:]
}

type entryHdr struct {
	fsize   int64  // length in bytes for regular files; system-dependent for others
	mtime   int64  // modification time in UnixMicro
	mode    uint32 // file mode bits
	nameLen uint32
}

const (
	entryHdrLen = 24
)

func (p *entryHdr) read(b []byte) ([]byte, error) {
	if len(b) < entryHdrLen {
		return nil, ErrEntryHdrNoEnough
	}
	p.fsize = int64(binary.LittleEndian.Uint64(b))
	p.mtime = int64(binary.LittleEndian.Uint64(b[8:]))
	p.mode = binary.LittleEndian.Uint32(b[16:])
	p.nameLen = binary.LittleEndian.Uint32(b[20:])
	return b[24:], nil
}

func WriteFileInfo(b []byte, fi fs.FileInfo) []byte {
	binary.LittleEndian.PutUint64(b, uint64(fi.Size()))
	binary.LittleEndian.PutUint64(b[8:], uint64(fi.ModTime().UnixMicro()))
	binary.LittleEndian.PutUint32(b[16:], uint32(fi.Mode()))
	name := fi.Name()
	binary.LittleEndian.PutUint32(b[20:], uint32(len(name)))
	copy(b[24:], name)
	return b[24+len(name):]
}

type fileInfo struct {
	d    entryHdr
	name []byte // base name of the file
}

func (p *fileInfo) read(b []byte) (avail []byte, err error) {
	b, err = p.d.read(b)
	if err != nil {
		return nil, err
	}
	p.name, avail, err = readBytes(b, int(p.d.nameLen))
	return
}

func (p *fileInfo) Name() string {
	return string(p.name)
}

func (p *fileInfo) Size() int64 {
	return p.d.fsize
}

func (p *fileInfo) Mode() fs.FileMode {
	return fs.FileMode(p.d.mode)
}

func (p *fileInfo) ModTime() time.Time {
	return time.UnixMicro(p.d.mtime)
}

func (p *fileInfo) IsDir() bool {
	return p.Mode().IsDir()
}

func (p *fileInfo) Sys() interface{} {
	return nil
}

// -----------------------------------------------------------------------------------------

func ReadFileInfos(b []byte) (fis []fs.FileInfo, err error) {
	var h cacheHdr
	if b, err = h.read(b); err != nil {
		return
	}
	fis = make([]fs.FileInfo, h.count)
	for i := range fis {
		var fi fileInfo
		if b, err = fi.read(b); err != nil {
			return
		}
		fis[i] = &fi
	}
	return
}

func SizeFileInfo(fi fs.FileInfo) int {
	return entryHdrLen + len(fi.Name())
}

func SizeFileInfos(fis []fs.FileInfo) int {
	namesLen := 0
	for _, fi := range fis {
		namesLen += len(fi.Name())
	}
	return cacheHdrLen + entryHdrLen*len(fis) + namesLen
}

// -----------------------------------------------------------------------------------------

func BytesFileInfo(fi fs.FileInfo) []byte {
	n := SizeFileInfo(fi)
	b := make([]byte, n)
	WriteFileInfo(b, fi)
	return b
}

func FileInfoFrom(b []byte) (ret fs.FileInfo, err error) {
	var fi fileInfo
	if _, err = fi.read(b); err != nil {
		return
	}
	return &fi, nil
}

// -----------------------------------------------------------------------------------------
