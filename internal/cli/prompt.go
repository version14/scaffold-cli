package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/version14/dot/internal/flow"
)

// ErrAborted is returned when the user presses Escape / Ctrl-C to quit.
var ErrAborted = errors.New("cli: user aborted")

// HuhFormRunner implements flow.FlowRunner using Charm's Huh form library.
//
// It is the only type in the project that imports Huh. The strategy is
// "one form = one flow": the runner pre-walks the full flow graph, allocates
// Huh fields for every reachable question, wraps each in a Group with a
// WithHideFunc reactive-hide condition, then runs a single huh.NewForm call.
//
// Because every question lives in the same form, Huh's native back-navigation
// works across all questions without any engine-level history stack.
//
// LoopQuestions appear in the main form as numeric "how many?" inputs. After
// the form completes, the runner executes body sub-forms (one per iteration)
// to collect the per-iteration answers.
type HuhFormRunner struct {
	Hooks     *flow.HookRegistry
	Fragments *flow.FragmentRegistry
}

func NewHuhFormRunner() *HuhFormRunner {
	return &HuhFormRunner{
		Hooks:     flow.NewHookRegistry(),
		Fragments: flow.NewFragmentRegistry(),
	}
}

// Run implements flow.FlowRunner.
func (r *HuhFormRunner) Run(root flow.Question) (*flow.FlowContext, error) {
	// ── Step 1: Pre-walk the flow graph ──────────────────────────────────────
	walker := newFormWalker(r.Hooks, r.Fragments)
	walker.walk(root)

	// ── Step 2: Allocate live answer pointers ─────────────────────────────────
	store := newLiveStore()

	// ── Step 3: Build Huh groups (one per slot) ───────────────────────────────
	groups, err := r.buildGroups(walker.slots, store)
	if err != nil {
		return nil, err
	}

	// ── Step 4: Run the single main form ──────────────────────────────────────
	if len(groups) > 0 {
		form := huh.NewForm(groups...)
		if runErr := form.Run(); runErr != nil {
			if errors.Is(runErr, huh.ErrUserAborted) {
				return nil, ErrAborted
			}
			return nil, fmt.Errorf("cli: form: %w", runErr)
		}
	}

	// ── Step 5: Collect answers from active (non-hidden) slots ────────────────
	ctx := &flow.FlowContext{
		Answers:       make(map[string]flow.AnswerNode),
		LoopStack:     []flow.LoopFrame{},
		LoadedPlugins: pluginNames(r.Hooks.Plugins()),
	}

	for _, slot := range walker.slots {
		hideFunc := buildHideFunc(slot.conditions, store)
		if hideFunc() {
			continue // slot was hidden — not on the active path
		}

		id := slot.question.ID()
		ctx.VisitedNodes = append(ctx.VisitedNodes, id)

		switch q := slot.question.(type) {
		case *flow.TextQuestion:
			val := store.getString(id)
			if val == "" {
				val = q.Default
			}
			ctx.Answers[id] = val

		case *flow.OptionQuestion:
			if q.Multiple {
				ctx.Answers[id] = store.getStrSlice(id)
			} else {
				ctx.Answers[id] = store.getString(id)
			}

		case *flow.ConfirmQuestion:
			ctx.Answers[id] = store.getBool(id)

		case *flow.LoopQuestion:
			// Count was captured by the main form input; body is handled below.
			// We store the final []map[string]Answer in step 6.
		}
	}

	// ── Step 6: Execute loop body sub-forms ───────────────────────────────────
	for _, barrier := range walker.loops {
		slot := walker.slots[barrier.slotIdx]
		hideFunc := buildHideFunc(slot.conditions, store)
		if hideFunc() {
			continue // loop itself was hidden; skip
		}

		lq := barrier.question
		count := parseCount(store.getString(lq.ID()))
		iterations, iterErr := r.runLoopSubForms(lq, count, ctx)
		if iterErr != nil {
			return nil, iterErr
		}
		ctx.Answers[lq.ID()] = iterations
	}

	// ── Step 7: Post-loop Continue sub-forms ──────────────────────────────────
	// The main form walker deliberately skips LoopQuestion.Continue so that
	// post-loop questions (e.g. confirmGenerate) appear AFTER the body
	// sub-forms, not before. We run them here as a separate form now that all
	// iterations are complete.
	for _, barrier := range walker.loops {
		slot := walker.slots[barrier.slotIdx]
		if buildHideFunc(slot.conditions, store)() {
			continue // loop was hidden; its Continue is irrelevant
		}

		lq := barrier.question
		if lq.Continue == nil || lq.Continue.End || lq.Continue.Question == nil {
			continue
		}

		continueStore := newLiveStore()
		contWalker := newFormWalker(r.Hooks, r.Fragments)
		contWalker.walk(lq.Continue.Question)

		contGroups, err := r.buildGroups(contWalker.slots, continueStore)
		if err != nil {
			return nil, fmt.Errorf("cli: post-loop continue %q: %w", lq.ID(), err)
		}

		if len(contGroups) > 0 {
			form := huh.NewForm(contGroups...)
			if runErr := form.Run(); runErr != nil {
				if errors.Is(runErr, huh.ErrUserAborted) {
					return nil, ErrAborted
				}
				return nil, fmt.Errorf("cli: post-loop continue %q: %w", lq.ID(), runErr)
			}
		}

		for _, cSlot := range contWalker.slots {
			if buildHideFunc(cSlot.conditions, continueStore)() {
				continue
			}
			id := cSlot.question.ID()
			ctx.VisitedNodes = append(ctx.VisitedNodes, id)
			switch q := cSlot.question.(type) {
			case *flow.TextQuestion:
				val := continueStore.getString(id)
				if val == "" {
					val = q.Default
				}
				ctx.Answers[id] = val
			case *flow.OptionQuestion:
				if q.Multiple {
					ctx.Answers[id] = continueStore.getStrSlice(id)
				} else {
					ctx.Answers[id] = continueStore.getString(id)
				}
			case *flow.ConfirmQuestion:
				ctx.Answers[id] = continueStore.getBool(id)
			}
		}
	}

	return ctx, nil
}

// buildGroups converts a slice of formSlots into Huh groups. Each group
// contains exactly one field so that Huh's back/forward moves by question.
func (r *HuhFormRunner) buildGroups(slots []*formSlot, store *liveStore) ([]*huh.Group, error) {
	groups := make([]*huh.Group, 0, len(slots))

	for _, slot := range slots {
		field, fieldErr := r.buildField(slot.question, store)
		if fieldErr != nil {
			return nil, fieldErr
		}
		if field == nil {
			continue // IfQuestion — no UI
		}

		g := huh.NewGroup(field).WithHideFunc(buildHideFunc(slot.conditions, store))
		groups = append(groups, g)
	}
	return groups, nil
}

// buildField creates the Huh field for a question, wiring it to the liveStore
// pointer so hide functions see updates immediately.
func (r *HuhFormRunner) buildField(q flow.Question, store *liveStore) (huh.Field, error) {
	switch typed := q.(type) {

	case *flow.TextQuestion:
		ptr := store.allocString(typed.ID())
		*ptr = typed.Default
		inp := huh.NewInput().
			Title(typed.Label).
			Description(typed.Description).
			Placeholder(typed.Default).
			Value(ptr)
		if typed.Validate != nil {
			inp = inp.Validate(typed.Validate)
		}
		return inp, nil

	case *flow.OptionQuestion:
		if typed.Multiple {
			ptr := store.allocStrSlice(typed.ID())
			opts := toHuhOptions(typed.Options)
			return huh.NewMultiSelect[string]().
				Title(typed.Label).
				Description(typed.Description).
				Options(opts...).
				Value(ptr), nil
		}
		ptr := store.allocString(typed.ID())
		opts := toHuhOptions(typed.Options)
		return huh.NewSelect[string]().
			Title(typed.Label).
			Description(typed.Description).
			Options(opts...).
			Value(ptr), nil

	case *flow.ConfirmQuestion:
		ptr := store.allocBool(typed.ID())
		*ptr = typed.Default
		return huh.NewConfirm().
			Title(typed.Label).
			Description(typed.Description).
			Value(ptr), nil

	case *flow.LoopQuestion:
		// Rendered as a count input in the main form.
		ptr := store.allocString(typed.ID())
		return huh.NewInput().
			Title(fmt.Sprintf("%s — how many?", typed.Label)).
			Placeholder("0").
			Validate(validatePositiveInt).
			Value(ptr), nil

	case *flow.IfQuestion:
		return nil, nil // no user input; handled by walker conditions

	default:
		return nil, fmt.Errorf("cli: unknown question type %T", q)
	}
}

// runLoopSubForms executes `count` sub-forms for the body of a LoopQuestion.
// Each iteration pre-walks the body sub-graph (same approach as the main form
// walker) so conditional questions within the body get proper hide functions.
func (r *HuhFormRunner) runLoopSubForms(
	lq *flow.LoopQuestion,
	count int,
	ctx *flow.FlowContext,
) ([]map[string]flow.Answer, error) {
	results := make([]map[string]flow.Answer, count)

	for i := 0; i < count; i++ {
		PrintProgress(i+1, count, lq.Label)

		bodyStore := newLiveStore()

		subWalker := newFormWalker(r.Hooks, r.Fragments)
		for _, q := range lq.Body {
			subWalker.walk(q)
		}

		bodyGroups, err := r.buildGroups(subWalker.slots, bodyStore)
		if err != nil {
			return nil, fmt.Errorf("cli: loop %q iteration %d: %w", lq.ID(), i+1, err)
		}

		if len(bodyGroups) > 0 {
			form := huh.NewForm(bodyGroups...)
			if runErr := form.Run(); runErr != nil {
				if errors.Is(runErr, huh.ErrUserAborted) {
					return nil, ErrAborted
				}
				return nil, fmt.Errorf("cli: loop %q iteration %d: %w", lq.ID(), i+1, runErr)
			}
		}

		iter := make(map[string]flow.Answer, len(subWalker.slots))
		for _, slot := range subWalker.slots {
			if buildHideFunc(slot.conditions, bodyStore)() {
				continue
			}
			id := slot.question.ID()
			switch q := slot.question.(type) {
			case *flow.TextQuestion:
				val := bodyStore.getString(id)
				if val == "" {
					val = q.Default
				}
				iter[id] = val
			case *flow.OptionQuestion:
				if q.Multiple {
					iter[id] = bodyStore.getStrSlice(id)
				} else {
					iter[id] = bodyStore.getString(id)
				}
			case *flow.ConfirmQuestion:
				iter[id] = bodyStore.getBool(id)
			}
		}
		results[i] = iter
	}

	return results, nil
}

// toHuhOptions converts flow.Option slice to huh.Option[string] slice.
func toHuhOptions(opts []*flow.Option) []huh.Option[string] {
	out := make([]huh.Option[string], len(opts))
	for i, o := range opts {
		out[i] = huh.NewOption(o.Label, o.Value)
	}
	return out
}

// pluginNames converts []flow.PluginID to []string for FlowContext.
func pluginNames(ids []flow.PluginID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

// validatePositiveInt is a Huh Validate func for loop count inputs.
func validatePositiveInt(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("must be a number")
	}
	if n < 0 {
		return fmt.Errorf("must be zero or positive")
	}
	return nil
}

// parseCount converts a raw string count to int; blank or invalid → 0.
func parseCount(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
