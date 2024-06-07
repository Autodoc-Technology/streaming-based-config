package sbc

import (
	"context"
	"testing"
	"time"
)

func TestHolderGetValue(t *testing.T) {
	holder := NewHolder[int](10)
	value := holder.GetValue()
	if value != 10 {
		t.Fail()
	}
}

func TestHolderSetValue(t *testing.T) {
	holder := NewHolder[int](10)
	holder.setValue(20)
	value := holder.GetValue()
	if value != 20 {
		t.Fail()
	}
}

func TestHolderUpdatesNoChange(t *testing.T) {
	holder := NewHolder[int](10)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	updates := holder.Updates(ctx)
	<-updates
	_, ok := <-updates
	if ok {
		t.Fail()
	}
}
