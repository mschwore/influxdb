package query

type Plan struct {
	ready map[Node]struct{}
	want  map[*ReadEdge]struct{}
}

func NewPlan() *Plan {
	return &Plan{
		ready: make(map[Node]struct{}),
		want:  make(map[*ReadEdge]struct{}),
	}
}

type ExecuteNodeFn func(n Node) error

func ExecuteTargets(outputs []*ReadEdge, fn ExecuteNodeFn) error {
	plan := NewPlan()
	for _, e := range outputs {
		plan.AddTarget(e)
	}

	for {
		node := plan.FindWork()
		if node == nil {
			return nil
		}

		if err := fn(node); err != nil {
			return err
		}
		plan.NodeFinished(node)
	}
}

func (p *Plan) AddTarget(e *ReadEdge) {
	if _, ok := p.want[e]; ok {
		return
	}
	p.want[e] = struct{}{}

	if node, ok := e.Input.Node.(OptimizableNode); ok {
		node.Optimize()
	}

	if inputs := e.Input.Node.Inputs(); len(inputs) == 0 {
		p.ready[e.Input.Node] = struct{}{}
		return
	} else {
		for _, input := range inputs {
			p.AddTarget(input)
		}
	}
}

func (p *Plan) FindWork() Node {
	for n := range p.ready {
		delete(p.ready, n)
		return n
	}
	return nil
}

// EdgeFinished runs when notified that an Edge has finished running so the
// Edge's output Nodes can be checked to see if their output edges are now
// ready to be run.
func (p *Plan) NodeFinished(n Node) {
	for _, e := range n.Outputs() {
		// The nodes are now considered ready. Check if their output edge is
		// now ready to be executed (if they have one).
		if e.Output.Node != nil && AllInputsReady(e.Output.Node) {
			p.ready[e.Output.Node] = struct{}{}
		}
	}
}
