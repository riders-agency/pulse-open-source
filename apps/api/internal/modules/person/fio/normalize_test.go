package fio

import "testing"

func TestNormalizeEmpty(t *testing.T) {
	got := Normalize("")
	if got.Mode != "empty" || got.IsValid {
		t.Fatalf("got=%+v", got)
	}
}

func TestNormalizeSingleName(t *testing.T) {
	got := Normalize("Иван")
	if got.Mode != "single_name" || got.UName != "Иван" {
		t.Fatalf("got=%+v", got)
	}
}

func TestNormalizeTwoParts(t *testing.T) {
	got := Normalize("Иван Петров")
	if got.UName != "Иван" || got.USurname != "Петров" {
		t.Fatalf("got=%+v", got)
	}
}

func TestNormalizeSwapNameLast(t *testing.T) {
	got := Normalize("Петров Иван")
	if got.UName != "Иван" || got.USurname != "Петров" || !got.Changed {
		t.Fatalf("got=%+v", got)
	}
}

func TestNormalizeUnknownSurnameFirstBySuffix(t *testing.T) {
	got := Normalize("Махмудов Мехрадж")
	if got.UName != "Мехрадж" || got.USurname != "Махмудов" || !got.Changed {
		t.Fatalf("got=%+v", got)
	}
	if got.Confidence == "low" {
		t.Fatalf("expected non-low confidence, got=%+v", got)
	}
}

func TestNormalizeThreePartsSurnameFirst(t *testing.T) {
	got := Normalize("Петров Иван Сергеевич")
	if got.UName != "Иван" || got.USurname != "Петров" || got.USname != "Сергеевич" || !got.Changed {
		t.Fatalf("got=%+v", got)
	}
}

func TestNormalizeAmbiguousTwoPartsWarns(t *testing.T) {
	got := Normalize("River Stone")
	if got.UName != "River" || got.USurname != "Stone" {
		t.Fatalf("got=%+v", got)
	}
	if got.Confidence != "low" || len(got.Warnings) == 0 {
		t.Fatalf("expected low confidence warning, got=%+v", got)
	}
}
