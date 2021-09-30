package git

import (
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/stretchr/testify/assert"
)

var (
	t1, _ = time.Parse(time.RFC3339, "2020-10-12T09:00:00-04:00")
	t2, _ = time.Parse(time.RFC3339, "2020-10-22T16:00:00-04:00")

	JohnDoe = Signature{
		Name:  "John Doe",
		Email: "john@doe.com",
		Time:  t1,
	}

	JaneDoe = Signature{
		Name:  "Jane Doe",
		Email: "jane@doe.com",
		Time:  t2,
	}

	commit1 = Commit{
		Hash:      "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378",
		Author:    JohnDoe,
		Committer: JohnDoe,
		Message:   "foo",
	}

	commit2 = Commit{
		Hash:      "0251a422d2038967eeaaaa5c8aa76c7067fdef05",
		Author:    JaneDoe,
		Committer: JaneDoe,
		Message:   "bar",
	}

	tag1 = Tag{
		Type:   Lightweight,
		Hash:   "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378",
		Name:   "v0.1.0",
		Commit: commit1,
	}

	tag2Message = "Release v0.2.0"
	tag2        = Tag{
		Type: Annotated,
		Hash: "4ff025213432eee78526e2a75f2e043d34962b5a",
		Name: "v0.2.0",
		Tagger: &Signature{
			Name:  "John Doe",
			Email: "john@doe.com",
			Time:  t1,
		},
		Message: &tag2Message,
		Commit:  commit2,
	}
)

func TestSignature(t *testing.T) {
	tests := []struct {
		name            string
		s1, s2          Signature
		expectedBefore  bool
		expectedAfter   bool
		expectedString1 string
		expectedString2 string
	}{
		{
			name:            "Before",
			s1:              JohnDoe,
			s2:              JaneDoe,
			expectedBefore:  true,
			expectedAfter:   false,
			expectedString1: "John Doe <john@doe.com> 2020-10-12T09:00:00-04:00",
			expectedString2: "Jane Doe <jane@doe.com> 2020-10-22T16:00:00-04:00",
		},
		{
			name:            "After",
			s1:              JaneDoe,
			s2:              JohnDoe,
			expectedBefore:  false,
			expectedAfter:   true,
			expectedString1: "Jane Doe <jane@doe.com> 2020-10-22T16:00:00-04:00",
			expectedString2: "John Doe <john@doe.com> 2020-10-12T09:00:00-04:00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedBefore, tc.s1.Before(tc.s2))
			assert.Equal(t, tc.expectedAfter, tc.s1.After(tc.s2))
			assert.Equal(t, tc.expectedString1, tc.s1.String())
			assert.Equal(t, tc.expectedString2, tc.s2.String())
		})
	}
}

func TestToCommit(t *testing.T) {
	tests := []struct {
		name           string
		commitObj      *object.Commit
		expectedCommit Commit
	}{
		{
			name: "OK",
			commitObj: &object.Commit{
				Hash: plumbing.NewHash("25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378"),
				Author: object.Signature{
					Name:  "John Doe",
					Email: "john@doe.com",
					When:  t1,
				},
				Committer: object.Signature{
					Name:  "John Doe",
					Email: "john@doe.com",
					When:  t1,
				},
				Message: "foo",
			},
			expectedCommit: commit1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			commit := toCommit(tc.commitObj)

			assert.Equal(t, tc.expectedCommit, commit)
		})
	}
}

func TestCommit(t *testing.T) {
	tests := []struct {
		name                 string
		c                    Commit
		expectedShortMessage string
		expectedString       string
		expectedText         string
	}{
		{
			name:                 "Commit1",
			c:                    commit1,
			expectedShortMessage: "foo",
			expectedString:       "25aa2bd foo",
			expectedText:         "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378\nAuthor:    John Doe <john@doe.com> 2020-10-12T09:00:00-04:00\nCommitter: John Doe <john@doe.com> 2020-10-12T09:00:00-04:00\nfoo",
		},
		{
			name:                 "Commit2",
			c:                    commit2,
			expectedShortMessage: "bar",
			expectedString:       "0251a42 bar",
			expectedText:         "0251a422d2038967eeaaaa5c8aa76c7067fdef05\nAuthor:    Jane Doe <jane@doe.com> 2020-10-22T16:00:00-04:00\nCommitter: Jane Doe <jane@doe.com> 2020-10-22T16:00:00-04:00\nbar",
		},
		{
			name: "LongMessage",
			c: Commit{
				Hash:      "c414d1004154c6c324bd78c69d10ee101e676059",
				Author:    JohnDoe,
				Committer: JaneDoe,
				Message:   "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
			},
			expectedShortMessage: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore  ...",
			expectedString:       "c414d10 Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore  ...",
			expectedText:         "c414d1004154c6c324bd78c69d10ee101e676059\nAuthor:    John Doe <john@doe.com> 2020-10-12T09:00:00-04:00\nCommitter: Jane Doe <jane@doe.com> 2020-10-22T16:00:00-04:00\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedShortMessage, tc.c.ShortMessage())
			assert.Equal(t, tc.expectedString, tc.c.String())
			assert.Equal(t, tc.expectedText, tc.c.Text())
		})
	}
}

func TestCommit_Comparison(t *testing.T) {
	tests := []struct {
		name           string
		c1, c2         Commit
		expectedEqual  bool
		expectedBefore bool
		expectedAfter  bool
	}{
		{
			name:           "Equal",
			c1:             commit1,
			c2:             commit1,
			expectedEqual:  true,
			expectedBefore: false,
			expectedAfter:  false,
		},
		{
			name:           "Before",
			c1:             commit1,
			c2:             commit2,
			expectedEqual:  false,
			expectedBefore: true,
			expectedAfter:  false,
		},
		{
			name:           "After",
			c1:             commit2,
			c2:             commit1,
			expectedEqual:  false,
			expectedBefore: false,
			expectedAfter:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedEqual, tc.c1.Equal(tc.c2))
			assert.Equal(t, tc.expectedBefore, tc.c1.Before(tc.c2))
			assert.Equal(t, tc.expectedAfter, tc.c1.After(tc.c2))
		})
	}
}

func TestTagType(t *testing.T) {
	tests := []struct {
		name           string
		t              TagType
		expectedString string
	}{
		{
			name:           "Lightweight",
			t:              Lightweight,
			expectedString: "Lightweight",
		},
		{
			name:           "Annotated",
			t:              Annotated,
			expectedString: "Annotated",
		},
		{
			name:           "Invalid",
			t:              TagType(-1),
			expectedString: "Invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedString, tc.t.String())
		})
	}
}

func TestToLightweightTag(t *testing.T) {
	tests := []struct {
		name        string
		ref         *plumbing.Reference
		commitObj   *object.Commit
		expectedTag Tag
	}{
		{
			name: "OK",
			ref: plumbing.NewHashReference(
				plumbing.ReferenceName("v0.1.0"),
				plumbing.NewHash("25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378"),
			),
			commitObj: &object.Commit{
				Hash: plumbing.NewHash("25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378"),
				Author: object.Signature{
					Name:  "John Doe",
					Email: "john@doe.com",
					When:  t1,
				},
				Committer: object.Signature{
					Name:  "John Doe",
					Email: "john@doe.com",
					When:  t1,
				},
				Message: "foo",
			},
			expectedTag: tag1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag := toLightweightTag(tc.ref, tc.commitObj)

			assert.Equal(t, tc.expectedTag, tag)
		})
	}
}

func TestToAnnotatedTag(t *testing.T) {
	t1, _ := time.Parse(time.RFC3339, "2020-10-12T09:00:00-04:00")
	t2, _ := time.Parse(time.RFC3339, "2020-10-22T16:00:00-04:00")

	tests := []struct {
		name        string
		tagObj      *object.Tag
		commitObj   *object.Commit
		expectedTag Tag
	}{
		{
			name: "OK",
			tagObj: &object.Tag{
				Hash: plumbing.NewHash("4ff025213432eee78526e2a75f2e043d34962b5a"),
				Name: "v0.2.0",
				Tagger: object.Signature{
					Name:  "John Doe",
					Email: "john@doe.com",
					When:  t1,
				},
				Message: "Release v0.2.0",
			},
			commitObj: &object.Commit{
				Hash: plumbing.NewHash("0251a422d2038967eeaaaa5c8aa76c7067fdef05"),
				Author: object.Signature{
					Name:  "Jane Doe",
					Email: "jane@doe.com",
					When:  t2,
				},
				Committer: object.Signature{
					Name:  "Jane Doe",
					Email: "jane@doe.com",
					When:  t2,
				},
				Message: "bar",
			},
			expectedTag: tag2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag := toAnnotatedTag(tc.tagObj, tc.commitObj)

			assert.Equal(t, tc.expectedTag, tag)
		})
	}
}

func TestTag(t *testing.T) {
	tests := []struct {
		name           string
		t              Tag
		expectedIsZero bool
		expectedString string
	}{
		{
			name:           "Zero",
			t:              Tag{},
			expectedIsZero: true,
			expectedString: "",
		},
		{
			name:           "Tag1",
			t:              tag1,
			expectedIsZero: false,
			expectedString: "Lightweight 25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 v0.1.0 Commit[25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 foo]",
		},
		{
			name:           "Tag2",
			t:              tag2,
			expectedIsZero: false,
			expectedString: "Annotated 4ff025213432eee78526e2a75f2e043d34962b5a v0.2.0 Commit[0251a422d2038967eeaaaa5c8aa76c7067fdef05 bar]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedIsZero, tc.t.IsZero())
			assert.Equal(t, tc.expectedString, tc.t.String())
		})
	}
}

func TestTag_Comparison(t *testing.T) {
	tests := []struct {
		name           string
		t1, t2         Tag
		expectedEqual  bool
		expectedBefore bool
		expectedAfter  bool
	}{
		{
			name:           "Zero",
			t1:             Tag{},
			t2:             Tag{},
			expectedEqual:  true,
			expectedBefore: false,
			expectedAfter:  false,
		},
		{
			name:           "Equal",
			t1:             tag1,
			t2:             tag1,
			expectedEqual:  true,
			expectedBefore: false,
			expectedAfter:  false,
		},
		{
			name:           "Before",
			t1:             tag1,
			t2:             tag2,
			expectedEqual:  false,
			expectedBefore: true,
			expectedAfter:  false,
		},
		{
			name:           "After",
			t1:             tag2,
			t2:             tag1,
			expectedEqual:  false,
			expectedBefore: false,
			expectedAfter:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedEqual, tc.t1.Equal(tc.t2))
			assert.Equal(t, tc.expectedBefore, tc.t1.Before(tc.t2))
			assert.Equal(t, tc.expectedAfter, tc.t1.After(tc.t2))
		})
	}
}

func TestTags_First(t *testing.T) {
	tests := []struct {
		name        string
		tags        Tags
		f           func(Tag) bool
		expectedTag Tag
		expectedOK  bool
	}{
		{
			name:        "NoTagNoPredicate",
			tags:        nil,
			f:           nil,
			expectedTag: Tag{},
			expectedOK:  false,
		},
		{
			name:        "NoPredicate",
			tags:        Tags{tag2, tag1},
			f:           nil,
			expectedTag: tag2,
			expectedOK:  true,
		},
		{
			name: "Found",
			tags: Tags{tag2, tag1},
			f: func(t Tag) bool {
				return t.Name == "v0.1.0"
			},
			expectedTag: tag1,
			expectedOK:  true,
		},
		{
			name: "NotFound",
			tags: Tags{tag2, tag1},
			f: func(t Tag) bool {
				return t.Name == "v0.3.0"
			},
			expectedTag: Tag{},
			expectedOK:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag, ok := tc.tags.First(tc.f)

			assert.Equal(t, tc.expectedTag, tag)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestTags_Last(t *testing.T) {
	tests := []struct {
		name        string
		tags        Tags
		f           func(Tag) bool
		expectedTag Tag
		expectedOK  bool
	}{
		{
			name:        "NoTagNoPredicate",
			tags:        nil,
			f:           nil,
			expectedTag: Tag{},
			expectedOK:  false,
		},
		{
			name:        "NoPredicate",
			tags:        Tags{tag2, tag1},
			f:           nil,
			expectedTag: tag1,
			expectedOK:  true,
		},
		{
			name: "Found",
			tags: Tags{tag2, tag1},
			f: func(t Tag) bool {
				return t.Name == "v0.2.0"
			},
			expectedTag: tag2,
			expectedOK:  true,
		},
		{
			name: "NotFound",
			tags: Tags{tag2, tag1},
			f: func(t Tag) bool {
				return t.Name == "v0.3.0"
			},
			expectedTag: Tag{},
			expectedOK:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag, ok := tc.tags.Last(tc.f)

			assert.Equal(t, tc.expectedTag, tag)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestTags_Select(t *testing.T) {
	tests := []struct {
		name               string
		tags               Tags
		f                  func(Tag) bool
		expectedSelected   Tags
		expectedUnselected Tags
	}{
		{
			name: "OK",
			tags: Tags{tag2, tag1},
			f: func(t Tag) bool {
				return t.Type == Annotated
			},
			expectedSelected:   Tags{tag2},
			expectedUnselected: Tags{tag1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			selected, unselected := tc.tags.Select(tc.f)

			assert.Equal(t, tc.expectedSelected, selected)
			assert.Equal(t, tc.expectedUnselected, unselected)
		})
	}
}
