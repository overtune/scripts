package meta

import (
	"errors"
	"testing"
)

func TestParseBashBlockWithArgs(t *testing.T) {
	content := []byte(`#!/bin/bash
# @meta
# name: portkill
# description: Kill the process on a TCP port
# category: net
# args:
#   - name: port
#     required: true
#     help: TCP port number
# @end

echo hi
`)
	m, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "portkill" || m.Category != "net" {
		t.Fatalf("bad meta: %+v", m)
	}
	if !m.Documented {
		t.Fatalf("expected Documented true")
	}
	if len(m.Args) != 1 || m.Args[0].Name != "port" || !m.Args[0].Required {
		t.Fatalf("bad args: %+v", m.Args)
	}
}

func TestParseNodeComments(t *testing.T) {
	content := []byte(`#!/usr/bin/env node
// @meta
// name: hello
// description: say hi
// category: dev
// @end
console.log("hi")
`)
	m, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "hello" || m.Category != "dev" {
		t.Fatalf("bad meta: %+v", m)
	}
	if len(m.Args) != 0 {
		t.Fatalf("expected no args, got %+v", m.Args)
	}
}

func TestParseNoBlock(t *testing.T) {
	_, err := Parse([]byte("#!/bin/bash\necho hi\n"))
	if !errors.Is(err, ErrNoMetaBlock) {
		t.Fatalf("expected ErrNoMetaBlock, got %v", err)
	}
}

func TestParseMalformedYAML(t *testing.T) {
	content := []byte(`# @meta
# name: bad
#   description: : : broken
# args: [unterminated
# @end
`)
	if _, err := Parse(content); err == nil {
		t.Fatalf("expected a parse error for malformed yaml")
	}
}
