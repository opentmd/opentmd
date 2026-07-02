package agent

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
)

const (
	loopThreshold     = 3
	loopHardThreshold = 6
	loopWindow        = 32
)

var stateChangingTools = map[string]bool{
	"edit_file":       true,
	"write_file":      true,
	"search_replace":  true,
}

type loopEntry struct {
	key        string
	outputHash uint64
	success    bool
}

type loopGuard struct {
	recent []loopEntry
}

func newLoopGuard() *loopGuard {
	return &loopGuard{}
}

func (g *loopGuard) check(name, args string) (blocked bool, msg string) {
	key := loopKey(name, args)
	var matches []loopEntry
	for _, e := range g.recent {
		if e.key == key {
			matches = append(matches, e)
		}
	}

	if len(matches) >= loopHardThreshold-1 {
		return true, fmt.Sprintf(
			"[Loop guard] `%s` with identical arguments has run %d time(s) already in this turn. "+
				"Output varies slightly each run but you keep issuing the same query — that's a loop. "+
				"Try a different approach: change the command, edit code to make progress, or stop and ask the user.",
			name, len(matches),
		)
	}

	if len(matches) < loopThreshold-1 {
		return false, ""
	}

	first := matches[0]
	allSame := true
	for _, e := range matches {
		if e.outputHash != first.outputHash || e.success != first.success {
			allSame = false
			break
		}
	}
	if !allSame {
		return false, ""
	}

	return true, fmt.Sprintf(
		"[Loop guard] This exact tool call (`%s` with identical arguments) has run %d time(s) already "+
			"in this turn with the same output and no intervening state change. Try a different approach.",
		name, len(matches),
	)
}

func (g *loopGuard) record(name, args, output string, success bool) {
	key := loopKey(name, args)
	isStateChanging := stateChangingTools[name]
	keyIsNew := true
	for _, e := range g.recent {
		if e.key == key {
			keyIsNew = false
			break
		}
	}
	if isStateChanging && success && keyIsNew {
		g.recent = g.recent[:0]
	}

	g.recent = append(g.recent, loopEntry{
		key:        key,
		outputHash: hashStr(output),
		success:    success,
	})
	if len(g.recent) > loopWindow {
		g.recent = g.recent[1:]
	}
}

func loopKey(name, args string) string {
	normalised := args
	var v any
	if err := json.Unmarshal([]byte(args), &v); err == nil {
		if b, err := json.Marshal(v); err == nil {
			normalised = string(b)
		}
	}
	return name + "\x00" + normalised
}

func hashStr(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}
