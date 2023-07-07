import { Gcp } from 'k6/x/gcp';

const jsonKey = JSON.parse(open('credentials.json'))

const gcp = new Gcp({
  key: jsonKey,
  scope: ['https://www.googleapis.com/auth/cloud-platform'] // Default value
})
export default function() {
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
  console.log(result) // Monitoring API Query Result
}
