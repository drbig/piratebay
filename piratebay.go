package piratebay

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const (
	ROOTURI        = "http://thepiratebay.org"
	CATEGORYURI    = "/recent"
	CATEGORYREGEXP = "<opt.*? (.*?)=\"(.*?)\">([A-Za-z- ()/]+)?<?"
	SEARCHURI      = "/search/%s/0/%s/%s"
	SEARCHREGEXP   = "<a href=\"/torrent/(\\d+)/.*?\" class=\"detLink\" title=\".*?\">(.*?)</a>.</div>.<a href=\"(.*?)\".*?Uploaded <?b?>?(.*?)<?/?b?>?, Size (.*?), ULed.*?\"right\">(.*?)</td>.*?\"right\">(.*?)</td>"
	INFOURI        = "/ajax_details_filelist.php?id=%s"
	INFOREGEXP     = "<td align=\"left\">(.*?)</td>"
)

type Category struct {
	Group string
	Title string
	ID    string
}

func (c *Category) String() string {
  return fmt.Sprintf("%s/%s", c.Group, c.Title)
}

type File struct {
	Title string
	Size  int64
}

type Site struct {
	RootURI        string
	CategoryURI    string
	SearchURI      string
	InfoURI        string
	CategoryREGEXP *regexp.Regexp
	SearchREGEXP   *regexp.Regexp
	InfoREGEXP     *regexp.Regexp
	Categories     map[string]map[string]string
	Client         *http.Client
}

func NewSite() *Site {
	return &Site{
		RootURI:        ROOTURI,
		CategoryURI:    CATEGORYURI,
		SearchURI:      SEARCHURI,
		InfoURI:        INFOURI,
		CategoryREGEXP: regexp.MustCompile(CATEGORYREGEXP),
		SearchREGEXP:   regexp.MustCompile(SEARCHREGEXP),
		InfoREGEXP:     regexp.MustCompile(INFOREGEXP),
		Categories:     nil,
		Client:         &http.Client{},
	}
}

func (s *Site) String() string {
	return fmt.Sprintf("%s", s.RootURI)
}

func (s *Site) UpdateCategories() error {
	res, err := s.Client.Get(s.RootURI + s.CategoryURI)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("Unsuccessful request: %d", res.StatusCode)
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	s.parseCategories(string(data))
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
      ID: value,
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
          ID: value,
        }
      }
    }
  }
  return foundCat, nil
}

type Torrent struct {
	Site
	Category
	ID       string
	Title    string
	Magnet   string
	Uploaded string
	User     string
	VIPUser  bool
	Size     int64
	Seeders  int
	Leechers int
	Files    []File
}
