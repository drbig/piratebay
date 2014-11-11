// See LICENSE.txt for licensing information.

package piratebay

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

const (
	VERSION   = "0.0.1"
	FILTERSEP = ":"
)

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
