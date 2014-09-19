package piratebay

import (
	//"fmt"
	"testing"
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

func TestOrderingFake(t *testing.T) {
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
    "type": "13",
    "name": "1",
    "uploaded": "3",
    "size": "5",
    "seeders": "7",
    "leechers": "9",
    "uled by": "11",
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
