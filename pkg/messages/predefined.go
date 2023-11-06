/*
Copyright 2022 The Dapr Authors
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

package messages

import (
	"net/http"

	grpcCodes "google.golang.org/grpc/codes"
)

const (
	// Generic.
	ErrInfoDomain = "dapr.io"

	// Error Reasons prefixes INIT
	ReasonPrefixInitCli        = "DAPR_CLI_INIT_"
	ReasonPrefixInitSelfHosted = "DAPR_SELF_HOSTED_INIT_"
	ReasonPrefixInitK8s        = "DAPR_K8S_INIT_"
	ReasonPrefixInitInvoke     = "DAPR_INVOKE_INIT_"

	// Error Reasons prefixes RUNTIME
	ReasonPrefixRuntimeCli = "DAPR_RUNTIME_CLI_"
	ReasonPrefixSelfHosted = "DAPR_SELF_HOSTED_"
	ReasonPrefixDaprToDapr = "DAPR_RUNTIME_GRPC_"

	// Error Reasons COMPONENTS
	ReasonPrefixPubSub             = "DAPR_PUBSUB_"
	ReasonPrefixStateStore         = "DAPR_STATE_"
	ReasonPrefixBindings           = "DAPR_BINDING_"
	ReasonPrefixSecretStore        = "DAPR_SECRET_"
	ReasonPrefixConfigurationStore = "DAPR_CONFIGURATION_"
	ReasonPrefixLock               = "DAPR_LOCK_"
	ReasonPrefixNameResolution     = "DAPR_NAME_RESOLUTION_"
	ReasonPrefixMiddleware         = "DAPR_MIDDLEWARE_"

	// Resource types.

	// Http.
	ErrNotFound             = "method %q is not found"
	ErrMalformedRequestData = "can't serialize request data field: %s"

	// State.
	ErrStateGet        = "fail to get %s from state store %s: %s"
	ErrStateDelete     = "failed deleting state with key %s: %s"
	ErrStateSave       = "failed saving state in state store %s: %s"
	ErrStateDeleteBulk = "failed deleting state in state store %s: %s"

	// StateTransaction.
	ErrStateStoreNotSupported     = "state store %s doesn't support transaction"
	ErrNotSupportedStateOperation = "operation type %s not supported"
	ErrStateTransaction           = "error while executing state transaction: %s"

	// Binding.
	ErrInvokeOutputBinding = "error invoking output binding %s: %s"

	// PubSub.
	ErrPubsubNotConfigured      = "no pubsub is configured"
	ErrPubsubEmpty              = "pubsub name is empty"
	ErrPubsubNotFound           = "pubsub %s not found"
	ErrTopicEmpty               = "topic is empty in pubsub %s"
	ErrPubsubCloudEventsSer     = "error when marshalling cloud event envelope for topic %s pubsub %s: %s"
	ErrPubsubPublishMessage     = "error when publish to topic %s in pubsub %s: %s"
	ErrPubsubForbidden          = "topic %s is not allowed for app id %s"
	ErrPubsubCloudEventCreation = "cannot create cloudevent: %s"
	ErrPubsubUnmarshal          = "error when unmarshaling the request for topic %s pubsub %s: %s"
	ErrPubsubMarshal            = "error marshaling events to bytes for topic %s pubsub %s: %s"
	ErrPubsubGetSubscriptions   = "unable to get app subscriptions %s"
	ErrPublishOutbox            = "error while publishing outbox message: %s"

	// AppChannel.
	ErrChannelNotFound       = "app channel is not initialized"
	ErrInternalInvokeRequest = "parsing InternalInvokeRequest error: %s"
	ErrChannelInvoke         = "error invoking app channel: %s"

	// AppHealth.
	ErrAppUnhealthy = "app is not in a healthy state"

	// Actor.
	ErrActorInstanceMissing      = "actor instance is missing"
	ErrActorInvoke               = "error invoke actor method: %s"
	ErrActorReminderCreate       = "error creating actor reminder: %s"
	ErrActorReminderGet          = "error getting actor reminder: %s"
	ErrActorReminderDelete       = "error deleting actor reminder: %s"
	ErrActorTimerCreate          = "error creating actor timer: %s"
	ErrActorTimerDelete          = "error deleting actor timer: %s"
	ErrActorStateGet             = "error getting actor state: %s"
	ErrActorStateTransactionSave = "error saving actor transaction state: %s"

	// Configuration.
	ErrConfigurationStoresNotConfigured = "configuration stores not configured"
	ErrConfigurationStoreNotFound       = "configuration store %s not found"
	ErrConfigurationGet                 = "failed to get %s from Configuration store %s: %v"
	ErrConfigurationSubscribe           = "failed to subscribe %s from Configuration store %s: %v"
	ErrConfigurationUnsubscribe         = "failed to unsubscribe to configuration request %s: %v"
)

var (
	// Generic.
	ErrBadRequest       = APIError{message: "invalid request: %v", tag: "ERR_BAD_REQUEST", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrAPIUnimplemented = APIError{message: "this API is currently not implemented", tag: "ERR_API_UNIMPLEMENTED", httpCode: http.StatusNotImplemented, grpcCode: grpcCodes.Unimplemented}

	// HTTP.
	ErrBodyRead         = APIError{message: "failed to read request body: %v", tag: "ERR_BODY_READ", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrMalformedRequest = APIError{message: "failed deserializing HTTP body: %v", tag: "ERR_MALFORMED_REQUEST", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}

	// DirectMessaging.
	ErrDirectInvoke         = APIError{message: "failed to invoke, id: %s, err: %v", tag: "ERR_DIRECT_INVOKE", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrDirectInvokeNoAppID  = APIError{message: "failed getting app id either from the URL path or the header dapr-app-id", tag: "ERR_DIRECT_INVOKE", httpCode: http.StatusNotFound, grpcCode: grpcCodes.NotFound}
	ErrDirectInvokeNotReady = APIError{message: "invoke API is not ready", tag: "ERR_DIRECT_INVOKE", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}

	// Healthz.
	ErrHealthNotReady         = APIError{message: "dapr is not ready", tag: "ERR_HEALTH_NOT_READY", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrOutboundHealthNotReady = APIError{message: "dapr outbound is not ready", tag: "ERR_OUTBOUND_HEALTH_NOT_READY", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrHealthAppIDNotMatch    = APIError{message: "dapr app-id does not match", tag: "ERR_HEALTH_APPID_NOT_MATCH", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}

	// State.
	ErrStateStoresNotConfigured = APIError{
		message:       "state store is not configured",
		tag:           "ERR_STATE_STORE_NOT_CONFIGURED",
		httpCode:      http.StatusInternalServerError,
		grpcCode:      grpcCodes.FailedPrecondition,
		errInfoReason: ReasonPrefixStateStore + "NOT_CONFIGURED",
	}
	ErrStateStoreNotFound = APIError{
		message:       "state store %s is not found",
		tag:           "ERR_STATE_STORE_NOT_FOUND",
		httpCode:      http.StatusBadRequest,
		grpcCode:      grpcCodes.InvalidArgument,
		errInfoReason: ReasonPrefixStateStore + "NOT_FOUND",
	}
	ErrStateQueryFailed = APIError{
		message:       "failed query in state store %s: %s",
		tag:           "ERR_STATE_QUERY",
		httpCode:      http.StatusInternalServerError,
		grpcCode:      grpcCodes.Internal,
		errInfoReason: ReasonPrefixStateStore + "QUERY_FAILED",
	}
	ErrStateQueryUnsupported = APIError{
		message:       "state store does not support querying",
		tag:           "ERR_STATE_STORE_NOT_SUPPORTED",
		httpCode:      http.StatusInternalServerError,
		grpcCode:      grpcCodes.Internal,
		errInfoReason: ReasonPrefixStateStore + "QUERY_UNSUPPORTED",
	}
	ErrStateTooManyTransactionalOp = APIError{
		message:       "the transaction contains %d operations, which is more than what the state store supports: %d",
		tag:           "ERR_STATE_STORE_TOO_MANY_TRANSACTIONS",
		httpCode:      http.StatusBadRequest,
		grpcCode:      grpcCodes.InvalidArgument,
		errInfoReason: ReasonPrefixStateStore + "TOO_MANY_TRANSACTIONS",
	}

	// PubSub.
	ErrPubSubMetadataDeserialize = APIError{message: "failed deserializing metadata: %v", tag: "ERR_PUBSUB_REQUEST_METADATA", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}

	// Secrets.
	ErrSecretStoreNotConfigured = APIError{message: "secret store is not configured", tag: "ERR_SECRET_STORES_NOT_CONFIGURED", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.FailedPrecondition}
	ErrSecretStoreNotFound      = APIError{message: "failed finding secret store with key %s", tag: "ERR_SECRET_STORE_NOT_FOUND", httpCode: http.StatusUnauthorized, grpcCode: grpcCodes.InvalidArgument}
	ErrSecretPermissionDenied   = APIError{message: "access denied by policy to get %q from %q", tag: "ERR_PERMISSION_DENIED", httpCode: http.StatusForbidden, grpcCode: grpcCodes.PermissionDenied}
	ErrSecretGet                = APIError{message: "failed getting secret with key %s from secret store %s: %s", tag: "ERR_SECRET_GET", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrBulkSecretGet            = APIError{message: "failed getting secrets from secret store %s: %v", tag: "ERR_SECRET_GET", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}

	// Crypto.
	ErrCryptoProvidersNotConfigured = APIError{message: "crypto providers not configured", tag: "ERR_CRYPTO_PROVIDERS_NOT_CONFIGURED", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrCryptoProviderNotFound       = APIError{message: "crypto provider %s not found", tag: "ERR_CRYPTO_PROVIDER_NOT_FOUND", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrCryptoGetKey                 = APIError{message: "failed to retrieve key %s: %v", tag: "ERR_CRYPTO_KEY", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrCryptoOperation              = APIError{message: "failed to perform operation: %v", tag: "ERR_CRYPTO", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}

	// Actor.
	ErrActorReminderOpActorNotHosted = APIError{message: "operations on actor reminders are only possible on hosted actor types", tag: "ERR_ACTOR_REMINDER_NON_HOSTED", httpCode: http.StatusForbidden, grpcCode: grpcCodes.PermissionDenied}
	ErrActorRuntimeNotFound          = APIError{message: `the state store is not configured to use the actor runtime. Have you set the - name: actorStateStore value: "true" in your state store component file?`, tag: "ERR_ACTOR_RUNTIME_NOT_FOUND", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}

	// Lock.
	ErrLockStoresNotConfigured    = APIError{message: "lock store is not configured", tag: "ERR_LOCK_STORE_NOT_CONFIGURED", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.FailedPrecondition}
	ErrResourceIDEmpty            = APIError{message: "ResourceId is empty in lock store %s", tag: "ERR_MALFORMED_REQUEST", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrLockOwnerEmpty             = APIError{message: "LockOwner is empty in lock store %s", tag: "ERR_MALFORMED_REQUEST", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrExpiryInSecondsNotPositive = APIError{message: "ExpiryInSeconds is not positive in lock store %s", tag: "ERR_MALFORMED_REQUEST", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrLockStoreNotFound          = APIError{message: "lock store %s not found", tag: "ERR_LOCK_STORE_NOT_FOUND", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrTryLockFailed              = APIError{message: "failed to try acquiring lock: %s", tag: "ERR_TRY_LOCK", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrUnlockFailed               = APIError{message: "failed to release lock: %s", tag: "ERR_UNLOCK", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}

	// Workflow.
	ErrStartWorkflow                 = APIError{message: "error starting workflow '%s': %s", tag: "ERR_START_WORKFLOW", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrWorkflowGetResponse           = APIError{message: "error while getting workflow info on instance '%s': %s", tag: "ERR_GET_WORKFLOW", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrWorkflowNameMissing           = APIError{message: "workflow name is not configured", tag: "ERR_WORKFLOW_NAME_MISSING", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrInstanceIDTooLong             = APIError{message: "workflow instance ID exceeds the max length of %d characters", tag: "ERR_INSTANCE_ID_TOO_LONG", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrInvalidInstanceID             = APIError{message: "workflow instance ID '%s' is invalid: only alphanumeric and underscore characters are allowed", tag: "ERR_INSTANCE_ID_INVALID", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrWorkflowComponentDoesNotExist = APIError{message: "workflow component '%s' does not exist", tag: "ERR_WORKFLOW_COMPONENT_NOT_FOUND", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrMissingOrEmptyInstance        = APIError{message: "no instance ID was provided", tag: "ERR_INSTANCE_ID_PROVIDED_MISSING", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrWorkflowInstanceNotFound      = APIError{message: "unable to find workflow with the provided instance ID: %s", tag: "ERR_INSTANCE_ID_NOT_FOUND", httpCode: http.StatusNotFound, grpcCode: grpcCodes.NotFound}
	ErrNoOrMissingWorkflowComponent  = APIError{message: "no workflow component was provided", tag: "ERR_WORKFLOW_COMPONENT_MISSING", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrTerminateWorkflow             = APIError{message: "error terminating workflow '%s': %s", tag: "ERR_TERMINATE_WORKFLOW", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrMissingWorkflowEventName      = APIError{message: "missing workflow event name", tag: "ERR_WORKFLOW_EVENT_NAME_MISSING", httpCode: http.StatusBadRequest, grpcCode: grpcCodes.InvalidArgument}
	ErrRaiseEventWorkflow            = APIError{message: "error raising event on workflow '%s': %s", tag: "ERR_RAISE_EVENT_WORKFLOW", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrPauseWorkflow                 = APIError{message: "error pausing workflow %s: %s", tag: "ERR_PAUSE_WORKFLOW", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrResumeWorkflow                = APIError{message: "error resuming workflow %s: %s", tag: "ERR_RESUME_WORKFLOW", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
	ErrPurgeWorkflow                 = APIError{message: "error purging workflow %s: %s", tag: "ERR_PURGE_WORKFLOW", httpCode: http.StatusInternalServerError, grpcCode: grpcCodes.Internal}
)
