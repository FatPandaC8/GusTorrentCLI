package metadata

import (
	"testing"
)

func TestGetMetadata_SingleFile_Minimal(t *testing.T) {
	data := []byte(
		"d8:announce55:http://bittorrent-test-tracker.codecrafters.io/announce" +
		"4:infod" +
		"12:piece lengthi4e" +
		"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
		"4:name8:file.txt" +
		"6:lengthi100e" +
		"ee",
	)

	info, _, err := GetMetadata(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Length == nil || *info.Length != 100 {
		t.Fatalf("expected length 100")
	}

	if info.Name != "file.txt" {
		t.Fatalf("expected name file.txt")
	}
}

func TestGetMetadata_MissingInfo(t *testing.T) {
	data := []byte(
		"d8:announce11:http://test" + 
		"4:name8:file.txte",
	)

	_, _, err := GetMetadata(data)
	if err == nil {
		t.Fatal("expected error for missing info")
	}
}

func TestGetMetadata_MissingName(t *testing.T) {
	data := []byte(
		"d8:announce11:http://test" + 
		"4:infod" +
			"12:piece lengthi4e" +
			"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
			"6:lengthi200e" +
			"ee",
	)

	_, _, err := GetMetadata(data)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestGetMetadata_MissingPieceLength(t *testing.T) {
	data := []byte(
		"d8:announce11:http://test" +
			"4:infod" +
			"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
			"4:name8:file.txt" +
			"6:lengthi200e" +
			"ee",
	)

	_, _, err := GetMetadata(data)
	if err == nil {
		t.Fatal("expected error for missing piece length")
	}
}

func TestGetMetadata_MissingPieces(t *testing.T) {
	data := []byte(
		"d" +
			"8:announce11:http://test" +
			"4:infod" +
				"12:piece lengthi4e" +
				"4:name8:file.txt" +
				"6:lengthi200e" +
			"ee",
	)

	_, _, err := GetMetadata(data)
	if err == nil || err.Error() != "missing pieces" {
		t.Fatalf("expected missing pieces error, got %v", err)
	}
}

func TestGetMetadata_PieceLengthNotMod20(t *testing.T) {
	data := []byte(
		"d8:announce11:http://test" +
			"4:infod" +
			"12:piece lengthi4e" + 
			"6:pieces21:aaaaaaaaaaaaaaaaaaaaa" +
			"4:name8:file.txt" +
			"6:lengthi200e" +
			"ee",
	)

	_, _, err := GetMetadata(data)
	if err == nil {
		t.Fatal("expected error for missing pieces")
	}
}

func TestGetMetadata_NotFileNorLength(t *testing.T) {
	data := []byte(
		"d8:announce11:http://test" +
			"4:infod" +
			"12:piece lengthi4e" + 
			"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
			"4:name8:file.txt" +
			"ee",
	)

	_, _, err := GetMetadata(data)
	if err == nil {
		t.Fatal("expected error for not file or length")
	}
}

func TestParseFiles_CoversLengthAndPathOK(t *testing.T) {
	data := []byte(
		"d8:announce55:http://bittorrent-test-tracker.codecrafters.io/announce" + 
		"4:infod" +
		"12:piece lengthi512e" +
		"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
		"4:name8:testname" +
		"5:filesl" +
		"d6:lengthi123e4:pathl8:file.txteee" +
		"ee",
	)

	info, _, err := GetMetadata(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(info.Files) != 1 {
		t.Fatalf("expected 1 file")
	}
	if info.Files[0].Length != 123 {
		t.Fatalf("wrong length")
	}
	if info.Files[0].Path[0] != "file.txt" {
		t.Fatalf("wrong path")
	}
}

func TestGetMetadata_InvalidRootType(t *testing.T) {
	data := []byte("i123e")

	_, _, err := GetMetadata(data)
	if err == nil {
		t.Fatal("expected error for invalid root type")
	}
}

func TestGetMetadata_InvalidLength(t *testing.T) {
	data := []byte(
		"d8:announce55:http://bittorrent-test-tracker.codecrafters.io/announce" +
		"4:infod" +
		"12:piece lengthi4e" +
		"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
		"4:name8:file.txt" +
		"6:lengthee" +
		"ee",
	)

	_, _, err := GetMetadata(data)
	if err == nil {
		t.Fatal("expected error for invalid length")
	}
}

func TestGetMetadata_SingleFile_WithOptional(t *testing.T) {
	data := []byte(
		"d8:announce55:http://bittorrent-test-tracker.codecrafters.io/announce" +
		"4:infod" +
		"12:piece lengthi4e" +
		"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
		"4:name8:file.txt" +
		"6:lengthi200e" +
		"7:privatei1e" +
		"6:md5sum3:abc" +
		"ee",
	)

	info, _, err := GetMetadata(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Private == nil || *info.Private != 1 {
		t.Fatalf("expected private=1")
	}
	if info.Md5sum == nil || *info.Md5sum != "abc" {
		t.Fatalf("expected md5sum")
	}
}

func TestGetMetadata_MultiFile(t *testing.T) {
	data := []byte(
		"d8:announce55:http://bittorrent-test-tracker.codecrafters.io/announce" +
		"4:infod" +
		"12:piece lengthi4e" +
		"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
		"4:name6:folder" +
		"5:filesl" +
			"d6:lengthi50e4:pathl5:a.txtee" +
		"ee" +
		"ee",
	)

	info, _, err := GetMetadata(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(info.Files) != 1 {
		t.Fatalf("expected 1 file")
	}
	if info.Length != nil {
		t.Fatalf("should not be single-file")
	}
}

func TestGetMetadata_MultiFile_WithMD5(t *testing.T) {
	data := []byte(
		"d8:announce55:http://bittorrent-test-tracker.codecrafters.io/announce" +
		"4:infod" +
		"12:piece lengthi4e" +
		"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
		"4:name6:folder" +
		"5:filesl" +
			"d6:lengthi50e4:pathl5:a.txt" +
			"e6:md5sum3:xyze" +
		"e" +
		"ee",
	)

	info, _, err := GetMetadata(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Files[0].Md5sum == nil {
		t.Fatalf("expected md5sum")
	}
}

func TestGetMetadata_InvalidPieces(t *testing.T) {
	data := []byte(
		"d4:infod" +
		"12:piece lengthi4e" +
		"6:pieces5:short" +
		"4:name4:file" +
		"6:lengthi10e" +
		"ee",
	)

	_, _, err := GetMetadata(data)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetMetadata_Invalid_NoMode(t *testing.T) {
	data := []byte(
		"d4:infod" +
		"12:piece lengthi4e" +
		"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
		"4:name4:file" +
		"ee",
	)

	_, _, err := GetMetadata(data)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetMetadata_InvalidRoot(t *testing.T) {
	_, _, err := GetMetadata([]byte("i123e"))
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetMetadata_NoInfo(t *testing.T) {
	data := []byte("d3:foo3:bare")
	_, _, err := GetMetadata(data)
	if err == nil {
		t.Fatalf("expected error")
	}
}