package main

import (
	"fmt"
	"gustorrent/internal/metadata"
	pr "gustorrent/internal/peer"
	"gustorrent/internal/tracker"
	"os"
	"sync"
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

	// compute total size (for last piece)
	totalSize := 0
	if info.Length != nil {
		totalSize = *info.Length
	} else {
		for _, f := range info.Files {
			totalSize += f.Length
		}
	}
	lastPieceLength := totalSize - (totalPieces-1)*info.PieceLength

	peers, err := tracker.GetPeers(data)
	if err != nil {
		panic(err)
	}

	fmt.Println("Peers:", len(peers))
	if len(peers) == 0 {
		panic("no peers")
	}

	maxPeers := 20
	sem := make(chan struct{}, maxPeers)

	var wg sync.WaitGroup

	for _, p := range peers {
		addr := p.String()

		sem <- struct{}{} // acquire slot
		wg.Add(1)

		go func() {
			defer wg.Done()
			defer func() { <-sem }() // release slot (like a counting semaphore)

			pr.PeerWorker(addr, pm, data, &info, lastPieceLength)
		}()
	}

	// progress monitor (optional but very useful)
	go func() {
		for {
			done := pm.Done()
			fmt.Printf("\rProgress: %d / %d", done, totalPieces)

			if done >= totalPieces {
				fmt.Println("\nDownload complete!")
				return
			}

			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()

	fmt.Println("\nAll pieces downloaded!")
}