package filter

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/pkg/wildcard"
)

// checkFilter tries to match filter against announce
func (s *service) checkFilter(filter domain.Filter, announce domain.Announce) bool {
	announce.TorrentName = cleanReleaseName(announce.TorrentName)

	if !filter.Enabled {
		return false
	}

	if filter.Scene && announce.Scene != filter.Scene {
		return false
	}

	if filter.Freeleech && announce.Freeleech != filter.Freeleech {
		return false
	}

	if filter.FreeleechPercent != "" && !checkFreeleechPercent(announce.FreeleechPercent, filter.FreeleechPercent) {
		return false
	}

	if filter.Shows != "" && !checkFilterStrings(announce.TorrentName, filter.Shows) {
		return false
	}

	//if filter.Seasons != "" && !checkFilterStrings(announce.TorrentName, filter.Seasons) {
	//	return false
	//}
	//
	//if filter.Episodes != "" && !checkFilterStrings(announce.TorrentName, filter.Episodes) {
	//	return false
	//}

	// matchRelease
	if filter.MatchReleases != "" && !checkFilterStrings(announce.TorrentName, filter.MatchReleases) {
		return false
	}

	if filter.MatchReleaseGroups != "" && !checkFilterStrings(announce.TorrentName, filter.MatchReleaseGroups) {
		return false
	}

	if filter.ExceptReleaseGroups != "" && checkFilterStrings(announce.TorrentName, filter.ExceptReleaseGroups) {
		return false
	}

	if filter.MatchUploaders != "" && !checkFilterStrings(announce.Uploader, filter.MatchUploaders) {
		return false
	}

	if filter.ExceptUploaders != "" && checkFilterStrings(announce.Uploader, filter.ExceptUploaders) {
		return false
	}

	if len(filter.Resolutions) > 0 && !checkFilterSlice(announce.TorrentName, filter.Resolutions) {
		return false
	}

	if len(filter.Codecs) > 0 && !checkFilterSlice(announce.TorrentName, filter.Codecs) {
		return false
	}

	if len(filter.Sources) > 0 && !checkFilterSlice(announce.TorrentName, filter.Sources) {
		return false
	}

	if len(filter.Containers) > 0 && !checkFilterSlice(announce.TorrentName, filter.Containers) {
		return false
	}

	if filter.Years != "" && !checkFilterStrings(announce.TorrentName, filter.Years) {
		return false
	}

	if filter.MatchCategories != "" && !checkFilterStrings(announce.Category, filter.MatchCategories) {
		return false
	}

	if filter.ExceptCategories != "" && checkFilterStrings(announce.Category, filter.ExceptCategories) {
		return false
	}

	if filter.Tags != "" && !checkFilterStrings(announce.Tags, filter.Tags) {
		return false
	}

	if filter.ExceptTags != "" && checkFilterStrings(announce.Tags, filter.ExceptTags) {
		return false
	}

	return true
}

func checkFilterSlice(name string, filterList []string) bool {
	name = strings.ToLower(name)

	for _, filter := range filterList {
		filter = strings.ToLower(filter)
		filter = strings.Trim(filter, " ")
		// check if line contains * or ?, if so try wildcard match, otherwise try substring match
		a := strings.ContainsAny(filter, "?|*")
		if a {
			match := wildcard.Match(filter, name)
			if match {
				return true
			}
		} else {
			b := strings.Contains(name, filter)
			if b {
				return true
			}
		}
	}

	return false
}

func checkFilterStrings(name string, filterList string) bool {
	filterSplit := strings.Split(filterList, ",")
	name = strings.ToLower(name)

	for _, s := range filterSplit {
		s = strings.ToLower(s)
		s = strings.Trim(s, " ")
		// check if line contains * or ?, if so try wildcard match, otherwise try substring match
		a := strings.ContainsAny(s, "?|*")
		if a {
			match := wildcard.Match(s, name)
			if match {
				return true
			}
		} else {
			b := strings.Contains(name, s)
			if b {
				return true
			}
		}

	}

	return false
}

func checkFreeleechPercent(announcePercent string, filterPercent string) bool {
	filters := strings.Split(filterPercent, ",")

	// remove % and trim spaces
	announcePercent = strings.Replace(announcePercent, "%", "", -1)
	announcePercent = strings.Trim(announcePercent, " ")

	announcePercentInt, err := strconv.ParseInt(announcePercent, 10, 32)
	if err != nil {
		return false
	}

	for _, s := range filters {
		s = strings.Replace(s, "%", "", -1)
		s = strings.Trim(s, " ")

		if strings.Contains(s, "-") {
			minMax := strings.Split(s, "-")

			//compare, err := strconv.ParseInt(announcePercent, 10, 32)
			//if err != nil {
			//	return false
			//}

			// to int
			min, err := strconv.ParseInt(minMax[0], 10, 32)
			if err != nil {
				return false
			}

			max, err := strconv.ParseInt(minMax[1], 10, 32)
			if err != nil {
				return false
			}

			if min > max {
				// handle error
				return false
			} else {
				// if announcePercent is greater than min and less than max return true
				if announcePercentInt >= min && announcePercentInt <= max {
					return true
				}
			}
		}

		filterPercentInt, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return false
		}

		if filterPercentInt == announcePercentInt {
			return true
		}
	}

	return false
}

func cleanReleaseName(input string) string {
	// Make a Regex to say we only want letters and numbers
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		//log.Fatal(err)
	}
	processedString := reg.ReplaceAllString(input, " ")

	return processedString
}
