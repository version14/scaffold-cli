package flow

// Question interface + all types: OptionQuestion, TextQuestion, ConfirmQuestion, LoopQuestion, IfQuestion

// Question is the engine-facing interface for a flow node.
// It does NOT know about Huh or any terminal library.
type Question interface {
	ID() string
	// Next returns the edge to follow given a user's answer.
	// Called by the engine after the adapter collects the answer.
	Next(answer Answer) *Next
}

// QuestionBase holds the common fields every question type shares.
// Embed it to get ID() and Next() for free.
type QuestionBase struct {
	ID_   string
	Next_ *Next
}

func (b QuestionBase) ID() string { return b.ID_ }

// Each type embeds it
type OptionQuestion struct {
	QuestionBase
	Label       string
	Description string
	Multiple    bool
	Options     []*Option
}

type Option struct {
	Label string
	Value string
	Next  *Next // branch to follow if this option is chosen (single-select only)
}

type TextQuestion struct {
	QuestionBase
	Label       string
	Description string
	Default     string
	Validate    func(string) error
}

type ConfirmQuestion struct {
	QuestionBase
	Label       string
	Description string
	Default     bool
	Then        *Next
	Else        *Next
}

type LoopQuestion struct {
	QuestionBase
	Label    string
	Body     []Question
	Continue *Next
}

type IfQuestion struct {
	QuestionBase
	Label     string
	Condition func(ctx *FlowContext) bool
	Then      *Next
	Else      *Next
}

func (q *OptionQuestion) Next(answer Answer) *Next {
	if q.Multiple {
		return q.Next_
	}
	val, ok := answer.(string)
	if !ok {
		return nil
	}
	for _, opt := range q.Options {
		if opt.Value == val {
			return opt.Next
		}
	}
	return nil
}

func (q *TextQuestion) Next(answer Answer) *Next {
	return q.Next_
}

func (q *ConfirmQuestion) Next(answer Answer) *Next {
	val, ok := answer.(bool)
	if !ok || !val {
		return q.Else
	}
	return q.Then
}

func (q *LoopQuestion) Next(answer Answer) *Next {
	return q.Continue
}

func (q *IfQuestion) Next(answer Answer) *Next {
	// IfQuestion has no user input — engine calls this with nil
	// Condition is evaluated by the engine directly, not via Next()
	return nil
}
