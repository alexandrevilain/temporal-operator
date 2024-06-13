// Licensed to Alexandre VILAIN under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Alexandre VILAIN licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package v1beta1

import (
	"github.com/google/uuid"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScheduleIntervalSpec matches times that can be expressed as:
//
//	Epoch + (n * every) + offset
//
//	where n is all integers ≥ 0.
//
// For example, an `every` of 1 hour with `offset` of zero would match every hour, on the hour. The same `every` but an `offset`
// of 19 minutes would match every `xx:19:00`. An `every` of 28 days with `offset` zero would match `2022-02-17T00:00:00Z`
// (among other times). The same `every` with `offset` of 3 days, 5 hours, and 23 minutes would match `2022-02-20T05:23:00Z`
// instead.
type ScheduleIntervalSpec struct {
	// Every describes the period to repeat the interval.
	//
	// +kubebuilder:validation:Required
	Every metav1.Duration `json:"every"`

	// Offset is a fixed offset added to the intervals period.
	// Defaults to 0.
	//
	// +optional
	Offset *metav1.Duration `json:"offset,omitempty"`
}

// If end < start, then end is interpreted as
// equal to start. This means you can use a Range with start set to a value, and
// end and step unset to represent a single value.
type ScheduleSecondMinuteRange struct {
	// Start of the range (inclusive).
	// Defaults to 0.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=59
	Start int32 `json:"start,omitempty"`

	// End of the range (inclusive).
	// Defaults to start.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=59
	End int32 `json:"end,omitempty"`

	// Step to be take between each value.
	// Defaults to 1.
	//
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=59
	Step int32 `json:"step,omitempty"`
}

func (r ScheduleSecondMinuteRange) GetStart() int32 { return r.Start }
func (r ScheduleSecondMinuteRange) GetEnd() int32   { return r.End }
func (r ScheduleSecondMinuteRange) GetStep() int32  { return r.Step }

// If end < start, then end is interpreted as
// equal to start. This means you can use a Range with start set to a value, and
// end and step unset to represent a single value.
type ScheduleHourRange struct {
	// Start of the range (inclusive).
	// Defaults to 0.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=23
	Start int32 `json:"start,omitempty"`

	// End of the range (inclusive).
	// Defaults to start.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=23
	End int32 `json:"end,omitempty"`

	// Step to be take between each value.
	// Defaults to 1.
	//
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=23
	Step int32 `json:"step,omitempty"`
}

func (r ScheduleHourRange) GetStart() int32 { return r.Start }
func (r ScheduleHourRange) GetEnd() int32   { return r.End }
func (r ScheduleHourRange) GetStep() int32  { return r.Step }

// If end < start, then end is interpreted as
// equal to start. This means you can use a Range with start set to a value, and
// end and step unset to represent a single value.
type ScheduleDayOfMonthRange struct {
	// Start of the range (inclusive).
	// Defaults to 1.
	//
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=31
	Start int32 `json:"start,omitempty"`

	// End of the range (inclusive).
	// Defaults to start.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=31
	End int32 `json:"end,omitempty"`

	// Step to be take between each value.
	// Defaults to 1.
	//
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=31
	Step int32 `json:"step,omitempty"`
}

func (r ScheduleDayOfMonthRange) GetStart() int32 { return r.Start }
func (r ScheduleDayOfMonthRange) GetEnd() int32   { return r.End }
func (r ScheduleDayOfMonthRange) GetStep() int32  { return r.Step }

// If end < start, then end is interpreted as
// equal to start. This means you can use a Range with start set to a value, and
// end and step unset to represent a single value.
type ScheduleMonthRange struct {
	// Start of the range (inclusive).
	// Defaults to 1.
	//
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=12
	Start int32 `json:"start,omitempty"`

	// End of the range (inclusive).
	// Defaults to start.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=12
	End int32 `json:"end,omitempty"`

	// Step to be take between each value.
	// Defaults to 1.
	//
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=12
	Step int32 `json:"step,omitempty"`
}

func (r ScheduleMonthRange) GetStart() int32 { return r.Start }
func (r ScheduleMonthRange) GetEnd() int32   { return r.End }
func (r ScheduleMonthRange) GetStep() int32  { return r.Step }

// If end < start, then end is interpreted as
// equal to start. This means you can use a Range with start set to a value, and
// end and step unset to represent a single value.
type ScheduleYearRange struct {
	// Start of the range (inclusive).
	//
	// +optional
	// +kubebuilder:validation:Minimum=1970
	Start int32 `json:"start,omitempty"`

	// End of the range (inclusive).
	// Defaults to start.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1970
	End int32 `json:"end,omitempty"`

	// Step to be take between each value.
	// Defaults to 1.
	//
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	Step int32 `json:"step,omitempty"`
}

func (r ScheduleYearRange) GetStart() int32 { return r.Start }
func (r ScheduleYearRange) GetEnd() int32   { return r.End }
func (r ScheduleYearRange) GetStep() int32  { return r.Step }

// If end < start, then end is interpreted as
// equal to start. This means you can use a Range with start set to a value, and
// end and step unset to represent a single value.
type ScheduleDayOfWeekRange struct {
	// Start of the range (inclusive).
	// Defaults to 0.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=6
	Start int32 `json:"start,omitempty"`

	// End of the range (inclusive).
	// Defaults to start.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=6
	End int32 `json:"end,omitempty"`

	// Step to be take between each value.
	// Defaults to 1.
	//
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=6
	Step int32 `json:"step,omitempty"`
}

func (r ScheduleDayOfWeekRange) GetStart() int32 { return r.Start }
func (r ScheduleDayOfWeekRange) GetEnd() int32   { return r.End }
func (r ScheduleDayOfWeekRange) GetStep() int32  { return r.Step }

// ScheduleCalendarSpec is an event specification relative to the calendar, similar to a traditional cron specification.
// A timestamp matches if at least one range of each field matches the
// corresponding fields of the timestamp, except for year: if year is missing,
// that means all years match. For all fields besides year, at least one Range must be present to match anything.
type ScheduleCalendarSpec struct {
	// Second range to match (0-59).
	// Defaults to 0.
	//
	// +optional
	// +kubebuilder:default={{"start": 0}}
	Second []ScheduleSecondMinuteRange `json:"second,omitempty"`

	// Minute range to match (0-59).
	// Defaults to 0.
	//
	// +optional
	// +kubebuilder:default={{"start": 0}}
	Minute []ScheduleSecondMinuteRange `json:"minute,omitempty"`

	// Hour range to match (0-23).
	// Defaults to 0.
	//
	// +optional
	// +kubebuilder:default={{"start": 0}}
	Hour []ScheduleHourRange `json:"hour,omitempty"`

	// DayOfMonth range to match (1-31)
	// Defaults to match all days.
	//
	// +optional
	// +kubebuilder:default={{"end": 31}}
	DayOfMonth []ScheduleDayOfMonthRange `json:"dayOfMonth,omitempty"`

	// Month range to match (1-12).
	// Defaults to match all months.
	//
	// +optional
	// +kubebuilder:default={{"end": 12}}
	Month []ScheduleMonthRange `json:"month,omitempty"`

	// Year range to match.
	// Defaults to match all years.
	//
	// +optional
	Year []ScheduleYearRange `json:"year,omitempty"`

	// DayOfWeek range to match (0-6; 0 is Sunday)
	// Defaults to match all days of the week.
	//
	// +optional
	// +kubebuilder:default={{"end": 6}}
	DayOfWeek []ScheduleDayOfWeekRange `json:"dayOfWeek,omitempty"`

	// Comment describes the intention of this schedule.
	//
	// +optional
	Comment string `json:"comment,omitempty"`
}

// ScheduleSpec is a complete description of a set of absolute timestamps.
type ScheduleSpec struct {
	// Calendars represents calendar-based specifications of times.
	//
	// +optional
	Calendars []ScheduleCalendarSpec `json:"calendars,omitempty"`

	// Intervals represents interval-based specifications of times.
	//
	// +optional
	Intervals []ScheduleIntervalSpec `json:"intervals,omitempty"`

	// Crons are cron based specifications of times.
	// Crons is provided for easy migration from legacy Cron Workflows. For new
	// use cases, we recommend using ScheduleSpec.Calendars or ScheduleSpec.
	// Intervals for readability and maintainability. Once a schedule is created all
	// expressions in Crons will be translated to ScheduleSpec.Calendars on the server.
	//
	// For example, `0 12 * * MON-WED,FRI` is every M/Tu/W/F at noon
	//
	// The string can have 5, 6, or 7 fields, separated by spaces, and they are interpreted in the
	// same way as a ScheduleCalendarSpec:
	//
	//	- 5 fields:         Minute, Hour, DayOfMonth, Month, DayOfWeek
	//
	//	- 6 fields:         Minute, Hour, DayOfMonth, Month, DayOfWeek, Year
	//
	//	- 7 fields: Second, Minute, Hour, DayOfMonth, Month, DayOfWeek, Year
	//
	// Notes:
	//	- If Year is not given, it defaults to *.
	//	- If Second is not given, it defaults to 0.
	//	- Shorthands @yearly, @monthly, @weekly, @daily, and @hourly are also
	//		accepted instead of the 5-7 time fields.
	//	- @every <interval>[/<phase>] is accepted and gets compiled into an
	//		IntervalSpec instead. <interval> and <phase> should be a decimal integer
	//		with a unit suffix s, m, h, or d.
	//	- Optionally, the string can be preceded by CRON_TZ=<time zone name> or
	//		TZ=<time zone name>, which will get copied to ScheduleSpec.TimeZoneName. (In which case the ScheduleSpec.TimeZone field should be left empty.)
	//	- Optionally, "#" followed by a comment can appear at the end of the string.
	//	- Note that the special case that some cron implementations have for
	//		treating DayOfMonth and DayOfWeek as "or" instead of "and" when both
	//		are set is not implemented.
	//
	// +optional
	Crons []string `json:"crons,omitempty"`

	// ExcludeCalendars defines any matching times that will be skipped.
	//
	// All fields of the ScheduleCalendarSpec including seconds must match a time for the time to be skipped.
	//
	// +optional
	ExcludeCalendars []ScheduleCalendarSpec `json:"excludeCalendars,omitempty"`

	// StartAt represents the start of the schedule. Any times before `startAt` will be skipped.
	// Together, `startAt` and `endAt` make an inclusive interval.
	// Defaults to the beginning of time.
	// For example: 2024-05-13T00:00:00Z
	//
	// +optional
	StartAt *metav1.Time `json:"startAt,omitempty"`

	// EndAt represents the end of the schedule. Any times after `endAt` will be skipped.
	// Defaults to the end of time.
	// For example: 2024-05-13T00:00:00Z
	//
	// +optional
	EndAt *metav1.Time `json:"endAt,omitempty"`

	// Jitter represents a duration that is used to apply a jitter to scheduled times.
	// All times will be incremented by a random value from 0 to this amount of jitter, capped
	// by the time until the next schedule.
	// Defaults to 0.
	//
	// +optional
	Jitter *metav1.Duration `json:"jitter,omitempty"`

	// TimeZoneName represents the IANA time zone name, for example `US/Pacific`.
	//
	// The definition will be loaded by Temporal Server from the environment it runs in.
	//
	// Calendar spec matching is based on literal matching of the clock time
	// with no special handling of DST: if you write a calendar spec that fires
	// at 2:30am and specify a time zone that follows DST, that action will not
	// be triggered on the day that has no 2:30am. Similarly, an action that
	// fires at 1:30am will be triggered twice on the day that has two 1:30s.
	//
	// Note: No actions are taken on leap-seconds (e.g. 23:59:60 UTC).
	// Defaults to UTC.
	//
	// +optional
	TimeZoneName string `json:"timezoneName,omitempty"`
}

const (
	ScheduleOverlapPolicyUnspecified    = "unspecified"
	ScheduleOverlapPolicySkip           = "skip"
	ScheduleOverlapPolicyBufferOne      = "bufferOne"
	ScheduleOverlapPolicyBufferAll      = "bufferAll"
	ScheduleOverlapPolicyCancelOther    = "cancelOther"
	ScheduleOverlapPolicyTerminateOther = "terminateOther"
	ScheduleOverlapPolicyAllowAll       = "allowAll"
)

// Overlap controls what happens when an Action would be started by a
// Schedule at the same time that an older Action is still running.
//
// Supported values:
//
// "skip" - Default. Nothing happens; the Workflow Execution is not started.
//
// "bufferOne" - Starts the Workflow Execution as soon as the current one completes.
// The buffer is limited to one. If another Workflow Execution is supposed to start,
// but one is already in the buffer, only the one in the buffer eventually starts.
//
// "bufferAll" - Allows an unlimited number of Workflows to buffer. They are started sequentially.
//
// "cancelOther" - Cancels the running Workflow Execution, and then starts the new one
// after the old one completes cancellation.
//
// "terminateOther" - Terminates the running Workflow Execution and starts the new one immediately.
//
// "allowAll" - Starts any number of concurrent Workflow Executions.
// With this policy (and only this policy), more than one Workflow Execution,
// started by the Schedule, can run simultaneously.
//
// +kubebuilder:validation:Enum=skip;bufferOne;bufferAll;cancelOther;terminateOther;allowAll
type ScheduleOverlapPolicy string

// SchedulePolicies represent policies for overlaps, catchups, pause on failure, and workflow ID.
type SchedulePolicies struct {
	// +optional
	Overlap ScheduleOverlapPolicy `json:"overlap,omitempty"`

	// CatchupWindow The Temporal Server might be down or unavailable at the time
	// when a Schedule should take an Action. When the Server comes back up,
	// CatchupWindow controls which missed Actions should be taken at that point.
	//
	// +optional
	CatchupWindow *metav1.Duration `json:"catchupWindow,omitempty"`

	// PauseOnFailure if true, and a workflow run fails or times out, turn on "paused".
	// This applies after retry policies: the full chain of retries must fail to trigger a pause here.
	//
	// +optional
	PauseOnFailure bool `json:"pauseOnFailure,omitempty"`

	// NOTE: This does not appear to be supported at the moment
	// // KeepOriginalWorkflowId if true, and the action would start a workflow, a timestamp will not be appended
	// // to the scheduled workflow id.
	// //
	// // +optional
	// KeepOriginalWorkflowId bool `json:"keepOriginalWorkflowId,omitempty"`
}

// RetryPolicy defines how retries ought to be handled, usable by both workflows and activities.
type RetryPolicy struct {
	// Interval of the first retry. If retryBackoffCoefficient is 1.0 then it is used for all retries.
	//
	// +optional
	InitialInterval *metav1.Duration `json:"initialInterval,omitempty"`

	// Coefficient used to calculate the next retry interval.
	// The next retry interval is previous interval multiplied by the coefficient.
	// Must be 1 or larger.
	//
	// +optional
	BackoffCoefficient *resource.Quantity `json:"backoffCoefficient,omitempty"`

	// Maximum interval between retries. Exponential backoff leads to interval increase.
	// This value is the cap of the increase. Default is 100x of the initial interval.
	//
	// +optional
	MaximumInterval *metav1.Duration `json:"maximumInterval,omitempty"`

	// Maximum number of attempts. When exceeded the retries stop even if not expired yet.
	// 1 disables retries. 0 means unlimited (up to the timeouts).
	//
	// +optional
	MaximumAttempts int32 `json:"maximumAttempts,omitempty"`

	// Non-Retryable errors types. Will stop retrying if the error type matches this list. Note that
	// this is not a substring match, the error *type* (not message) must match exactly.
	//
	// +optional
	NonRetryableErrorTypes []string `json:"nonRetryableErrorTypes,omitempty"`
}

// ScheduleWorkflowAction describes a workflow to launch.
type ScheduleWorkflowAction struct {
	// WorkflowID represents the business identifier of the workflow execution.
	// The WorkflowID of the started workflow may not match this exactly,
	// it may have a timestamp appended for uniqueness.
	// Defaults to a uuid.
	//
	// +optional
	WorkflowID string `json:"id,omitempty"`

	// WorkflowType represents the identifier used by a workflow author to define the workflow
	// Workflow type name.
	//
	// +kubebuilder:validation:Required
	WorkflowType string `json:"type"`

	// TaskQueue represents a workflow task queue.
	// This is also the name of the activity task queue on which activities are scheduled.
	//
	// +kubebuilder:validation:Required
	TaskQueue string `json:"taskQueue"`

	// Inputs contains arguments to pass to the workflow.
	//
	// +optional
	Inputs *apiextensionsv1.JSON `json:"inputs,omitempty"`

	// WorkflowExecutionTimeout is the timeout for duration of workflow execution.
	//
	// +optional
	WorkflowExecutionTimeout *metav1.Duration `json:"executionTimeout,omitempty"`

	// WorkflowRunTimeout is the timeout for duration of a single workflow run.
	WorkflowRunTimeout *metav1.Duration `json:"runTimeout,omitempty"`

	// WorkflowTaskTimeout is The timeout for processing workflow task from the time the worker
	// pulled this task.
	//
	// +optional
	WorkflowTaskTimeout *metav1.Duration `json:"taskTimeout,omitempty"`

	// NOTE: This is not supported on scheduled workflows
	// WorkflowIdReusePolicy `json:"idReusePolicy,omitempty"`

	// RetryPolicy is the retry policy for the workflow. If a retry policy is specified,
	// in case of workflow failure server will start new workflow execution if
	// needed based on the retry policy.
	//
	// +optional
	RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`

	// Memo is optional non-indexed info that will be shown in list workflow.
	//
	// +optional
	// +kubebuilder:validation:Type=object
	Memo *apiextensionsv1.JSON `json:"memo,omitempty"`

	// SearchAttributes is optional indexed info that can be used in query of List/Scan/Count workflow APIs. The key
	// and value type must be registered on Temporal server side. For supported operations on different server versions
	// see [Visibility].
	//
	// [Visibility]: https://docs.temporal.io/visibility
	//
	// +optional
	// +kubebuilder:validation:Type=object
	SearchAttributes *apiextensionsv1.JSON `json:"searchAttributes,omitempty"`
}

func (action *ScheduleWorkflowAction) GetWorkflowID() string {
	if action.WorkflowID == "" {
		return uuid.NewString()
	}

	return action.WorkflowID
}

// ScheduleAction contains the actions that the schedule should perform.
type ScheduleAction struct {
	// +kubebuilder:validation:Required
	Workflow ScheduleWorkflowAction `json:"workflow"`
}

// ScheduleState describes the current state of a schedule.
type ScheduleState struct {
	// Note is an informative human-readable message with contextual notes, e.g. the reason
	// a Schedule is paused. The system may overwrite this message on certain
	// conditions, e.g. when pause-on-failure happens.
	//
	// +optional
	Note string `json:"notes,omitempty"`

	// Paused is true if the schedule is paused.
	//
	// +optional
	Paused bool `json:"paused,omitempty"`

	// LimitedActions limits actions. While true RemainingActions will be decremented for each action taken.
	// Skipped actions (due to overlap policy) do not count against remaining actions.
	//
	// +optional
	LimitedActions bool `json:"limitedActions,omitempty"`

	// RemainingActions represents the Actions remaining in this Schedule.
	// Once this number hits 0, no further Actions are taken.
	// manual actions through backfill or ScheduleHandle.Trigger still run.
	//
	// +optional
	RemainingActions int64 `json:"remainingActions,omitempty"`
}

// Schedule contains all fields related to a schedule.
type Schedule struct {
	Action ScheduleAction    `json:"action,omitempty"`
	Spec   ScheduleSpec      `json:"spec,omitempty"`
	Policy *SchedulePolicies `json:"policy,omitempty"`
	State  *ScheduleState    `json:"state,omitempty"`
}

// TemporalScheduleSpec defines the desired state of Schedule.
type TemporalScheduleSpec struct {
	// Reference to the temporal namespace the schedule will be created in.
	NamespaceRef ObjectReference `json:"namespaceRef"`

	Schedule Schedule `json:"schedule"`

	// Memo is optional non-indexed info that will be shown in list workflow.
	//
	// +optional
	// +kubebuilder:validation:Type=object
	Memo *apiextensionsv1.JSON `json:"memo,omitempty"`

	// SearchAttributes is optional indexed info that can be used in query of List/Scan/Count workflow APIs. The key
	// and value type must be registered on Temporal server side. For supported operations on different server versions
	// see [Visibility].
	//
	// [Visibility]: https://docs.temporal.io/visibility
	//
	// +optional
	// +kubebuilder:validation:Type=object
	SearchAttributes *apiextensionsv1.JSON `json:"searchAttributes,omitempty"`

	// AllowDeletion makes the controller delete the Temporal schedule if the
	// CRD is deleted.
	//
	// +optional
	AllowDeletion bool `json:"allowDeletion,omitempty"`
}

// TemporalScheduleStatus defines the observed state of Schedule.
type TemporalScheduleStatus struct {
	// Conditions represent the latest available observations of the Schedule state.
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].status"
// +kubebuilder:printcolumn:name="ReconcileSuccess",type="string",JSONPath=".status.conditions[?(@.type == 'ReconcileSuccess')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// A TemporalSchedule creates a schedule in the targeted temporal cluster.
type TemporalSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemporalScheduleSpec   `json:"spec,omitempty"`
	Status TemporalScheduleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TemporalSchedule contains a list of Schedule.
type TemporalScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TemporalSchedule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TemporalSchedule{}, &TemporalScheduleList{})
}
