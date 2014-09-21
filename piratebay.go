package piratebay

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	VERSION = "0.0.1"
  FILTERSEP = ":"
)

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

var (
	killHTMLRegexp = regexp.MustCompile(`<.*?>`)
)

var (
	filters map[string]Filter
)

type FilterFunc func(*Torrent) bool

type Filter struct {
	Name string
	Args string
	Desc string
	Init func(string, string) (FilterFunc, error)
}

func (f Filter) String() string {
  return fmt.Sprintf("%s(%s) - %s", f.Name, f.Args, f.Desc)
}

func RegisterFilter(f Filter) {
	if _, present := filters[f.Name]; present {
		panic(fmt.Sprintf("Filter '%s' already registered", f.Name))
	}
	filters[f.Name] = f
}

func init() {
	filters = make(map[string]Filter)
	initFilters()
}

func initFilters() {
	RegisterFilter(Filter{
		Name: "seeders",
    Args: "min - int, max - int",
    Desc: "Filter by torrent min/max seeders",
		Init: func(arg, value string) (FilterFunc, error) {
			valueInt, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			switch arg {
			case "min":
				return func(tr *Torrent) bool {
					return (tr.Seeders >= valueInt)
				}, nil
			case "max":
				return func(tr *Torrent) bool {
					return tr.Seeders <= valueInt
				}, nil
			default:
				return nil, fmt.Errorf("Unknown arg '%s'", arg)
			}
		},
	})
}

func SetupFilters(fs []string) ([]FilterFunc, error) {
  var out []FilterFunc
  for _, f := range fs {
    parts := strings.Split(f, FILTERSEP)
    switch len(parts) {
    case 1:
      filter, present := filters[parts[0]]
      if !present {
        return out, fmt.Errorf("Filter '%s' not found", parts[0])
      }
      fFunc, err := filter.Init("", "")
      if err != nil {
        return out, fmt.Errorf("Setup failed for filter '%s': %s", parts[0], err)
      }
      out = append(out, fFunc)
    case 3:
      filter, present := filters[parts[0]]
      if !present {
        return out, fmt.Errorf("Filter '%s' not found", parts[0])
      }
      fFunc, err := filter.Init(parts[1], parts[2])
      if err != nil {
        return out, fmt.Errorf("Setup failed for filter '%s': %s", parts[0], err)
      }
      out = append(out, fFunc)
    default:
      return out, fmt.Errorf("Wrong format of '%s'", f)
    }
  }
  return out, nil
}

func GetFilters() []Filter {
  var fs []Filter
  for _, f := range filters {
    fs = append(fs, f)
  }
  return fs
}

func ApplyFilters(trs []*Torrent, fs []FilterFunc) []*Torrent {
  var out []*Torrent
  for ti := 0; ti < len(trs); ti++ {
    for fi := 0; fi < len(fs); fi++ {
      if !fs[fi](trs[ti]) {
        break
      }
      out = append(out, trs[ti])
    }
  }
  return out
}

type Category struct {
	Group string
	Title string
	ID    string
}

func (c *Category) String() string {
	return fmt.Sprintf("%s/%s", c.Group, c.Title)
}

type Ordering struct {
	Title string
	ID    string
}

func (o *Ordering) String() string {
	return fmt.Sprintf("%s", o.Title)
}

type File struct {
	Path    string
	SizeStr string
	SizeInt int64
}

func (f *File) String() string {
	return fmt.Sprintf("%s", f.Path)
}

type Site struct {
	RootURI        string
	InfraURI       string
	SearchURI      string
	InfoURI        string
	FilesURI       string
	CategoryREGEXP *regexp.Regexp
	OrderingREGEXP *regexp.Regexp
	SearchREGEXP   *regexp.Regexp
	InfoREGEXP     *regexp.Regexp
	FilesREGEXP    *regexp.Regexp
	Categories     map[string]map[string]string
	Orderings      map[string]string
	Client         *http.Client
	Logger         *log.Logger

	infraData string
}

func NewSite() *Site {
	return &Site{
		RootURI:        ROOTURI,
		InfraURI:       INFRAURI,
		SearchURI:      SEARCHURI,
		InfoURI:        INFOURI,
		FilesURI:       FILESURI,
		CategoryREGEXP: regexp.MustCompile(CATEGORYREGEXP),
		OrderingREGEXP: regexp.MustCompile(ORDERINGREGEXP),
		SearchREGEXP:   regexp.MustCompile(SEARCHREGEXP),
		InfoREGEXP:     regexp.MustCompile(INFOREGEXP),
		FilesREGEXP:    regexp.MustCompile(FILESREGEXP),
		Categories:     nil,
		Orderings:      nil,
		Client:         &http.Client{},
		Logger:         log.New(os.Stderr, "DEBUG: ", log.Lshortfile),
	}
}

func (s *Site) String() string {
	return fmt.Sprintf("%s", s.RootURI)
}

func (s *Site) makeRequest(uri string) (string, error) {
	s.Logger.Printf("Making request for %s", uri)
	res, err := s.Client.Get(uri)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("Unsuccessful request for '%s': %d", uri, res.StatusCode)
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *Site) getInfraData() (string, error) {
	if s.infraData != "" {
		s.Logger.Println("Using cached infraData")
		return s.infraData, nil
	}
	data, err := s.makeRequest(s.RootURI + s.InfraURI)
	if err != nil {
		return "", nil
	}
	s.infraData = data
	return data, nil
}

func (s *Site) UpdateCategories() error {
	data, err := s.getInfraData()
	if err != nil {
		return err
	}
	s.parseCategories(data)
	return nil
}

func (s *Site) UpdateOrderings() error {
	data, err := s.getInfraData()
	if err != nil {
		return err
	}
	s.parseOrderings(data)
	return nil
}

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
	return
}

func (s *Site) parseOrderings(input string) {
	s.Orderings = make(map[string]string, 9)
	for _, match := range s.OrderingREGEXP.FindAllStringSubmatch(input, -1) {
		ordering := strings.ToLower(match[2])
		s.Orderings[ordering] = match[1]
	}
	return
}

func (s *Site) FindCategory(group string, category string) (*Category, error) {
	if s.Categories == nil {
		return nil, fmt.Errorf("Categories not loaded")
	}
	if category == "" {
		return nil, fmt.Errorf("Category not specified")
	}
	if group != "" {
		categories, present := s.Categories[group]
		if !present {
			return nil, fmt.Errorf("Category group '%s' not found", group)
		}
		value, present := categories[category]
		if !present {
			return nil, fmt.Errorf("Category '%s/%s' not found", group, category)
		}
		return &Category{
			Group: group,
			Title: category,
			ID:    value,
		}, nil
	}
	var foundCat *Category
	for group, categories := range s.Categories {
		for cat, value := range categories {
			if cat == category {
				if foundCat != nil {
					return nil, fmt.Errorf("Category '%s' is ambiguous, please specify group", category)
				}
				foundCat = &Category{
					Group: group,
					Title: category,
					ID:    value,
				}
			}
		}
	}
	if foundCat == nil {
		return nil, fmt.Errorf("Category '%s' not found", category)
	}
	return foundCat, nil
}

func (s *Site) FindOrdering(ordering string) (*Ordering, error) {
	if s.Orderings == nil {
		return nil, fmt.Errorf("Orderings not loaded")
	}
	if ordering == "" {
		return nil, fmt.Errorf("Ordering not specified")
	}
	value, present := s.Orderings[ordering]
	if !present {
		return nil, fmt.Errorf("Ordering '%s' not found", ordering)
	}
	return &Ordering{
		Title: ordering,
		ID:    value,
	}, nil
}

func (s *Site) Search(query string, c *Category, o *Ordering) ([]*Torrent, error) {
	var torrents []*Torrent
	data, err := s.makeRequest(s.RootURI + fmt.Sprintf(s.SearchURI, query, o.ID, c.ID))
	if err != nil {
		return torrents, err
	}
	return s.parseSearch(data), nil
}

type Torrent struct {
	Site
	Category
	ID       string
	Title    string
	Magnet   string
	Uploaded time.Time
	User     string
	VIPUser  bool
	SizeStr  string
	SizeInt  int64
	Seeders  int
	Leechers int
	Files    []*File

	detailed bool
}

func (t *Torrent) String() string {
	return fmt.Sprintf("%s (%s)", t.Title, t.ID)
}

func (t *Torrent) InfoURI() string {
	return t.Site.RootURI + fmt.Sprintf(t.Site.InfoURI, t.ID)
}

func (t *Torrent) GetDetails() error {
	if t.detailed {
		t.Site.Logger.Println("Torrent already had details")
		return nil
	}
	data, err := t.Site.makeRequest(t.Site.RootURI + fmt.Sprintf(t.Site.InfoURI, t.ID))
	if err != nil {
		return err
	}
	t.parseDetails(data)
	return nil
}

func (t *Torrent) GetFiles() error {
	if len(t.Files) > 0 {
		t.Site.Logger.Println("Torrent already had files")
		return nil
	}
	data, err := t.Site.makeRequest(t.Site.RootURI + fmt.Sprintf(t.Site.FilesURI, t.ID))
	if err != nil {
		return err
	}
	return t.parseFiles(data)
}

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

func removeHTML(input string) string {
	output := killHTMLRegexp.ReplaceAllString(input, "")
	return strings.Replace(output, "&nbsp;", " ", -1)
}

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
