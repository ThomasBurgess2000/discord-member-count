package main

import (
	"encoding/csv"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/geziyor/geziyor/export"
	"github.com/prometheus/common/log"
)

func readInvites() []string {
	log.Info("Reading CSV...")
	var links []string
	csvfile, err := os.Open("serverlinks.csv")
	if err != nil {
		log.Error(err)
	}
	reader := csv.NewReader(csvfile)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error(err)
		}
		links = append(links, record[1])
	}
	return links
}

func startScrape(links []string) {
	log.Info("Starting scrape...")
	geziyor.NewGeziyor(&geziyor.Options{
		StartURLs: links,
		ParseFunc: quotesParse,
		Exporters: []export.Exporter{&export.JSONLine{FileName: "servers.jl"}},
	}).Start()
}

func main() {
	links := readInvites()
	startScrape(links)
	log.Info("Starting 5 minute cycle...")
	for range time.Tick(time.Minute * 5) {
		links = readInvites()
		startScrape(links)
	}
}

type DiscordItem struct {
	MemberCount int    `json:"member_count"`
	ScrapeDate  string `json:"scrape_date"`
	InviteID    string `json:"invite_id"`
	ServerName  string `json:"server_name"`
	ServerIcon  string `json:"server_icon"`
}

func quotesParse(g *geziyor.Geziyor, r *client.Response) {
	body := string(r.Body)

	// Member Count
	reg := regexp.MustCompile(`hang out with\s*(.*?)\s*other members`)
	matches := reg.FindAllStringSubmatch(body, -1)
	var memberCountString string
	if len(matches) < 1 {
		r.HTMLDoc.Find("meta").Each(func(i int, s *goquery.Selection) {
			if name, _ := s.Attr("name"); name == "description" {
				description, _ := s.Attr("content")
				descriptionSplit := strings.Split(description, "| ")
				memberCountString = strings.ReplaceAll(descriptionSplit[1], " members", "")
			}
		})
	} else {
		memberCountString = matches[0][1]
	}
	memberCountString = strings.ReplaceAll(memberCountString, ",", "")
	memberCount, err := strconv.Atoi(memberCountString)
	if err != nil {
		log.Error(err)
	}

	// Invite ID
	reg = regexp.MustCompile(`discord.com/invite/\s*(.*?)\s*"`)
	matches = reg.FindAllStringSubmatch(body, -1)
	inviteID := matches[0][1]

	// Server name
	reg = regexp.MustCompile(`Check out the \s*(.*?)\s* community on Discord`)
	matches = reg.FindAllStringSubmatch(body, -1)
	if len(matches) < 1 {
		reg = regexp.MustCompile(`Join the \s*(.*?)\s* Discord Server`)
		matches = reg.FindAllStringSubmatch(body, -1)
	}
	serverName := matches[0][1]

	// Server Icon
	reg = regexp.MustCompile(`property="og:image" content="\s*(.*?)\s*"`)
	matches = reg.FindAllStringSubmatch(body, -1)
	serverIcon := matches[0][1]

	// Scrape time
	now := time.Now()
	nowString := now.Format("2006-01-02 15:04:05")

	discordServer := DiscordItem{
		MemberCount: memberCount,
		ScrapeDate:  nowString,
		InviteID:    inviteID,
		ServerName:  serverName,
		ServerIcon:  serverIcon,
	}

	g.Exports <- discordServer
}
