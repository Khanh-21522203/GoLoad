package logic

import (
	"context"
	"io"
	"log"
	"net/http"
)

const (
	HTTPResponseHeaderContentType = "Content-Type"
	HTTPMetadataKeyContentType    = "content-type"
)

type Downloader interface {
	Download(ctx context.Context, writer io.Writer) (map[string]any, error)
}
type HTTPDownloader struct {
	url string
}

func NewHTTPDownloader(url string) Downloader {
	return &HTTPDownloader{
		url: url,
	}
}
func (h HTTPDownloader) Download(ctx context.Context, writer io.Writer) (map[string]any, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, h.url, http.NoBody)
	if err != nil {
		log.Printf("failed to create http get request")
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("failed to make http get request")
		return nil, err
	}
	defer response.Body.Close()
	_, err = io.Copy(writer, response.Body)
	if err != nil {
		log.Printf("failed to read response and write to writer")
		return nil, err
	}
	metadata := map[string]any{
		HTTPMetadataKeyContentType: response.Header.Get(HTTPResponseHeaderContentType),
	}
	return metadata, nil
}
