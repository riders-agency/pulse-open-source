package main

import "testing"

func TestExtractAPIPathsStopsAtEscapedLineBreak(t *testing.T) {
	paths := extractAPIPaths(`code: 'POST /api/v1/site-chat-conversations/{id}/messages\nContent-Type: application/json'`)
	if len(paths) != 1 || paths[0] != "/api/v1/site-chat-conversations/{id}/messages" {
		t.Fatalf("paths=%q", paths)
	}
}
