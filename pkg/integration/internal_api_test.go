package integration

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"

	"github.com/mutagen-io/mutagen/pkg/forwarding"
	"github.com/mutagen-io/mutagen/pkg/forwarding/endpoint/local"
	"github.com/mutagen-io/mutagen/pkg/integration/fixtures/constants"
	"github.com/mutagen-io/mutagen/pkg/integration/protocols/netpipe"
	"github.com/mutagen-io/mutagen/pkg/prompting"
	"github.com/mutagen-io/mutagen/pkg/selection"
	"github.com/mutagen-io/mutagen/pkg/synchronization"
	"github.com/mutagen-io/mutagen/pkg/synchronization/compression"
	"github.com/mutagen-io/mutagen/pkg/synchronization/core"
	"github.com/mutagen-io/mutagen/pkg/synchronization/hashing"
	"github.com/mutagen-io/mutagen/pkg/url"
)

func waitForSuccessfulSynchronizationCycle(ctx context.Context, sessionID string, allowScanProblems, allowConflicts, allowTransitionProblems bool) error {
	// Create a session selection specification.
	selection := &selection.Selection{
		Specifications: []string{sessionID},
	}

	// Perform waiting.
	var previousStateIndex uint64
	var states []*synchronization.State
	var err error
	for {
		previousStateIndex, states, err = synchronizationManager.List(ctx, selection, previousStateIndex)
		if err != nil {
			return fmt.Errorf("unable to list session states: %w", err)
		} else if len(states) != 1 {
			return errors.New("invalid number of session states returned")
		} else if states[0].SuccessfulCycles > 0 {
			if !allowScanProblems && (len(states[0].AlphaState.ScanProblems) > 0 || len(states[0].BetaState.ScanProblems) > 0) {
				return errors.New("scan problems detected (and disallowed)")
			} else if !allowConflicts && len(states[0].Conflicts) > 0 {
				return errors.New("conflicts detected (and disallowed)")
			} else if !allowTransitionProblems && (len(states[0].AlphaState.TransitionProblems) > 0 || len(states[0].BetaState.TransitionProblems) > 0) {
				return errors.New("transition problems detected (and disallowed)")
			}
			return nil
		}
	}
}

func testSessionLifecycle(ctx context.Context, prompter string, alpha, beta *url.URL, configuration *synchronization.Configuration, allowScanProblems, allowConflicts, allowTransitionProblems bool) error {
	// Create a session.
	sessionID, err := synchronizationManager.Create(
		ctx,
		alpha, beta,
		configuration,
		&synchronization.Configuration{},
		&synchronization.Configuration{},
		"testSynchronizationSession",
		nil,
		false,
		prompter,
	)
	if err != nil {
		return fmt.Errorf("unable to create session: %w", err)
	}

	// Wait for the session to have at least one successful synchronization
	// cycle.
	// TODO: Should we add a timeout on this?
	if err := waitForSuccessfulSynchronizationCycle(ctx, sessionID, allowScanProblems, allowConflicts, allowTransitionProblems); err != nil {
		return fmt.Errorf("unable to wait for successful synchronization: %w", err)
	}

	// TODO: Add hook for verifying file contents.

	// TODO: Add hook for verifying presence/absence of particular
	// conflicts/problems and remove that monitoring from
	// waitForSuccessfulSynchronizationCycle (maybe have it pass back the
	// relevant state).

	// Create a session selection specification.
	selection := &selection.Selection{
		Specifications: []string{sessionID},
	}

	// Pause the session.
	if err := synchronizationManager.Pause(ctx, selection, ""); err != nil {
		return fmt.Errorf("unable to pause session: %w", err)
	}

	// Resume the session.
	if err := synchronizationManager.Resume(ctx, selection, ""); err != nil {
		return fmt.Errorf("unable to resume session: %w", err)
	}

	// Wait for the session to have at least one additional synchronization
	// cycle.
	if err := waitForSuccessfulSynchronizationCycle(ctx, sessionID, allowScanProblems, allowConflicts, allowTransitionProblems); err != nil {
		return fmt.Errorf("unable to wait for additional synchronization: %w", err)
	}

	// Attempt an additional resume (this should be a no-op).
	if err := synchronizationManager.Resume(ctx, selection, ""); err != nil {
		return fmt.Errorf("unable to perform additional resume: %w", err)
	}

	// Terminate the session.
	if err := synchronizationManager.Terminate(ctx, selection, ""); err != nil {
		return fmt.Errorf("unable to terminate session: %w", err)
	}

	// TODO: Verify that cleanup took place.

	// Success.
	return nil
}

func TestSynchronizationBothRootsNil(t *testing.T) {
	// Allow this test to run in parallel.
	t.Parallel()

	// Calculate alpha and beta paths.
	directory := t.TempDir()
	alphaRoot := filepath.Join(directory, "alpha")
	betaRoot := filepath.Join(directory, "beta")

	// Compute alpha and beta URLs.
	alphaURL := &url.URL{Path: alphaRoot}
	betaURL := &url.URL{Path: betaRoot}

	// Compute configuration. We use defaults for everything.
	configuration := &synchronization.Configuration{}

	// Test the session lifecycle.
	if err := testSessionLifecycle(context.Background(), "", alphaURL, betaURL, configuration, false, false, false); err != nil {
		t.Error("session lifecycle test failed:", err)
	}
}

func TestSynchronizationGOROOTSrcToBeta(t *testing.T) {
	// Check the end-to-end test mode and compute the source synchronization
	// root accordingly. If no mode has been specified, then skip the test.
	endToEndTestMode := os.Getenv("MUTAGEN_TEST_END_TO_END")
	var sourceRoot string
	if endToEndTestMode == "" {
		t.Skip()
	} else if endToEndTestMode == "full" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src")
	} else if endToEndTestMode == "slim" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src", "bufio")
	} else {
		t.Fatal("unknown end-to-end test mode specified:", endToEndTestMode)
	}

	// Allow the test to run in parallel.
	t.Parallel()

	// Calculate alpha and beta paths.
	alphaRoot := sourceRoot
	betaRoot := filepath.Join(t.TempDir(), "beta")

	// Compute alpha and beta URLs.
	alphaURL := &url.URL{Path: alphaRoot}
	betaURL := &url.URL{Path: betaRoot}

	// Compute configuration. We use defaults for everything.
	configuration := &synchronization.Configuration{}

	// Test the session lifecycle.
	if err := testSessionLifecycle(context.Background(), "", alphaURL, betaURL, configuration, false, false, false); err != nil {
		t.Error("session lifecycle test failed:", err)
	}
}

func TestSynchronizationGOROOTSrcToAlpha(t *testing.T) {
	// Check the end-to-end test mode and compute the source synchronization
	// root accordingly. If no mode has been specified, then skip the test.
	endToEndTestMode := os.Getenv("MUTAGEN_TEST_END_TO_END")
	var sourceRoot string
	if endToEndTestMode == "" {
		t.Skip()
	} else if endToEndTestMode == "full" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src")
	} else if endToEndTestMode == "slim" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src", "bufio")
	} else {
		t.Fatal("unknown end-to-end test mode specified:", endToEndTestMode)
	}

	// Allow the test to run in parallel.
	t.Parallel()

	// Calculate alpha and beta paths.
	alphaRoot := filepath.Join(t.TempDir(), "alpha")
	betaRoot := sourceRoot

	// Compute alpha and beta URLs.
	alphaURL := &url.URL{Path: alphaRoot}
	betaURL := &url.URL{Path: betaRoot}

	// Compute configuration. We use defaults for everything.
	configuration := &synchronization.Configuration{}

	// Test the session lifecycle.
	if err := testSessionLifecycle(context.Background(), "", alphaURL, betaURL, configuration, false, false, false); err != nil {
		t.Error("session lifecycle test failed:", err)
	}
}

func TestSynchronizationGOROOTSrcToBetaInMemory(t *testing.T) {
	// Define configuration variations.
	testCases := []*synchronization.Configuration{
		{},
		{
			CompressionAlgorithm: compression.Algorithm_AlgorithmNone,
		},
		{
			HashingAlgorithm: hashing.Algorithm_AlgorithmSHA256,
		},
	}
	if hashing.Algorithm_AlgorithmXXH128.SupportStatus() == hashing.AlgorithmSupportStatusSupported {
		testCases = append(testCases, &synchronization.Configuration{
			HashingAlgorithm: hashing.Algorithm_AlgorithmXXH128,
		})
	}
	if compression.Algorithm_AlgorithmZstandard.SupportStatus() == compression.AlgorithmSupportStatusSupported {
		testCases = append(testCases, &synchronization.Configuration{
			CompressionAlgorithm: compression.Algorithm_AlgorithmZstandard,
		})
	}

	// Check the end-to-end test mode and compute the source synchronization
	// root accordingly. If no mode has been specified, then skip the test.
	endToEndTestMode := os.Getenv("MUTAGEN_TEST_END_TO_END")
	var sourceRoot string
	if endToEndTestMode == "" {
		t.Skip()
	} else if endToEndTestMode == "full" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src")
	} else if endToEndTestMode == "slim" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src", "bufio")
	} else {
		t.Fatal("unknown end-to-end test mode specified:", endToEndTestMode)
	}

	// Allow the test to run in parallel.
	t.Parallel()

	// Loop over configurations and test the session lifecycle.
	for _, configuration := range testCases {
		// Calculate alpha and beta paths.
		alphaRoot := sourceRoot
		betaRoot := filepath.Join(t.TempDir(), "beta")

		// Compute alpha and beta URLs. We use a special protocol with a custom
		// handler to indicate an in-memory connection.
		alphaURL := &url.URL{Path: alphaRoot}
		betaURL := &url.URL{
			Protocol: netpipe.Protocol_Netpipe,
			Path:     betaRoot,
		}

		// Test the session lifecycle.
		if err := testSessionLifecycle(context.Background(), "", alphaURL, betaURL, configuration, false, false, false); err != nil {
			t.Error("session lifecycle test failed:", err)
		}
	}
}

func TestSynchronizationGOROOTSrcToBetaOverSSH(t *testing.T) {
	// If localhost SSH support isn't available, then skip this test.
	if os.Getenv("MUTAGEN_TEST_SSH") != "true" {
		t.Skip()
	}

	// Check the end-to-end test mode and compute the source synchronization
	// root accordingly. If no mode has been specified, then skip the test.
	endToEndTestMode := os.Getenv("MUTAGEN_TEST_END_TO_END")
	var sourceRoot string
	if endToEndTestMode == "" {
		t.Skip()
	} else if endToEndTestMode == "full" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src")
	} else if endToEndTestMode == "slim" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src", "bufio")
	} else {
		t.Fatal("unknown end-to-end test mode specified:", endToEndTestMode)
	}

	// Allow the test to run in parallel.
	t.Parallel()

	// Calculate alpha and beta paths.
	alphaRoot := sourceRoot
	betaRoot := filepath.Join(t.TempDir(), "beta")

	// Compute alpha and beta URLs.
	alphaURL := &url.URL{Path: alphaRoot}
	betaURL := &url.URL{
		Protocol: url.Protocol_SSH,
		Host:     "localhost",
		Path:     betaRoot,
	}

	// Compute configuration. We use defaults for everything.
	configuration := &synchronization.Configuration{}

	// Test the session lifecycle.
	if err := testSessionLifecycle(context.Background(), "", alphaURL, betaURL, configuration, false, false, false); err != nil {
		t.Error("session lifecycle test failed:", err)
	}
}

// testWindowsDockerTransportPrompter is a prompting.Prompter implementation
// that will answer "yes" to all prompts. It's needed to confirm container
// restart behavior in the Docker transport on Windows.
type testWindowsDockerTransportPrompter struct{}

func (t *testWindowsDockerTransportPrompter) Message(_ string) error {
	return nil
}

func (t *testWindowsDockerTransportPrompter) Prompt(_ string) (string, error) {
	return "yes", nil
}

func TestSynchronizationGOROOTSrcToBetaOverDocker(t *testing.T) {
	// If Docker test support isn't available, then skip this test.
	if os.Getenv("MUTAGEN_TEST_DOCKER") != "true" {
		t.Skip()
	}

	// Check the end-to-end test mode and compute the source synchronization
	// root accordingly. If no mode has been specified, then skip the test.
	endToEndTestMode := os.Getenv("MUTAGEN_TEST_END_TO_END")
	var sourceRoot string
	if endToEndTestMode == "" {
		t.Skip()
	} else if endToEndTestMode == "full" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src")
	} else if endToEndTestMode == "slim" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src", "bufio")
	} else {
		t.Fatal("unknown end-to-end test mode specified:", endToEndTestMode)
	}

	// If we're on a POSIX system, then allow this test to run concurrently with
	// other tests. On Windows, agent installation into Docker containers
	// requires temporarily halting the container, meaning that multiple
	// simultaneous Docker tests could conflict with each other, so we don't
	// allow Docker-based tests to run concurrently on Windows.
	if runtime.GOOS != "windows" {
		t.Parallel()
	}

	// If we're on Windows, register a prompter that will answer yes to
	// questions about stoping and restarting containers.
	var prompter string
	if runtime.GOOS == "windows" {
		if p, err := prompting.RegisterPrompter(&testWindowsDockerTransportPrompter{}); err != nil {
			t.Fatal("unable to register prompter:", err)
		} else {
			prompter = p
			defer prompting.UnregisterPrompter(prompter)
		}
	}

	// Create a unique directory name for synchronization into the container. We
	// don't clean it up, because it will be wiped out when the test container
	// is deleted.
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		t.Fatal("unable to create random directory UUID:", err)
	}

	// Calculate alpha and beta paths.
	alphaRoot := sourceRoot
	betaRoot := "~/" + randomUUID.String()

	// Grab Docker environment variables.
	environment := make(map[string]string, len(url.DockerEnvironmentVariables))
	for _, variable := range url.DockerEnvironmentVariables {
		environment[variable] = os.Getenv(variable)
	}

	// Compute alpha and beta URLs.
	alphaURL := &url.URL{Path: alphaRoot}
	betaURL := &url.URL{
		Protocol:    url.Protocol_Docker,
		User:        os.Getenv("MUTAGEN_TEST_DOCKER_USERNAME"),
		Host:        os.Getenv("MUTAGEN_TEST_DOCKER_CONTAINER_NAME"),
		Path:        betaRoot,
		Environment: environment,
	}

	// Verify that the beta URL is valid (this will validate the test
	// environment variables as well).
	if err := betaURL.EnsureValid(); err != nil {
		t.Fatal("beta URL is invalid:", err)
	}

	// Compute configuration. We use defaults for everything.
	configuration := &synchronization.Configuration{}

	// Test the session lifecycle.
	if err := testSessionLifecycle(context.Background(), prompter, alphaURL, betaURL, configuration, false, false, false); err != nil {
		t.Error("session lifecycle test failed:", err)
	}
}

func init() {
	// HACK: Disable lazy listener initialization since it makes test
	// coordination difficult.
	local.DisableLazyListenerInitialization = true
}

func TestForwardingToHTTPDemo(t *testing.T) {
	// If Docker test support isn't available, then skip this test.
	if os.Getenv("MUTAGEN_TEST_DOCKER") != "true" {
		t.Skip()
	}

	// If we're on a POSIX system, then allow this test to run concurrently with
	// other tests. On Windows, agent installation into Docker containers
	// requires temporarily halting the container, meaning that multiple
	// simultaneous Docker tests could conflict with each other, so we don't
	// allow Docker-based tests to run concurrently on Windows.
	if runtime.GOOS != "windows" {
		t.Parallel()
	}

	// If we're on Windows, register a prompter that will answer yes to
	// questions about stoping and restarting containers.
	var prompter string
	if runtime.GOOS == "windows" {
		if p, err := prompting.RegisterPrompter(&testWindowsDockerTransportPrompter{}); err != nil {
			t.Fatal("unable to register prompter:", err)
		} else {
			prompter = p
			defer prompting.UnregisterPrompter(prompter)
		}
	}

	// Pick a local listener address.
	listenerProtocol := "tcp"
	listenerAddress := "localhost:7070"

	// Compute source and destination URLs.
	source := &url.URL{
		Kind:     url.Kind_Forwarding,
		Protocol: url.Protocol_Local,
		Path:     listenerProtocol + ":" + listenerAddress,
	}
	destination := &url.URL{
		Kind:     url.Kind_Forwarding,
		Protocol: url.Protocol_Docker,
		User:     os.Getenv("MUTAGEN_TEST_DOCKER_USERNAME"),
		Host:     os.Getenv("MUTAGEN_TEST_DOCKER_CONTAINER_NAME"),
		Path:     "tcp:" + constants.HTTPDemoBindAddress,
	}

	// Verify that the destination URL is valid (this will validate the test
	// environment variables as well).
	if err := destination.EnsureValid(); err != nil {
		t.Fatal("beta URL is invalid:", err)
	}

	// Create a function to perform a simple HTTP request and ensure that the
	// returned contents are as expected.
	performHTTPRequest := func() error {
		// Perform the request and defer closure of the response body.
		response, err := http.Get(fmt.Sprintf("http://%s/", listenerAddress))
		if err != nil {
			return fmt.Errorf("unable to perform HTTP GET: %w", err)
		}
		defer response.Body.Close()

		// Read the full body.
		message, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("unable to read response body: %w", err)
		}

		// Compare the message.
		if string(message) != constants.HTTPDemoResponse {
			return errors.New("response does not match expected")
		}

		// Success.
		return nil
	}

	// Create a context to regulate the test.
	ctx := context.Background()

	// Create a forwarding session. Note that we've disabled lazy listener
	// initialization using a private API in the init function above, so we can
	// be sure that the listener has been established (with some non-empty
	// backlog) by the time creation is complete.
	sessionID, err := forwardingManager.Create(
		ctx,
		source,
		destination,
		&forwarding.Configuration{},
		&forwarding.Configuration{},
		&forwarding.Configuration{},
		"testForwardingSession",
		nil,
		false,
		prompter,
	)
	if err != nil {
		t.Fatal("unable to create session:", err)
	}

	// Attempt an HTTP request.
	// TODO: Attempt a more complicated exchange here. Maybe gRPC?
	if err := performHTTPRequest(); err != nil {
		t.Error("error performing forwarded HTTP request:", err)
	}

	// Create a session selection specification.
	selection := &selection.Selection{
		Specifications: []string{sessionID},
	}

	// Pause the session.
	if err := forwardingManager.Pause(ctx, selection, ""); err != nil {
		t.Error("unable to pause session:", err)
	}

	// Resume the session.
	if err := forwardingManager.Resume(ctx, selection, ""); err != nil {
		t.Error("unable to resume session:", err)
	}

	// Attempt an HTTP request.
	// TODO: Attempt a more complicated exchange here. Maybe gRPC?
	if err := performHTTPRequest(); err != nil {
		t.Error("error performing forwarded HTTP request:", err)
	}

	// Attempt an additional resume (this should be a no-op).
	if err := forwardingManager.Resume(ctx, selection, ""); err != nil {
		t.Error("unable to perform additional resume:", err)
	}

	// Terminate the session.
	if err := forwardingManager.Terminate(ctx, selection, ""); err != nil {
		t.Error("unable to terminate session:", err)
	}

	// TODO: Verify that cleanup took place.
}

// TODO: Add forwarding tests using the netpipe protocol.

// waitForSynchronizationCycleWithConflicts waits for a synchronization cycle to
// complete and returns the session state, expecting conflicts to be present.
func waitForSynchronizationCycleWithConflicts(ctx context.Context, sessionID string) (*synchronization.State, error) {
	// Create a session selection specification.
	selection := &selection.Selection{
		Specifications: []string{sessionID},
	}

	// Perform waiting.
	var previousStateIndex uint64
	var states []*synchronization.State
	var err error
	for {
		previousStateIndex, states, err = synchronizationManager.List(ctx, selection, previousStateIndex)
		if err != nil {
			return nil, fmt.Errorf("unable to list session states: %w", err)
		} else if len(states) != 1 {
			return nil, errors.New("invalid number of session states returned")
		} else if states[0].SuccessfulCycles > 0 {
			return states[0], nil
		}
	}
}

// TestSynchronizationFlushResolveConflictsFor tests the resolve-conflicts-for
// option of the flush operation.
func TestSynchronizationFlushResolveConflictsFor(t *testing.T) {
	// Allow this test to run in parallel.
	t.Parallel()

	// Create a context for the test.
	ctx := context.Background()

	// Calculate alpha and beta paths.
	directory := t.TempDir()
	alphaRoot := filepath.Join(directory, "alpha")
	betaRoot := filepath.Join(directory, "beta")

	// Create the alpha and beta directories.
	if err := os.MkdirAll(alphaRoot, 0700); err != nil {
		t.Fatal("unable to create alpha directory:", err)
	}
	if err := os.MkdirAll(betaRoot, 0700); err != nil {
		t.Fatal("unable to create beta directory:", err)
	}

	// Compute alpha and beta URLs.
	alphaURL := &url.URL{Path: alphaRoot}
	betaURL := &url.URL{Path: betaRoot}

	// Compute configuration with two-way-safe mode (which creates conflicts
	// instead of auto-resolving them).
	configuration := &synchronization.Configuration{
		SynchronizationMode: core.SynchronizationMode_SynchronizationModeTwoWaySafe,
	}

	// Create a session.
	sessionID, err := synchronizationManager.Create(
		ctx,
		alphaURL, betaURL,
		configuration,
		&synchronization.Configuration{},
		&synchronization.Configuration{},
		"testResolveConflictsFor",
		nil,
		false,
		"",
	)
	if err != nil {
		t.Fatal("unable to create session:", err)
	}

	// Ensure session termination on test completion.
	defer func() {
		selection := &selection.Selection{Specifications: []string{sessionID}}
		synchronizationManager.Terminate(ctx, selection, "")
	}()

	// Wait for the initial synchronization cycle to complete.
	if err := waitForSuccessfulSynchronizationCycle(ctx, sessionID, false, false, false); err != nil {
		t.Fatal("unable to wait for initial synchronization:", err)
	}

	// Create a session selection specification.
	selection := &selection.Selection{
		Specifications: []string{sessionID},
	}

	// Pause the session so we can create conflicting content without
	// interference from the synchronization loop.
	if err := synchronizationManager.Pause(ctx, selection, ""); err != nil {
		t.Fatal("unable to pause session:", err)
	}

	// Create conflicting files on both sides.
	alphaContent := []byte("alpha content")
	betaContent := []byte("beta content")
	conflictFile := "conflict.txt"
	if err := os.WriteFile(filepath.Join(alphaRoot, conflictFile), alphaContent, 0600); err != nil {
		t.Fatal("unable to write alpha conflict file:", err)
	}
	if err := os.WriteFile(filepath.Join(betaRoot, conflictFile), betaContent, 0600); err != nil {
		t.Fatal("unable to write beta conflict file:", err)
	}

	// Resume the session.
	if err := synchronizationManager.Resume(ctx, selection, ""); err != nil {
		t.Fatal("unable to resume session:", err)
	}

	// Wait for a synchronization cycle and verify conflicts exist.
	state, err := waitForSynchronizationCycleWithConflicts(ctx, sessionID)
	if err != nil {
		t.Fatal("unable to wait for synchronization cycle:", err)
	}
	if len(state.Conflicts) == 0 {
		t.Fatal("expected conflicts but none were found")
	}

	// Test 1: Flush with resolveConflictsFor="alpha" - alpha should win.
	if err := synchronizationManager.Flush(ctx, selection, "", false, "alpha"); err != nil {
		t.Fatal("unable to flush with alpha resolution:", err)
	}

	// Wait for the flush to complete.
	state, err = waitForSynchronizationCycleWithConflicts(ctx, sessionID)
	if err != nil {
		t.Fatal("unable to wait for synchronization after alpha flush:", err)
	}

	// Verify no conflicts remain after alpha resolution.
	if len(state.Conflicts) > 0 {
		t.Errorf("expected no conflicts after alpha resolution, but found %d", len(state.Conflicts))
	}

	// Verify beta has alpha's content (alpha wins).
	betaFileContent, err := os.ReadFile(filepath.Join(betaRoot, conflictFile))
	if err != nil {
		t.Fatal("unable to read beta file after alpha resolution:", err)
	}
	if string(betaFileContent) != string(alphaContent) {
		t.Errorf("beta file should have alpha content after alpha resolution: got %q, want %q",
			string(betaFileContent), string(alphaContent))
	}

	// Pause the session again to create new conflicts.
	if err := synchronizationManager.Pause(ctx, selection, ""); err != nil {
		t.Fatal("unable to pause session for second test:", err)
	}

	// Create new conflicting files on both sides.
	alphaContent2 := []byte("alpha content 2")
	betaContent2 := []byte("beta content 2")
	conflictFile2 := "conflict2.txt"
	if err := os.WriteFile(filepath.Join(alphaRoot, conflictFile2), alphaContent2, 0600); err != nil {
		t.Fatal("unable to write alpha conflict file 2:", err)
	}
	if err := os.WriteFile(filepath.Join(betaRoot, conflictFile2), betaContent2, 0600); err != nil {
		t.Fatal("unable to write beta conflict file 2:", err)
	}

	// Resume the session.
	if err := synchronizationManager.Resume(ctx, selection, ""); err != nil {
		t.Fatal("unable to resume session for second test:", err)
	}

	// Wait for conflicts to appear.
	state, err = waitForSynchronizationCycleWithConflicts(ctx, sessionID)
	if err != nil {
		t.Fatal("unable to wait for second conflict:", err)
	}
	if len(state.Conflicts) == 0 {
		t.Fatal("expected conflicts for second test but none were found")
	}

	// Test 2: Flush with resolveConflictsFor="beta" - beta should win.
	if err := synchronizationManager.Flush(ctx, selection, "", false, "beta"); err != nil {
		t.Fatal("unable to flush with beta resolution:", err)
	}

	// Wait for the flush to complete.
	state, err = waitForSynchronizationCycleWithConflicts(ctx, sessionID)
	if err != nil {
		t.Fatal("unable to wait for synchronization after beta flush:", err)
	}

	// Verify no conflicts remain after beta resolution.
	if len(state.Conflicts) > 0 {
		t.Errorf("expected no conflicts after beta resolution, but found %d", len(state.Conflicts))
	}

	// Verify alpha has beta's content (beta wins).
	alphaFileContent, err := os.ReadFile(filepath.Join(alphaRoot, conflictFile2))
	if err != nil {
		t.Fatal("unable to read alpha file after beta resolution:", err)
	}
	if string(alphaFileContent) != string(betaContent2) {
		t.Errorf("alpha file should have beta content after beta resolution: got %q, want %q",
			string(alphaFileContent), string(betaContent2))
	}
}
