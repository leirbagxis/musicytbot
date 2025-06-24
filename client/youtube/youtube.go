package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type YoutubeClient struct {
	apiKey string
}

func New(apiKey string) *YoutubeClient {
	return &YoutubeClient{
		apiKey: apiKey,
	}
}

func (y *YoutubeClient) GetMusicDetails(title string) ([]YoutubeDetails, error) {
	encodedTitle := url.QueryEscape(title)
	url := "https://www.googleapis.com/youtube/v3/search?part=snippet&type=video&maxResults=20&q=%s&key=%s"
	resp, err := http.Get(fmt.Sprintf(url, encodedTitle, y.apiKey))

	if err != nil {
		return nil, fmt.Errorf("error fetching video details: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching video details: received status code %d", resp.StatusCode)
	}

	var youtubeResponse YoutubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&youtubeResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if len(youtubeResponse.Items) == 0 {
		return nil, fmt.Errorf("no videos found for title: %s", title)
	}

	var youtubeDetails []YoutubeDetails
	for _, item := range youtubeResponse.Items {
		youtubeDetails = append(youtubeDetails, YoutubeDetails{
			VideoID: item.ID.VideoID,
			Title:   item.Snippet.Title,
		})
	}

	return youtubeDetails, nil

}
