package migscheduling

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type MigScheduling struct {
	handle    framework.Handle
	clientSet *kubernetes.Clientset
	dynclient dynamic.Interface
}

var _ framework.FilterPlugin = &MigScheduling{}

const (
	Name           = "MigScheduling"
	RootKubeConfig = "/root/.kube/config"
)

func (pl *MigScheduling) Name() string {
	return Name
}

func (pl *MigScheduling) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	if node, ok := pod.Annotations["origin.node.name"]; ok {
		klog.Infof("Pod %s is migrated from node %s", pod.Name, node)
		if node == nodeInfo.Node().Name {
			klog.Infof("Pod %s is migrated to node %s", pod.Name, nodeInfo.Node().Name)
			return framework.NewStatus(framework.Unschedulable, "Pod is migrated to this node")
		}
	}
	return framework.NewStatus(framework.Success, "Node: "+nodeInfo.Node().Name)
}

// func (pl *MigScheduling) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
// 	klog.Infof("Start filtering node: %s, pod: %s", nodeInfo.Node().Name, pod.Name)
// 	migrators, err := listMigrators(ctx, pl.dynclient, "")
// 	if err != nil {
// 		klog.ErrorS(err, "Failed to list migrators")
// 		return framework.NewStatus(framework.Error, "")
// 	}
// 	if len(migrators.Items) == 0 {
// 		klog.Infof("No migrators found")
// 		return framework.NewStatus(framework.Success, "No migrators found")
// 	} else if len(migrators.Items) > 1 {
// 		klog.Infof("Too many migrators, dont know which nodes to be scheduled")
// 		return framework.NewStatus(framework.Error, "Too many migrators, dont know which nodes to be scheduled")
// 	}
// 	migrator := migrators.Items[0]
// 	// get pod
// 	podName := migrator.Spec.PodName
// 	podNamespace := migrator.Spec.PodNameSpace
// 	klog.Infof("Migrator %s PodName: %s, PodNamespace: %s", migrator.Name, podName, podNamespace)
// 	deployName := strings.Split(podName, "-")[0]
// 	newPodDeployName := strings.Split(pod.Name, "-")[0]
// 	if deployName == newPodDeployName && migrator.Spec.Destination != nodeInfo.Node().Name {
// 		klog.Infof("Migrate Destination: %s, Current Node: %s, return error", migrator.Spec.Destination, nodeInfo.Node().Name)
// 		return framework.NewStatus(framework.Unschedulable, "Failed to schedule pod to not destination node: "+nodeInfo.Node().Name)
// 	}

// 	klog.Infof("Migrate Destination: %s, Current Node: %s, return success", migrator.Spec.Destination, nodeInfo.Node().Name)
// 	return framework.NewStatus(framework.Success, "Node: "+nodeInfo.Node().Name)
// }

func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	klog.Infof("Initializing MigScheduling plugin")
	useInClusterConfig := true
	var k8sConfig *rest.Config

	if useInClusterConfig {
		klog.Infof("Using in cluster config")
		config, err := rest.InClusterConfig()
		if err != nil {
			klog.ErrorS(err, "Failed to get in cluster config")
			return nil, fmt.Errorf("failed to get in cluster config")
		}
		k8sConfig = config
	} else {
		klog.Infof("Using local dev mode")
		config, err := clientcmd.BuildConfigFromFlags("", RootKubeConfig)
		if err != nil {
			klog.ErrorS(err, "Failed to get kubeconfig")
			return nil, fmt.Errorf("failed to get root kubeconfig")
		}
		k8sConfig = config
	}

	dynclient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		klog.ErrorS(err, "Failed to create dynamic client")
	}
	clientSet, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		klog.ErrorS(err, "Failed to create clientset")
	}
	klog.Infof("Initializing MigScheduling plugin successful")
	return &MigScheduling{
		handle:    handle,
		clientSet: clientSet,
		dynclient: dynclient,
	}, nil
}

// func getMigrator(ctx context.Context, client dynamic.Interface, name, namespace string) (*Migrator, error) {
// 	migrator, err := client.Resource(getGVR("migrate.openeuler.org", "v1alpha1", "migrators")).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
// 	if err != nil {
// 		klog.ErrorS(err, "Failed to get migrator")
// 		return nil, err
// 	}
// 	var mig Migrator
// 	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(migrator.UnstructuredContent(), &mig); err != nil {
// 		klog.ErrorS(err, "Failed to convert unstructured to struct")
// 		return nil, err
// 	}

// 	return &mig, nil
// }

// func listMigrators(ctx context.Context, client dynamic.Interface, namespace string) (*MigratorList, error) {
// 	if namespace == "" {
// 		namespace = "default"
// 	}
// 	migrators, err := client.Resource(getGVR("migrate.openeuler.org", "v1alpha1", "migrators")).Namespace(namespace).List(ctx, metav1.ListOptions{})
// 	if err != nil {
// 		klog.ErrorS(err, "Failed to list migrators")
// 		return nil, err
// 	}
// 	data, err := migrators.MarshalJSON()
// 	if err != nil {
// 		klog.ErrorS(err, "Failed to marshal migrators")
// 		return nil, err
// 	}
// 	var migList MigratorList
// 	if err := json.Unmarshal(data, &migList); err != nil {
// 		klog.ErrorS(err, "Failed to unmarshal migrators")
// 		return nil, err
// 	}
// 	return &migList, nil
// }

// // getGVR :- gets GroupVersionResource for dynamic client
// func getGVR(group, version, resource string) schema.GroupVersionResource {
// 	return schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
// }
