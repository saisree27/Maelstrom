package engine

import "testing"

func TestNullMoveIncrementsFiftyMoveClock(t *testing.T) {
	InitializeEverythingExceptTTable()

	b := Board{}
	b.InitStartPos()
	b.plyCnt50 = 73

	b.MakeNullMove()
	if b.plyCnt50 != 74 {
		t.Fatalf("expected null move to increment fifty-move clock to 74, got %d", b.plyCnt50)
	}

	b.UndoNullMove()
	if b.plyCnt50 != 73 {
		t.Fatalf("expected undo null move to restore fifty-move clock to 73, got %d", b.plyCnt50)
	}
}

