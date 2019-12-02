package backend

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/takutakahashi/container-api-gateway/pkg/types"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubernetesBackend struct{}

func (b KubernetesBackend) Execute(e types.Endpoint) (*bytes.Buffer, *bytes.Buffer, error) {
	namespace := "default"
	config, err := rest.InClusterConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	jobsClient := clientset.BatchV1().Jobs(namespace)
	name := e.Path[1:] + "-" + uuid.New().String()
	sd := make(map[string]string)
	for _, key := range e.Env {
		sd[key] = os.Getenv(key)
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-env",
			Namespace: namespace,
		},
		StringData: sd,
		Type:       corev1.SecretTypeOpaque,
	}
	secretsClient := clientset.CoreV1().Secrets(secret.Namespace)
	if _, err = secretsClient.Create(secret); err != nil {
		if _, err := secretsClient.Get(secret.Name, metav1.GetOptions{}); err != nil {
			return nil, nil, err
		}
		if _, err = secretsClient.Update(secret); err != nil {
			return nil, nil, err
		}
	}
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
							EnvFrom: []corev1.EnvFromSource{
								corev1.EnvFromSource{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: secret.Name,
										},
									},
								},
							},
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
	if e.Async {
		go b.watchLog(job)
		return nil, nil, nil
	}
	return b.watchLog(job)
}

func (b KubernetesBackend) watchLog(job *batchv1.Job) (*bytes.Buffer, *bytes.Buffer, error) {
	config, err := rest.InClusterConfig()
	clientset, err := kubernetes.NewForConfig(config)
	jobsClient := clientset.BatchV1().Jobs(job.Namespace)
	secretsClient := clientset.CoreV1().Secrets(job.Namespace)
	for true {
		job, _ := jobsClient.Get(job.Name, metav1.GetOptions{})
		if job.Status.Succeeded > 0 {
			break
		}
		if job.Status.Failed > 0 {
			return nil, nil, errors.New("job failed")
		}
		time.Sleep(1 * time.Second)
	}
	podsClient := clientset.CoreV1().Pods(job.Namespace)
	req := podsClient.GetLogs(job.Name, &corev1.PodLogOptions{})
	podLogs, err := req.Stream()
	if err != nil {
		return nil, nil, err
	}
	defer podLogs.Close()
	secretsClient.Delete(job.Spec.Template.Spec.Containers[0].EnvFrom[0].SecretRef.Name, &metav1.DeleteOptions{})
	fmt.Println("job deleted")
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return nil, nil, err
	}
	return buf, nil, nil
}
