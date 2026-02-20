package commitplanner

import (
	"context"

	"cloud.google.com/go/spanner"
)

type Plan struct {
	muts []*spanner.Mutation
}

// Applier is the interface used by interactors to commit a Plan.
// Use this instead of *Committer so tests can inject a mock.
type Applier interface {
	Apply(ctx context.Context, p *Plan) error
}

type Committer struct {
	dbClient *spanner.Client
}

func NewPlan() *Plan {
	return &Plan{muts: []*spanner.Mutation{}}
}

func (p *Plan) Add(mut *spanner.Mutation) {
	p.muts = append(p.muts, mut)
}

func NewCommitter(client *spanner.Client) *Committer {
	return &Committer{dbClient: client}
}

func (c *Committer) Apply(ctx context.Context, p *Plan) error {
	_, err := c.dbClient.Apply(ctx, p.muts)
	return err
}
