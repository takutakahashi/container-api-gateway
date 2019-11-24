package backend

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
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
	name := e.Path[1:] + "-" + uuid.New().String()
	env := funk.Map(e.Env, func(key string) corev1.EnvVar {
		return corev1.EnvVar{Name: key, Value: os.Getenv(key)}
	}).([]corev1.EnvVar)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "job",
							Image:   e.Container.Image,
							Command: e.BuildCommand(),
							Env:     env,
						},
					},
				},
			},
		},
	}
	_, err = jobsClient.Create(job)
	if err != nil {
		return nil, nil, err
	}
	for true {
		job, _ := jobsClient.Get(name, metav1.GetOptions{})
		if job.Status.Succeeded > 0 {
			break
		}
		if job.Status.Failed > 0 {
			return nil, nil, errors.New("job failed")
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
