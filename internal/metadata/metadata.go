package metadata

import (
	"fmt"
	"gustorrent/internal/bencode"
)

type Info struct {
	PieceLength int
	Pieces      []byte
	Name        string

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

func GetMetadata(data []byte) (Info, error) {
	pos := 0 // only for recursive for decode
	rootVal, _, err := bencode.Decode(data, pos)
	if err != nil {
		return Info{}, err
	}

	root := rootVal.Dict

	infoVal, ok := root["info"]
	if !ok {
		return Info{}, fmt.Errorf("missing info dict")
	}

	return parseInfo(infoVal)
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
