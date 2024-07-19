package k6

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/ymktmk/apply-k6-crd/api/v1alpha1"
	"gopkg.in/yaml.v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type K6 struct {
	client         dynamic.Interface
	Resource       schema.GroupVersionResource
	UnstructuredK6 *unstructured.Unstructured
	CurrentK6      v1alpha1.TestRun
}

func NewK6(template, vus, duration, rps, parallelism, file string) (*K6, error) {

	client, err := NewClientSet()
	if err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(template)
	if err != nil {
		return nil, err
	}

	var currentK6 v1alpha1.TestRun
	err = yaml.Unmarshal(bytes, &currentK6)
	if err != nil {
		return nil, err
	}

	// Only create after this logic
	err = Validate(currentK6)
	if err != nil {
		return nil, err
	}

	currentK6.ObjectMeta.Name = generateRandomName("k6")

	numberOfJobs := currentK6.Spec.Parallelism
	if parallelism != "" {
		num, err := strconv.ParseInt(parallelism, 10, 32)
		if err != nil {
			return nil, err
		}
		numberOfJobs = int32(num)
	}

	// js
	jsFile := currentK6.Spec.Script.ConfigMap.File
	if file != "" {
		jsFile = file
	}

	k6Res := schema.GroupVersionResource{Group: "k6.io", Version: "v1alpha1", Resource: "testruns"}

	k6 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "k6.io/v1alpha1",
			"kind":       "TestRun",
			"metadata": map[string]interface{}{
				"name":      currentK6.ObjectMeta.Name,
				"namespace": currentK6.ObjectMeta.Namespace,
			},
			"spec": map[string]interface{}{
				// int64
				"arguments":   "",
				"parallelism": int64(numberOfJobs),
				"script": map[string]interface{}{
					"configMap": map[string]interface{}{
						"name": currentK6.Spec.Script.ConfigMap.Name,
						"file": jsFile,
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

	if len(currentK6.Spec.Runner.Env) != 0 {
		envList := getEnvList(currentK6.Spec.Runner.Env)
		err = unstructured.SetNestedSlice(runner, envList, "env")
		if err != nil {
			return nil, err
		}
	}

	// arguments into spec
	if len(currentK6.Spec.Arguments) != 0 {
		args := overrideArgs(currentK6.Spec.Arguments, vus, duration, rps)
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

func (k *K6) CreateK6() error {
	name := k.CurrentK6.ObjectMeta.Name
	namespace := k.CurrentK6.ObjectMeta.Namespace
	k6 := k.UnstructuredK6

	result, err := k.client.Resource(k.Resource).Namespace(namespace).Create(context.TODO(), k6, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	if result != nil {
		log.Printf("k6.k6.io/%q created\n", result.GetName())
	}

	// delete k6 crd
	managementInformerFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(k.client, 0, metav1.NamespaceAll, nil)

	gvrMachineSet := schema.GroupVersionResource{
		Group:    "k6.io",
		Version:  "v1alpha1",
		Resource: "testruns",
	}

	stopCh := make(chan struct{})
	closeCh := make(chan bool)

	machineSetInformer := managementInformerFactory.ForResource(gvrMachineSet)

	machineSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			status := newObj.(*unstructured.Unstructured).Object["status"].(map[string]interface{})["stage"]
			log.Printf("K6 Status: %s\n", status)
			if status == "finished" {
				err := k.client.Resource(k.Resource).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
				if err != nil {
					log.Printf("Error: %s\n", err)
					os.Exit(1)
				}
				log.Printf("k6.k6.io/%q deleted\n", name)
				closeCh <- true
			}
		},
	})

	for {
		go func() {
			managementInformerFactory.Start(stopCh)
		}()
		close := <-closeCh
		if close {
			break
		}
	}

	return nil
}

func Validate(k6 v1alpha1.TestRun) error {

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

func overrideArgs(args, vus, duration, rps string) string {
	array := strings.Split(args, " ")
	if vus != "" {
		array = append(array, "--vus "+vus)
	}
	if duration != "" {
		array = append(array, "--duration "+duration+"s")
	}
	if rps != "" {
		array = append(array, "--rps "+rps)
	}
	args = strings.Join(array, " ")
	return args
}

func getEnvList(envVar []v1alpha1.EnvVar) []interface{} {
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
				"name": v.Name,
				"valueFrom": map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": v.ValueFrom.SecretKeyRef.Name,
						"key":  v.ValueFrom.SecretKeyRef.Key,
					},
				},
			}
			envList = append(envList, env)
		}
	}
	return envList
}

func generateRandomName(name string) string {
	lengthToGenerate := math.Min(float64(62-len(name)), float64(32))
	return fmt.Sprintf("%s-%s", name, secureRandomStr(int(lengthToGenerate)/2))
}

func secureRandomStr(b int) string {
	k := make([]byte, b)
	if _, err := rand.Read(k); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", k)
}
