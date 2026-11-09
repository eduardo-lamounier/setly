package stack

import (
	"testing"
)

func TestStackPop(t *testing.T) {
	tests := []struct {
		name  string
		stack Stack[int]
		want  int
	}{
		{"empty stack", New[int](), 0},
		{"non-empty", New(0, 1, 2, 3), 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			startingLen := len(test.stack.data)

			got := test.stack.Pop()

			if got != test.want {
				t.Errorf("stack.Pop() = %d; want %d.", got, test.want)
			}

			if startingLen != 0 && startingLen-1 != len(test.stack.data) {
				t.Errorf(
					"len(test.stack.data) = %d; want %d.",
					len(test.stack.data), startingLen-1,
				)
			}
		})
	}
}

func TestStackTop(t *testing.T) {
	s := New(1, 2)

	got := s.Top()
	want := 2

	if got != want {
		t.Errorf(
			"Expected s.Top() to return %d (the element at the top), but got %d",
			want, got,
		)
	}
}

func TestStackPush(t *testing.T) {
	s := New(10)

	s.Push(20)

	got := s.data[len(s.data)-1]
	want := 20

	if got != want {
		t.Errorf(
			"Expected s.Push(%d) to append %d to the top of the stack, but got %d.",
			want, want, got,
		)
	}
}

func TestStackCount(t *testing.T) {
	s := New(10, 20, 30, 40, 50)

	got := s.Count()
	want := 5

	if got != want {
		t.Errorf("Expected s.Count() to return %d, but got %d.", want, got)
	}
}
