package spot

import (
	"context"
	"math"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// Name is the plugin name
	Name = "spot"

	capacityKey = "node.kubernetes.io/capacity"
	ondemand    = "on-demand"
	spot        = "spot"

	scoreWeight = 100
)

type SpotPlugin struct {
	frameworkHandler framework.Handle
	client           kubernetes.Interface
	ctx              context.Context
}

var _ framework.ScorePlugin = &SpotPlugin{}

// New initializes and returns a new PlacementPolicy plugin.
func New(ctx context.Context, obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	client := kubernetes.NewForConfigOrDie(handle.KubeConfig())

	plugin := &SpotPlugin{
		frameworkHandler: handle,
		client:           client,
		ctx:              ctx,
	}

	return plugin, nil
}

func (s *SpotPlugin) Name() string {
	return Name
}

func (s *SpotPlugin) Score(ctx context.Context, state *framework.CycleState, p *corev1.Pod, nodeName string) (int64, *framework.Status) {
	podList, err := s.GetPodsWithLabels(ctx, p.Labels)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, "failed to get pods with labels: "+err.Error())
	}

	hasNodeSelectPodNum := 0
	currentNodeSelctPodNum := 0
	hasOnDemandPodNum := 0
	for pi := range podList {
		if len(podList[pi].Spec.NodeName) != 0 {
			hasNodeSelectPodNum++
			if podList[pi].Spec.NodeName == nodeName {
				currentNodeSelctPodNum++
			}

			if podList[pi].Annotations[capacityKey] == ondemand {
				hasOnDemandPodNum++
			}
		}
	}

	nodeCap, err := s.GetNodeCapacity(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, "failed to get node capacity: "+err.Error())
	}

	switch hasNodeSelectPodNum {
	case 0:
		if nodeCap == ondemand {
			return scoreWeight, nil
		}
	default:
		if hasOnDemandPodNum == 0 {
			if nodeCap == ondemand {
				return scoreWeight, nil
			}
			return 0, nil
		}

		if nodeCap == spot {
			return int64(math.Ceil(scoreWeight/float64(currentNodeSelctPodNum) + 1)), nil
		}
	}

	return 0, nil
}

func (s *SpotPlugin) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// GetPodsWithLabels returns the pods with labels
func (s *SpotPlugin) GetPodsWithLabels(ctx context.Context, podLabels map[string]string) ([]*corev1.Pod, error) {
	return s.frameworkHandler.SharedInformerFactory().Core().V1().Pods().Lister().List(labels.Set(podLabels).AsSelector())
}

// GetNodeCapacity returns the node capacity
func (s *SpotPlugin) GetNodeCapacity(nodeName string) (string, error) {
	node, err := s.frameworkHandler.SharedInformerFactory().Core().V1().Nodes().Lister().Get(nodeName)
	if err != nil {
		return "", err
	}

	return node.Labels[capacityKey], nil
}

// AnnotatePodNodeCapacity annotates the node capacity of the pod
func (s *SpotPlugin) AnnotatePodNodeCapacity(ctx context.Context, pod *corev1.Pod, capacity string) (*corev1.Pod, error) {
	annotations := map[string]string{}
	if pod.Annotations != nil {
		annotations = pod.Annotations
	}

	annotations[capacityKey] = capacity
	pod.Annotations = annotations
	return s.client.CoreV1().Pods(pod.Namespace).Update(ctx, pod, metav1.UpdateOptions{})
}
