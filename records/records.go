package records

import (
	"crypto/sha512"
	"fmt"
	"hash/fnv"
	"sort"

	rpb "chronicler/records/proto"

	"github.com/lanseg/golang-commons/collections"
)

func nonNull[T any](a *T, b *T) *T {
	if a != nil {
		return a
	}
	return b
}

func nonEmpty(a string, b string) string {
	if a != "" {
		return a
	}
	return b
}

func hashSource(src *rpb.Source) string {
	if src == nil {
		return ""
	}
	checksum := []byte{}
	checksum = append(checksum, []byte(src.SenderId)...)
	checksum = append(checksum, []byte(src.ChannelId)...)
	checksum = append(checksum, []byte(src.MessageId)...)
	checksum = append(checksum, []byte(src.Url)...)
	checksum = append(checksum, byte(src.Type))
	return fmt.Sprintf("%x", sha512.Sum512(checksum))
}

func GetRecordId(record *rpb.Record) string {
	return fmt.Sprintf("%x", sha512.Sum512(
		[]byte(hashSource(record.Source)+hashSource(record.Parent))))
}

func GetRecordSetId(set *rpb.RecordSet) string {
	if set.Id != "" {
		return set.Id
	}
	if set.Request == nil {
		return ""
	}
	if set.Request.Parent != nil {
		return hashSource(set.Request.Parent)
	}
	if set.Request.Source != nil {
		return hashSource(set.Request.Source)
	}
	return ""
}

func fnv32(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func merge[T any](a []T, b []T, hash func(T) uint32, merger func(T, T) T) []T {
	result := []T{}

	if len(a) == 0 && len(b) == 0 {
		return result
	} else if len(a) > 0 && len(b) == 0 {
		return append(a)
	} else if len(a) == 0 && len(b) > 0 {
		return append(b)
	}

	aById := collections.GroupBy(a, hash)
	bById := collections.GroupBy(b, hash)
	for k, v := range bById {
		if _, ok := aById[k]; !ok {
			aById[k] = []T{}
		}
		aById[k] = append(aById[k], v...)
	}

	for _, v := range aById {
		first := v[0]
		for _, f := range v[1:] {
			first = merger(first, f)
		}
		result = append(result, first)
	}
	return result
}

func MergeFiles(a []*rpb.File, b []*rpb.File) []*rpb.File {
	result := merge(a, b, func(f *rpb.File) uint32 {
		return fnv32(f.FileId)
	}, func(af *rpb.File, bf *rpb.File) *rpb.File {
		if af.FileUrl == "" {
			return bf
		}
		return af
	})
	sort.Slice(result, func(a int, b int) bool {
		return result[a].FileId < result[b].FileId
	})
	return result
}

func MergeUserMetadata(a []*rpb.UserMetadata, b []*rpb.UserMetadata) []*rpb.UserMetadata {
	result := merge(a, b, func(u *rpb.UserMetadata) uint32 {
		return fnv32(u.Id)
	}, func(au *rpb.UserMetadata, bu *rpb.UserMetadata) *rpb.UserMetadata {
		return &rpb.UserMetadata{
			Id:       au.Id,
			Username: nonEmpty(au.Username, bu.Username),
			Quotes:   MergeStrings(au.Quotes, bu.Quotes),
		}
	})
	sort.Slice(result, func(a int, b int) bool {
		if result[a].Username < result[b].Username {
			return true
		}
		return result[a].Id < result[b].Id
	})
	return result
}

func MergeRecords(a []*rpb.Record, b []*rpb.Record) []*rpb.Record {
	result := merge(a, b, func(r *rpb.Record) uint32 {
		return fnv32(GetRecordId(r))
	}, func(ar *rpb.Record, br *rpb.Record) *rpb.Record {
		result := &rpb.Record{
			Source: nonNull(ar.Source, br.Source),
			Parent: nonNull(ar.Parent, br.Parent),
			Files:  MergeFiles(ar.Files, br.Files),
			Links:  MergeStrings(ar.Links, br.Links),
		}
		target := ar
		if nonEmpty(ar.TextContent, br.TextContent) == br.TextContent {
			target = br
		}
		result.TextContent = target.TextContent
		result.RawContent = target.RawContent
		result.Time = target.Time
		return result
	})
	sort.Slice(result, func(a int, b int) bool {
		return GetRecordId(result[a]) < GetRecordId(result[b])
	})
	return result
}

func MergeRecordSets(a *rpb.RecordSet, b *rpb.RecordSet) *rpb.RecordSet {
	return &rpb.RecordSet{
		Id:           nonEmpty(a.Id, b.Id),
		Request:      nonNull(a.Request, b.Request),
		Records:      MergeRecords(a.Records, b.Records),
		UserMetadata: MergeUserMetadata(a.UserMetadata, b.UserMetadata),
	}
}

func MergeStrings(a []string, b []string) []string {
	resultSet := collections.NewSet(a)
	resultSet.AddSet(collections.NewSet(b))
	result := resultSet.Values()
	sort.Strings(result)
	return result
}
