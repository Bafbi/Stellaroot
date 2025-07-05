## Seting up kind cluster for local testing

### Using kind for Local Testing
1. Install kind if you haven't already. You can find the installation instructions at [kind installation guide](https://kind.sigs.k8s.io/docs/user/quick-start/#installation).
2. Create a kind cluster with the following command:
   ```bash
   kind create cluster --name stellaroot-local --config kind-config.yaml
   ```
   This will create a new kind cluster named `aixpa-env`, and gonna merge the kubeconfig file with your current `.kube/config`.
3. Add some local DNS entries to your `/etc/hosts` file (or equivalent on Windows `C:\Windows\System32\drivers\etc\hosts`):
   ```plaintext
   # Added for local testing of aixpa cluster
   127.0.0.1 argocd.local
   127.0.0.1 test.local
   ```
