/*
Copyright 2023 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package actors

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/microsoft/durabletask-go/api"
	"github.com/microsoft/durabletask-go/backend"

	"github.com/dapr/dapr/pkg/actors"
	diag "github.com/dapr/dapr/pkg/diagnostics"
	invokev1 "github.com/dapr/dapr/pkg/messaging/v1"
	internalsv1pb "github.com/dapr/dapr/pkg/proto/internals/v1"
	internalv1pb "github.com/dapr/dapr/pkg/proto/internals/v1"
)

const ExecuteActivityMethod = "Execute"

const CompletePendingActivityMethod = "CompletePendingActivity"

var ErrDuplicateInvocation = errors.New("duplicate invocation")

const activityStateKey = "activityState"

type activityActor struct {
	actorID                   string
	actorRuntime              actors.Actors
	scheduler                 activityScheduler
	state                     *activityState
	cachingDisabled           bool
	defaultTimeout            time.Duration
	reminderInterval          time.Duration
	config                    actorsBackendConfig
	onPendingActivityResponse onPendingActivityResponse
}

// ActivityRequest represents a request by a worklow to invoke an activity.
type ActivityRequest struct {
	HistoryEvent []byte
}

type activityState struct {
	EventPayload []byte
}

type ActivityResult struct {
	ActivityEvent []byte
	Success       bool
	ResultEvent   []byte
}

// activityScheduler is a func interface for pushing activity work items into the backend
type activityScheduler func(ctx context.Context, wi *backend.ActivityWorkItem) error

type onPendingActivityResponse func(*backend.ActivityResponse) error

type activityActorOpts struct {
	cachingDisabled  bool
	defaultTimeout   time.Duration
	reminderInterval time.Duration
}

// NewActivityActor creates an internal activity actor for executing workflow activity logic.
func NewActivityActor(scheduler activityScheduler, onPendingActivityResponse onPendingActivityResponse, backendConfig actorsBackendConfig, opts *activityActorOpts) actors.InternalActorFactory {
	return func(actorType string, actorID string, actors actors.Actors) actors.InternalActor {
		a := &activityActor{
			actorID:                   actorID,
			actorRuntime:              actors,
			scheduler:                 scheduler,
			defaultTimeout:            1 * time.Hour,
			reminderInterval:          1 * time.Minute,
			config:                    backendConfig,
			cachingDisabled:           opts.cachingDisabled,
			onPendingActivityResponse: onPendingActivityResponse,
		}

		if opts.defaultTimeout > 0 {
			a.defaultTimeout = opts.defaultTimeout
		}
		if opts.reminderInterval > 0 {
			a.reminderInterval = opts.reminderInterval
		}

		return a
	}
}

// InvokeMethod implements actors.InternalActor and schedules the background execution of a workflow activity.
// Activities are scheduled by workflows and can execute for arbitrary lengths of time. Instead of executing
// activity logic directly, InvokeMethod creates a reminder that executes the activity logic. InvokeMethod
// returns immediately after creating the reminder, enabling the workflow to continue processing other events
// in parallel.
func (a *activityActor) InvokeMethod(ctx context.Context, req *internalv1pb.InternalInvokeRequest, md map[string][]string) ([]byte, error) {
	wfLogger.Debugf("Activity actor '%s': invoking method '%s'", req.GetActor().GetActorId(), req.GetMessage().GetMethod())

	if req.GetMessage().GetMethod() == CompletePendingActivityMethod {
		err := a.executeOnPendingActivity(req.GetMessage().GetData().GetValue())
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	var ar ActivityRequest
	if err := actors.DecodeInternalActorData(bytes.NewReader(req.GetMessage().GetData().GetValue()), &ar); err != nil {
		return nil, fmt.Errorf("failed to decode activity request: %w", err)
	}

	// // Try to load activity state. If we find any, that means the activity invocation is a duplicate.
	// if _, err := a.loadActivityState(ctx); err != nil {
	// 	return nil, err
	// }

	if req.GetMessage().GetMethod() == "PurgeWorkflowState" {
		return nil, a.purgeActivityState(ctx)
	}

	if req.GetMessage().GetMethod() != ExecuteActivityMethod {
		return nil, fmt.Errorf("unknown method %v", req.GetMessage().GetMethod())
	}

	// Save the request details to the state store in case we need it after recovering from a failure.
	err := a.saveActivityState(ctx, &activityState{
		EventPayload: ar.HistoryEvent,
	})
	if err != nil {
		return nil, err
	}

	// The actual execution is triggered by a reminder
	err = a.createReliableReminder(ctx, nil, 0)
	return nil, err
}

func (a *activityActor) executeOnPendingActivity(response []byte) error {
	res, err := backend.UnmarshalActivityResponse(response)
	if err != nil {
		return fmt.Errorf("error decoding activity response %w", err)
	}
	err = a.onPendingActivityResponse(res)
	if err != nil {
		return err
	}
	return nil
}

// InvokeReminder implements actors.InternalActor and executes the activity logic.
func (a *activityActor) InvokeReminder(ctx context.Context, reminder actors.InternalActorReminder, metadata map[string][]string) error {
	wfLogger.Debugf("Activity actor '%s': invoking reminder '%s'", reminder.ActorID, reminder.Name)

	var err error
	switch reminder.Name {
	case "run-activity":
		err = a.executeActivity(ctx, reminder.Name)
	case "run-activity-sync":
		err = a.executeActivitySync(ctx, reminder)
	default:
		panic(fmt.Sprintf("unexpected reminder name: %s", reminder.Name))
	}

	var recoverableErr *recoverableError
	// Returning nil signals that we want the execution to be retried in the next period interval
	switch {
	case err == nil:
		// We delete the reminder on success and on non-recoverable errors.
		return actors.ErrReminderCanceled
	case errors.Is(err, context.DeadlineExceeded):
		wfLogger.Warnf("%s: execution of '%s' timed-out and will be retried later: %v", reminder.ActorID, reminder.Name, err)
		return nil
	case errors.Is(err, context.Canceled):
		wfLogger.Warnf("%s: received cancellation signal while waiting for activity execution '%s'", reminder.ActorID, reminder.Name)
		return nil
	case errors.As(err, &recoverableErr):
		wfLogger.Warnf("%s: execution failed with a recoverable error and will be retried later: %v", reminder.ActorID, err)
		return nil
	default: // Other error
		wfLogger.Errorf("%s: execution failed with a non-recoverable error: %v", reminder.ActorID, err)
		// TODO: Reply with a failure - this requires support from durabletask-go to produce TaskFailure results
		return actors.ErrReminderCanceled
	}
}

func (a *activityActor) getWorkflowID() (string, error) {
	endIndex := strings.Index(a.actorID, "::")
	if endIndex < 0 {
		return "", fmt.Errorf("invalid activity actor ID: '%s'", a.actorID)
	}
	return a.actorID[0:endIndex], nil
}

func (a *activityActor) executeActivity(ctx context.Context, reminderName string) error {
	state, err := a.loadActivityState(ctx)
	if err != nil {
		return err
	}
	if state == nil {
		return errors.New("no activity state found")
	}

	eventPayload := state.EventPayload

	taskEvent, err := backend.UnmarshalHistoryEvent(eventPayload)
	if err != nil {
		return err
	}

	workflowID, err := a.getWorkflowID()
	if err != nil {
		return err
	}

	wi := &backend.ActivityWorkItem{
		SequenceNumber: int64(taskEvent.GetEventId()),
		InstanceID:     api.InstanceID(workflowID),
		NewEvent:       taskEvent,
		Properties:     make(map[string]interface{}),
	}

	wfLogger.Debugf("Activity actor '%s': scheduling activity '%s' for workflow with instanceId '%s'", a.actorID, reminderName, wi.InstanceID)
	scheduleCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err = a.scheduler(scheduleCtx, wi)
	if errors.Is(err, context.DeadlineExceeded) {
		return newRecoverableError(fmt.Errorf("timed-out trying to schedule an activity execution - this can happen if too many activities are running in parallel or if the workflow engine isn't running: %w", err))
	} else if err != nil {
		return newRecoverableError(fmt.Errorf("failed to schedule an activity execution: %w", err))
	}

	protoEvent, err := backend.MarshalHistoryEvent(wi.NewEvent)
	if err != nil {
		return err
	}

	dataEnc, err := json.Marshal(ActivityResult{
		ActivityEvent: protoEvent,
		Success:       false,
	})
	if err != nil {
		return fmt.Errorf("failed to encode data as JSON: %w", err)
	}
	err = a.actorRuntime.CreateReminder(ctx, &actors.CreateReminderRequest{
		ActorType: a.config.activityActorType,
		ActorID:   a.actorID,
		Data:      dataEnc,
		DueTime:   a.defaultTimeout.String(),
		Name:      "run-activity-sync",
		Period:    a.reminderInterval.String(),
	})
	if err != nil {
		return newRecoverableError(err)
	}
	return nil
}

func (a *activityActor) executeActivitySync(ctx context.Context, reminder actors.InternalActorReminder) error {

	var activityResult ActivityResult
	err := json.Unmarshal(reminder.Data, &activityResult)
	if err != nil {
		return fmt.Errorf("error unmarshaling reminder json %v", err)
	}

	taskEvent, err := backend.UnmarshalHistoryEvent(activityResult.ActivityEvent)
	if err != nil {
		return err
	}
	activityName := ""
	if ts := taskEvent.GetTaskScheduled(); ts != nil {
		activityName = ts.GetName()
	} else {
		return fmt.Errorf("invalid activity task event: '%s'", taskEvent.String())
	}

	workflowID, err := a.getWorkflowID()
	if err != nil {
		return err
	}

	// Activity execution started
	executionStatus := ""
	elapsed := float64(0)
	// Record metrics on exit
	defer func() {
		if executionStatus != "" {
			diag.DefaultWorkflowMonitoring.ActivityExecutionEvent(ctx, activityName, executionStatus, elapsed)
		}
	}()

	if !activityResult.Success {
		// Activity execution failed with recoverable error
		executionStatus = diag.StatusRecoverable

		err = a.createReliableReminder(ctx, nil, a.reminderInterval)
		if err != nil {
			wfLogger.Errorf("error re-creating activity reminder %v", err)
			return newRecoverableError(errExecutionAborted) // AbandonActivityWorkItem was called
		}
		// aborted execution has been successfully processed, the activity execution will start from the beginning with the reminder that has been created
		return nil
	}

	wfLogger.Debugf("Activity actor '%s': activity completed for workflow with instanceId '%s' activityName '%s'", a.actorID, workflowID, activityName)

	resultEvent, err := backend.UnmarshalHistoryEvent(activityResult.ResultEvent)
	if err != nil {
		return fmt.Errorf("error unmarshaling result event %v", err)
	}

	// publish the result back to the workflow actor as a new event to be processed
	resultData, err := backend.MarshalHistoryEvent(resultEvent)
	if err != nil {
		// Returning non-recoverable error
		executionStatus = diag.StatusFailed
		return err
	}
	req := internalsv1pb.
		NewInternalInvokeRequest(AddWorkflowEventMethod).
		WithActor(a.config.workflowActorType, workflowID).
		WithData(resultData).
		WithContentType(invokev1.OctetStreamContentType)

	_, err = a.actorRuntime.Call(ctx, req)
	switch {
	case err != nil:
		// Returning recoverable error, record metrics
		executionStatus = diag.StatusRecoverable
		return newRecoverableError(fmt.Errorf("failed to invoke '%s' method on workflow actor: %w", AddWorkflowEventMethod, err))
	case resultEvent.GetTaskCompleted() != nil:
		// Activity execution completed successfully
		executionStatus = diag.StatusSuccess
	case resultEvent.GetTaskFailed() != nil:
		// Activity execution failed
		executionStatus = diag.StatusFailed
	}
	return nil
}

// InvokeTimer implements actors.InternalActor
func (*activityActor) InvokeTimer(ctx context.Context, timer actors.InternalActorReminder, metadata map[string][]string) error {
	return errors.New("timers are not implemented")
}

// DeactivateActor implements actors.InternalActor
func (a *activityActor) DeactivateActor(ctx context.Context) error {
	wfLogger.Debugf("Activity actor '%s': deactivating", a.actorID)
	a.state = nil // A bit of extra caution, shouldn't be necessary
	return nil
}

func (a *activityActor) loadActivityState(ctx context.Context) (*activityState, error) {
	// See if the state for this actor is already cached in memory.
	if a.state != nil {
		return a.state, nil
	}

	// Loading from the state store is only expected in process failure recovery scenarios.
	wfLogger.Debugf("Activity actor '%s': loading activity state", a.actorID)

	req := actors.GetStateRequest{
		ActorType: a.config.activityActorType,
		ActorID:   a.actorID,
		Key:       activityStateKey,
	}
	res, err := a.actorRuntime.GetState(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to load activity state: %w", err)
	}

	if len(res.Data) == 0 {
		// no data was found - this is expected on the initial invocation of the activity actor.
		return nil, nil
	}

	state := &activityState{}
	err = json.Unmarshal(res.Data, state)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal activity state: %w", err)
	}
	return state, nil
}

func (a *activityActor) saveActivityState(ctx context.Context, state *activityState) error {
	req := actors.TransactionalRequest{
		ActorType: a.config.activityActorType,
		ActorID:   a.actorID,
		Operations: []actors.TransactionalOperation{{
			Operation: actors.Upsert,
			Request: actors.TransactionalUpsert{
				Key:   activityStateKey,
				Value: state,
			},
		}},
	}
	if err := a.actorRuntime.TransactionalStateOperation(ctx, &req); err != nil {
		return fmt.Errorf("failed to save activity state: %w", err)
	}

	if !a.cachingDisabled {
		a.state = state
	}
	return nil
}

func (a *activityActor) purgeActivityState(ctx context.Context) error {
	wfLogger.Debugf("Activity actor '%s': purging activity state", a.actorID)
	err := a.actorRuntime.TransactionalStateOperation(ctx, &actors.TransactionalRequest{
		ActorType: a.config.activityActorType,
		ActorID:   a.actorID,
		Operations: []actors.TransactionalOperation{{
			Operation: actors.Delete,
			Request: actors.TransactionalDelete{
				Key: activityStateKey,
			},
		}},
	})
	if err != nil {
		return fmt.Errorf("failed to delete activity state with error: %w", err)
	}
	a.state = nil

	return nil
}

func (a *activityActor) createReliableReminder(ctx context.Context, data any, delay time.Duration) error {
	const reminderName = "run-activity"
	wfLogger.Debugf("Activity actor '%s': creating reminder '%s' for immediate execution", a.actorID, reminderName)
	dataEnc, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode data as JSON: %w", err)
	}
	return a.actorRuntime.CreateReminder(ctx, &actors.CreateReminderRequest{
		ActorType: a.config.activityActorType,
		ActorID:   a.actorID,
		Data:      dataEnc,
		DueTime:   delay.String(),
		Name:      reminderName,
		Period:    a.reminderInterval.String(),
	})
}
