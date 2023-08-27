package keepass

import (
	"path"

	keepass "github.com/tobischo/gokeepasslib/v3"
	"github.com/zostay/go-std/slices"
)

type groupDir struct {
	group *keepass.Group
	dir   string
}

// KeepassWalker represents a tool for walking Keepass records.
type KeepassWalker struct {
	groups  []groupDir
	entries []*keepass.Entry

	currentDir   string
	currentGroup *keepass.Group
	currentEntry *keepass.Entry

	walkEntries bool
}

// Walker creates an iterator for walking through the Keepass database records.
func (k *Keepass) Walker(walkEntries bool) *KeepassWalker {
	w := &KeepassWalker{
		groups: []groupDir{
			{
				group: &k.db.Content.Root.Groups[0],
				dir:   "",
			},
		},
		entries:     []*keepass.Entry{},
		walkEntries: walkEntries,
	}

	return w
}

// pushGroups pushes a pointer to each group onto the open list in reverse
// order.
func (w *KeepassWalker) pushGroups(groups []keepass.Group) {
	for i := len(groups) - 1; i >= 0; i-- {
		thisDir := path.Join(w.currentDir, groups[i].Name)
		w.groups = slices.Push(w.groups, groupDir{&groups[i], thisDir})
	}
}

// pushEntries pushes a pointer to each entry onto the open list in reverse
// order.
func (w *KeepassWalker) pushEntries(entries []keepass.Entry) {
	for i := len(entries) - 1; i >= 0; i-- {
		w.entries = slices.Push(w.entries, &entries[i])
	}
}

// Next returns the next record for iteration. If walkEntries was set to true,
// this will return true if another entry is found in the tree. Otherwise, this
// will return false if another group is found in the tree. Returns false if no
// records are left for iteration.
func (w *KeepassWalker) Next() bool {
	if w.walkEntries {
		return w.nextEntry()
	}
	return w.nextGroup()
}

// nextEntry sets the cursor on the next available entry. If no such entry is
// found, this returns false, otherwise returns true.
func (w *KeepassWalker) nextEntry() bool {
	for len(w.entries) == 0 {
		if len(w.groups) == 0 {
			return false
		}

		var currentGroupDir groupDir
		currentGroupDir, w.groups = slices.Pop(w.groups)
		w.currentGroup = currentGroupDir.group
		w.currentDir = currentGroupDir.dir

		if len(w.currentGroup.Entries) > 0 {
			w.pushGroups(w.currentGroup.Groups)
			w.pushEntries(w.currentGroup.Entries)
			break
		}
	}

	w.currentEntry, w.entries = slices.Pop(w.entries)
	return true
}

// nextGroup sets the cursor on the next available group. If no such group is
// found, this returns false, otherwise returns true.
func (w *KeepassWalker) nextGroup() bool {
	if len(w.groups) == 0 {
		return false
	}

	var currentGroupDir groupDir
	currentGroupDir, w.groups = slices.Pop(w.groups)
	w.currentGroup = currentGroupDir.group
	w.currentDir = currentGroupDir.dir

	w.pushGroups(w.currentGroup.Groups)
	return true
}

// Entry retrieves the current entry to inspect during iteration.
func (w *KeepassWalker) Entry() *keepass.Entry {
	return w.currentEntry
}

// Group retrieves the current group to inspect during iteration.
func (w *KeepassWalker) Group() *keepass.Group {
	return w.currentGroup
}

// Dir retrieves the name of the location of the current group as a path.
func (w *KeepassWalker) Dir() string {
	return w.currentDir
}
