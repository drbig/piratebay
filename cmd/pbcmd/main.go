// See LICENSE.txt for licensing information.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/drbig/piratebay"
)

const (
	VERSION    = "0.0.3"
	TIMELAYOUT = "2006-01-02 15:04:05 MST"
	FILTERSEP  = ";"
)

var (
	flagOrder          string
	flagCategory       string
	flagFilters        string
	flagShowFilters    bool
	flagShowOrders     bool
	flagShowCategories bool
	flagFirst          bool
	flagMagnet         bool
	flagDetails        bool
	flagDebug          bool
	flagVersion        bool
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] query query...\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Won't run any queries if any of -sf, -so, and -sc options have been supplied.\n")
		fmt.Fprintf(os.Stderr, "Running a query or using -so or -sc requires a connection to PirateBay.\n\n")
		flag.PrintDefaults()
	}
	flag.StringVar(&flagOrder, "o", "seeders", "sorting order (always descending)")
	flag.StringVar(&flagCategory, "c", "all", "category filter ('unique category' or 'group/category')")
	flag.StringVar(&flagFilters, "filters", "", "filters to apply (in sequence)")
	flag.BoolVar(&flagShowFilters, "sf", false, "print available filters")
	flag.BoolVar(&flagShowOrders, "so", false, "fetch and print available orderings")
	flag.BoolVar(&flagShowCategories, "sc", false, "fetch and print available categories")
	flag.BoolVar(&flagFirst, "f", false, "only print first match")
	flag.BoolVar(&flagMagnet, "m", false, "only print magnet link")
	flag.BoolVar(&flagDetails, "d", false, "print details for each torrent")
	flag.BoolVar(&flagDebug, "debug", false, "enable library debug output")
	flag.BoolVar(&flagVersion, "version", false, "show version and exit")
}

func main() {
	flag.Parse()
	if flagVersion {
		fmt.Fprintf(os.Stderr, "pbcmd command version: %s\n", VERSION)
		fmt.Fprintf(os.Stderr, "piratebay library version: %s\n", piratebay.VERSION)
		fmt.Fprintln(os.Stderr, "See LICENSE.txt for legal details.")
		os.Exit(0)
	}
	if flagShowFilters {
		fmt.Println("Available filters:")
		for _, f := range piratebay.GetFilters() {
			fmt.Println(f)
		}
	}
	pb := piratebay.NewSite()
	if !flagDebug {
		pb.Logger = log.New(ioutil.Discard, "", 0)
	}
	if flagShowOrders {
		loadOrderings(pb)
		fmt.Println("Available sort orders:")
		for o, id := range pb.Orderings {
			if flagDebug {
				fmt.Printf("%s (%s)\n", o, id)
			} else {
				fmt.Println(o)
			}
		}
	}
	if flagShowCategories {
		loadCategories(pb)
		fmt.Println("Available categories:")
		for group, cats := range pb.Categories {
			for c, id := range cats {
				if flagDebug {
					fmt.Printf("%s/%s (%s)\n", group, c, id)
				} else {
					fmt.Printf("%s/%s\n", group, c)
				}
			}
		}
	}
	if flagShowFilters || flagShowOrders || flagShowCategories {
		os.Exit(0)
	}
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	loadOrderings(pb)
	loadCategories(pb)
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
	var filters []piratebay.FilterFunc
	if flagFilters != "" {
		parts := strings.Split(flagFilters, FILTERSEP)
		filters, err = piratebay.SetupFilters(parts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting up filters: %s\n", err)
			os.Exit(2)
		}
	}

	for i, query := range flag.Args() {
		torrents, err := pb.Search(query, category, order)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error for query '%s': %s\n", query, err)
			continue
		}
		if len(torrents) < 1 {
			fmt.Fprintf(os.Stderr, "Nothing found for query '%s' (raw)\n", query)
			continue
		}
		if len(filters) != 0 {
			torrents = piratebay.ApplyFilters(torrents, filters)
		}
		if len(torrents) < 1 {
			fmt.Fprintf(os.Stderr, "Nothing found for query '%s' (filtered)\n", query)
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

func loadOrderings(pb *piratebay.Site) {
	err := pb.UpdateOrderings()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't load orderings: %s\n", err)
		os.Exit(2)
	}
}

func loadCategories(pb *piratebay.Site) {
	err := pb.UpdateCategories()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't load categories: %s\n", err)
		os.Exit(2)
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
