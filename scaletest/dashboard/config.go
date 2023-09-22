package dashboard

import (
	"github.com/google/uuid"
	"time"

	"golang.org/x/xerrors"

	"cdr.dev/slog"
)

type Config struct {
	// MinWait is the minimum interval between fetches.
	MinWait time.Duration `json:"duration_min"`
	// MaxWait is the maximum interval between fetches.
	MaxWait time.Duration `json:"duration_max"`
	// Trace is whether to trace the requests.
	Trace bool `json:"trace"`
	// Logger is the logger to use.
	Logger slog.Logger `json:"-"`
	// RollTable is the set of actions to perform
	RollTable RollTable `json:"roll_table"`
	// Headless controls headless mode for chromedp.
	Headless bool `json:"no_headless"`
	// UserID is the user as which to run the test
	UserID uuid.UUID `json:"user_id"`
}

func (c Config) Validate() error {
	if c.MinWait <= 0 {
		return xerrors.Errorf("validate duration_min: must be greater than zero")
	}

	if c.MaxWait <= 0 {
		return xerrors.Errorf("validate duration_max: must be greater than zero")
	}

	if c.MinWait > c.MaxWait {
		return xerrors.Errorf("validate duration_min: must be less than duration_max")
	}

	if c.UserID == uuid.Nil {
		return xerrors.Errorf("validate user_id: must not be nil")
	}

	return nil
}
