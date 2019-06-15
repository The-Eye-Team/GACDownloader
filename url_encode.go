package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ncw/rclone/lib/pacer"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var cantFindImageError = errors.New("cant find Image")

type Item struct {
	tileInfo  TileInfo
	imageName string
	thumbnailUri string
	urlKey string
}

func NewItem(itemUrl string) (*Item, error) {
	uri, err := url.Parse(itemUrl)
	if err != nil {
		return nil, err
	}

	pathParts := strings.Split(uri.Path, "/")
	imageID := pathParts[len(pathParts)-1]
	imageSlug := pathParts[len(pathParts)-2]

	imageName := fmt.Sprintf("%s - %s", imageSlug, imageID)

	itemPage, err := requestAsBytes(itemUrl)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(itemPage))
	if err != nil {
		return nil, err
	}

	thumbnailUri, exists := doc.Find("[property=\"og:image\"]").Attr("content")
	if !exists {
		return nil, cantFindImageError
	}

	thumbnailUriWithoutProto := strings.Split(thumbnailUri, ":")[1]
	regex, err := regexp.Compile("\"" + thumbnailUriWithoutProto + "\",\"(?P<key>[^\"]+)\"")
	if err != nil {
		return nil, err
	}

	key := regex.FindStringSubmatch(string(itemPage))[1]

	xmlData, err := requestAsBytes(thumbnailUri + "=g")
	if err != nil {
		return nil, err
	}

	var tileInfo TileInfo
	err = xml.Unmarshal(xmlData, &tileInfo)
	if err != nil {
		return nil, err
	}

	return &Item{
		tileInfo,
		imageName,
		thumbnailUri,
		key,
	}, nil
}

func (item *Item) encodeUrl(x, y, zoom int) (encodedUrl string, err error) {
	thumbnailID := strings.Split(item.thumbnailUri, "/")[3]

	signPath := fmt.Sprintf("%s=x%d-y%d-z%d-t%s", thumbnailID, x, y, zoom, item.urlKey)

	IV, _ := hex.DecodeString("7b2b4e23de2cc5c5")

	mac := hmac.New(sha1.New, IV)
	mac.Write([]byte(signPath))
	digest := mac.Sum(nil)

	encoded := base64.URLEncoding.EncodeToString(digest)

	encoded = strings.ReplaceAll(encoded, "-", "_")

	encodedUrl = fmt.Sprintf("%s=x%d-y%d-z%d-t%s", item.thumbnailUri, x, y, zoom, encoded[:len(encoded)-1])

	return encodedUrl, nil
}


var Pacer = pacer.New()
func requestAsBytes(url string) ([]byte, error) {
	var response *http.Response
	var err error

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36")

	err = Pacer.Call(func() (bool, error) {
		response, err = client.Do(request)
		return true, err
	})

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}
