package k6

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/ymktmk/apply-k6-crd/api/v1alpha1"
	"gopkg.in/yaml.v2"
	// corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func NewK6(
	template,
	vus,
	duration,
	rps,
	parallelism string) (*K6, error) {

	client, err := NewClientSet()
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadFile(template)
	if err != nil {
		return nil, err
	}

	var currentK6 v1alpha1.K6
	err = yaml.Unmarshal(bytes, &currentK6)
	if err != nil {
		return nil, err
	}

	// Only create after this logic
	err = Validate(currentK6)
	if err != nil {
		return nil, err
	}

	numberOfJobs := currentK6.Spec.Parallelism
	if parallelism != "" {
		num, err := strconv.ParseInt(parallelism, 10, 32)
		if err != nil {
			return nil, err
		}
		numberOfJobs = num
	}

	k6Res := schema.GroupVersionResource{Group: "k6.io", Version: "v1alpha1", Resource: "k6s"}

	k6 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "k6.io/v1alpha1",
			"kind":       "K6",
			"metadata": map[string]interface{}{
				"name":      currentK6.ObjectMeta.Name,
				"namespace": currentK6.ObjectMeta.Namespace,
			},
			"spec": map[string]interface{}{
				"parallelism": numberOfJobs,
				"script": map[string]interface{}{
					"configMap": map[string]interface{}{
						"name": currentK6.Spec.Script.ConfigMap.Name,
						"file": currentK6.Spec.Script.ConfigMap.File,
					},
				},
				"runner": map[string]interface{}{},
			},
		},
	}

	// get spec
	spec, _, err := unstructured.NestedMap(k6.Object, "spec")
	if err != nil {
		return nil, err
	}

	// get runner
	runner, _, err := unstructured.NestedMap(k6.Object, "spec", "runner")
	if err != nil {
		return nil, err
	}

	// envがあればmanifestから取得して詰める
	if len(currentK6.Spec.Runner.Env) != 0 {
		envList := GetEnvList(currentK6.Spec.Runner.Env)
		err = unstructured.SetNestedSlice(runner, envList, "env")
		if err != nil {
			return nil, err
		}
	}

	// arguments into spec
	if len(currentK6.Spec.Arguments) != 0 {
		args := OverrideArgs(currentK6.Spec.Arguments, vus, duration, rps)
		log.Println(args)
		err = unstructured.SetNestedField(spec, args, "arguments")
		if err != nil {
			return nil, err
		}
	}

	// runner into spec
	err = unstructured.SetNestedMap(spec, runner, "runner")
	if err != nil {
		return nil, err
	}

	// spec into k6.Object
	err = unstructured.SetNestedMap(k6.Object, spec, "spec")
	if err != nil {
		return nil, err
	}

	log.Println(k6)

	return &K6{
		client:         client,
		Resource:       k6Res,
		UnstructuredK6: k6,
		CurrentK6:      currentK6,
	}, nil

}

type K6 struct {
	client         dynamic.Interface
	Resource       schema.GroupVersionResource
	UnstructuredK6 *unstructured.Unstructured
	CurrentK6      v1alpha1.K6
}

func (k *K6) CreateK6() error {
	namespace := k.CurrentK6.ObjectMeta.Namespace
	k6 := k.UnstructuredK6

	result, err := k.client.Resource(k.Resource).Namespace(namespace).Create(context.TODO(), k6, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	if result != nil {
		log.Printf("k6.k6.io/%q created\n", result.GetName())
	}

	return nil
}

func (k *K6) DeleteK6() error {
	name := k.CurrentK6.ObjectMeta.Name
	namespace := k.CurrentK6.ObjectMeta.Namespace

	err := k.client.Resource(k.Resource).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	log.Printf("k6.k6.io/%q deleted\n", name)

	return nil
}

func Validate(k6 v1alpha1.K6) error {

	if len(k6.ObjectMeta.Name) == 0 {
		return fmt.Errorf("metadata.name is not found")
	}

	if len(k6.ObjectMeta.Namespace) == 0 {
		return fmt.Errorf("metadata.namespace is not found")
	}

	if k6.Spec.Parallelism == 0 {
		return fmt.Errorf("set spec.parallelism to 1 or more")
	}

	if len(k6.Spec.Script.ConfigMap.Name) == 0 {
		return fmt.Errorf("spec.script.configmap.name is not found")
	}

	if len(k6.Spec.Script.ConfigMap.File) == 0 {
		return fmt.Errorf("spec.script.configmap.file is not found")
	}

	// env validate
	if len(k6.Spec.Runner.Env) != 0 {
		for _, v := range k6.Spec.Runner.Env {
			if v.Name == "" {
				return fmt.Errorf("spec.runner.env.name is not found")
			}
			if v.Value == "" && v.ValueFrom.SecretKeyRef.Name == "" && v.ValueFrom.SecretKeyRef.Key == "" {
				return fmt.Errorf("spec.runner.env.value is valueFrom not found")
			}
		}
	}

	return nil
}

func OverrideArgs(args, vus, duration, rps string) string {
	array := strings.Split(args, " ")
	for i, s := range array {
		if s == "--vus" && vus != "" {
			array[i+1] = vus
		}
		if s == "--duration" && duration != "" {
			array[i+1] = duration + "s"
		}
		if s == "--rps" && rps != "" {
			array[i+1] = rps
		}
	}
	args = strings.Join(array, " ")
	return args
}

func GetEnvList(envVar []v1alpha1.EnvVar) []interface{} {

	var envList []interface{}

	for _, v := range envVar {
		if v.Name != "" {
			env := map[string]interface{}{
				"name":  v.Name,
				"value": v.Value,
			}
			envList = append(envList, env)
		}
		if v.ValueFrom.SecretKeyRef.Name != "" {
			env := map[string]interface{}{
				"name":  v.Name,
				"valueFrom": map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": v.ValueFrom.SecretKeyRef.Name,
						"key": v.ValueFrom.SecretKeyRef.Key,
					},
				},
			}
			envList = append(envList, env)
		}
	}

	return envList
}
