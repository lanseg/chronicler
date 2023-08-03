package records

import (
	"crypto/sha512"
	"fmt"
	"hash/fnv"
	"sort"

	rpb "chronicler/records/proto"

	"github.com/lanseg/golang-commons/collections"
)

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
	sort.Slice(result, func(ia int, ib int) bool {
		return hash(result[ia]) < hash(result[ib])
	})
	return result
}

func MergeFiles(a []*rpb.File, b []*rpb.File) []*rpb.File {
	return merge(a, b, func(f *rpb.File) uint32 {
		return fnv32(f.FileId)
	}, func(af *rpb.File, bf *rpb.File) *rpb.File {
		if af.FileUrl == "" {
			return bf
		}
		return af
	})
}
