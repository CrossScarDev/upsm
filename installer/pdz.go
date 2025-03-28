// Port of https://github.com/cranksters/playdate-reverse-engineering/blob/main/tools/pdz.py to Go

package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	PDZ_IDENT = []byte("Playdate PDZ")

	fileTypes = map[uint8]string{
		1: "luac",
		2: "pdi",
		3: "pdt",
		4: "pdv",
		5: "pda",
		6: "pds",
		7: "pft",
	}

	fileIdents = map[string][]byte{
		"pdi": []byte("Playdate IMG"),
		"pdt": []byte("Playdate IMT"),
		"pdv": []byte("Playdate VID"),
		"pda": []byte("Playdate AUD"),
		"pds": []byte("Playdate STR"),
		"pft": []byte("Playdate FNT"),
	}
)

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type Entry struct {
	Name             string
	Type             string
	Data             []byte
	Size             uint32
	Compressed       bool
	DecompressedSize uint32
	AudioRate        uint32
	AudioFormat      uint32
}

type PDZ struct {
	buffer     *os.File
	Entries    map[string]*Entry
	NumEntries int
}

func (pdz *PDZ) readHeader() error {
	pdz.buffer.Seek(0, io.SeekStart)
	magic := make([]byte, 16)
	if _, err := pdz.buffer.Read(magic); err != nil {
		return err
	}
	magic = bytes.TrimRight(magic, "\x00")
	if !bytes.Equal(magic, PDZ_IDENT) {
		return fmt.Errorf("invalid PDZ file ident")
	}

	pdz.buffer.Seek(12, io.SeekStart)
	flagsBytes := make([]byte, 4)
	if _, err := pdz.buffer.Read(flagsBytes); err != nil {
		return err
	}
	flags := binary.LittleEndian.Uint32(flagsBytes)
	if (flags & 0x40000000) != 0 {
		return fmt.Errorf("PDZ file is encrypted")
	}
	return nil
}

func readNullTerminated(r io.Reader) (string, error) {
	var buf bytes.Buffer
	b := make([]byte, 1)
	for {
		if _, err := r.Read(b); err != nil {
			return "", err
		}
		if b[0] == 0 {
			break
		}
		buf.Write(b)
	}
	return buf.String(), nil
}

func (pdz *PDZ) readEntries() error {
	fileInfo, err := pdz.buffer.Stat()
	if err != nil {
		return err
	}
	pdzLength := fileInfo.Size()
	ptr := int64(0x10)

	for ptr < pdzLength {
		pdz.buffer.Seek(ptr, io.SeekStart)
		headBytes := make([]byte, 4)
		if _, err := io.ReadFull(pdz.buffer, headBytes); err != nil {
			return err
		}
		head := binary.LittleEndian.Uint32(headBytes)
		flags := uint8(head & 0xFF)
		entryLen := (head >> 8) & 0x00FFFFFF
		isCompressed := (flags >> 7) & 1
		fileTypeCode := flags & 0x0F
		fileType, ok := fileTypes[fileTypeCode]
		if !ok {
			return fmt.Errorf("unknown file type code %d", fileTypeCode)
		}

		fileName, err := readNullTerminated(pdz.buffer)
		if err != nil {
			return err
		}

		currentOffset, err := pdz.buffer.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		newOffset := (currentOffset + 3) &^ 3
		if _, err := pdz.buffer.Seek(newOffset, io.SeekStart); err != nil {
			return err
		}

		var audioRate, audioFormat uint32
		if fileType == "pda" {
			audioInfoBytes := make([]byte, 4)
			if _, err := io.ReadFull(pdz.buffer, audioInfoBytes); err != nil {
				return err
			}
			audioInfo := binary.LittleEndian.Uint32(audioInfoBytes)
			audioRate = audioInfo & 0x00FFFFFF
			audioFormat = (audioInfo >> 24) & 0xFF
			entryLen -= 4
		}

		var decompressedSize uint32
		if isCompressed != 0 {
			decompressedSizeBytes := make([]byte, 4)
			if _, err := io.ReadFull(pdz.buffer, decompressedSizeBytes); err != nil {
				return err
			}
			decompressedSize = binary.LittleEndian.Uint32(decompressedSizeBytes)
			entryLen -= 4
		} else {
			decompressedSize = entryLen
		}

		data := make([]byte, entryLen)
		if _, err := io.ReadFull(pdz.buffer, data); err != nil {
			return err
		}

		ptr, err = pdz.buffer.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}

		entry := &Entry{
			Name:             fileName,
			Type:             fileType,
			Data:             data,
			Size:             entryLen,
			Compressed:       isCompressed != 0,
			DecompressedSize: decompressedSize,
		}
		if fileType == "pda" {
			entry.AudioRate = audioRate
			entry.AudioFormat = audioFormat
		}

		pdz.Entries[fileName] = entry
		pdz.NumEntries++
	}
	return nil
}

func (pdz *PDZ) getEntryData(name string) ([]byte, error) {
	entry, ok := pdz.Entries[name]
	if !ok {
		return nil, fmt.Errorf("entry %s not found", name)
	}
	if entry.Compressed {
		r, err := zlib.NewReader(bytes.NewReader(entry.Data))
		if err != nil {
			return nil, err
		}
		defer r.Close()
		return io.ReadAll(r)
	}
	return entry.Data, nil
}

func (pdz *PDZ) constructEntryHeader(name string) ([]byte, error) {
	entry, ok := pdz.Entries[name]
	if !ok {
		return nil, fmt.Errorf("entry %s not found", name)
	}
	fileType := entry.Type
	ident, ok := fileIdents[fileType]
	if !ok {
		return nil, fmt.Errorf("file type %s has no ident", fileType)
	}

	buf := new(bytes.Buffer)
	if fileType == "pda" {
		audioInfo := (entry.AudioFormat << 24) | entry.AudioRate
		buf.Write(ident)
		if err := binary.Write(buf, binary.LittleEndian, audioInfo); err != nil {
			return nil, err
		}
	} else {
		flags := uint32(0)
		if entry.Compressed {
			flags = 0x80000000
		}
		buf.Write(ident)
		if err := binary.Write(buf, binary.LittleEndian, flags); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (pdz *PDZ) saveEntryData(name, outDir string, genHeader bool) error {
	entry, ok := pdz.Entries[name]
	if !ok {
		return fmt.Errorf("entry %s not found", name)
	}

	data, err := pdz.getEntryData(name)
	if err != nil {
		return err
	}

	filePath := filepath.Join(outDir, entry.Name+"."+entry.Type)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if genHeader {
		if _, ok := fileIdents[entry.Type]; ok {
			header, err := pdz.constructEntryHeader(name)
			if err != nil {
				return err
			}
			if _, err := file.Write(header); err != nil {
				return err
			}
		}
	}

	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}

func (pdz *PDZ) saveEntries(outDir string, genHeader bool) error {
	for name := range pdz.Entries {
		if err := pdz.saveEntryData(name, outDir, genHeader); err != nil {
			return err
		}
	}
	return nil
}

func (pdz *PDZ) printEntries() {
	for name, entry := range pdz.Entries {
		fmt.Printf("%s: %s\n", name, entry.Type)
	}
}
