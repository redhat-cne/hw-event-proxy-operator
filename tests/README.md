# Testing the Hardware Event Proxy Operator

## End to End Tests Using Kuttl

### Prerequisites
- [kubectl-kuttl](https://kuttl.dev/docs/#install-kuttl-cli)
- Point `$KUBECONFIG` to an existing Openshift cluster

### Run Test
```bash
 make test-ci
```

### Test Cases
0. Deploy hw-event-proxy-operator-controller-manager. 
    - Verify the pod is in Ready state.
1. Deploy hw-event-proxy, a sample HardwareEvent CR.
    - Since there is no Redfish secret present, the deployment should be blocked.
    - Verify the hw-event-proxy pod is not created.
2. Deploy Redfish secret.
    - Verify the hw-event-proxy pod is created and in Ready state.
3. Delete hw-event-proxy deployment directly by `oc delete deployment <name of deployment>`.
    - The operator controller manager should detect the missing deployment and redeploy it.
    - Verify the hw-event-proxy pod is recreated and in Ready state.
4. Cleanup. Delete hw-event-proxy and Redfish secret by `oc delete -f <yaml of CR>`.
    - Verify the resources are cleaned up. 
