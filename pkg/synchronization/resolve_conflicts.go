package synchronization

import (
	"context"
	"fmt"
	"time"

	"github.com/mutagen-io/mutagen/pkg/selection"
	"github.com/mutagen-io/mutagen/pkg/synchronization/core"
	"github.com/mutagen-io/mutagen/pkg/synchronization/endpoint/local"
	"github.com/mutagen-io/mutagen/pkg/synchronization/rsync"
)

func (m *Manager) ResolveConflicts(ctx context.Context, selection *selection.Selection, favorAlpha bool) error {
	// Lock the session registry for the duration of the operation.
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()

	// Grab the session and ensure it's still valid.
	session, err := s.findSession(request.Selection)
	if err != nil {
		return nil, err
	}

	// Lock the session for the duration of the operation.
	session.Lock()
	defer session.Unlock()

	// Ensure the session is ready for synchronization.
	if session.State != core.SessionStateReady {
		return nil, fmt.Errorf("session not ready for synchronization")
	}

	// Get the current synchronization state.
	ancestorChanges, alphaChanges, betaChanges, err := core.GetStagingStatus(session)
	if err != nil {
		return nil, fmt.Errorf("unable to compute staging status: %w", err)
	}

	// Identify conflicts.
	conflicts := core.FindConflicts(ancestorChanges, alphaChanges, betaChanges)

	// Resolve conflicts in favor of the specified side.
	var winningChanges, losingChanges *core.Changes
	if request.FavorAlpha {
		winningChanges = alphaChanges
		losingChanges = betaChanges
	} else {
		winningChanges = betaChanges
		losingChanges = alphaChanges
	}

	// Apply resolutions.
	for _, conflict := range conflicts {
		if winningChange, ok := winningChanges.Contents[conflict.Path]; ok {
			losingChanges.Contents[conflict.Path] = winningChange
		} else {
			delete(losingChanges.Contents, conflict.Path)
		}
	}

	// Create rsync engines for both endpoints.
	alphaEngine := rsync.NewEngine()
	betaEngine := rsync.NewEngine()

	// Perform synchronization to propagate changes.
	if err := core.Synchronize(
		ctx,
		session.Alpha.(local.Endpoint),
		session.Beta.(local.Endpoint),
		ancestorChanges,
		alphaChanges,
		betaChanges,
		alphaEngine,
		betaEngine,
		session.Configuration.SynchronizationMode,
	); err != nil {
		return nil, fmt.Errorf("unable to perform synchronization: %w", err)
	}

	// Update the session's synchronization state.
	session.LastSynchronizationTime = time.Now()

	// Save the session.
	if err := s.saveSession(session); err != nil {
		return nil, fmt.Errorf("unable to save session: %w", err)
	}

	return &ResolveConflictsResponse{}, nil
}
