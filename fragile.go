// See LICENSE.txt for licensing information.

package piratebay

import (
	"regexp"
)

// Following consts define URIs and Regexps that are used to interact with PirateBay site
// and extract relevant information. The idea is that as long as the site's HTML layout
// doesn't change too much one will only need to make tweaks here.
// At least that's the idea.
const (
	ROOTURI        = `http://thepiratebay.org`                                                                                                                                                                                        // PirateBay root URI
	INFRAURI       = `/search/a/0/99/0`                                                                                                                                                                                               // URI for fetching 'infrastructure' data
	CATEGORYREGEXP = `<opt.*? (.*?)="(.*?)">([A-Za-z0-9- ()/]+)?<?`                                                                                                                                                                   // Regexp for Category data extraction
	ORDERINGREGEXP = `/(\d+)/0" title="Order by (.*?)"`                                                                                                                                                                               // Regexp for Ordering data extraction
	SEARCHURI      = `/search/%s/0/%s/%s`                                                                                                                                                                                             // URI for search queries
	SEARCHREGEXP   = `(?s)category">(.*?)</a>.*?/browse/(\d+)".*?category">(.*?)</a>.*?torrent/(\d+)/.*?>(.*?)</a>.*?(magnet.*?)".*?(vip|11x11).*?Uploaded (.*?), Size (.*?), ULed by .*?>(.*?)<.*?right">(\d+)<.*?right">(\d+)</td>` // Regexp for extracting search results
	INFOURI        = `/torrent/%s`                                                                                                                                                                                                    // URI for fetching Torrent details
	INFOREGEXP     = `(?s)Size:.*?\((.*?)&nbsp;Bytes\).*?Uploaded:.*?d>(.*?)</d`                                                                                                                                                      // Regexp for Torrent details extraction
	FILESURI       = `/ajax_details_filelist.php?id=%s`                                                                                                                                                                               // URI for fetching Torrent Files data
	FILESREGEXP    = `left">(.*?)</td.*?right">(.*?)<`                                                                                                                                                                                // Regexp for extracting File data
)

// This should be treated as a const.
var (
	killHTMLRegexp = regexp.MustCompile(`<.*?>`) // Regexp used for removing HTML
)
