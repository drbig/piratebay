package main

import (
	"flag"
	"fmt"
	"github.com/drbig/piratebay"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	VERSION    = "0.0.1"
	TIMELAYOUT = "2006-01-02 15:04:05 MST"
)

var (
	flagOrder    string
	flagCategory string
	flagFirst    bool
	flagMagnet   bool
	flagDetails  bool
	flagDebug    bool
	flagVersion  bool
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] query query...\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&flagOrder, "o", "seeders", "sorting order")
	flag.StringVar(&flagCategory, "c", "all", "category filter")
	flag.BoolVar(&flagFirst, "f", false, "only print first match")
	flag.BoolVar(&flagMagnet, "m", false, "only print magnet link")
	flag.BoolVar(&flagDetails, "d", false, "print details for each torrent")
	flag.BoolVar(&flagDebug, "debug", false, "enable debug logging")
	flag.BoolVar(&flagVersion, "version", false, "show version and exit")
}

func main() {
	flag.Parse()
	if flagVersion {
		fmt.Fprintf(os.Stderr, "pbcmd command version: %s\n", VERSION)
		fmt.Fprintf(os.Stderr, "piratebay library version: %s\n", piratebay.VERSION)
		return
	}
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	pb := piratebay.NewSite()
	if !flagDebug {
		pb.Logger = log.New(ioutil.Discard, "", 0)
	}
	err := pb.UpdateOrderings()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't load orderings: %s\n", err)
		os.Exit(2)
	}
	err = pb.UpdateCategories()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't load categories: %s\n", err)
		os.Exit(2)
	}
	order, err := pb.FindOrdering(flagOrder)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't find ordering: %s\n", err)
		os.Exit(2)
	}
	parts := strings.Split(flagCategory, "/")
	var category *piratebay.Category
	switch len(parts) {
	case 1:
		category, err = pb.FindCategory("", parts[0])
	case 2:
		category, err = pb.FindCategory(parts[0], parts[1])
	default:
		fmt.Fprintf(os.Stderr, "Can't parse '%s' as a category\n", flagCategory)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't find category: %s\n", err)
		os.Exit(2)
	}

	for i, query := range flag.Args() {
		torrents, err := pb.Search(query, category, order)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error for query '%s': %s\n", query, err)
			continue
		}
		if len(torrents) < 1 {
			fmt.Fprintf(os.Stderr, "Nothing found for query '%s'\n", query)
			continue
		}
		if flagFirst {
			torrents = torrents[0:1]
		}
		for j, tr := range torrents {
			if flagMagnet {
				fmt.Println(tr.Magnet)
				continue
			}
			// 2 + 1 + 2 + 2 + 64 + 2 + 4 = 77 < 80 == good
			fmt.Printf("%2d %2d  %-64s  %4d\n", i+1, j+1, tr.Title, tr.Seeders)
			if flagDetails {
				printDetails(tr)
			}
		}
	}
}

func printDetails(tr *piratebay.Torrent) {
	tr.GetDetails()
	tr.GetFiles()
	fmt.Printf(
		"       %-10s  %s  %-27s  %4d\n",
		tr.SizeStr,
		tr.Uploaded.Format(TIMELAYOUT),
		tr.User,
		tr.Leechers,
	)
	fmt.Printf("       %s\n", tr.InfoURI())
  fmt.Printf("       Files: _______________________________________________________________\n")
	for idx, file := range tr.Files {
		fmt.Printf("  %3d  %-58s  %10s\n", idx+1, file.Path, file.SizeStr)
	}
	fmt.Println()
}