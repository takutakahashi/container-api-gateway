package backend

import (
	"bytes"
	"io"
	"time"

	"github.com/takutakahashi/container-api-gateway/pkg/types"
	"github.com/thoas/go-funk"
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
	jobsClient := clientset.BatchV1().Jobs("default")
	name := e.Path[1:] + "-" + funk.RandomString(10)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "job",
							Image:   e.Container.Image,
							Command: e.Container.Command,
						},
					},
				},
			},
		},
	}
	jobsClient.Create(job)
	for true {
		job, err := jobsClient.Get(name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}
		if job.Status.Succeeded > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	podsClient := clientset.CoreV1().Pods(job.Namespace)
	req := podsClient.GetLogs(name, &corev1.PodLogOptions{})
	podLogs, err := req.Stream()
	if err != nil {
		return nil, nil, err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return nil, nil, err
	}
	return buf, nil, nil
}
