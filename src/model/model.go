package model

type Searchapi struct {
	VideoId     string `json:"VideoId"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
	ChannelId   string `json:"ChannelId"`
	ChannelName string `json:"ChannelName"`
	PublishedAt string `json:"PublishedAt"`
	Etag        string `json:"Etag"`
}
