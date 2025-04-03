package backend

import (
	"bytes"
	"context"
	"fmt"
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
	"k8s.io/utils/ptr"
)

type KubernetesBackend struct{}

func (b KubernetesBackend) Execute(e types.Endpoint) (*bytes.Buffer, *bytes.Buffer, error) {
	namespace := "default"
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}
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
	ctx := context.Background()
	if _, err = secretsClient.Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		if _, err := secretsClient.Get(ctx, secret.Name, metav1.GetOptions{}); err != nil {
			return nil, nil, err
		}
		if _, err = secretsClient.Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
			return nil, nil, err
		}
	}
	containers := funk.Map(e.Containers, func(c types.Container) corev1.Container {
		return corev1.Container{
			Name:    c.Name,
			Image:   c.Image,
			Command: e.BuildCommand(c),
			EnvFrom: []corev1.EnvFromSource{
				{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: secret.Name,
						},
					},
				},
			},
		}
	}).([]corev1.Container)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: ptr.To(int32(60)),
			BackoffLimit:            ptr.To(int32(0)),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers:    containers,
				},
			},
		},
	}
	_, err = jobsClient.Create(ctx, job, metav1.CreateOptions{})
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
	if err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	jobsClient := clientset.BatchV1().Jobs(job.Namespace)
	secretsClient := clientset.CoreV1().Secrets(job.Namespace)
	ctx := context.Background()
	for {
		j, err := jobsClient.Get(ctx, job.Name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}
		if j.Status.Succeeded > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("job succeeded")
	podsClient := clientset.CoreV1().Pods(job.Namespace)
	pods, err := podsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	var pod corev1.Pod
	for _, p := range pods.Items {
		fmt.Println(p.Labels)
		if val, ok := p.Labels["job-name"]; ok && val == job.Name {
			pod = p
			break
		}
	}
	req := podsClient.GetLogs(pod.Name, &corev1.PodLogOptions{Container: "main"})

	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer podLogs.Close()
	secretsClient.Delete(ctx, job.Spec.Template.Spec.Containers[0].EnvFrom[0].SecretRef.Name, metav1.DeleteOptions{})
	fmt.Println("job deleted")
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return nil, nil, err
	}
	return buf, nil, nil
}
