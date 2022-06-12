package set

import "testing"

func newSet(size int) *HashSet[int] {
	var slice []int
	for i := 0; i < size; i++ {
		slice = append(slice, i)
	}
	for i := 0; i < size; i++ {
		slice = append(slice, i)
	}
	return NewSlice(slice)
}

func TestHashSet_Add(t *testing.T) {
	type test struct {
		valA string
		valB int
	}
	set := New[test]()
	val := test{
		valA: "test",
		valB: 0,
	}
	set.Add(val)
	if _, ok := set.items[val]; !ok {
		t.Errorf("add val: %v, but fail!", val)
	}
}

func TestHashSet_Remove(t *testing.T) {
	set := newSet(10)
	tests := []struct {
		name      string
		deleteVal int
	}{
		{"delete 0", 0},
		{"delete 0 again", 0},
		{"delete 1", 1},
		{"delete 3", 3},
		{"delete 10", 10},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			set.Remove(test.deleteVal)
			if _, ok := set.items[test.deleteVal]; ok {
				t.Errorf("delete %v fail!", test.deleteVal)
			}
		})
	}

}

func TestHashSet_Contains(t *testing.T) {
	set := newSet(10)
	tests := []struct {
		name   string
		val    int
		result bool
	}{
		{"contains 0", 0, true},
		{"contains 1", 1, true},
		{"contains 10", 10, false},
		{"contains 20", 20, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := set.Contains(test.val); got != test.result {
				t.Errorf("got: %v, except: %v", got, test.result)
			}
		})
	}
}

func TestHashSet_Clear(t *testing.T) {
	set := newSet(100)
	if len(set.items) == 0 {
		panic("create set fail!")
	}
	set.Clear()
	if len(set.items) != 0 {
		t.Errorf("clear set fail!")
	}
}

func TestHashSet_Len(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"create 0", 0},
		{"create 1", 1},
		{"create 10", 10},
		{"create 100", 100},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			set := newSet(test.size)
			if got := set.Len(); got != test.size {
				t.Errorf("got: %v, except: %v", got, test.size)
			}
		})
	}
}
