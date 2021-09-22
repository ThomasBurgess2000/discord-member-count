package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/geziyor/geziyor/export"
	"github.com/prometheus/common/log"
)

type CollectionItem struct {
	CollectionName  string  `json:"collection_name"`
	Slug            string  `json:"slug"`
	Logo            string  `json:"logo"`
	FloorPrice      float32 `json:"floor_price"`
	MarketCap       float32 `json:"market_cap"`
	NumOwners       float32 `json:"num_owners"`
	TotalSupply     float32 `json:"total_supply"`
	SevenDayChange  float32 `json:"seven_day_change"`
	SevenDayVolume  float32 `json:"seven_day_volume"`
	OneDayChange    float32 `json:"one_day_change"`
	OneDayVolume    float32 `json:"one_day_volume"`
	ThirtyDayChange float32 `json:"thirty_day_change"`
	ThirtyDayVolume float32 `json:"thirty_day_volume"`
	TotalVolume     float32 `json:"total_volume"`
	DiscordSlug     string  `json:"discord_slug"`
	TwitterUsername string  `json:"twitter_username"`
}

func main() {
	log.Info("Starting scrape...")
	geziyor.NewGeziyor(&geziyor.Options{
		StartRequestsFunc: func(g *geziyor.Geziyor) {
			g.GetRendered("https://opensea.io/rankings", g.Opt.ParseFunc)
		},
		ParseFunc:          rankingsParse,
		Exporters:          []export.Exporter{&export.JSONLine{FileName: "nfts.jl"}},
		UserAgent:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:92.0) Gecko/20100101 Firefox/92.0",
		RobotsTxtDisabled:  true,
		RetryHTTPCodes:     []int{500, 502, 503, 504, 522, 524, 408, 429},
		RetryTimes:         10,
		CookiesDisabled:    false,
		ConcurrentRequests: 1,
	}).Start()
}

func rankingsParse(g *geziyor.Geziyor, r *client.Response) {
	reg := regexp.MustCompile(`"json":{"data":{"collections":{"edges":\s*(.*?)\s*"data":{"collections":{"edges"`)
	matches := reg.FindAllStringSubmatch(string(r.Body), -1)
	jsonString := matches[0][1]
	jsonStringSplit := strings.Split(jsonString, ",\"pageInfo")
	jsonString = jsonStringSplit[0]
	var rankings Rankings
	err := json.Unmarshal([]byte(jsonString), &rankings)
	if err != nil {
		log.Error(err)
	}
	for _, collection := range rankings {
		collectionItem := CollectionItem{
			CollectionName:  collection.Node.Name,
			Slug:            collection.Node.Slug,
			Logo:            collection.Node.Logo,
			FloorPrice:      float32(collection.Node.Stats.FloorPrice),
			MarketCap:       float32(collection.Node.Stats.MarketCap),
			NumOwners:       float32(collection.Node.Stats.NumOwners),
			TotalSupply:     float32(collection.Node.Stats.TotalSupply),
			SevenDayChange:  float32(collection.Node.Stats.SevenDayChange),
			SevenDayVolume:  float32(collection.Node.Stats.SevenDayVolume),
			OneDayChange:    float32(collection.Node.Stats.OneDayChange),
			OneDayVolume:    float32(collection.Node.Stats.OneDayVolume),
			ThirtyDayChange: float32(collection.Node.Stats.ThirtyDayChange),
			ThirtyDayVolume: float32(collection.Node.Stats.ThirtyDayVolume),
			TotalVolume:     float32(collection.Node.Stats.TotalVolume),
		}

		req, err := http.NewRequest("GET", "https://opensea.io/collection/"+collectionItem.Slug, nil)
		if err != nil {
			log.Error(err)
		}

		request := client.Request{
			Request: req,
			Meta: map[string]interface{}{
				"collectionItem": collectionItem,
			},
			Rendered: true,
		}

		g.Do(&request, parseCollection)
	}
}

func parseCollection(g *geziyor.Geziyor, r *client.Response) {
	collectionItem := r.Request.Meta["collectionItem"].(CollectionItem)

	// Discord
	reg := regexp.MustCompile(`discord.gg\s*(.*?)\s*","`)
	matches := reg.FindAllStringSubmatch(string(r.Body), -1)
	if len(matches) > 0 {
		discordSlug := matches[0][1]
		discordSlug = strings.ReplaceAll(discordSlug, "\\u002F", "")
		collectionItem.DiscordSlug = discordSlug
	}

	// Twitter
	reg = regexp.MustCompile(`twitterUsername":"\s*(.*?)\s*","`)
	matches = reg.FindAllStringSubmatch(string(r.Body), -1)
	if len(matches) > 0 {
		twitterSlug := matches[0][1]
		collectionItem.TwitterUsername = twitterSlug
	}

	g.Exports <- collectionItem
}

type Rankings []struct {
	Node struct {
		CreatedDate string `json:"createdDate"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Logo        string `json:"logo"`
		Stats       struct {
			FloorPrice      float32 `json:"floorPrice"`
			MarketCap       float64 `json:"marketCap"`
			NumOwners       int     `json:"numOwners"`
			TotalSupply     int     `json:"totalSupply"`
			SevenDayChange  float64 `json:"sevenDayChange"`
			SevenDayVolume  float64 `json:"sevenDayVolume"`
			OneDayChange    float64 `json:"oneDayChange"`
			OneDayVolume    float64 `json:"oneDayVolume"`
			ThirtyDayChange float64 `json:"thirtyDayChange"`
			ThirtyDayVolume float64 `json:"thirtyDayVolume"`
			TotalVolume     float64 `json:"totalVolume"`
			ID              string  `json:"id"`
		} `json:"stats"`
		ID       string `json:"id"`
		Typename string `json:"__typename"`
	} `json:"node"`
	Cursor string `json:"cursor"`
}
