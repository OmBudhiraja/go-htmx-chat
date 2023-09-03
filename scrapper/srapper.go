package scrapper

import (
	"errors"
	"net/http"

	"github.com/OmBudhiraja/go-htmx-chat/utils"
	"github.com/PuerkitoBio/goquery"
)

type Metadata struct {
	Title       string
	Description string
	Image       string
	Url         string
}

func GetMetadata(url string) (Metadata, error) {

	var metadata Metadata

	if !utils.IsValidURL(url) {
		return metadata, errors.New("invalid URL")
	}

	res, err := http.Get(url)

	if err != nil {
		return metadata, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return metadata, err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)

	metadata.Title = doc.Find(`meta[property="og:title"]`).AttrOr("content", "")
	metadata.Description = doc.Find(`meta[property="og:description"]`).AttrOr("content", "")
	metadata.Image = doc.Find(`meta[property="og:image"]`).AttrOr("content", "/public/images/placeholder.png")
	metadata.Url = doc.Find(`meta[property="og:url"]`).AttrOr("content", url)

	if metadata.Title == "" {
		metadata.Title = doc.Find("title").Text()
	}

	if metadata.Description == "" {
		metadata.Description = doc.Find(`meta[name="description"]`).AttrOr("content", "")
	}

	if err != nil {
		return metadata, err
	}

	return metadata, nil

}
