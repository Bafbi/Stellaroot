## Seting up kind cluster for local testing

### Using kind for Local Testing
1. Install kind if you haven't already. You can find the installation instructions at [kind installation guide](https://kind.sigs.k8s.io/docs/user/quick-start/#installation).
1. Create a kind cluster with the following command:
   ```bash
   kind create cluster --name stellaroot-local --config kubernetes/kind-config.yaml
   ```
   This will create a new kind cluster named `stellaroot-local`, and gonna merge the kubeconfig file with your current `.kube/config`.
1. Download and install the `cloud-provider-kind` to get a load balancer. [doc](https://kind.sigs.k8s.io/docs/user/loadbalancer/)
<!-- 3. Add some local DNS entries to your `/etc/hosts` file (or equivalent on Windows `C:\Windows\System32\drivers\etc\hosts`):
   ```plaintext
   # Added for local testing of stellaroot cluster
   127.0.0.1 argocd.local
   127.0.0.1 test.local
   ``` -->

