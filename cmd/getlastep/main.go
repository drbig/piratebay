package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"github.com/drbig/piratebay"
	"github.com/drbig/transmission_rpc"
	"github.com/drbig/tvrage"
)

const (
	VERSION = "0.0.1"
)

var (
	flagClient  string
	flagVersion bool
	filterMap   = []string{"seeders:min:1", "size:min:400000000", "files:include:.*\\.mkv"}
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options...] show show...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
	}
	flag.BoolVar(&flagVersion, "v", false, "print version and exit")
	flag.StringVar(&flagClient, "c", "", "full Transmission RPC URL")
}

func main() {
	flag.Parse()
	if flagVersion {
		fmt.Fprintf(os.Stderr, "getlastep                version: %s\n", VERSION)
		fmt.Fprintf(os.Stderr, "tvrage           library version: %s\n", tvrage.VERSION)
		fmt.Fprintf(os.Stderr, "piratebay        library version: %s\n", piratebay.VERSION)
		fmt.Fprintf(os.Stderr, "transmission_rpc library version: %s\n", transmission_rpc.VERSION)
		os.Exit(0)
	}
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(0)
	}
	var client *transmission_rpc.Client
	if len(flagClient) > 0 {
		clientURL, err := url.Parse(flagClient)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't parse client URL: %s\n", err)
			os.Exit(1)
		}
		client = transmission_rpc.NewClient(fmt.Sprintf("%s://%s", clientURL.Scheme, clientURL.Host))
		if clientURL.User != nil {
			pass, set := clientURL.User.Password()
			if !set {
				pass = ""
			}
			client.SetAuth(clientURL.User.Username(), pass)
		}
	}
	pb := piratebay.NewSite()
	pb.Logger = log.New(ioutil.Discard, "", 0)
	if err := pb.UpdateCategories(); err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't load categories: %s\n", err)
		os.Exit(1)
	}
	if err := pb.UpdateOrderings(); err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't load orderings: %s\n", err)
		os.Exit(1)
	}
	order, err := pb.FindOrdering("seeders")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't setup ordering: %s\n", err)
		os.Exit(1)
	}
	category, err := pb.FindCategory("video", "hd - tv shows")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't setup category: %s\n", err)
		os.Exit(1)
	}
	filters, err := piratebay.SetupFilters(filterMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't setup filters: %s\n", err)
		os.Exit(1)
	}

	for idx, title := range flag.Args() {
		shows, err := tvrage.Search(title)
		if err != nil {
			fmt.Fprintf(os.Stderr, `ERROR searching for "%s":\n%s\n\n`, title, err)
			continue
		}
		episodes, err := tvrage.EpisodeList(shows[0].ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, `ERROR fetching episodes for "%s":\n%s\n\n`, title, err)
			continue
		}
		ep, found := episodes.Last()
		if !found {
			fmt.Printf("No last eposide found, sorry.\n\n")
			continue
		}
		fmt.Printf("%2d. %s\n    %s (%s, %s)\n", idx+1, shows[0], ep, ep.AirDate.Format(`2006-02-01`), ep.DeltaDays())
		query := fmt.Sprintf("%s S%02dE%02d x264", shows[0].Name, ep.Season, ep.Number)
		torrents, err := pb.Search(query, category, order)
		if err != nil {
			fmt.Fprintf(os.Stderr, `ERROR searching piratebay:\n%s\n\n`, err)
			continue
		}
		if len(torrents) < 1 {
			fmt.Printf("    No torrent found, sorry :(\n\n")
			continue
		}
		filtered := piratebay.ApplyFilters(torrents, filters)
		var best *piratebay.Torrent
		if len(filtered) > 0 {
			best = filtered[0]
		} else {
			best = torrents[0]
		}
		fmt.Printf("  - Found you a torrent:\n    %s\n", best)
		added := false
		if client != nil {
			_, err := client.Request("torrent-add", map[string]string{"filename": best.Magnet})
			if err != nil {
				fmt.Printf("    Couldn't add the torrent :(\n")
			} else {
				fmt.Printf("  - Added torrent :)\n")
				added = true
			}
		}
		if !added {
			fmt.Printf("  - Magnet link:\n%s\n", best.Magnet)
		}
		fmt.Println()
	}
}
