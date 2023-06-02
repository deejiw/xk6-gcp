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

const gcpClient = new Gcp()
export default function() {
}
```
