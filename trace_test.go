package tab_test

import (
	"context"
	"testing"

	"github.com/devigned/tab"
)

func TestFor(t *testing.T) {
	tab.For(context.Background()).Debug("should no op")
}
