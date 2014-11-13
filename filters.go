// See LICENSE.txt for licensing information.

package piratebay

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	FILTERSEP = ":" // separator for filter arguments
)

// FilterFunc is a signature of a Torrent filtering function.
// The function should return true if the Torrent passes the filter
// conditions, false otherwise.
type FilterFunc func(*Torrent) bool

// Filter represents a named filter.
type Filter struct {
	Name string
	Args string
	Desc string
	Init func(string, string) (FilterFunc, error)
}

var (
	filters map[string]Filter // global slice of filters
)

func init() {
	filters = make(map[string]Filter)
	initFilters()
}

// String returns a pretty string representation of a Filter.
// The returned string contains the 'signature' needed to use
// the filter.
func (f Filter) String() string {
	return fmt.Sprintf("%s(%s) - %s", f.Name, f.Args, f.Desc)
}

// RegisterFilter registers the Filter in the global slice of Filters.
func RegisterFilter(f Filter) {
	if _, present := filters[f.Name]; present {
		panic(fmt.Sprintf("Filter '%s' already registered", f.Name))
	}
	filters[f.Name] = f
}

// SetupFilters parses a string description of Filters and returns
// a corresponding slice of initialised FilterFuncs.
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

// GetFilters returns a slice of currently registered Filters.
func GetFilters() []Filter {
	var fs []Filter
	for _, f := range filters {
		fs = append(fs, f)
	}
	return fs
}

// ApplyFilters filters a slice of Torrents by applying a slice of FilterFuncs.
func ApplyFilters(trs []*Torrent, fs []FilterFunc) []*Torrent {
	var out []*Torrent
	for ti := 0; ti < len(trs); ti++ {
		toPass := len(fs)
		for fi := 0; fi < len(fs); fi++ {
			if !fs[fi](trs[ti]) {
				break
			}
			toPass -= 1
		}
		if toPass == 0 {
			out = append(out, trs[ti])
		}
	}
	return out
}

// initFilters registers the currently defined Filters.
// This may change in the future.
// TODO: Figure out a nicer way of adding Filters.
func initFilters() {
	RegisterFilter(Filter{
		Name: "seeders",
		Args: "min - int | max - int",
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

	RegisterFilter(Filter{
		Name: "leechers",
		Args: "min - int | max - int",
		Desc: "Filter by torrent min/max leechers",
		Init: func(arg, value string) (FilterFunc, error) {
			valueInt, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			switch arg {
			case "min":
				return func(tr *Torrent) bool {
					return (tr.Leechers >= valueInt)
				}, nil
			case "max":
				return func(tr *Torrent) bool {
					return tr.Leechers <= valueInt
				}, nil
			default:
				return nil, fmt.Errorf("Unknown arg '%s'", arg)
			}
		},
	})

	RegisterFilter(Filter{
		Name: "size",
		Args: "min - int | max - int",
		Desc: "Filter by torrent total min/max size",
		Init: func(arg, value string) (FilterFunc, error) {
			valueInt, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}
			switch arg {
			case "min":
				return func(tr *Torrent) bool {
					return (tr.SizeInt >= valueInt)
				}, nil
			case "max":
				return func(tr *Torrent) bool {
					return tr.SizeInt <= valueInt
				}, nil
			default:
				return nil, fmt.Errorf("Unknown arg '%s'", arg)
			}
		},
	})

	RegisterFilter(Filter{
		Name: "files",
		Args: "include - regexp | exclude - regexp",
		Desc: "Filter by torrent files' name include/exclude",
		Init: func(arg, value string) (FilterFunc, error) {
			regexp, err := regexp.Compile(value)
			if err != nil {
				return nil, err
			}
			switch arg {
			case "exclude":
				return func(tr *Torrent) bool {
					tr.GetFiles()
					ok := true
					for _, f := range tr.Files {
						if regexp.MatchString(f.Path) {
							ok = false
							break
						}
					}
					return ok
				}, nil
			case "include":
				return func(tr *Torrent) bool {
					tr.GetFiles()
					ok := false
					for _, f := range tr.Files {
						if regexp.MatchString(f.Path) {
							ok = true
							break
						}
					}
					return ok
				}, nil
			default:
				return nil, fmt.Errorf("Unknown arg '%s'", arg)
			}
		},
	})


}
