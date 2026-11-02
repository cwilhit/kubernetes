/*
Copyright 2026 The kcp Authors.

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

package resourcequota

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	"testing"
)

func TestReplenishmentPreservesLogicalCluster(t *testing.T) {
	for _, tc := range []struct{ cluster, namespace, want string }{
		{"clusterone", "", "clusterone|"},
		{"clustertwo", "", "clustertwo|"},
		{"clusterone", "clusterone|admin", "clusterone|admin"},
		{"clusterone", "tenant", "clusterone|tenant"},
		{"", "tenant", "tenant"},
		{"", "", ""},
	} {
		t.Run(tc.cluster+"-"+tc.namespace, func(t *testing.T) {
			queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[*event]())
			defer queue.ShutDown()
			got := "unset"
			qm := &QuotaMonitor{resourceChanges: queue, replenishmentFunc: func(_ context.Context, _ schema.GroupResource, namespace string) { got = namespace }}
			obj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "deleted-workspace", Namespace: tc.namespace, Annotations: map[string]string{"kcp.io/cluster": tc.cluster}}}
			queue.Add(&event{eventType: deleteEvent, obj: obj, gvr: schema.GroupVersionResource{Group: "tenancy.kcp.io", Version: "v1alpha1", Resource: "workspaces"}})
			if !qm.processResourceChanges(context.Background()) {
				t.Fatal("monitor stopped")
			}
			if got != tc.want {
				t.Fatalf("replenished namespace %q, want %q", got, tc.want)
			}
		})
	}
}
