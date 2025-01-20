package common

import (
	"testing"
)

func TestUUID4(t *testing.T) {

	t.Run("UUID generation test", func(t *testing.T) {
		uuid := UUID4()
		if len(uuid) != 36 {
			t.Errorf("Incorrect UUID4: %s", uuid)
		}
	})

}

func TestUUID4For(t *testing.T) {

	for _, tc := range []struct {
		name string
		item interface{}
		want string
	}{
		{
			name: "for string",
			item: "some string a very long string whatever whatever blah",
			want: "b9d1969b-af0c-4f63-9dbe-cb02cf64ee57",
		},
		{
			name: "empty string",
			item: "",
			want: "0880945b-2dab-4be9-9aa0-733055270e4d",
		},
		{
			name: "nil interface",
			item: nil,
			want: "66ab441a-fa75-4277-b806-e89c611ded0e",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := UUID4For(tc.item)
			if result != tc.want {
				t.Errorf("Expected UUID4For(%v) = %s, but got %s", tc.item, tc.want, result)
			}
		})
	}
}
