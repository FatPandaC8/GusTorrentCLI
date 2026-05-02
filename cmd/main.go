package main

import (
	"fmt"
	"gustorrent/internal/metadata"
	pr "gustorrent/internal/peer"
	"gustorrent/internal/tracker"
	"io"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: app <torrent-file>")
		return
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	info, _, err := metadata.GetMetadata(data)
	if err != nil {
		panic(err)
	}

	totalPieces := len(info.Pieces) / 20
	pm := pr.NewPieceManager(totalPieces)

	// precompute total size + last piece
	totalSize := 0
	if info.Length != nil {
		totalSize = *info.Length
	} else {
		for _, f := range info.Files {
			totalSize += f.Length
		}
	}
	lastPieceLength := totalSize - (totalPieces-1) * info.PieceLength

	peers, err := tracker.GetPeers(data)
	if err != nil {
		panic(err)
	}

	fmt.Println("Peers:", len(peers))
	if len(peers) == 0 {
		panic("no peers")
	}

	completed := 0

	for completed < totalPieces {
		progress := false

		for _, p := range peers {
			if completed >= totalPieces {
				break
			}

			fmt.Println("Connecting to peer:", p)

			conn, err := net.DialTimeout("tcp", p.String(), 5*time.Second)
			if err != nil {
				continue
			}

			conn.SetDeadline(time.Now().Add(30 * time.Second))

			hs, err := tracker.BuildHandshake(data, "-GT0001-123456789012")
			if err != nil {
				conn.Close()
				continue // skip to the next peer
			}

			if _, err := conn.Write(hs); err != nil {
				conn.Close()
				continue
			}

			resp := make([]byte, 68)
			if _, err := io.ReadFull(conn, resp); err != nil {
				conn.Close()
				continue
			}

			fmt.Println("Handshake OK")

			var bitfield tracker.Bitfield

			for { // wait for bitfield
				id, payload, err := pr.ReadMessage(conn)
				if err != nil {
					break
				}
				if id == 5 {
					bitfield = payload
					break
				}
			}

			if bitfield == nil {
				fmt.Println("No bitfield -> skip")
				conn.Close()
				continue
			}

			if _, err := conn.Write(tracker.BuildInterested()); err != nil {
				conn.Close()
				continue
			}

			unchoked := false
			for {
				id, _, err := pr.ReadMessage(conn)
				if err != nil {
					break
				}
				if id == 1 {
					unchoked = true
					fmt.Println("Unchoked!")
					break
				}
			}

			if !unchoked {
				fmt.Println("Never unchoked -> skip")
				conn.Close()
				continue
			}

			for {
				if completed >= totalPieces {
					conn.Close()
					break
				}

				pieceIndex, ok := pm.PickPiece(bitfield)
				if !ok {
					fmt.Println("Peer exhausted")
					break
				}

				pieceLength := info.PieceLength
				if pieceIndex == totalPieces - 1 {
					pieceLength = lastPieceLength
				}

				fmt.Println("Downloading piece:", pieceIndex)

				piece, err := pr.DownloadPiece(conn, pieceIndex, pieceLength)
				fmt.Println("PIECE downloaded:", len(piece))
				if err != nil {
					fmt.Println("Failed piece:", pieceIndex)
					fmt.Println("Failed piece error:", err)
					pm.MarkFailed(pieceIndex)
					continue
				}

				pm.MarkDone(pieceIndex)
				completed++
				progress = true

				fmt.Println("Done piece:", pieceIndex, "(", completed, "/", totalPieces, ")")
			}

			conn.Close()
		}

		if !progress {
			fmt.Println("No progress -> retrying...")
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Println("\nAll pieces downloaded!")
}