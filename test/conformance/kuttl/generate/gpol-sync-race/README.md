# GeneratingPolicy Sync Race Condition Test

This test validates the fix for the race condition when a GeneratingPolicy has both
`spec.evaluation.synchronize.enabled: true` and `spec.evaluation.generateExisting.enabled: true`.

## Bug Summary

When both flags are enabled, two UpdateRequests were created concurrently:
1. UR with `cacheRestore=true` (for watcher cache restoration)
2. UR with `cacheRestore=false` (for generateExisting)

These URs raced, causing resources to be created and immediately deleted.

## Test Steps

### Manual Testing

1. Create KinD cluster:
```bash
make kind-create-cluster
```

2. Build and deploy Kyverno:
```bash
make kind-deploy-kyverno
```

3. Create test namespace and trigger resource:
```bash
kubectl create namespace test-gpol-sync
kubectl create deployment nginx --image=nginx -n test-gpol-sync
```

4. Apply the GeneratingPolicy:
```bash
kubectl apply -f policy.yaml
```

5. Watch ConfigMaps in the namespace:
```bash
kubectl get cm -n test-gpol-sync -w
```

**Expected**: ConfigMap `nginx-config` should be created and persist.
**Before fix**: ConfigMap would be created then immediately deleted.

6. Update the trigger to test UPDATE sync:
```bash
kubectl set env deployment/nginx TEST=1 -n test-gpol-sync
```

**Expected**: ConfigMap should remain (or be recreated if deleted).

7. Delete the trigger to test cleanup:
```bash
kubectl delete deployment nginx -n test-gpol-sync
```

**Expected**: ConfigMap should be deleted (sync cleanup).

8. Recreate trigger:
```bash
kubectl create deployment nginx --image=nginx -n test-gpol-sync
```

**Expected**: ConfigMap should be recreated.

## Cleanup

```bash
kubectl delete -f policy.yaml
kubectl delete namespace test-gpol-sync
```
