package viewer

import (
	"chronicler/common"
	"chronicler/iferr"
	opb "chronicler/proto"
	"chronicler/storage"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	objectFileName = "snapshot.json"
)

type Viewer struct {
	Root string
}

func formatObject(obj *opb.Object, prefix int) string {
	result := strings.Builder{}
	t := time.Unix(obj.CreatedAt.Seconds, int64(obj.CreatedAt.Nanos))
	result.WriteString(fmt.Sprintf("‣ [%s] %s: ", t.Format("2006-01-02 15:04"), obj.Generator[0].Name))
	for _, c := range obj.Content {
		txt := regexp.MustCompilePOSIX("[\n\t]*").ReplaceAllString(c.Text, "")
		txt = regexp.MustCompilePOSIX("(<br>|</p>)+").ReplaceAllString(txt, "\n")
		txt = regexp.MustCompilePOSIX("<[^>]*>").ReplaceAllString(txt, " ")
		result.WriteString(strings.TrimSpace(txt) + "\n")
	}
	prefixStr := ""
	for range prefix {
		prefixStr += "    "
	}
	lines := strings.Split(common.WrapText(result.String(), 80), "\n")
	for i := range lines {
		lines[i] = prefixStr + "   " + strings.TrimSpace(lines[i])
	}
	from := 0
	to := 0
	total := len(lines)
	for i := 0; i < total/2; i++ {
		if lines[i] == "" && (i == 0 || i == from+1) {
			from = i
		}
		if lines[total-i-1] == "" && (i == 0 || i == to+1) {
			to = i
		}
	}
	return strings.Join(lines[from:total-to-1], "\n")
}

func (v *Viewer) View(id string) error {
	store := storage.BlockStorage{
		Storage: iferr.Exit(storage.NewLocalStorage(filepath.Join(v.Root, id))),
	}

	result := &opb.Snapshot{}
	if err := store.GetObject(&storage.GetRequest{Url: objectFileName}, &result); err != nil {
		return err
	}

	objByParent := map[string]([]*opb.Object){}
	objById := map[string]*opb.Object{}
	for _, obj := range result.Objects {
		if _, ok := objByParent[obj.Parent]; !ok {
			objByParent[obj.Parent] = []*opb.Object{}
		}
		objByParent[obj.Parent] = append(objByParent[obj.Parent], obj)
		objById[obj.Id] = obj
	}
	toVisit := append([]*opb.Object{}, objByParent[""]...)
	for len(toVisit) > 0 {
		obj := toVisit[0]
		toVisit = toVisit[1:]
		prefix := 0

		p := obj.Parent
		for ; p != ""; prefix++ {
			p = objById[p].Parent
		}
		fmt.Println(formatObject(obj, prefix))
		children := objByParent[obj.Id]
		if len(children) == 0 {
			continue
		}
		sort.Slice(children, func(i, j int) bool {
			return children[i].CreatedAt.Seconds < children[j].CreatedAt.Seconds
		})
		toVisit = append(children, toVisit...)
	}
	return nil
}
