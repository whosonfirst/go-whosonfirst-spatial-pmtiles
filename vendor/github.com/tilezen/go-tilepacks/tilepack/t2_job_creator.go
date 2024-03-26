package tilepack

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/maptile"
)

func NewTapalcatl2JobGenerator(bucket string, requesterPays bool, pathTemplate string, layerName string, materializedZooms []maptile.Zoom, zooms []maptile.Zoom, bounds orb.Bound) (JobGenerator, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}

	downloader := s3manager.NewDownloader(sess)

	return &tapalcatl2JobGenerator{
		s3Client:          downloader,
		bucket:            bucket,
		requesterPays:     requesterPays,
		pathTemplate:      pathTemplate,
		layerName:         layerName,
		materializedZooms: materializedZooms,
		bounds:            bounds,
		zooms:             zooms,
	}, nil
}

type tapalcatl2JobGenerator struct {
	s3Client          *s3manager.Downloader
	bucket            string
	requesterPays     bool
	pathTemplate      string
	layerName         string
	materializedZooms []maptile.Zoom
	bounds            orb.Bound
	zooms             []maptile.Zoom
}

func arrayContains(needle maptile.Zoom, haystack []maptile.Zoom) bool {
	for _, z := range haystack {
		if z == needle {
			return true
		}
	}
	return false
}

func (x *tapalcatl2JobGenerator) CreateWorker() (func(id int, jobs chan *TileRequest, results chan *TileResponse), error) {
	f := func(id int, jobs chan *TileRequest, results chan *TileResponse) {
		for request := range jobs {
			// Download the Tapalcatl2 archive zip to a byte buffer
			compressedBytes := &aws.WriteAtBuffer{}
			input := &s3.GetObjectInput{
				Bucket: aws.String(x.bucket),
				Key:    aws.String(request.URL),
			}

			if x.requesterPays {
				input.RequestPayer = aws.String("requester")
			}

			numBytes, err := x.s3Client.Download(compressedBytes, input)
			if err != nil {
				log.Fatalf("Unable to download item %s: %+v", request.URL, err)
			}

			// Uncompress the archive
			readBytes := compressedBytes.Bytes()
			readBytesReader := bytes.NewReader(readBytes)
			zippedReader, err := zip.NewReader(readBytesReader, numBytes)
			if err != nil {
				log.Fatalf("Unable to unzip t2 archive %s: %+v", request.URL, err)
			}

			// Iterate over the contents of the zip and add them as TileResponses
			for _, zf := range zippedReader.File {
				var tileX, tileY uint32
				var tileZ maptile.Zoom
				if n, err := fmt.Sscanf(zf.Name, "%d/%d/%d@2x.png", &tileZ, &tileX, &tileY); err != nil || n != 3 {
					log.Fatalf("Couldn't scan t2 name")
				}

				t := maptile.New(tileX, tileY, tileZ)

				if !arrayContains(tileZ, x.zooms) {
					continue
				}

				if !x.bounds.Intersects(t.Bound()) {
					continue
				}

				// Read the data for the tile
				zfReader, err := zf.Open()
				if err != nil {
					log.Fatalf("Couldn't read zf %s: %+v", zf.Name, err)
				}

				b, err := io.ReadAll(zfReader)
				if err != nil {
					log.Fatalf("Couldn't read zf %s: %+v", zf.Name, err)
				}

				results <- &TileResponse{
					Data: b,
					Tile: t,
				}
			}
		}
	}

	return f, nil
}

func (x *tapalcatl2JobGenerator) CreateJobs(jobs chan *TileRequest) error {
	// Iterate over the list of materialized zooms
	for _, materializedZoom := range x.materializedZooms {
		// Generate requests for tiles in the bounding box at this materialized zoom
		GenerateTiles(&GenerateTilesOptions{
			Bounds:    x.bounds,
			InvertedY: false,
			Zooms:     []maptile.Zoom{materializedZoom},
			ConsumerFunc: func(t maptile.Tile) {
				hash := md5.Sum([]byte(fmt.Sprintf("%d/%d/%d.zip", t.Z, t.X, t.Y)))
				hashHex := hex.EncodeToString(hash[:])

				path := strings.NewReplacer(
					"{x}", fmt.Sprintf("%d", t.X),
					"{y}", fmt.Sprintf("%d", t.Y),
					"{z}", fmt.Sprintf("%d", t.Z),
					"{l}", x.layerName,
					"{h}", hashHex[:5]).Replace(x.pathTemplate)

				jobs <- &TileRequest{
					Tile: t,
					URL:  path,
				}
			},
		})
	}

	return nil
}
