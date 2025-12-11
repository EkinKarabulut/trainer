/*
Copyright 2025 The Kubeflow Authors.

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

package kaischeduler

import (
	"context"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/client"

	trainer "github.com/kubeflow/trainer/v2/pkg/apis/trainer/v1alpha1"
	"github.com/kubeflow/trainer/v2/pkg/runtime"
	"github.com/kubeflow/trainer/v2/pkg/runtime/framework"
)

const (
	// MinAvailableAnnotationKey is the annotation key used to pass the pre-calculated
	// MinAvailable value to KAI-Scheduler's pod-grouper plugin.
	MinAvailableAnnotationKey = "kai.scheduler/min-available"

	// QueueLabelKey is the label key used to specify the scheduling queue for pods.
	QueueLabelKey = "kai.scheduler/queue"
)

type KAIScheduler struct{}

var _ framework.EnforcePodGroupPolicyPlugin = (*KAIScheduler)(nil)

const Name = "KAIScheduler"

func New(_ context.Context, _ client.Client, _ client.FieldIndexer) (framework.Plugin, error) {
	return &KAIScheduler{}, nil
}

func (k *KAIScheduler) Name() string {
	return Name
}

func (k *KAIScheduler) EnforcePodGroupPolicy(info *runtime.Info, trainJob *trainer.TrainJob) error {
	if info == nil || info.RuntimePolicy.PodGroupPolicy == nil || trainJob == nil || info.RuntimePolicy.PodGroupPolicy.KAIScheduler == nil {
		return nil
	}

	if info.Scheduler.PodAnnotations == nil {
		info.Scheduler.PodAnnotations = map[string]string{}
	}
	if info.Scheduler.PodLabels == nil {
		info.Scheduler.PodLabels = map[string]string{}
	}

	var totalMembers int32
	for _, ps := range info.TemplateSpec.PodSets {
		if ps.Count != nil {
			totalMembers += *ps.Count
		}
	}

	info.Scheduler.PodAnnotations[MinAvailableAnnotationKey] = strconv.FormatInt(int64(totalMembers), 10)

	// Set the queue label from the KAIScheduler policy.
	// KAI's pod-grouper will use this label to determine the scheduling queue.
	if queue := info.RuntimePolicy.PodGroupPolicy.KAIScheduler.Queue; queue != nil && *queue != "" {
		info.Scheduler.PodLabels[QueueLabelKey] = *queue
	}

	return nil
}