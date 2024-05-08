*Sample run:*
- cd to projects/volsnap
- install CR-  `make install`
- create manifest - `make install`
- start controller - `make run`
- In another terminal run
  - pvc creation - `kubectl apply -f config/samples/example-pvc.yaml`
  - snapshot creation - `kubectl apply -f config/samples/vol_v1_volsnap.yaml`
 - check result -
     -  `kubectl get volumesnapshot`
  ```
/projects/volsnap$ kubectl get volumesnapshot volsnap-sample-try-3
NAME                   READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS            SNAPSHOTCONTENT                                    CREATIONTIME   AGE
volsnap-sample-try-3   true         csi-pvc                             1Gi           csi-hostpath-snapclass   snapcontent-b5041279-f0ac-4b35-a522-fbbc8314362b   21m            21m
```
