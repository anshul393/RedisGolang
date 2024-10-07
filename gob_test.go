package main

import (
	"encoding/gob"
	"os"
	"testing"
)

func TestGob(t *testing.T) {
	l := NewList()
	l.LPush(1)
	l.LPush(2)

	// Encode the list
	f, err := os.Create("demo.gob")
	if err != nil {
		t.Fatal(err)
	}

	// defer f.Close()

	// buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(f)

	if err := enc.Encode(l); err != nil {
		t.Fatal(err)
	}

	f.Close()

	f, err = os.Open("demo.gob")
	if err != nil {
		t.Fatal(err)
	}

	dec := gob.NewDecoder(f)

	list := &List{}
	if err := dec.Decode(list); err != nil {
		t.Fatal(err)
	}

	if list.Len() != 2 {
		t.Fatal("invalid length")
	}

	if list.Head.Value != 2 {
		t.Fatal("expected 2 got :", list.Head.Value)
	}

	if list.Head.Next.Value != 1 {
		t.Fatal("expected 1 got :", list.Head.Next.Value)
	}

}
