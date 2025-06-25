package youtube

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	url := "https://www.googleapis.com/youtube/v3/search?part=snippet&type=musicvideoCategoryId=10&maxResults=11&q=%s&key=%s"
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

	var VideoIDs []string
	for _, item := range youtubeResponse.Items {
		VideoIDs = append(VideoIDs, item.ID.VideoID)
	}

	durations, err := y.getMusicDuration(VideoIDs, y.apiKey)
	if err != nil {
		return nil, fmt.Errorf("error fetching music durations: %w", err)
	}

	var youtubeDetails []YoutubeDetails
	for _, item := range youtubeResponse.Items {
		duration := durations[item.ID.VideoID]
		youtubeDetails = append(youtubeDetails, YoutubeDetails{
			VideoID:  item.ID.VideoID,
			Title:    item.Snippet.Title,
			Duration: duration,
		})
	}

	return youtubeDetails, nil

}

func (y *YoutubeClient) getMusicDuration(videoIDs []string, apiKey string) (map[string]string, error) {
	ids := strings.Join(videoIDs, ",")
	url := "https://www.googleapis.com/youtube/v3/videos?part=contentDetails&id=%s&key=%s"
	resp, err := http.Get(fmt.Sprintf(url, ids, apiKey))
	if err != nil {
		return nil, fmt.Errorf("error fetching video durations: %w", err)
	}

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error fetching video durations: received status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var musicDetailsReponse MusicDetailsReponse
	err = json.NewDecoder(resp.Body).Decode(&musicDetailsReponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	durations := make(map[string]string)
	for _, item := range musicDetailsReponse.Items {
		durations[item.ID] = item.ContentDetails.Duration
	}

	return durations, nil

}
