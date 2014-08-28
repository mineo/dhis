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
// HeaderTempl is the template for the album header
const HeaderTempl = "[quote][b]%d[/b] [artist]%s[/artist] - [b][album artist=%s]%s[/album][/b] (%d)[/quote]\n"
// ImageTempl is the template for an image
const ImageTempl = "[align=center][url=https://musicbrainz.org/release/%s][img=http://coverartarchive.org/release/%s/front-250][/img][/url][/align]"

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

	for index, info := range lastFmImageInfos {
		fmt.Printf(HeaderTempl, index, info.artist, info.artist, info.album, info.plays)
		if info.mbid == nil {
			continue
			// fmt.Printf("%s by %s has no MBID in Last.fm\n", info.album, info.artist)
		} else if !info.hasCAAImage {
			continue
			// fmt.Printf("%s by %s has no image in the CAA\n", info.album, info.artist)
		} else {
		    fmt.Printf(ImageTempl, info.mbid.String(), info.mbid.String())
		}
	}
}
