// See LICENSE.txt for licensing information.

package piratebay

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseDetails parses and fills in Torrent details.
func (t *Torrent) parseDetails(input string) {
	match := t.Site.InfoREGEXP.FindStringSubmatch(input)
	if len(match) != 3 {
		t.Site.Logger.Printf("Error parsing details for %s\n", t)
		return
	}
	size, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		t.Site.Logger.Printf("Error parsing detailed size for %s from '%s'\n", t, match[1])
	} else {
		t.SizeInt = size
	}
	stamp, err := time.Parse("2006-01-02 15:04:05 MST", match[2])
	if err != nil {
		t.Site.Logger.Printf("Error parsing date for %s from '%s'\n", t, match[2])
	} else {
		t.Uploaded = stamp
	}
	t.detailed = true
	return
}

// parseFile parses and fills in Torrent Files slice.
func (t *Torrent) parseFiles(input string) error {
	for _, match := range t.Site.FilesREGEXP.FindAllStringSubmatch(input, -1) {
		sizeStr := removeHTML(match[2])
		sizeInt := parseSize(match[2])
		if sizeInt < 0 {
			t.Site.Logger.Printf("Error parsing size for %s from '%s'", t, match[2])
		}
		t.Files = append(t.Files, &File{Path: match[1], SizeStr: sizeStr, SizeInt: sizeInt})
	}
	if len(t.Files) < 1 {
		return fmt.Errorf("No files found")
	}
	return nil
}

// parseCategories parses and fills in Site Categories.
func (s *Site) parseCategories(input string) {
	var group string
	s.Categories = make(map[string]map[string]string, 8)
	s.Categories[""] = make(map[string]string, 1)
	for _, match := range s.CategoryREGEXP.FindAllStringSubmatch(input, -1) {
		switch match[1] {
		case "label":
			group = strings.ToLower(match[2])
			if _, present := s.Categories[group]; !present {
				s.Categories[group] = make(map[string]string, 8)
			}
		case "value":
			category := strings.ToLower(match[3])
			s.Categories[group][category] = match[2]
		}
	}

	// group/all IDs, unfortunately hard-coded for now
	s.Categories["audio"]["all"] = "100"
	s.Categories["video"]["all"] = "200"
	s.Categories["applications"]["all"] = "300"
	s.Categories["games"]["all"] = "400"
	s.Categories["porn"]["all"] = "500"
	s.Categories["other"]["all"] = "600"

	return
}

// parseOrderings parses and fills in Site Orderings.
func (s *Site) parseOrderings(input string) {
	s.Orderings = make(map[string]string, 9)
	for _, match := range s.OrderingREGEXP.FindAllStringSubmatch(input, -1) {
		ordering := strings.ToLower(match[2])
		s.Orderings[ordering] = match[1]
	}
	return
}

// parseSearch parses search query results and returns a slice of pointers
// to Torrents.
func (s *Site) parseSearch(input string) []*Torrent {
	var torrents []*Torrent
	var cat Category
	var isVIP bool
	for _, match := range s.SearchREGEXP.FindAllStringSubmatch(input, -1) {
		group := strings.ToLower(match[1])
		catID := match[2]
		category := strings.ToLower(match[3])
		cat = Category{
			Group: group,
			Title: category,
			ID:    catID,
		}
		id := match[4]
		title := match[5]
		magnet := match[6]
		if match[7] == "vip" {
			isVIP = true
		} else {
			isVIP = false
		}
		stamp, err := parseDate(match[8])
		if err != nil {
			s.Logger.Printf("Error parsing date from '%s': %s\n", match[8], err)
		}
		sizeStr := removeHTML(match[9])
		sizeInt := parseSize(match[9])
		if sizeInt < 0 {
			s.Logger.Printf("Error parsing size from '%s'\n", match[9])
		}
		uploader := match[10]
		seeders, err := strconv.Atoi(match[11])
		if err != nil {
			s.Logger.Printf("Error parsing seeders from '%s'\n", match[11])
			seeders = -1
		}
		leechers, err := strconv.Atoi(match[12])
		if err != nil {
			s.Logger.Printf("Error parsing leechers from '%s'\n", match[12])
			leechers = -1
		}
		torrents = append(torrents, &Torrent{
			Site:     *s,
			Category: cat,
			ID:       id,
			Title:    title,
			Magnet:   magnet,
			Uploaded: stamp,
			User:     uploader,
			VIPUser:  isVIP,
			SizeStr:  sizeStr,
			SizeInt:  sizeInt,
			Seeders:  seeders,
			Leechers: leechers,
		})
	}
	return torrents
}

// removeHTML is a helper function that removes HTML from a string.
// It uses the global killHTMLRegexp regexp.
func removeHTML(input string) string {
	output := killHTMLRegexp.ReplaceAllString(input, "")
	return strings.Replace(output, "&nbsp;", " ", -1)
}

// parseSize is a helper function that parses human-readable size
// string into an int64.
func parseSize(input string) int64 {
	input = removeHTML(input)
	multiplier := int64(1)
	parts := strings.Split(input, " ")
	if len(parts) != 2 {
		return -1
	}
	rawSize, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return -1
	}
	switch parts[1] {
	case "TiB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "GiB":
		multiplier = 1024 * 1024 * 1024
	case "MiB":
		multiplier = 1024 * 1024
	case "KiB":
		multiplier = 1024
	}
	return int64(rawSize * float64(multiplier))
}

// makeOffsetDate is a helper function for parsing relative dates.
func makeOffsetDate(ref time.Time, offset time.Duration, hour, minute int) time.Time {
	ref = ref.Add(offset)
	if hour == -1 || minute == -1 {
		return ref
	}
	return time.Date(
		ref.Year(),
		ref.Month(),
		ref.Day(),
		hour,
		minute,
		0, 0,
		ref.Location(),
	)
}

// parseDate is a helper function that parses a string representation
// of a date as used on PirateBay.
func parseDate(input string) (time.Time, error) {
	input = removeHTML(input)
	parts := strings.Split(input, " ")
	reference := time.Now()
	if len(parts) < 2 {
		return reference, fmt.Errorf("Not enough string parts")
	}
	if parts[len(parts)-1] == "ago" {
		mins, err := strconv.Atoi(parts[0])
		if err != nil {
			return reference, fmt.Errorf("Cloudn't parse minutes ago")
		}
		return reference.Add(time.Duration(-mins) * time.Minute), nil
	}
	if parts[0] == "Today" {
		parsed, err := time.Parse("15:04", parts[1])
		if err != nil {
			return reference, fmt.Errorf("Couldn't parse today")
		}
		return makeOffsetDate(reference, 0, parsed.Hour(), parsed.Minute()), nil
	}
	if parts[0] == "Y-day" {
		parsed, err := time.Parse("15:04", parts[1])
		if err != nil {
			return reference, fmt.Errorf("Couldn't parse y-day")
		}
		return makeOffsetDate(reference, -24*time.Hour, parsed.Hour(), parsed.Minute()), nil
	}
	parsed, err := time.Parse("01-02 15:04", input)
	if err == nil {
		return time.Date(
			reference.Year(),
			parsed.Month(),
			parsed.Day(),
			parsed.Hour(),
			parsed.Minute(),
			0, 0,
			reference.Location(),
		), nil
	}
	return time.Parse("01-02 2006", input)
}
