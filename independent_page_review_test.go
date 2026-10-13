package envai

import (
	"testing"
)

func TestIndependentNullCompletionPage(t *testing.T) {
	c := &Client{}
	if err := c.validateNegotiatedJSON("listCompletions", []byte(`{"data":null}`)); err == nil {
		t.Fatal("data:null page accepted as empty")
	}
}
