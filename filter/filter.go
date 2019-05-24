package filter

import (
	"strings"
	"errors"
	"fmt"
	"regexp"
	"time"
	"strconv"
)

func Filter(pattern string, all []string, m *map[string]bool, set_to bool) error {
	pattern_parts := strings.SplitN(pattern, ":", 2)
	if len(pattern_parts) != 2 {
		return errors.New(fmt.Sprintf("wrong pattern format : %s", pattern))
	}
	switch pattern_parts[0] {
	case "regexp":
		exp, err := regexp.Compile(pattern_parts[1])
		if err != nil {
			return errors.New(fmt.Sprintf("regex pattern compile failue : %s, %v", pattern, err))
		}
		for _, e := range all {
			if exp.MatchString(e) {
				(*m)[e] = set_to
			}
		}
	case "date":
		idx_v := 0
		lt_enabled := false
		gt_enabled := false
		eq_enabled := false
		for i, ch := range pattern_parts[1] {
			if ch == '<' {
				lt_enabled = true
			} else if ch == '>' {
				gt_enabled = true
			} else if ch == '=' {
				eq_enabled = true
			} else {
				idx_v = i
				break
			}
		}
		date_offset, err := strconv.Atoi(pattern_parts[1][idx_v:len(pattern_parts[1])])
		if err != nil {
			return errors.New(fmt.Sprintf("date offset not an integer : %s, %v", pattern, err))
		}
		target_date := time.Now().Add(time.Duration(date_offset) * time.Hour * 24)
		target_date = time.Date(target_date.Year(), target_date.Month(), target_date.Day(), 0, 0, 0, 0, time.Local)
		for _, e := range all {
			this_date, err := time.ParseInLocation("20060102", e, time.Local)
			if err != nil {
				continue
			}
			if lt_enabled && (this_date.Before(target_date)) {
				(*m)[e] = set_to
			}
			if gt_enabled && (this_date.After(target_date)) {
				(*m)[e] = set_to
			}
			if eq_enabled && (this_date.Equal(target_date)) {
				(*m)[e] = set_to
			}
		}
	case "equal":
		for _, e := range all {
			if e == pattern_parts[1] {
				(*m)[e] = set_to
			}
		}
	default:
		return errors.New(fmt.Sprintf("pattern op not found : %s", pattern_parts[0]))
	}
	return nil
}

func MultiFilter(select_patterns []string, except_patterns []string, all []string) (*map[string]bool, error) {
	m := map[string]bool{}
	var err error
	for _, e := range select_patterns {
		err = Filter(e, all, &m, true)
		if err != nil {
			return nil, err
		}
	}
	for _, e := range except_patterns {
		err = Filter(e, all, &m, false)
		if err != nil {
			return nil, err
		}
	}
	return &m, nil
}


