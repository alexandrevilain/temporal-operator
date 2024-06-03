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

package temporal

import (
	"encoding/json"
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/enumerable"
	"github.com/google/uuid"
	commonv1 "go.temporal.io/api/common/v1"
	enumsv1 "go.temporal.io/api/enums/v1"
	schedulev1 "go.temporal.io/api/schedule/v1"
	"go.temporal.io/api/taskqueue/v1"
	workflowv1 "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/server/common/payloads"
	"go.temporal.io/server/common/primitives/timestamp"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	ScheduleClientIdentity = "temporal-operator.temporal.io"
)

func unmarshalArrayFromJson(j *apiextensionsv1.JSON) (*[]interface{}, error) {
	if j == nil {
		return nil, nil
	}

	var object []interface{}
	if err := json.Unmarshal(j.Raw, &object); err != nil {
		return nil, fmt.Errorf("invalid type, expected array")
	}

	return &object, nil
}

func unmarshalMapFromJson(j *apiextensionsv1.JSON) (*map[string]interface{}, error) {
	if j == nil {
		return nil, nil
	}

	var object map[string]interface{}
	if err := json.Unmarshal(j.Raw, &object); err != nil {
		return nil, fmt.Errorf("invalid type, expected map")
	}

	return &object, nil
}

func encodePayloadArray(a *[]interface{}) (*commonv1.Payloads, error) {
	if a == nil || len(*a) == 0 {
		return nil, nil
	}

	re, err := payloads.Encode(*a...)
	if err != nil {
		return nil, fmt.Errorf("unable to encode input: %w", err)
	}

	return re, nil
}

func encodePayloadMap(p map[string]interface{}) (map[string]*commonv1.Payload, error) {
	if len(p) == 0 {
		return nil, nil
	}

	dc := converter.GetDefaultDataConverter()
	re := make(map[string]*commonv1.Payload, len(p))
	var err error
	for k, v := range p {
		re[k], err = dc.ToPayload(v)
		if err != nil {
			return nil, err
		}
	}

	return re, nil
}

func encodeMemo(memo map[string]interface{}) (*commonv1.Memo, error) {
	if len(memo) == 0 {
		return nil, nil
	}

	fields, err := encodePayloadMap(memo)
	if err != nil {
		return nil, err
	}

	return &commonv1.Memo{Fields: fields}, nil
}

func encodeSearchAttributes(sa map[string]interface{}) (*commonv1.SearchAttributes, error) {
	if len(sa) == 0 {
		return nil, nil
	}

	fields, err := encodePayloadMap(sa)
	if err != nil {
		return nil, err
	}

	return &commonv1.SearchAttributes{IndexedFields: fields}, nil
}

func buildRetryPolicy(policy *v1beta1.RetryPolicy) *commonv1.RetryPolicy {
	if policy == nil {
		return nil
	}

	re := commonv1.RetryPolicy{
		MaximumAttempts:        policy.MaximumAttempts,
		NonRetryableErrorTypes: policy.NonRetryableErrorTypes,
	}

	if policy.InitialInterval != nil {
		re.InitialInterval = timestamp.DurationPtr(policy.InitialInterval.Duration)
	}
	if policy.MaximumInterval != nil {
		re.MaximumInterval = timestamp.DurationPtr(policy.MaximumInterval.Duration)
	}
	if policy.BackoffCoefficient != nil {
		re.BackoffCoefficient = policy.BackoffCoefficient.AsApproximateFloat64()
	}

	return &re
}

func buildPayloads(j *apiextensionsv1.JSON) (*commonv1.Payloads, error) {
	if j == nil {
		return nil, nil
	}

	inputArray, err := unmarshalArrayFromJson(j)
	if err != nil {
		return nil, nil
	}
	inputs, err := encodePayloadArray(inputArray)
	if err != nil {
		return nil, err
	}

	return inputs, nil
}

func buildMemo(j *apiextensionsv1.JSON) (*commonv1.Memo, error) {
	if j == nil {
		return nil, nil
	}

	json, err := unmarshalMapFromJson(j)
	if err != nil {
		return nil, err
	}
	memo, err := encodeMemo(*json)
	if err != nil {
		return nil, err
	}

	return memo, nil
}

func buildSearchAttributes(j *apiextensionsv1.JSON) (*commonv1.SearchAttributes, error) {
	if j == nil {
		return nil, nil
	}

	json, err := unmarshalMapFromJson(j)
	if err != nil {
		return nil, err
	}
	searchAttributes, err := encodeSearchAttributes(*json)
	if err != nil {
		return nil, err
	}

	return searchAttributes, nil
}

func buildAction(action v1beta1.ScheduleAction) (*schedulev1.ScheduleAction, error) {
	workflow := action.Workflow

	inputs, err := buildPayloads(workflow.Inputs)
	if err != nil {
		return nil, err
	}
	memo, err := buildMemo(workflow.Memo)
	if err != nil {
		return nil, err
	}
	searchAttributes, err := buildSearchAttributes(workflow.SearchAttributes)
	if err != nil {
		return nil, err
	}

	startWorkflow := workflowv1.NewWorkflowExecutionInfo{
		WorkflowId: workflow.GetWorkflowId(),
		WorkflowType: &commonv1.WorkflowType{
			Name: workflow.WorkflowType,
		},
		TaskQueue: &taskqueue.TaskQueue{
			Name: workflow.TaskQueue,
		},
		Input: inputs,
		// NOTE: This is not supported on scheduled workflows
		// WorkflowIdReusePolicy: buildWorkflowIdReusePolicy(workflow.WorkflowIdReusePolicy),
		RetryPolicy:      buildRetryPolicy(workflow.RetryPolicy),
		Memo:             memo,
		SearchAttributes: searchAttributes,
	}

	if workflow.WorkflowExecutionTimeout != nil {
		startWorkflow.WorkflowExecutionTimeout = timestamp.DurationPtr(workflow.WorkflowExecutionTimeout.Duration)
	}
	if workflow.WorkflowRunTimeout != nil {
		startWorkflow.WorkflowRunTimeout = timestamp.DurationPtr(workflow.WorkflowRunTimeout.Duration)
	}
	if workflow.WorkflowTaskTimeout != nil {
		startWorkflow.WorkflowTaskTimeout = timestamp.DurationPtr(workflow.WorkflowTaskTimeout.Duration)
	}

	return &schedulev1.ScheduleAction{
		Action: &schedulev1.ScheduleAction_StartWorkflow{
			StartWorkflow: &startWorkflow,
		},
	}, nil
}

type ScheduleRange interface {
	GetStart() int32
	GetEnd() int32
	GetStep() int32
}

func buildRange[T ScheduleRange](r T) *schedulev1.Range {
	var ri ScheduleRange = r
	return &schedulev1.Range{
		Start: ri.GetStart(),
		End:   ri.GetEnd(),
		Step:  ri.GetStep(),
	}
}

func buildCalendar(calendars []v1beta1.ScheduleCalendarSpec) []*schedulev1.StructuredCalendarSpec {
	re := enumerable.Select(calendars, func(c v1beta1.ScheduleCalendarSpec) *schedulev1.StructuredCalendarSpec {
		return &schedulev1.StructuredCalendarSpec{
			Second:     enumerable.Select(c.Second, buildRange),
			Minute:     enumerable.Select(c.Minute, buildRange),
			Hour:       enumerable.Select(c.Hour, buildRange),
			DayOfMonth: enumerable.Select(c.DayOfMonth, buildRange),
			Month:      enumerable.Select(c.Month, buildRange),
			Year:       enumerable.Select(c.Year, buildRange),
			DayOfWeek:  enumerable.Select(c.DayOfWeek, buildRange),
			Comment:    c.Comment,
		}
	})

	return re
}

func buildIntervals(intervals []v1beta1.ScheduleIntervalSpec) []*schedulev1.IntervalSpec {
	return enumerable.Select(intervals, func(i v1beta1.ScheduleIntervalSpec) *schedulev1.IntervalSpec {
		re := schedulev1.IntervalSpec{
			Interval: timestamp.DurationPtr(i.Every.Duration),
		}

		if i.Offset != nil {
			re.Phase = timestamp.DurationPtr(i.Offset.Duration)
		}

		return &re
	})
}

func buildSpec(spec v1beta1.ScheduleSpec) (*schedulev1.ScheduleSpec, error) {
	re := schedulev1.ScheduleSpec{
		StructuredCalendar:        buildCalendar(spec.Calendars),
		CronString:                spec.Crons,
		Interval:                  buildIntervals(spec.Intervals),
		ExcludeStructuredCalendar: buildCalendar(spec.ExcludeCalendars),
		TimezoneName:              spec.TimeZoneName,
	}

	if spec.StartAt != nil {
		re.StartTime = timestamp.TimePtr(spec.StartAt.Time)
	}
	if spec.EndAt != nil {
		re.EndTime = timestamp.TimePtr(spec.EndAt.Time)
	}
	if spec.Jitter != nil {
		re.Jitter = timestamp.DurationPtr(spec.Jitter.Duration)
	}

	return &re, nil
}

func buildOverlapPolicy(p v1beta1.ScheduleOverlapPolicy) enumsv1.ScheduleOverlapPolicy {
	switch p {
	case v1beta1.ScheduleOverlapPolicySkip:
		return enumsv1.SCHEDULE_OVERLAP_POLICY_SKIP
	case v1beta1.ScheduleOverlapPolicyBufferOne:
		return enumsv1.SCHEDULE_OVERLAP_POLICY_BUFFER_ONE
	case v1beta1.ScheduleOverlapPolicyBufferAll:
		return enumsv1.SCHEDULE_OVERLAP_POLICY_BUFFER_ALL
	case v1beta1.ScheduleOverlapPolicyCancelOther:
		return enumsv1.SCHEDULE_OVERLAP_POLICY_CANCEL_OTHER
	case v1beta1.ScheduleOverlapPolicyTerminateOther:
		return enumsv1.SCHEDULE_OVERLAP_POLICY_TERMINATE_OTHER
	case v1beta1.ScheduleOverlapPolicyAllowAll:
		return enumsv1.SCHEDULE_OVERLAP_POLICY_ALLOW_ALL
	default:
		return enumsv1.SCHEDULE_OVERLAP_POLICY_UNSPECIFIED
	}
}

func buildPolicies(policies *v1beta1.SchedulePolicies) (*schedulev1.SchedulePolicies, error) {
	if policies == nil {
		return &schedulev1.SchedulePolicies{}, nil
	}

	re := schedulev1.SchedulePolicies{
		OverlapPolicy:  buildOverlapPolicy(policies.Overlap),
		PauseOnFailure: policies.PauseOnFailure,
		// NOTE: This does not appear to be supported at the moment
		// KeepOriginalWorkflowId: policies.KeepOriginalWorkflowId,
	}

	if policies.CatchupWindow != nil {
		re.CatchupWindow = timestamp.DurationPtr(policies.CatchupWindow.Duration)
	}

	return &re, nil
}

func buildState(state *v1beta1.ScheduleState) (*schedulev1.ScheduleState, error) {
	if state == nil {
		return &schedulev1.ScheduleState{}, nil
	}

	re := &schedulev1.ScheduleState{
		Notes:            state.Note,
		Paused:           state.Paused,
		LimitedActions:   state.LimitedActions,
		RemainingActions: int64(state.RemainingActions),
	}

	return re, nil
}

func buildSchedule(schedule *v1beta1.TemporalSchedule) (*schedulev1.Schedule, error) {
	if schedule == nil {
		return nil, fmt.Errorf("schedule is undefined")
	}

	action, err := buildAction(schedule.Spec.Schedule.Action)
	if err != nil {
		return nil, err
	}
	spec, err := buildSpec(schedule.Spec.Schedule.Spec)
	if err != nil {
		return nil, err
	}
	policies, err := buildPolicies(schedule.Spec.Schedule.Policy)
	if err != nil {
		return nil, err
	}
	state, err := buildState(schedule.Spec.Schedule.State)
	if err != nil {
		return nil, err
	}

	re := &schedulev1.Schedule{
		Action:   action,
		Spec:     spec,
		Policies: policies,
		State:    state,
	}

	return re, nil
}

func ScheduleToCreateScheduleRequest(cluster *v1beta1.TemporalCluster, schedule *v1beta1.TemporalSchedule) (*workflowservice.CreateScheduleRequest, error) {
	sch, err := buildSchedule(schedule)
	if err != nil {
		return nil, err
	}

	memo, err := buildMemo(schedule.Spec.Memo)
	if err != nil {
		return nil, err
	}
	searchAttributes, err := buildSearchAttributes(schedule.Spec.SearchAttributes)
	if err != nil {
		return nil, err
	}

	// TODO: Consider if InitialPatch should be supported
	//	InitialPatch: &schedulev1.SchedulePatch{
	//		// TriggerImmediately: ,
	//		// BackfillRequest: ,
	//		// Pause: ,
	//		// Unpause: ,
	//	},

	re := &workflowservice.CreateScheduleRequest{
		Namespace:        schedule.Spec.NamespaceRef.Name,
		ScheduleId:       schedule.GetName(),
		Schedule:         sch,
		Identity:         ScheduleClientIdentity,
		RequestId:        uuid.NewString(),
		Memo:             memo,
		SearchAttributes: searchAttributes,
	}

	return re, nil
}

func ScheduleToUpdateScheduleRequest(schedule *v1beta1.TemporalSchedule) (*workflowservice.UpdateScheduleRequest, error) {
	sch, err := buildSchedule(schedule)
	if err != nil {
		return nil, err
	}

	// TODO: Memo and search attributes cannot be changed. Should this be handled by a webhook?
	re := &workflowservice.UpdateScheduleRequest{
		Namespace:  schedule.Spec.NamespaceRef.Name,
		ScheduleId: schedule.GetName(),
		Schedule:   sch,
		Identity:   ScheduleClientIdentity,
		RequestId:  uuid.NewString(),
	}

	return re, nil
}

func ScheduleToDeleteScheduleRequest(schedule *v1beta1.TemporalSchedule) *workflowservice.DeleteScheduleRequest {
	return &workflowservice.DeleteScheduleRequest{
		Namespace:  schedule.Spec.NamespaceRef.Name,
		ScheduleId: schedule.GetName(),
		Identity:   ScheduleClientIdentity,
	}
}
