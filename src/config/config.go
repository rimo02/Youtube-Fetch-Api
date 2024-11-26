package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	AddrPort         string
	Etag             string
	MaxVideosFetched int
	LimitPerPage     int
	Query            string
	ValidAPIKey      string
	DeleteApiKeys    int
	Mongouri         string
}

var config Config

func InitConfig() {
	flag.StringVar(&config.AddrPort, "addrport", os.Getenv("ADDR_PORT"), "port where server is running")

	flag.StringVar(&config.Mongouri, "mongodburi", os.Getenv("MONGO_URI"), "mongodb uri for connection")

	flag.StringVar(&config.Query, "QUERY", os.Getenv("QUERY"), "search query parameter")

	flag.StringVar(&config.ValidAPIKey, "API_KEY", os.Getenv("ValidAPIKey"), "API key for authentication")

	maxVideosFetched, err := strconv.Atoi(os.Getenv("MAX_VIDEOS_FETCHED"))
	if err != nil {
		maxVideosFetched = 10 // default value if environment variable is not set or invalid
	}
	flag.IntVar(&config.MaxVideosFetched, "MAX_VIDEOS_FETCHED", maxVideosFetched, "maximum number of videos to fetch")

	flag.IntVar(&config.LimitPerPage, "PER_PAGE_LIMIT", 10, "limit of videos per page")

	config.Etag = ""
	flag.Parse()
}

func GetQuery() string {
	return config.Query
}
func GetEtag() string {
	return config.Etag
}
func SetEtag(etag string) {
	config.Etag = etag
}
func MaxVideosFetched() int {
	return config.MaxVideosFetched
}
func LimitPerPage() int {
	return config.LimitPerPage
}

func GetApiKey() string {
	return config.ValidAPIKey
}
func SetApiKey(key string) {
	config.ValidAPIKey = key
}
func GetMaxVideos() int {
	return config.MaxVideosFetched
}
