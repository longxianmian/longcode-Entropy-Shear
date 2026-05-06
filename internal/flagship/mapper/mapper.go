// Package mapper turns multi-source flagship inputs into the five-element
// state vector (Goal / Fact / Evidence / Constraint / Action).
//
// The mapping is deterministic: no LLM, no randomness, no external calls.
// Element states live in [0, 1] and feed the 5x5 evaluation matrix in
// internal/flagship/state.
package mapper

// Goal describes the intended outcome to evaluate.
type Goal struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	Priority    string `json:"priority,omitempty"`
}

// Fact is an atomic ground claim about the world.
type Fact struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Value  interface{} `json:"value"`
	Source string      `json:"source,omitempty"`
}

// Evidence supports or refutes facts and goals.
type Evidence struct {
	ID         string  `json:"id"`
	Kind       string  `json:"kind"`
	Reference  string  `json:"reference,omitempty"`
	Confidence float64 `json:"confidence"`
	Verified   bool    `json:"verified"`
}

// Constraint represents permission, compliance, governance or policy gates.
// Severity is "soft" or "hard"; Satisfied being nil means unknown.
type Constraint struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Severity  string `json:"severity"`
	Statement string `json:"statement,omitempty"`
	Satisfied *bool  `json:"satisfied,omitempty"`
}

// Action is a proposed step to execute.
type Action struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Reversible bool    `json:"reversible"`
	Cost       float64 `json:"cost,omitempty"`
}

// ElementStates is the five-element state vector. Each value is clamped to
// [0, 1] by Map.
type ElementStates struct {
	Goal       float64 `json:"goal"`
	Fact       float64 `json:"fact"`
	Evidence   float64 `json:"evidence"`
	Constraint float64 `json:"constraint"`
	Action     float64 `json:"action"`
}

// Map computes ElementStates from multi-source input. The rules are:
//
//   - Goal:       0 if missing; 0.5 if id only; 1.0 with id + description.
//   - Fact:       0 with no facts; otherwise the share of facts that have a
//                 non-empty key and a non-nil value.
//   - Evidence:   0 with no evidence; otherwise the mean of
//                 confidence * (1.0 if verified else 0.5).
//   - Constraint: 1.0 with no constraints; otherwise satisfied / total,
//                 where Satisfied == nil counts as unsatisfied.
//   - Action:     0 with no actions; otherwise mean of (1 - cost), clamped.
func Map(g *Goal, facts []Fact, ev []Evidence, cs []Constraint, acts []Action) ElementStates {
	return ElementStates{
		Goal:       goalState(g),
		Fact:       factState(facts),
		Evidence:   evidenceState(ev),
		Constraint: constraintState(cs),
		Action:     actionState(acts),
	}
}

func goalState(g *Goal) float64 {
	if g == nil || g.ID == "" {
		return 0
	}
	if g.Description == "" {
		return 0.5
	}
	return 1.0
}

func factState(facts []Fact) float64 {
	if len(facts) == 0 {
		return 0
	}
	var ok int
	for _, f := range facts {
		if f.Key != "" && f.Value != nil {
			ok++
		}
	}
	return clamp01(float64(ok) / float64(len(facts)))
}

func evidenceState(ev []Evidence) float64 {
	if len(ev) == 0 {
		return 0
	}
	var sum float64
	for _, e := range ev {
		c := clamp01(e.Confidence)
		if e.Verified {
			sum += c
		} else {
			sum += c * 0.5
		}
	}
	return clamp01(sum / float64(len(ev)))
}

func constraintState(cs []Constraint) float64 {
	if len(cs) == 0 {
		return 1.0
	}
	var ok int
	for _, c := range cs {
		if c.Satisfied != nil && *c.Satisfied {
			ok++
		}
	}
	return clamp01(float64(ok) / float64(len(cs)))
}

func actionState(acts []Action) float64 {
	if len(acts) == 0 {
		return 0
	}
	var sum float64
	for _, a := range acts {
		sum += clamp01(1.0 - clamp01(a.Cost))
	}
	return clamp01(sum / float64(len(acts)))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
