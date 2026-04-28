package metadata

import (
	"crypto/sha1"
	"fmt"
	"gustorrent/internal/bencode"
)

type Info struct {
	PieceLength int
	Pieces      []byte
	Name        string
	Announce 	string // url to the tracker

	// optional
	Private *int // being a pointer meaning it could be nil

	// single-file
	Length *int
	Md5sum *string

	// multi-file
	Files []FileInfo
}

type FileInfo struct {
	Length int
	Md5sum *string
	Path   []string
}

func GetMetadata(data []byte) (Info, [20]byte, error) {
	rootVal, _, err := bencode.Decode(data, 0)
	if err != nil {
		return Info{}, [20]byte{}, err
	}

	root := rootVal.Dict
	announceVal, ok := root["announce"]
	if !ok {
		return Info{}, [20]byte{}, fmt.Errorf("missing announce")
	}

	infoVal, ok := root["info"]
	if !ok {
		return Info{}, [20]byte{}, fmt.Errorf("missing info dict")
	}

	info, err := parseInfo(infoVal)
	if err != nil {
		return Info{}, [20]byte{}, err
	}

	infoHash := infoHash(data, infoVal)

	info.Announce = string(announceVal.Str)
	
	return info, infoHash, nil
}

func parseInfo(v bencode.Value) (Info, error) {
	m := v.Dict

	info := Info{}

	// required
	PieceLengthVal, ok := m["piece length"]
	if !ok {
		return Info{}, fmt.Errorf("missing length")
	}
	PiecesVal, ok := m["pieces"]
	if !ok {
		return Info{}, fmt.Errorf("missing pieces")
	}
	NameVal, ok := m["name"]
	if !ok {
		return Info{}, fmt.Errorf("missing name")
	}
	info.PieceLength = PieceLengthVal.Int
	info.Pieces = PiecesVal.Str
	info.Name = string(NameVal.Str)

	// validate pieces
	if len(info.Pieces)%20 != 0 {
		return Info{}, fmt.Errorf("invalid pieces length")
	}

	// optional: private
	if val, ok := m["private"]; ok {
		x := val.Int
		info.Private = &x
	}

	// mode detection
	if filesVal, ok := m["files"]; ok {
		files := parseFiles(filesVal)
		info.Files = files

	} else if lengthVal, ok := m["length"]; ok {
		x := lengthVal.Int
		info.Length = &x

		if md5, ok := m["md5sum"]; ok {
			s := string(md5.Str)
			info.Md5sum = &s
		}

	} else {
		return Info{}, fmt.Errorf("invalid: neither files nor length")
	}

	return info, nil
}

func parseFiles(v bencode.Value) []FileInfo {
	var result []FileInfo

	for _, item := range v.List {
		m := item.Dict

		lengthVal := m["length"]

		pathVal := m["path"]

		f := FileInfo{
			Length: lengthVal.Int,
			Path:   toStringSlice(pathVal.List),
		}

		if md5, ok := m["md5sum"]; ok {
			s := string(md5.Str)
			f.Md5sum = &s
		}

		result = append(result, f)
	}

	return result
}

func toStringSlice(list []bencode.Value) []string {
	res := make([]string, len(list))
	for i, v := range list {
		res[i] = string(v.Str)
	}
	return res
}

func infoHash(data []byte, infoVal bencode.Value) [20]byte {
	raw := data[infoVal.Start:infoVal.End]
	return sha1.Sum(raw)
}

// [test] Used specially for benchmarking, don't use for production or anything else
func getMetadataNoHash(data []byte) (Info, error) {
    rootVal, _, err := bencode.Decode(data, 0)
    if err != nil {
        return Info{}, err
    }

    infoVal := rootVal.Dict["info"]

    info, err := parseInfo(infoVal)
    if err != nil {
        return Info{}, err
    }

    return info, nil
}