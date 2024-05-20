**The project contains 2 CRDs:**
- volsnap- responsible for creating backup/volumesnapshot of given PVC
- volrestore- responsible for creating PVC from given volsnap

**Sample run**
- cd to projects/volsnap
- install CR-  `make install`
- create manifest - `make install`
- start controller - `make run`

- In another terminal run
  - pvc creation - `kubectl apply -f config/samples/example-pvc.yaml`

***Creating Volsnap***
 - snapshot creation - `kubectl apply -f config/samples/vol_v1_volsnap.yaml`
 - check result -
     -  `kubectl get volumesnapshot`
  ```
/projects/volsnap$ kubectl get volumesnapshot volsnap-sample-try-3
NAME                   READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS            SNAPSHOTCONTENT                                    CREATIONTIME   AGE
volsnap-sample-try-3   true         csi-pvc                             1Gi           csi-hostpath-snapclass   snapcontent-b5041279-f0ac-4b35-a522-fbbc8314362b   21m            21m
 ```

**Restoring volsnap using volrestore:**
- restore creation - `kubectl apply -f config/samples/vol_v1_volrestore.yaml`
- check result-
  
```
  /go/src/github.com/projects/volsnap$ kubectl get volrestore
NAME                  AGE
volrestore-sample-4   85s

/go/src/github.com/projects/volsnap$ kubectl get pvc restore-volrestore-sample-4
NAME                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      VOLUMEATTRIBUTESCLASS   AGE
restore-volrestore-sample-4   Bound    pvc-a1f909b0-012a-4366-81dc-27ca6ebc0b50   2          RWO            csi-hostpath-sc   <unset>                 3d23h

```
