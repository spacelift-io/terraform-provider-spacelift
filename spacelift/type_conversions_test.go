package spacelift

import "testing"

func TestListToOptionalStringListEmptyListClearsField(t *testing.T) {
	t.Run("unset stays nil", func(t *testing.T) {
		if got := listToOptionalStringList(nil); got != nil {
			t.Fatalf("unset list should stay nil, got %v", *got)
		}
	})

	t.Run("empty list is sent so the remote field can be cleared", func(t *testing.T) {
		got := listToOptionalStringList([]any{})
		if got == nil {
			t.Fatal("empty list returned nil, so the field is omitted and a previously configured value cannot be cleared")
		}
		if len(*got) != 0 {
			t.Fatalf("expected an empty list, got %v", *got)
		}
	})

	t.Run("non-empty list is preserved", func(t *testing.T) {
		got := listToOptionalStringList([]any{"a", "b"})
		if got == nil || len(*got) != 2 || (*got)[0] != "a" || (*got)[1] != "b" {
			t.Fatalf("unexpected list: %v", got)
		}
	})
}
