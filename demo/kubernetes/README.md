
# Rook on Kubernetes
- [Quickstart](#quickstart)
- [Design](#design)

## Quickstart
This example shows how to build a simple, multi-tier web application on Kubernetes using persistent volumes enabled by Rook.

### Prerequisites

This example requires a running Kubernetes cluster with access to `modprobe` by the kubelet. To make sure you have a Kubernetes cluster that is ready for `rook`, you can [follow these quick instructions](pre-reqs.md).

Note that we are striving for even more smooth integration with Kubernetes in the future such that `rook` will work out of the box with any Kubernetes cluster.

### Deploy Rook

With your Kubernetes cluster running, Rook can be setup and deployed by simply creating the [rook-operator](rook-operator.yaml) deployment and creating a [rook cluster](rook-cluster.yaml).

```
cd demo/kubernetes
kubectl create -f rook-operator.yaml
kubectl create -f rook-cluster.yaml
```

Use `kubectl` to list pods in the rook namespace. You should be able to see the following: 

```
$ kubectl -n rook get pod
NAME                            READY     STATUS    RESTARTS   AGE
mon0                            1/1       Running   0          1m
mon1                            1/1       Running   0          1m
mon2                            1/1       Running   0          1m
osd-n1sm3                       1/1       Running   0          1m
osd-pb0sh                       1/1       Running   0          1m
osd-rth3q                       1/1       Running   0          1m
rgw-1785797224-9xb4r            1/1       Running   0          1m
rgw-1785797224-vbg8d            1/1       Running   0          1m
rook-api-4184191414-l0wmw       1/1       Running   0          1m
rook-operator-349747813-c3dmm   1/1       Running   0          1m
```
**NOTE:** RGW (object storage gateway) is currently deployed by default but in the future will be done only when needed (see [#413](https://github.com/rook/rook/issues/413))

### Provision Storage
Before Rook can start provisioning storage, a StorageClass needs to be created. This is used to specify information needed for Kubernetes to interoperate with Rook for provisioning persistent volumes.  Rook already creates a default admin and demo user, whose secrets are already specified in the sample [rook-storageclass.yaml](rook-storageclass.yaml).

Now we just need to specify the Ceph monitor endpoints (requires `jq`):

```
export MONS=$(kubectl -n rook get pod mon0 mon1 mon2 -o json|jq ".items[].status.podIP"|tr -d "\""|sed -e 's/$/:6790/'|paste -s -d, -)
sed 's#INSERT_HERE#'$MONS'#' rook-storageclass.yaml | kubectl create -f -
``` 
**NOTE:** In the v0.4 release we plan to expose monitors via DNS/service names instead of IP address (see [#355](https://github.com/rook/rook/issues/355)), which will streamline the experience and remove the need for this step.

### Consume the storage

Now that rook is running and integrated with Kubernetes, we can create a sample app to consume the block storage provisioned by rook. We will create the classic wordpress and mysql apps.
Both these apps will make use of block volumes provisioned by rook.

Start mysql and wordpress:

```
kubectl create -f mysql.yaml
kubectl create -f wordpress.yaml
```

Both of these apps create a block volume and mount it to their respective pod. You can see the Kubernetes volume claims by running the following:

```
$ kubectl get pvc
NAME             STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
mysql-pv-claim   Bound     pvc-95402dbc-efc0-11e6-bc9a-0cc47a3459ee   20Gi       RWO           1m
wp-pv-claim      Bound     pvc-39e43169-efc1-11e6-bc9a-0cc47a3459ee   20Gi       RWO           1m
```

Once the wordpress and mysql pods are in the `Running` state, get the cluster IP of the wordpress app and enter it in your brower:

```
$ kubectl get svc wordpress
NAME        CLUSTER-IP   EXTERNAL-IP   PORT(S)        AGE
wordpress   10.3.0.155   <pending>     80:30841/TCP   2m
```

You should see the wordpress app running.  

**NOTE:** When running in a vagrant environment, there will be no external IP address to reach wordpress with.  You will only be able to reach wordpress via the `CLUSTER-IP` from inside the Kubernetes cluster.

### Rook Client
You also have the option to use the `rook` client tool directly by running it in a pod that can be started in the cluster with:
```
kubectl create -f rook-client/rook-client.yml
```  

Starting the rook-client pod will take a bit of time to download the container, so you can check to see when it's ready with (it should be in the `Running` state):
```
kubectl -n rook get pod rook-client
```

Connect to the rook-client pod and verify the `rook` client can talk to the cluster:
```
kubectl -n rook exec rook-client -it bash
rook node ls
```

At this point, you can use the `rook` tool along with some [simple steps to create and manage block, file and object storage](../client/README.md).

## Design

With Rook running in the Kubernetes cluster, Kubernetes applications can
mount block devices and filesystems managed by Rook, or can use the S3/Swift API for object storage. The Rook operator 
automates configuration of the Ceph storage components and monitors the cluster to ensure the storage remains available
and healthy. There is also a REST API service for configuring the Rook storage and a command line tool called `rook`.

![Rook Architecture on Kubernetes](/Documentation/media/kubernetes.png)

The Rook operator is a simple container containing the `rook-operator` binary that has all that is needed to bootstrap
and monitor the storage cluster. The operator will start and monitor ceph monitor pods and a daemonset for the OSDs to
which provides basic RADOS storage, as well as a deployment for a RESTful API service. When requested through the api service,
object storage (S3/Swift) is enabled by starting a deployment for RGW, while a shared file system is enabled with a deployment for MDS.

The operator will monitor the storage daemons to ensure the cluster is healthy, to start more ceph mons when necessary, to 
failover mons, make adjustments as the cluster grows or shrinks, etc. The operator will also watch for desired state changes 
requested by the api service and apply the changes.

The Rook daemons (Mons, OSDs, RGW, and MDS) are compiled to a single binary `rookd`, and included in a minimal container.
`rookd` uses an embedded version of Ceph for storing all data -- there are no changes to the data path. 
Rook does not attempt to maintain full fidelity with Ceph. Many of the Ceph concepts like placement groups and crush maps 
are hidden so you don't have to worry about them. Instead Rook creates a much simplified UX for admins that is in terms 
of physical resources, pools, volumes, filesystems, and buckets.

Rook is implemented in golang. Ceph is implemented in C++ where the data path is highly optimized. We believe
this combination offers the best of both worlds.

See [Design](https://github.com/rook/rook/wiki/Design) wiki for more details.