package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"sync/atomic"
)

const URL = "https://artsandculture.google.com/asset/wedding-supper/8wGaVfLW0exl6A"

const MaxDownloads = 20

var client = &http.Client{}

var workers int64

func main() {
	item, err := NewItem(URL)
	if err != nil {
		logrus.Fatal(err)
	}

	var worker sync.WaitGroup
	for zoomLevel := range item.tileInfo.PyramidLevel {
		worker.Add(1)

		go func(zoomLevel int) {
			logrus.Infof("Starting Zoom level %d", zoomLevel)

			defer worker.Done()
			item.DownloadZoomLevel(zoomLevel, func(x, y, z int, item *Item) bool {

				folderPath := path.Join("Downloads", item.imageName, fmt.Sprintf("Zoom %d", zoomLevel))
				imageName := fmt.Sprintf("x%d-y%d-z%d.jpg", x, y, zoomLevel)
				filePath := path.Join(folderPath, imageName)

				if _, err := os.Stat(filePath); !os.IsNotExist(err) {
					logrus.Infof("Skipping File  `%s`", imageName)
					return false
				}

				return true
			}, func(x, y, zoomLevel int, data []byte, item *Item) {
				folderPath := path.Join("Downloads", item.imageName, fmt.Sprintf("Zoom %d", zoomLevel))

				err := os.MkdirAll(folderPath, 0755)
				if err != nil {
					logrus.Fatal(err)
				}

				imageName := fmt.Sprintf("x%d-y%d-z%d.jpg", x, y, zoomLevel)
				logrus.Infof("Finished `%s`", imageName)

				filePath := path.Join(folderPath, imageName)

				err = ioutil.WriteFile(filePath, data, 0775)
				if err != nil {
					logrus.Fatal(err)
				}
			})
		}(zoomLevel)
	}

	worker.Wait()
}

func (item *Item) DownloadZoomLevel(zoomLevel int, checkFunc func(x, y, z int, item *Item) bool, downloadedFunc func(x, y, z int, data []byte, item *Item)) {
	level := item.tileInfo.PyramidLevel[zoomLevel]

	var worker sync.WaitGroup
	for x := 0; x < level.NumTilesX; x++ {
		for y := 0; y < level.NumTilesY; y++ {
			if atomic.LoadInt64(&workers) >= MaxDownloads {
				worker.Wait()
			}

			if !checkFunc(x, y, zoomLevel, item) {
				continue
			}

			logrus.Infof("Starting X %d Y %d", x, y)
			worker.Add(1)
			atomic.AddInt64(&workers, 1)

			go func(x, y int) {
				defer atomic.AddInt64(&workers, -1)
				defer worker.Done()

				encodedUrl, err := item.encodeUrl(x, y, zoomLevel)
				if err != nil {
					logrus.Fatal(err)
				}

				content, err := requestAsBytes(encodedUrl)
				if err != nil {
					logrus.Fatal(err)
				}

				decoded, err := decodeImage(content)
				if err != nil {
					logrus.Info(encodedUrl)
					logrus.Fatal(err)
				}

				downloadedFunc(x, y, zoomLevel, decoded, item)

				//img1, _, err := image.Decode(decoded)
			}(x, y)
		}
	}

	worker.Wait()

}
