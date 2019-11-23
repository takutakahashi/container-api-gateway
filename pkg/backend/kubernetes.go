package backend

import (
	"bytes"

	"github.com/takutakahashi/container-api-gateway/pkg/types"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubernetesBackend struct{}

func (b KubernetesBackend) Execute(e types.Endpoint) (*bytes.Buffer, *bytes.Buffer, error) {
	config, err := rest.InClusterConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	jobsClient := clientset.BatchV1().Jobs("gitlab")
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-job",
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "demo",
							Image: "myimage",
						},
					},
				},
			},
		},
	}
	jobsClient.Create(job)
	return nil, nil, nil
}
