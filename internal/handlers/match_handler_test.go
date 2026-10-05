package handlers

import "testing"

func TestJoinAnswers(t *testing.T) {
	got := JoinAnswers([]string{"Mi comida favorita es el asado", " Soy hincha de River Plate. ", "Rap."})
	want := "Mi comida favorita es el asado. Soy hincha de River Plate. Rap."
	if got != want {
		t.Fatalf("JoinAnswers() = %q, want %q", got, want)
	}
}
