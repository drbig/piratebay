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
      "all":"0",
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
  cat, err := s.FindCategory("", "other")
  if err == nil {
    t.Errorf("Found non-unique category 'other': %s", cat)
  }
  cat, err = s.FindCategory("video", "other")
  if err != nil {
    t.Errorf("Didn'f find fully-qualified category 'video/other': %s", err)
  }

  // category finder - broken queries
  cat, err = s.FindCategory("", "")
  if err == nil {
    t.Errorf("Didn't fail with empty categories loaded")
  }
  cat, err = s.FindCategory("whatever", "whatever")
  if err == nil {
    t.Errorf("Didn't fail with nonexistent category group")
  }
  cat, err = s.FindCategory("video", "whatever")
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
