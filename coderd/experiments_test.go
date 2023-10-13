package coderd_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/coder/v2/coderd/coderdtest"
	"github.com/coder/coder/v2/coderd/httpmw"
	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/coder/v2/testutil"
)

func Test_Experiments(t *testing.T) {
	t.Parallel()
	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		cfg := coderdtest.DeploymentValues(t)
		client := coderdtest.New(t, &coderdtest.Options{
			DeploymentValues: cfg,
		})
		_ = coderdtest.CreateFirstUser(t, client)

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancel()

		experiments, err := client.Experiments(ctx)
		require.NoError(t, err)
		require.NotNil(t, experiments)
		require.Empty(t, experiments.Enabled)
	})

	t.Run("multiple features", func(t *testing.T) {
		t.Parallel()
		cfg := coderdtest.DeploymentValues(t)
		cfg.Experiments = []string{"foo", "BAR"}
		client := coderdtest.New(t, &coderdtest.Options{
			DeploymentValues: cfg,
		})
		_ = coderdtest.CreateFirstUser(t, client)

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancel()

		experiments, err := client.Experiments(ctx)
		require.NoError(t, err)
		require.NotNil(t, experiments)
		// Should be lower-cased.
		require.ElementsMatch(t, []codersdk.Experiment{"foo", "bar"}, experiments)
	})

	t.Run("wildcard", func(t *testing.T) {
		t.Parallel()
		cfg := coderdtest.DeploymentValues(t)
		cfg.Experiments = []string{"*"}
		client := coderdtest.New(t, &coderdtest.Options{
			DeploymentValues: cfg,
		})
		_ = coderdtest.CreateFirstUser(t, client)

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancel()

		experiments, err := client.Experiments(ctx)
		require.NoError(t, err)
		require.NotNil(t, experiments)
		require.ElementsMatch(t, codersdk.ExperimentsAll, experiments)
	})

	t.Run("alternate wildcard with manual opt-in", func(t *testing.T) {
		t.Parallel()
		cfg := coderdtest.DeploymentValues(t)
		cfg.Experiments = []string{"*", "dAnGeR"}
		client := coderdtest.New(t, &coderdtest.Options{
			DeploymentValues: cfg,
		})
		_ = coderdtest.CreateFirstUser(t, client)

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancel()

		experiments, err := client.Experiments(ctx)
		require.NoError(t, err)
		require.NotNil(t, experiments)
		require.ElementsMatch(t, append(codersdk.ExperimentsAll, "danger"), experiments)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		t.Parallel()
		cfg := coderdtest.DeploymentValues(t)
		cfg.Experiments = []string{"*"}
		client := coderdtest.New(t, &coderdtest.Options{
			DeploymentValues: cfg,
		})
		// Explicitly omit creating a user so we're unauthorized.
		// _ = coderdtest.CreateFirstUser(t, client)

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancel()

		_, err := client.Experiments(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, httpmw.SignedOutErrorMessage)
	})
}
