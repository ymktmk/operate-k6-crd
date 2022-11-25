## operate-k6-crd

### 🚨 still in development 🚨

A GitHub Actions tool for working with K6 CRDs.

vus, duration, rps, parallelism can be overridden

* Some env setting are not supported

### Example of use

```yaml
- name: Create K6 CRD
  uses: ymktmk/operate-k6-crd@main
  with: 
    method: create
    parallelism: 1
    template: ./example/k6.yaml
```

### K6 CRD Example

```yaml
apiVersion: k6.io/v1alpha1
kind: K6
metadata:
  name: k6-sample
  namespace: k6-operator-system
spec:
  parallelism: 3
  script:
    configMap:
      name: crocodile-stress-test
      file: test.js
  arguments: --vus 4 --duration 30s --rps 10
  runner:
    env:
    - name: URL
      value: "https://test.k6.io"
    - name: SLACK_TOKEN
      valueFrom: 
        secretKeyRef:
          name: secret
          key: token
```
