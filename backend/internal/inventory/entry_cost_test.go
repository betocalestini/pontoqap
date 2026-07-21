package inventory

import "testing"

func TestEntryUnitCostCents(t *testing.T) {
	tests := []struct {
		name    string
		paid    int64
		other   int64
		qty     int
		want    int64
		wantErr bool
	}{
		{"even split", 1000, 0, 10, 100, false},
		{"with other expenses", 1000, 500, 10, 150, false},
		{"rounds up", 100, 0, 3, 33, false},
		{"zero total", 0, 0, 5, 0, false},
		{"invalid qty", 100, 0, 0, 0, true},
		{"negative paid", -1, 0, 1, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := EntryUnitCostCents(tc.paid, tc.other, tc.qty)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}
}
