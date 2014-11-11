// See LICENSE.txt for licensing information.

package piratebay

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type FilterFunc func(*Torrent) bool

type Filter struct {
	Name string
	Args string
	Desc string
	Init func(string, string) (FilterFunc, error)
}

var (
	filters map[string]Filter
)

func init() {
	filters = make(map[string]Filter)
	initFilters()
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
