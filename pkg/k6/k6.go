package k6

import (
	"context"
	// "io/ioutil"
	"log"
	"os"

	// "strconv"
	"strings"

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

	// bytes, err := ioutil.ReadFile(template)
	// if err != nil {
	// 	return nil, err
	// }

	// var currentK6 map[string]interface{}
	// err = yaml.Unmarshal(bytes, &currentK6)
	// if err != nil {
	// 	return nil, err
	// }

	f, err := os.Open(template)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    d := yaml.NewDecoder(f)
    
	var currentK6 map[string]interface{}
    if err := d.Decode(&currentK6); err != nil {
        log.Fatal(err)
    }

	k6Res := schema.GroupVersionResource{Group: "k6.io", Version: "v1alpha1", Resource: "k6s"}

	k6 := &unstructured.Unstructured{
		Object: currentK6,
	}

	if len(k6.Object["spec"].(map[interface{}]interface{})["arguments"].(string)) != 0 {
		args := OverrideArgs(k6.Object["spec"].(map[interface{}]interface{})["arguments"].(string), vus, duration, rps)
		k6.Object["spec"].(map[interface{}]interface{})["arguments"] = args
	}

	if len(parallelism) != 0 {
		k6.Object["spec"].(map[interface{}]interface{})["parallelism"] = parallelism
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
	CurrentK6      map[string]interface{}
}

func (k *K6) CreateK6() error {
	namespace := k.CurrentK6["metadata"].(map[interface{}]interface{})["namespace"].(string)
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
	name := k.CurrentK6["metadata"].(map[interface{}]interface{})["name"].(string)
	namespace := k.CurrentK6["metadata"].(map[interface{}]interface{})["namespace"].(string)

	err := k.client.Resource(k.Resource).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	log.Printf("k6.k6.io/%q deleted\n", name)

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
