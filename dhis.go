// Package main provides ...
package main

import (
	"fmt"
	"strconv"
	"sync"

	"code.google.com/p/go-uuid/uuid"

	"github.com/mineo/gocaa"
	"github.com/shkh/lastfm-go/lastfm"
)

const apiKey = "ed572ca7123d746483dd797a6d72bb88"

func getCAAInfo(client *caa.CAAClient, mbid uuid.UUID) (info *caa.CoverArtInfo, err error) {
	info, err = client.GetReleaseInfo(mbid)
	return
}

type lastFMImageInfo struct {
	artist      string
	album       string
	mbid        uuid.UUID
	plays       int
	hasCAAImage bool
}

func main() {
	user := "DasMineo"
	lfm := lastfm.New(apiKey, "")
	caaClient := caa.NewCAAClient("dhis")

	p := lastfm.P{
		"user":  user,
		"limit": 25,
	}
	res, err := lfm.User.GetTopAlbums(p)

	if err != nil {
		fmt.Println(err)
		return
	}

	var lastFmImageInfos [25]*lastFMImageInfo

	var wg sync.WaitGroup

	// Check for each album if it has an image in the CAA
	for i, album := range res.Albums {
		plays, _ := strconv.Atoi(album.PlayCount)

		lfmInfo := lastFMImageInfo{
			artist: album.Artist.Name,
			album:  album.Name,
			plays:  plays,
		}

		lastFmImageInfos[i] = &lfmInfo

		// Continuing makes no sense because last.fm doesn't have an MBID for
		// this album
		if album.Mbid == "" {
			continue
		}

		lfmInfo.mbid = uuid.Parse(album.Mbid)

		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			info, err := getCAAInfo(caaClient, lfmInfo.mbid)

			if err != nil {
				fmt.Printf("%s: %s\n", lfmInfo.mbid, err.Error())
				return
			}

			for _, imageInfo := range info.Images {
				if imageInfo.Front {
					lastFmImageInfos[index].hasCAAImage = true
					break
				}
			}
		}(i)
	}

	wg.Wait()

	for _, info := range lastFmImageInfos {
		if info.mbid == nil {
			fmt.Printf("%s by %s has no MBID in Last.fm\n", info.album, info.artist)
			continue
		} else if !info.hasCAAImage {
			fmt.Printf("%s by %s has no image in the CAA\n", info.album, info.artist)
			continue
		}

		fmt.Printf("%s by %s (%d plays)\n", info.album, info.artist, info.plays)
		fmt.Printf("http://coverartarchive.org/release/%s/front-500\n", info.mbid.String())
	}
}
