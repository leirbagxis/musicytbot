package youtube

type YoutubeResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
	} `json:"items"`
}

type YoutubeDetails struct {
	VideoID string `json:"videoId"`
	Title   string `json:"title"`
}
