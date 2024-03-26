package records

import (
	"sort"

	rpb "chronicler/records/proto"
)

// CompareRecords compare two records and returns -1 if a<b, 0 if a == b, 1 if a > b
func CompareRecords(a *rpb.Record, b *rpb.Record, srt *rpb.Sorting) int {
	if srt.Order == rpb.Sorting_ASC {
		return directCompareRecords(a, b, srt)
	}
	return directCompareRecords(b, a, srt)
}

func directCompareRecords(a *rpb.Record, b *rpb.Record, srt *rpb.Sorting) int {
	// Non-nil before nil
	if a == nil && b == nil {
		return 0
	}
	if a != nil && b == nil {
		return -1
	}
	if a == nil && b != nil {
		return 1
	}

	// Parent before child
	if a.Parent != nil && b.Parent != nil {
		if hashSource(a.Parent) == hashSource(b.Source) {
			return 1
		}
		if hashSource(b.Parent) == hashSource(a.Source) {
			return -1
		}
	}

	// Comparison according to srt
	switch srt.Field {
	case rpb.Sorting_FETCH_TIME:
		return int(a.FetchTime - b.FetchTime)
	case rpb.Sorting_CREATE_TIME:
		return int(a.Time - b.Time)
	}

	// If nothing happened then compare by create time
	return int(a.Time - b.Time)
}

func SortRecords(r []*rpb.Record, sorting *rpb.Sorting) []*rpb.Record {
	if r == nil {
		return r
	}

	sort.Slice(r, func(i int, j int) bool {
		return CompareRecords(r[i], r[j], sorting) < 0
	})
	return r
}

func SortRecordSets(rs []*rpb.RecordSet, sorting *rpb.Sorting) []*rpb.RecordSet {
	if rs == nil {
		return rs
	}
	for _, rset := range rs {
		rset.Records = SortRecords(rset.Records, sorting)
	}
	sort.Slice(rs, func(i int, j int) bool {
		if len(rs[i].Records) == 0 && len(rs[j].Records) == 0 {
			return rs[i].Id < rs[j].Id
		}
		if len(rs[i].Records) == 0 && len(rs[j].Records) != 0 {
			return true
		}
		if len(rs[i].Records) != 0 && len(rs[j].Records) == 0 {
			return false
		}
		return CompareRecords(rs[i].Records[0], rs[j].Records[0], sorting) < 0
	})
	return rs
}
