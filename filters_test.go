// See LICENSE.txt for licensing information.

package piratebay

import (
	"io/ioutil"
	"log"
	"testing"
)

type filterTest struct {
	call   []string
	fails  bool
	input  []*Torrent
	outlen int
}

func TestFilterSeeders(t *testing.T) {
	cases := [...]filterTest{
		{[]string{"seeders:x:broken"}, true, nil, 0},
		{[]string{"seeders:broken:1"}, true, nil, 0},
		{
			[]string{"seeders:min:2"},
			false,
			[]*Torrent{
				&Torrent{Seeders: 3},
				&Torrent{Seeders: 2},
				&Torrent{Seeders: 1},
			},
			2,
		},
		{
			[]string{"seeders:max:2"},
			false,
			[]*Torrent{
				&Torrent{Seeders: 3},
				&Torrent{Seeders: 2},
				&Torrent{Seeders: 1},
			},
			2,
		},
		{
			[]string{"seeders:max:2", "seeders:min:2"},
			false,
			[]*Torrent{
				&Torrent{Seeders: 3},
				&Torrent{Seeders: 2},
				&Torrent{Seeders: 1},
			},
			1,
		},
	}

	for idx, test := range cases {
		fs, err := SetupFilters(test.call)
		if (err != nil) == !test.fails {
			t.Errorf("(%d) Couldn't setup filter '%s'", idx+1, test.call)
		} else {
			if test.input != nil {
				res := ApplyFilters(test.input, fs)
				if len(res) != test.outlen {
					t.Errorf("(%d) Output length mismatch: %d != %d", idx+1, len(res), test.outlen)
				}
			}
		}
	}
}

func TestFilterLeechers(t *testing.T) {
	cases := [...]filterTest{
		{[]string{"leechers:x:broken"}, true, nil, 0},
		{[]string{"leechers:broken:1"}, true, nil, 0},
		{
			[]string{"leechers:min:100"},
			false,
			[]*Torrent{
				&Torrent{Leechers: 232},
				&Torrent{Leechers: 13},
				&Torrent{Leechers: 123},
				&Torrent{Leechers: 12},
				&Torrent{Leechers: 113},
			},
			3,
		},
		{
			[]string{"leechers:max:100"},
			false,
			[]*Torrent{
				&Torrent{Leechers: 232},
				&Torrent{Leechers: 13},
				&Torrent{Leechers: 123},
				&Torrent{Leechers: 12},
				&Torrent{Leechers: 113},
			},
			2,
		},
		{
			[]string{"leechers:max:150", "leechers:min:100"},
			false,
			[]*Torrent{
				&Torrent{Leechers: 232},
				&Torrent{Leechers: 13},
				&Torrent{Leechers: 123},
				&Torrent{Leechers: 12},
				&Torrent{Leechers: 113},
			},
			2,
		},
	}

	for idx, test := range cases {
		fs, err := SetupFilters(test.call)
		if (err != nil) == !test.fails {
			t.Errorf("(%d) Couldn't setup filter '%s'", idx+1, test.call)
		} else {
			if test.input != nil {
				res := ApplyFilters(test.input, fs)
				if len(res) != test.outlen {
					t.Errorf("(%d) Output length mismatch: %d != %d", idx+1, len(res), test.outlen)
				}
			}
		}
	}
}

func TestFilterFiles(t *testing.T) {
	pb := NewSite()
	pb.Logger = log.New(ioutil.Discard, "", 0)
	cases := [...]filterTest{
		{[]string{"files:x:("}, true, nil, 0},
		{[]string{"files:x:x"}, true, nil, 0},
		{
			[]string{"files:exclude:.*\\.iso"},
			false,
			[]*Torrent{
				&Torrent{
					Site: *pb,
					Files: []*File{
						&File{Path: "something.iso"},
						&File{Path: "whatever.txt"},
					},
				},
				&Torrent{
					Site: *pb,
					Files: []*File{
						&File{Path: "something.img"},
						&File{Path: "whatever.nfo"},
					},
				},
				&Torrent{
					Site: *pb,
					Files: []*File{
						&File{Path: "something.exe"},
						&File{Path: "whatever.zip"},
					},
				},
			},
			2,
		},
		{
			[]string{"files:include:.*\\.iso"},
			false,
			[]*Torrent{
				&Torrent{
					Site: *pb,
					Files: []*File{
						&File{Path: "something.iso"},
						&File{Path: "whatever.txt"},
					},
				},
				&Torrent{
					Site: *pb,
					Files: []*File{
						&File{Path: "something.img"},
						&File{Path: "whatever.nfo"},
					},
				},
				&Torrent{
					Site: *pb,
					Files: []*File{
						&File{Path: "something.exe"},
						&File{Path: "whatever.zip"},
					},
				},
			},
			1,
		},
	}

	for idx, test := range cases {
		fs, err := SetupFilters(test.call)
		if (err != nil) == !test.fails {
			t.Errorf("(%d) Couldn't setup filter '%s'", idx+1, test.call)
		} else {
			if test.input != nil {
				res := ApplyFilters(test.input, fs)
				if len(res) != test.outlen {
					t.Errorf("(%d) Output length mismatch: %d != %d", idx+1, len(res), test.outlen)
				}
			}
		}
	}
}
