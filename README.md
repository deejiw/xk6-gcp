# xk6-gcp

This is a [k6](https://k6.io) extension using the [xk6](https://github.com/grafana/xk6) system.

## Build

To build a `k6` binary with this extension, first ensure you have the prerequisites:

- [Go toolchain](https://go101.org/article/go-toolchain.html)
- Git

Then:

1. Install `xk6`:
  ```shell
  $ go install go.k6.io/xk6/cmd/xk6@latest
  ```

2. Build the binary:
  ```shell
  $ xk6 build --with github.com/deejiw/xk6-gcp@latest
  ```

## Example

```javascript
import { Gcp } from 'k6/x/gcp';

const jsonKey = JSON.parse(open('credentials.json'))

const gcp = new Gcp({
  key: jsonKey
  scope: ['https://www.googleapis.com/auth/cloud-platform'] // Default value
})
export default function() {
  const accessToken = gcp.getOAuth2AccessToken()
  console.log(accessToken['AccessToken'])

  const query = `fetch k8s_container
| metric 'kubernetes.io/container/cpu/limit_utilization'
| filter (resource.cluster_name == 'CLUSTER_NAME' &&
          resource.namespace_name == 'NAMESPACE_NAME' &&
          resource.pod_name =~ 'POD_NAME')
| group_by 1m, [value_limit_utilization_max: max(value.limit_utilization)]
| {
    top 2 | value [is_default_value: false()]
  ;
    ident
  }
| outer_join true(), _
| filter is_default_value
| value drop [is_default_value]
| every 1m
| condition val(0) > 0.73 '1'
`

  const result = gcp.queryTimeSeries('my-project-id', query)
  console.log(result)

}
```

## Command
k6 run script.js
