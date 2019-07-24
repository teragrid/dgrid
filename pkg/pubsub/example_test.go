package pubsub_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/teragrid/dgrid/pkg/log"

	"github.com/tgrid/tgrid/pkg/pubsub"
	"github.com/tgrid/tgrid/pkg/pubsub/query"
)

func TestExample(t *testing.T) {
	s := pubsub.NewServer()
	s.SetLogger(log.TestingLogger())
	s.Start()
	defer s.Stop()

	ctx := context.Background()
	subscription, err := s.Subscribe(ctx, "example-client", query.MustParse("asura.account.name='John'"))
	require.NoError(t, err)
	err = s.PublishWithTags(ctx, "Tombstone", map[string]string{"asura.account.name": "John"})
	require.NoError(t, err)
	assertReceive(t, "Tombstone", subscription.Out())
}
