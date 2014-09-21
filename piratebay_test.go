package piratebay

import (
	"fmt"
	"testing"
	"time"
)

func TestCategoriesFake(t *testing.T) {
	input := `
			<select id="category" name="category" onchange="javascript:setAll();">
       	<option value="0">All</option>
				<optgroup label="Audio">
					<option value="101">Music</option>
					<option value="102">Audio books</option>
					<option value="103">Sound clips</option>
					<option value="104">FLAC</option>
					<option value="199">Other</option>
				</optgroup>
				<optgroup label="Video">
					<option value="201">Movies</option>
					<option value="202">Movies DVDR</option>
					<option value="203">Music videos</option>
					<option value="204">Movie clips</option>
					<option value="205">TV shows</option>
					<option value="206">Handheld</option>
					<option value="207">HD - Movies</option>
					<option value="208">HD - TV shows</option>
					<option value="209">3D</option>
					<option value="299">Other</option>
				</optgroup>
				<optgroup label="Applications">
					<option value="301">Windows</option>
					<option value="302">Mac</option>
					<option value="303">UNIX</option>
					<option value="304">Handheld</option>
					<option value="305">IOS (iPad/iPhone)</option>
					<option value="306">Android</option>
					<option value="399">Other OS</option>
				</optgroup>
`
	catsRaw := map[string]map[string]string{
		"": map[string]string{
			"all": "0",
		},
		"audio": map[string]string{
			"music":       "101",
			"audio books": "102",
		},
		"video": map[string]string{
			"hd - movies":   "207",
			"hd - tv shows": "208",
		},
		"applications": map[string]string{
			"ios (ipad/iphone)": "305",
		},
	}

	// raw categories parsing
	s := NewSite()
	s.parseCategories(input)
	for group, props := range catsRaw {
		if _, present := s.Categories[group]; !present {
			t.Errorf("Raw category group '%s' not found", group)
			continue
		}
		for k, v := range props {
			value, present := s.Categories[group][k]
			if !present {
				t.Errorf("Raw category '%s/%s' not found", group, k)
				continue
			}
			if value != v {
				t.Errorf("Raw category '%s/%s' value mismatch: %s != %s", group, k, value, v)
			}
		}
	}

	// category finder - uniques
	for group, props := range catsRaw {
		if _, present := s.Categories[group]; !present {
			continue
		}
		for k, v := range props {
			cat, err := s.FindCategory("", k)
			if err != nil {
				t.Errorf("Didn't find unique category '%s/%s': %s", group, k, err)
				continue
			}
			if cat.ID != v {
				t.Errorf("Category '%s/%s' value mismatch: %s != %s", group, k, v, cat.ID)
			}
		}
	}

	// category finder - non-uniques
	cat, err := s.FindCategory("", "whatever")
	if err == nil {
		t.Errorf("Found non-unique nonexistent category 'whatever': %s", cat)
	}
	cat, err = s.FindCategory("", "other")
	if err == nil {
		t.Errorf("Found non-unique category 'other': %s", cat)
	}
	cat, err = s.FindCategory("video", "other")
	if err != nil {
		t.Errorf("Didn'f find fully-qualified category 'video/other': %s", err)
	}

	// category finder - broken queries
	_, err = s.FindCategory("audio", "")
	if err == nil {
		t.Errorf("Didn't fail with empty category (1): %s", cat)
	}
	_, err = s.FindCategory("", "")
	if err == nil {
		t.Errorf("Didn't fail with empty category (2): %s", cat)
	}
	_, err = s.FindCategory("whatever", "whatever")
	if err == nil {
		t.Errorf("Didn't fail with nonexistent category group")
	}
	_, err = s.FindCategory("video", "whatever")
	if err == nil {
		t.Errorf("Didn't fail with nonexistent category")
	}

	// category finder - no categories
	s = NewSite()
	_, err = s.FindCategory("", "all")
	if err == nil {
		t.Errorf("Didn't fail with no categories loaded")
	}
}

func TestOrderingsFake(t *testing.T) {
	input := `
	<thead id="tableHead">
		<tr class="header">
			<th><a href="/search/a/0/13/0" title="Order by Type">Type</a></th>
			<th><div class="sortby"><a href="/search/a/0/1/0" title="Order by Name">Name</a> (Order by: <a href="/search/a/0/3/0" title="Order by Uploaded">Uploaded</a>, <a href="/search/a/0/5/0" title="Order by Size">Size</a>, <span style="white-space: nowrap;"><a href="/search/a/0/11/0" title="Order by ULed by">ULed by</a></span>, <a href="/search/a/0/7/0" title="Order by Seeders">SE</a>, <a href="/search/a/0/9/0" title="Order by Leechers">LE</a>)</div><div class="viewswitch"> View: <a href="/switchview.php?view=s">Single</a> / Double&nbsp;</div></th>
			<th><abbr title="Seeders"><a href="/search/a/0/7/0" title="Order by Seeders">SE</a></abbr></th>
			<th><abbr title="Leechers"><a href="/search/a/0/9/0" title="Order by Leechers">LE</a></abbr></th>
		</tr>
	</thead>
`
	ordrRaw := map[string]string{
		"type":     "13",
		"name":     "1",
		"uploaded": "3",
		"size":     "5",
		"seeders":  "7",
		"leechers": "9",
		"uled by":  "11",
	}

	s := NewSite()
	s.parseOrderings(input)

	// raw orderings parsing
	for title, v := range ordrRaw {
		value, present := s.Orderings[title]
		if !present {
			t.Errorf("Raw ordering '%s' not found", title)
			continue
		}
		if value != v {
			t.Errorf("Raw ordering '%s' value mismatch: %s != %s", title, value, v)
		}
	}

	// ordering finder
	for title, v := range ordrRaw {
		ordr, err := s.FindOrdering(title)
		if err != nil {
			t.Errorf("Didn't find ordering '%s'", title)
			continue
		}
		if ordr.ID != v {
			t.Errorf("Ordering '%s' value mismatch: %s != %s", title, ordr.ID, v)
		}
	}

	// ordering finder - broken queries
	_, err := s.FindOrdering("")
	if err == nil {
		t.Errorf("Didn't fail with empty ordering")
	}
	_, err = s.FindOrdering("whatever")
	if err == nil {
		t.Errorf("Didn't fail with nonexistent ordering")
	}

	// ordering finder - no orderings
	s = NewSite()
	_, err = s.FindOrdering("whatever")
	if err == nil {
		t.Errorf("Didn't fail with no orderings loaded")
	}
}

type dateTest struct {
	in     string
	out    string
	broken bool
}

func TestParseDate(t *testing.T) {
	reference := time.Now()
	layout := "01-02 15:04 2006"
	cases := [...]dateTest{
		{`<b>11&nbsp;mins&nbsp;ago</b>`, makeOffsetDate(reference, -11*time.Minute, -1, -1).Format(layout), false},
		{`<b>30&nbsp;mins&nbsp;ago</b>`, makeOffsetDate(reference, -30*time.Minute, -1, -1).Format(layout), false},
		{`Today&nbsp;23:59`, makeOffsetDate(reference, 0, 23, 59).Format(layout), false},
		{`Today&nbsp;14:23`, makeOffsetDate(reference, 0, 14, 23).Format(layout), false},
		{`Today&nbsp;02:11`, makeOffsetDate(reference, 0, 2, 11).Format(layout), false},
		{`Y-day&nbsp;03:00`, makeOffsetDate(reference, -24*time.Hour, 3, 0).Format(layout), false},
		{`05-12&nbsp;11:09`, fmt.Sprintf("05-12 11:09 %d", reference.Year()), false},
		{`12-01&nbsp;23:59`, fmt.Sprintf("12-01 23:59 %d", reference.Year()), false},
		{`07-25&nbsp;2011`, "07-25 00:00 2011", false},
		{`01-01&nbsp;1998`, "01-01 00:00 1998", false},
		{`whatever`, "", true},
		{`abc minutes ago`, "", true},
		{`Today xx:xx`, "", true},
		{`Y-day aa:ee`, "", true},
	}

	for idx, test := range cases {
		parsed, err := parseDate(test.in)
		if err != nil && !test.broken {
			t.Errorf("(%d) Erred: %s", idx+1, err)
			continue
		}
		if err == nil && test.broken {
			t.Errorf("(%d) Didn't err while it should")
			continue
		}
		if !test.broken {
			str := parsed.Format(layout)
			if str != test.out {
				t.Errorf("(%d) Output mismatch: %s != %s", idx+1, str, test.out)
			}
		}
	}
}

func TestTorrentDetailsFake(t *testing.T) {
	input := `
		<dt>Size:</dt>
		<dd>1.37&nbsp;GiB&nbsp;(1469073700&nbsp;Bytes)</dd>

			<dt>Spoken language(s):</dt>
			<dd>English</dd>
		
		
			</dl>
	<dl class="col2">
		<dt>Uploaded:</dt>
		<dd>2008-01-12 00:09:20 GMT</dd>
		<dt>By:</dt>
		<dd>
		<a href="/user/flareup/" title="Browse flareup">flareup</a></dd>
		<dt>Seeders:</dt>
		<dd>0</dd>

		<dt>Leechers:</dt>
		<dd>2</dd>

		<dt>Comments</dt>
		<dd><span id="NumComments">2</span>
				&nbsp;
				</dd>

                <br />
                <dt>Info Hash:</dt><dd>&nbsp;</dd>
                F827F00809B195A168B6B88D1DAC6695E0B93418	</dl>
`
	s := NewSite()
	tr := &Torrent{Site: *s}
	layout := "2006-01-02 15:04:05 MST"
	tr.parseDetails(input)
	if !tr.detailed {
		t.Errorf("Parsing details failed")
		return
	}
	if tr.SizeInt != 1469073700 {
		t.Errorf("Size mismatch: %d != 1469073700", tr.SizeInt)
	}
	if str := tr.Uploaded.Format(layout); str != "2008-01-12 00:09:20 GMT" {
		t.Errorf("Uploaded mismatch: %s != 2008-01-12 00:09:20 GMT", str)
	}
}

type filesTest struct {
	path string
	size int64
}

func TestTorrentFilesFake(t *testing.T) {
	input := `
<div style="background:#FFFFFF none repeat scroll 0%clear:left;margin:0;min-height:0px;padding:0;width:100%;">
<table style="border:0pt none;width:100%;font-family:verdana,Arial,Helvetica,sans-serif;font-size:11px;">
<tr><td align="left">Cowboy Bebop - 23 - Brain Scratch.mp4</td><td align="right">516.27&nbsp;MiB</tr>
<tr><td align="left">Cowboy Bebop - 18 - Speak Like A Child.mp4</td><td align="right">332.62&nbsp;MiB</tr>
<tr><td align="left">Cowboy Bebop - 20 - Pierrot le Fou.mp4</td><td align="right">324.77&nbsp;MiB</tr>
</table>
`
	output := [...]filesTest{
		{`Cowboy Bebop - 23 - Brain Scratch.mp4`, 541348331},
		{`Cowboy Bebop - 18 - Speak Like A Child.mp4`, 348777349},
		{`Cowboy Bebop - 20 - Pierrot le Fou.mp4`, 340546027},
	}

	s := NewSite()
	tr := &Torrent{Site: *s}
	err := tr.parseFiles(input)
	if err != nil {
		t.Errorf("Parsing files failed")
		return
	}
	if len(tr.Files) != len(output) {
		t.Errorf("Parsed files length mismatch: %d != %d", len(tr.Files), len(output))
		return
	}
	for idx, file := range tr.Files {
		if file.Path != output[idx].path {
			t.Errorf("(%d) Path mismatch: %s != %s", file.Path, output[idx].path)
		}
		if file.SizeInt != output[idx].size {
			t.Errorf("(%d) Size mismatch: %d != %d", file.SizeInt, output[idx].size)
		}
	}
}

func TestSearchFake(t *testing.T) {
	input := `
<h2><span>Search results: a</span>&nbsp;Displaying hits from 1 to 30 (approx 999 found)</h2>
<div id="SearchResults"><div id="content">
			<div id="sky-right">
				 <iframe src="//cdn2.adexprt.com/exo_na/sky2.html" width="160" height="600" frameborder="0" scrolling="no"></iframe>
			</div>
	<div id="main-content">

		 <iframe src="//cdn1.adexprt.com/exo_na/center.html" width="728" height="90" frameborder="0" scrolling="no"></iframe>
	<table id="searchResult">
	<thead id="tableHead">
		<tr class="header">
			<th><a href="/search/a/0/13/0" title="Order by Type">Type</a></th>
			<th><div class="sortby"><a href="/search/a/0/1/0" title="Order by Name">Name</a> (Order by: <a href="/search/a/0/3/0" title="Order by Uploaded">Uploaded</a>, <a href="/search/a/0/5/0" title="Order by Size">Size</a>, <span style="white-space: nowrap;"><a href="/search/a/0/11/0" title="Order by ULed by">ULed by</a></span>, <a href="/search/a/0/7/0" title="Order by Seeders">SE</a>, <a href="/search/a/0/9/0" title="Order by Leechers">LE</a>)</div><div class="viewswitch"> View: <a href="/switchview.php?view=s">Single</a> / Double&nbsp;</div></th>
			<th><abbr title="Seeders"><a href="/search/a/0/7/0" title="Order by Seeders">SE</a></abbr></th>
			<th><abbr title="Leechers"><a href="/search/a/0/9/0" title="Order by Leechers">LE</a></abbr></th>
		</tr>
	</thead>
	<tr>
		<td class="vertTh">
			<center>
				<a href="/browse/200" title="More from this category">Video</a><br />
				(<a href="/browse/205" title="More from this category">TV shows</a>)
			</center>
		</td>
		<td>
<div class="detName">			<a href="/torrent/11068355/Would.I.Lie.To.You.S08E02.HDTV.XviD-AFG" class="detLink" title="Details for Would.I.Lie.To.You.S08E02.HDTV.XviD-AFG">Would.I.Lie.To.You.S08E02.HDTV.XviD-AFG</a>
</div>
<a href="magnet:?xt=urn:btih:14cf93721298e1b6694205019fce360dfbcf4164&dn=Would.I.Lie.To.You.S08E02.HDTV.XviD-AFG&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80&tr=udp%3A%2F%2Ftracker.publicbt.com%3A80&tr=udp%3A%2F%2Ftracker.istole.it%3A6969&tr=udp%3A%2F%2Fopen.demonii.com%3A1337" title="Download this torrent using magnet"><img src="/static/img/icon-magnet.gif" alt="Magnet link" /></a>			<a href="//piratebaytorrents.info/11068355/Would.I.Lie.To.You.S08E02.HDTV.XviD-AFG.11068355.TPB.torrent" title="Download this torrent"><img src="/static/img/dl.gif" class="dl" alt="Download" /></a><a href="/user/TvTeam"><img src="/static/img/vip.gif" alt="VIP" title="VIP" style="width:11px;" border='0' /></a><img src="/static/img/11x11p.png" />
			<font class="detDesc">Uploaded <b>11&nbsp;mins&nbsp;ago</b>, Size 244.08&nbsp;MiB, ULed by <a class="detDesc" href="/user/TvTeam/" title="Browse TvTeam">TvTeam</a></font>
		</td>
		<td align="right">0</td>
		<td align="right">0</td>
	</tr>
	<tr>
		<td class="vertTh">
			<center>
				<a href="/browse/600" title="More from this category">Other</a><br />
				(<a href="/browse/699" title="More from this category">Other</a>)
			</center>
		</td>
		<td>
<div class="detName">			<a href="/torrent/11068354/Nayma_-_Responsive_Multi-Purpose_WordPress_Theme" class="detLink" title="Details for Nayma - Responsive Multi-Purpose WordPress Theme">Nayma - Responsive Multi-Purpose WordPress Theme</a>
</div>
<a href="magnet:?xt=urn:btih:55bc118cd26376b888ac1ebc8c2fbbc250c4ea02&dn=Nayma+-+Responsive+Multi-Purpose+WordPress+Theme&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80&tr=udp%3A%2F%2Ftracker.publicbt.com%3A80&tr=udp%3A%2F%2Ftracker.istole.it%3A6969&tr=udp%3A%2F%2Fopen.demonii.com%3A1337" title="Download this torrent using magnet"><img src="/static/img/icon-magnet.gif" alt="Magnet link" /></a>			<a href="//piratebaytorrents.info/11068354/Nayma_-_Responsive_Multi-Purpose_WordPress_Theme.11068354.TPB.torrent" title="Download this torrent"><img src="/static/img/dl.gif" class="dl" alt="Download" /></a><img src="/static/img/icon_image.gif" alt="This torrent has a cover image" title="This torrent has a cover image" /><img src="/static/img/11x11p.png" /><img src="/static/img/11x11p.png" />
			<font class="detDesc">Uploaded <b>15&nbsp;mins&nbsp;ago</b>, Size 23.63&nbsp;MiB, ULed by <a class="detDesc" href="/user/nulledGOD/" title="Browse nulledGOD">nulledGOD</a></font>
		</td>
		<td align="right">0</td>
		<td align="right">0</td>
	</tr>
	<tr>
`
	s := NewSite()
	output := [...]*Torrent{
		&Torrent{
			Site: *s,
			Category: Category{
				Group: "video",
				Title: "tv shows",
				ID:    "205",
			},
			ID:       "11608355",
			Title:    "Would.I.Lie.To.You.S08E02.HDTV.XviD-AFG",
			Magnet:   "magnet:?xt=urn:btih:14cf93721298e1b6694205019fce360dfbcf4164&dn=Would.I.Lie.To.You.S08E02.HDTV.XviD-AFG&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80&tr=udp%3A%2F%2Ftracker.publicbt.com%3A80&tr=udp%3A%2F%2Ftracker.istole.it%3A6969&tr=udp%3A%2F%2Fopen.demonii.com%3A1337",
			Uploaded: time.Now().Add(-11 * time.Minute),
			User:     "TvTeam",
			VIPUser:  true,
			SizeInt:  255936430,
			Seeders:  0,
			Leechers: 0,
		},
		&Torrent{
			Site: *s,
			Category: Category{
				Group: "other",
				Title: "other",
				ID:    "699",
			},
			ID:       "11068354",
			Title:    "Nayma - Responsive Multi-Purpose WordPress Theme",
			Magnet:   "agnet:?xt=urn:btih:55bc118cd26376b888ac1ebc8c2fbbc250c4ea02&dn=Nayma+-+Responsive+Multi-Purpose+WordPress+Theme&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80&tr=udp%3A%2F%2Ftracker.publicbt.com%3A80&tr=udp%3A%2F%2Ftracker.istole.it%3A6969&tr=udp%3A%2F%2Fopen.demonii.com%3A1337",
			Uploaded: time.Now().Add(-15 * time.Minute),
			User:     "nulledGOD",
			VIPUser:  false,
			SizeInt:  24777850,
			Seeders:  0,
			Leechers: 0,
		},
	}
	layout := "01-02 15:04 2006"

	torrents := s.parseSearch(input)
	if len(torrents) != 2 {
		t.Errorf("Parsed torrents length mismatch: %d != 2", len(torrents))
		fmt.Println("ugly dump:")
		for _, tr := range torrents {
			fmt.Println(tr)
		}
		return
	}
	for idx, tr := range output {
		broken := false
		if torrents[idx].Title != tr.Title {
			t.Errorf("Size mismatch %d != %d", torrents[idx].Title, tr.Title)
			broken = true
		}
		if torrents[idx].SizeInt != tr.SizeInt {
			t.Errorf("Size mismatch %d != %d", torrents[idx].SizeInt, tr.SizeInt)
			broken = true
		}
		if torrents[idx].Uploaded.Format(layout) != tr.Uploaded.Format(layout) {
			t.Errorf("Uploaded mismatch %s != %s", torrents[idx].Uploaded, tr.Uploaded)
			broken = true
		}
		if torrents[idx].VIPUser != tr.VIPUser {
			t.Errorf("VIPUser mismatch %d != %d", torrents[idx].VIPUser, tr.VIPUser)
			broken = true
		}
		if torrents[idx].Category.ID != tr.Category.ID {
			t.Errorf("Category.ID mismatch %d != %d", torrents[idx].Category.ID, tr.Category.ID)
			broken = true
		}
		if broken {
			torrentFullDump(tr)
		}
	}
}

func torrentFullDump(t *Torrent) {
	fmt.Println("Category.Group: ", t.Category.Group)
	fmt.Println("Category.Title: ", t.Category.Title)
	fmt.Println("Category.ID:    ", t.Category.ID)
	fmt.Println("ID:             ", t.ID)
	fmt.Println("Title:          ", t.Title)
	fmt.Println("Magnet:         ", t.Magnet)
	fmt.Println("Uploaded:       ", t.Uploaded)
	fmt.Println("User:           ", t.User)
	fmt.Println("VIPUser:        ", t.VIPUser)
	fmt.Println("SizeStr:        ", t.SizeStr)
	fmt.Println("SizeInt:        ", t.SizeInt)
	fmt.Println("Seeders:        ", t.Seeders)
	fmt.Println("Leechers:       ", t.Leechers)
	return
}

func TestStringers(t *testing.T) {
	s := NewSite()
	if s.String() != fmt.Sprintf("%s", ROOTURI) {
		t.Errorf("Site stringer mismatch")
	}
	c := &Category{Group: "test", Title: "test"}
	if c.String() != "test/test" {
		t.Errorf("Category stringer mismatch")
	}
	o := &Ordering{Title: "test"}
	if o.String() != "test" {
		t.Errorf("Ordering stringer mismatch")
	}
	f := &File{Path: "/test.txt"}
	if f.String() != "/test.txt" {
		t.Errorf("File stringer mismatch")
	}
	torrent := &Torrent{Title: "test", ID: "1"}
	if torrent.String() != "test (1)" {
		t.Errorf("Torrent stringer mismatch")
	}
}

func TestCategoriesReal(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	s := NewSite()
	err := s.UpdateCategories()
	if err != nil {
		t.Errorf("Network error: %s", err)
		return
	}
	if _, p := s.Categories[""]["all"]; !p {
		t.Errorf("Default '/all' category not found?")
	}
	if _, err := s.FindCategory("", "hd - tv shows"); err != nil {
		t.Errorf("Unique category '/hd - tv shows' not found?")
	}
}

func TestOrderingsReal(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	s := NewSite()
	err := s.UpdateOrderings()
	if err != nil {
		t.Errorf("Network error: %s", err)
		return
	}
	if _, p := s.Orderings["name"]; !p {
		t.Errorf("Raw ordering 'name' not found?")
	}
	if _, err := s.FindOrdering("seeders"); err != nil {
		t.Errorf("Ordering 'seeders' not found?")
	}
}
