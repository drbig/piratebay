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
	ROOTURI        = `http://thepiratebay.org`
	INFRAURI       = `/search/a/0/99/0`
	CATEGORYREGEXP = `<opt.*? (.*?)="(.*?)">([A-Za-z- ()/]+)?<?`
	ORDERINGREGEXP = `/(\d+)/0" title="Order by (.*?)"`
	SEARCHURI      = `/search/%s/0/%s/%s`
	SEARCHREGEXP   = `(?s)category">(.*?)</a>.*?/browse/(\d+)".*?category">(.*?)</a>.*?torrent/(\d+)/.*?>(.*?)</a>.*?(magnet.*?)".*?(vip|11x11).*?Uploaded (.*?), Size (.*?), ULed by .*?>(.*?)<.*?right">(\d+)<.*?right">(\d+)</td>`
	INFOURI        = `/torrent/%s`
	INFOREGEXP     = `(?s)Size:.*?\((.*?)&nbsp;Bytes\).*?Uploaded:.*?d>(.*?)</d`
	FILESURI       = `/ajax_details_filelist.php?id=%s`
	FILESREGEXP    = `left">(.*?)</td.*?right">(.*?)<`
)

// This should be treated as a const.
var (
	killHTMLRegexp = regexp.MustCompile(`<.*?>`)
)
