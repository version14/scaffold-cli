package cli

import (
	"fmt"

	"github.com/version14/dot/internal/flow"
)

// ----------------------------------------------------------------------------
// pathCond — AND of condClauses, OR of pathConds per slot
// ----------------------------------------------------------------------------

// pathCond is a conjunction of clauses that must ALL be true for a slot to be
// considered "on the active path". Multiple pathConds on the same slot are
// OR-ed: the slot is visible when at least one holds.
type pathCond []condClause

func (pc pathCond) satisfied(store *liveStore) bool {
	for _, c := range pc {
		if !c.satisfied(store) {
			return false
		}
	}
	return true
}

func (pc pathCond) equals(other pathCond) bool {
	if len(pc) != len(other) {
		return false
	}
	for i := range pc {
		if !pc[i].equals(other[i]) {
			return false
		}
	}
	return true
}

type clauseKind int

const (
	clauseSelectEq  clauseKind = iota // string answer for questionID == value
	clauseConfirmEq                   // bool answer for questionID == boolVal
	clauseIfEq                        // IfQuestion condition result == boolVal
)

type condClause struct {
	kind       clauseKind
	questionID string
	value      string                       // clauseSelectEq
	boolVal    bool                         // clauseConfirmEq / clauseIfEq (expected)
	ifCond     func(*flow.FlowContext) bool // clauseIfEq
}

func (c condClause) satisfied(store *liveStore) bool {
	switch c.kind {
	case clauseSelectEq:
		return store.getString(c.questionID) == c.value
	case clauseConfirmEq:
		return store.getBool(c.questionID) == c.boolVal
	case clauseIfEq:
		// Build a partial FlowContext from live answers so the condition can
		// inspect them. The condition must be nil-safe for partial contexts.
		if c.ifCond == nil {
			return false
		}
		ctx := store.partialContext()
		return c.ifCond(ctx) == c.boolVal
	}
	return false
}

func (c condClause) equals(other condClause) bool {
	if c.kind != other.kind || c.questionID != other.questionID {
		return false
	}
	switch c.kind {
	case clauseSelectEq:
		return c.value == other.value
	case clauseConfirmEq, clauseIfEq:
		if c.boolVal != other.boolVal {
			return false
		}
		if c.kind == clauseIfEq {
			return fmt.Sprintf("%p", c.ifCond) == fmt.Sprintf("%p", other.ifCond)
		}
		return true
	}
	return false
}

// buildHideFunc returns a Huh-compatible hide function for a slot.
// The slot is hidden (true) when none of its pathConditions are satisfied.
func buildHideFunc(conds []pathCond, store *liveStore) func() bool {
	// No conditions at all → always visible.
	if len(conds) == 0 {
		return func() bool { return false }
	}
	// Any empty pathCond (no clauses) is unconditionally true → always visible.
	for _, pc := range conds {
		if len(pc) == 0 {
			return func() bool { return false }
		}
	}
	return func() bool {
		for _, pc := range conds {
			if pc.satisfied(store) {
				return false // at least one condition holds → show
			}
		}
		return true // all conditions fail → hide
	}
}

// ----------------------------------------------------------------------------
// liveStore — live answer pointers shared between Huh fields and hide funcs
// ----------------------------------------------------------------------------

// liveStore holds pointer-sized live values that Huh fields write to and
// hide functions read from. Because both sides share the same pointer, Huh's
// reactive updates automatically trigger the correct hide behaviour.
type liveStore struct {
	strings   map[string]*string
	bools     map[string]*bool
	strSlices map[string]*[]string
}

func newLiveStore() *liveStore {
	return &liveStore{
		strings:   make(map[string]*string),
		bools:     make(map[string]*bool),
		strSlices: make(map[string]*[]string),
	}
}

func (s *liveStore) allocString(id string) *string {
	if p, ok := s.strings[id]; ok {
		return p
	}
	p := new(string)
	s.strings[id] = p
	return p
}

func (s *liveStore) allocBool(id string) *bool {
	if p, ok := s.bools[id]; ok {
		return p
	}
	p := new(bool)
	s.bools[id] = p
	return p
}

func (s *liveStore) allocStrSlice(id string) *[]string {
	if p, ok := s.strSlices[id]; ok {
		return p
	}
	var sl []string
	p := &sl
	s.strSlices[id] = p
	return p
}

func (s *liveStore) getString(id string) string {
	if p, ok := s.strings[id]; ok {
		return *p
	}
	return ""
}

func (s *liveStore) getBool(id string) bool {
	if p, ok := s.bools[id]; ok {
		return *p
	}
	return false
}

func (s *liveStore) getStrSlice(id string) []string {
	if p, ok := s.strSlices[id]; ok {
		return *p
	}
	return nil
}

// partialContext builds a FlowContext from the current live answers.
// Used by clauseIfEq to evaluate IfQuestion conditions mid-form.
func (s *liveStore) partialContext() *flow.FlowContext {
	ctx := &flow.FlowContext{
		Answers:   make(map[string]flow.AnswerNode),
		LoopStack: []flow.LoopFrame{},
	}
	for id, p := range s.strings {
		ctx.Answers[id] = *p
	}
	for id, p := range s.bools {
		ctx.Answers[id] = *p
	}
	for id, p := range s.strSlices {
		ctx.Answers[id] = *p
	}
	return ctx
}

// ----------------------------------------------------------------------------
// formSlot — one potential question in the pre-built Huh form
// ----------------------------------------------------------------------------

// formSlot is one node collected during the pre-walk. It pairs a question with
// the set of pathConditions (OR-ed) that determine when it is visible.
type formSlot struct {
	question   flow.Question
	conditions []pathCond
}

// loopBarrier marks a LoopQuestion in the slot list so the runner can find it
// after the main form completes and execute the body sub-forms.
type loopBarrier struct {
	slotIdx  int
	question *flow.LoopQuestion
}

// ----------------------------------------------------------------------------
// formWalker — DFS pre-walk of the flow graph
// ----------------------------------------------------------------------------

// formWalker traverses the flow graph before rendering, collecting formSlots in
// display order and recording any LoopQuestion barriers for post-form handling.
//
// Rules:
//   - IfQuestion: not added as a slot (no UI); its condition is appended to the
//     path conditions of downstream nodes.
//   - LoopQuestion: added as a slot (shows "how many?" input); body is NOT
//     walked (handled as sub-forms in HuhFormRunner.Run).
//   - InsertAfter injections: spliced between the target and its natural next,
//     sharing the target's path condition.
//   - Already-visited nodes: merge the new pathCond (OR) and re-walk children
//     to propagate the updated visibility condition down the graph.
//   - Cycle detection: re-walking only occurs if the new condition is not
//     already registered for that node.
type formWalker struct {
	hooks     *flow.HookRegistry
	fragments *flow.FragmentRegistry

	slots   []*formSlot
	visited map[string]int // questionID → slot index
	loops   []*loopBarrier
}

func newFormWalker(hooks *flow.HookRegistry, fragments *flow.FragmentRegistry) *formWalker {

	return &formWalker{
		hooks:     hooks,
		fragments: fragments,
		visited:   make(map[string]int),
	}
}

func (w *formWalker) walk(root flow.Question) {
	w.walkQ(root, nil)
}

func (w *formWalker) walkQ(q flow.Question, cond pathCond) {
	if q == nil {
		return
	}

	// Apply Replace injection (first registered Replace wins).
	q = w.applyReplace(q)

	// IfQuestion has no UI: thread its condition into downstream nodes and return.
	if ifq, ok := q.(*flow.IfQuestion); ok {
		thenCond := appendCond(cond, condClause{
			kind:    clauseIfEq,
			boolVal: true,
			ifCond:  ifq.Condition,
		})
		elseCond := appendCond(cond, condClause{
			kind:    clauseIfEq,
			boolVal: false,
			ifCond:  ifq.Condition,
		})
		w.walkNext(ifq.Then, thenCond)
		w.walkNext(ifq.Else, elseCond)
		return
	}

	id := q.ID()

	// Merge condition if already visited.
	idx, alreadyVisited := w.visited[id]
	if alreadyVisited {
		// Avoid redundant re-walks if we already have this exact condition.
		for _, existing := range w.slots[idx].conditions {
			if existing.equals(cond) {
				return
			}
		}
		w.slots[idx].conditions = append(w.slots[idx].conditions, cond)
	} else {
		// Register slot.
		idx = len(w.slots)
		w.visited[id] = idx
		slot := &formSlot{question: q, conditions: []pathCond{cond}}
		w.slots = append(w.slots, slot)
	}

	// Collect InsertAfter injections for this node.
	var inserts []flow.Question
	if w.hooks != nil {
		_, _, inserts = w.hooks.ForKind(id)
	}

	switch typed := q.(type) {

	case *flow.OptionQuestion:
		// Walk inserts first (same condition as target — always shown after target).
		for _, ins := range inserts {
			w.walkQ(ins, cloneCond(cond))
		}
		if typed.Multiple {
			w.walkNext(typed.Next_, cond)
		} else {
			// Merge plugin-added options so we walk their branches too.
			merged := w.mergeOptions(typed)
			for _, opt := range merged {
				branchCond := appendCond(cond, condClause{
					kind:       clauseSelectEq,
					questionID: id,
					value:      opt.Value,
				})
				w.walkNext(opt.Next, branchCond)
			}
		}

	case *flow.ConfirmQuestion:
		for _, ins := range inserts {
			w.walkQ(ins, cloneCond(cond))
		}
		thenCond := appendCond(cond, condClause{kind: clauseConfirmEq, questionID: id, boolVal: true})
		elseCond := appendCond(cond, condClause{kind: clauseConfirmEq, questionID: id, boolVal: false})
		w.walkNext(typed.Then, thenCond)
		w.walkNext(typed.Else, elseCond)

	case *flow.TextQuestion:
		for _, ins := range inserts {
			w.walkQ(ins, cloneCond(cond))
		}
		w.walkNext(typed.Next_, cond)

	case *flow.LoopQuestion:
		// Record the barrier only once.
		if !alreadyVisited {
			w.loops = append(w.loops, &loopBarrier{slotIdx: idx, question: typed})
		}
		// Continue is deferred: HuhFormRunner runs it as a post-loop sub-form
		// after all iterations complete. Walking it here would insert the
		// Continue questions (e.g. confirmGenerate) into the main form BEFORE
		// the body sub-forms execute, reversing the intended order.
		// Do NOT walk typed.Body — those run as sub-forms.
	}
}

func (w *formWalker) walkNext(next *flow.Next, cond pathCond) {
	if next == nil || next.End {
		return
	}
	if next.Question != nil {
		w.walkQ(next.Question, cond)
		return
	}
	if next.Fragment != "" && w.fragments != nil {
		// Pass nil context: fragments that need answers at pre-walk time must
		// handle nil gracefully (they cannot branch on live answers here).
		resolved := w.fragments.Resolve(next.Fragment, nil)
		if resolved != nil {
			w.walkNext(resolved, cond)
		}
	}
}

// applyReplace returns the first registered Replace question for q.ID(),
// or q unchanged if none.
func (w *formWalker) applyReplace(q flow.Question) flow.Question {
	if w.hooks == nil {
		return q
	}
	replace, _, _ := w.hooks.ForKind(q.ID())
	if len(replace) == 0 {
		return q
	}
	return replace[0]
}

// mergeOptions returns the full option list for an OptionQuestion after
// appending any AddOption injections.
func (w *formWalker) mergeOptions(q *flow.OptionQuestion) []*flow.Option {
	if w.hooks == nil {
		return q.Options
	}
	_, extras, _ := w.hooks.ForKind(q.ID())
	if len(extras) == 0 {
		return q.Options
	}
	merged := make([]*flow.Option, len(q.Options)+len(extras))
	copy(merged, q.Options)
	copy(merged[len(q.Options):], extras)
	return merged
}

// cloneCond returns a copy of c so appending to it never mutates the original.
func cloneCond(c pathCond) pathCond {
	if len(c) == 0 {
		return nil
	}
	out := make(pathCond, len(c))
	copy(out, c)
	return out
}

// appendCond returns a NEW pathCond with clause appended to c.
func appendCond(c pathCond, clause condClause) pathCond {
	out := make(pathCond, len(c)+1)
	copy(out, c)
	out[len(c)] = clause
	return out
}
