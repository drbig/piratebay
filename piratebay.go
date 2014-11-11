// See LICENSE.txt for licensing information.

// Package piratebay implements a robust and comprehensive interface to PirateBay.
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
	VERSION = "0.0.1" // current version of the library
)

// Category represents a fully qualified PirateBay's category,
// e.g. "HD - TV shows". Categories are scraped.
type Category struct {
	Group string
	Title string
	ID    string
}

// Ordering represents a PirateBay's ordering. Orderings are scraped.
type Ordering struct {
	Title string
	ID    string
}

// Torrent represents a torrent and all its data that were possible to scrape.
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

// File represents a torrent's file. For convenience size is kept as both
// a int64 and as the scraped string.
type File struct {
	Path    string
	SizeStr string
	SizeInt int64
}

// Site gathers together all the information needed to interact with PirateBay.
// You may have several of this with different settings, each can then be used
// in parallel.
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

// String returns a pretty string representation of a Category.
func (c *Category) String() string {
	return fmt.Sprintf("%s/%s", c.Group, c.Title)
}

// String returns a pretty string representing an Ordering.
func (o *Ordering) String() string {
	return fmt.Sprintf("%s", o.Title)
}

// String returns a pretty string representation of a Torrent.
func (t *Torrent) String() string {
	return fmt.Sprintf("%s (%s)", t.Title, t.ID)
}

// String returns a pretty string representation of a File.
func (f *File) String() string {
	return fmt.Sprintf("%s", f.Path)
}

// String returns a pretty string representation of a Site.
func (s *Site) String() string {
	return fmt.Sprintf("%s", s.RootURI)
}

// InfoURI returns a string containing a URI to PirateBay's page with
// the details of the given Torrent.
func (t *Torrent) InfoURI() string {
	return t.Site.RootURI + fmt.Sprintf(t.Site.InfoURI, t.ID)
}

// GetDetails updates the Torrent data with additional information
// available only by scraping of the Torrent's details page.
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

// GetFiles updates the given Torrent slice of Files, by scraping the file list
// page.
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

// NewSite returns a Site with default settings.
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

// makeRequest makes a HTTP request using Site configuration,
// and returns response body on success.
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

// getInfraData fetches 'infrastructure' data, such as possible orderings and
// available categories.
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

// UpdateCategories updates available Categories.
func (s *Site) UpdateCategories() error {
	data, err := s.getInfraData()
	if err != nil {
		return err
	}
	s.parseCategories(data)
	return nil
}

// UpdateOrderings updates available Orderings.
func (s *Site) UpdateOrderings() error {
	data, err := s.getInfraData()
	if err != nil {
		return err
	}
	s.parseOrderings(data)
	return nil
}

// FindCategory returns the best matching PirateBay's Category for given group
// and category strings.
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

// FindOrderings returns the best matching PirateBay's Ordering for given
// ordering string.
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

// Search executes a search query.
func (s *Site) Search(query string, c *Category, o *Ordering) ([]*Torrent, error) {
	var torrents []*Torrent
	data, err := s.makeRequest(s.RootURI + fmt.Sprintf(s.SearchURI, query, o.ID, c.ID))
	if err != nil {
		return torrents, err
	}
	return s.parseSearch(data), nil
}
