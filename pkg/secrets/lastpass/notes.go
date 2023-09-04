package lastpass

import (
	"bufio"
	"fmt"
	"strings"
)

func parseNotes(notes string) map[string]string {
	lines := bufio.NewScanner(strings.NewReader(notes))
	res := map[string]string{}
	for lines.Scan() {
		line := lines.Text()
		vs := strings.SplitN(line, ":", 2)

		if len(vs) != 2 {
			return nil
		}

		res[vs[0]] = vs[1]
	}
	return res
}

func writeNotes(typ string, notes map[string]string) string {
	_, hasNotes := notes["Notes"]
	if typ == "" && len(notes) == 1 && hasNotes {
		return notes["Notes"]
	}

	res := &strings.Builder{}
	fmt.Fprintf(res, "NoteType:%s\n", typ)
	for k, v := range notes {
		fmt.Fprintf(res, "%s:%s\n", k, v)
	}
	return res.String()
}
