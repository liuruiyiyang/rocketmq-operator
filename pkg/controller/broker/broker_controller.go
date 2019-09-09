/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package broker

import (
	"context"
	"k8s.io/apimachinery/pkg/api/resource"

	cachev1alpha1 "github.com/operator-sdk-samples/rocketmq-operator/pkg/apis/cache/v1alpha1"
	cons "github.com/operator-sdk-samples/rocketmq-operator/pkg/constants"
	"github.com/operator-sdk-samples/rocketmq-operator/pkg/share"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strconv"
	"time"
)

var log = logf.Log.WithName("controller_broker")

// Add creates a new Broker Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBroker{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("broker-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Broker
	err = c.Watch(&source.Kind{Type: &cachev1alpha1.Broker{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Broker
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cachev1alpha1.Broker{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileBroker{}

// ReconcileBroker reconciles a Broker object
type ReconcileBroker struct {
	// TODO: Clarify the split client
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Broker object and makes changes based on the state read
// and what is in the Broker.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Broker StatefulSet for each Broker CR
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBroker) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Broker.")

	// Fetch the Broker instance
	broker := &cachev1alpha1.Broker{}
	err := r.client.Get(context.TODO(), request.NamespacedName, broker)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Broker resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Broker.")
		return reconcile.Result{}, err
	}



	share.GroupNum = int(broker.Spec.Size)
	slavePerGroup := broker.Spec.SlavePerGroup
	reqLogger.Info("brokerGroupNum=" + strconv.Itoa(share.GroupNum) + ", slavePerGroup=" + strconv.Itoa(slavePerGroup))

	for brokerClusterIndex := 0; brokerClusterIndex < share.GroupNum; brokerClusterIndex++ {
		reqLogger.Info("Check Broker cluster " + strconv.Itoa(brokerClusterIndex+1) + "/" + strconv.Itoa(share.GroupNum))
		// prepare pv and pvc for master broker
		pvName := genPVName(broker, brokerClusterIndex, false, 0)
		pvcName := genPVCName(broker, brokerClusterIndex, false, 0)
		pvReconcileError := r.reconcilePV(broker, brokerClusterIndex, false, 0)
		if pvReconcileError != nil {
			reqLogger.Error(pvReconcileError, "Failed to reconcile PV " + broker.Namespace + "/" + pvName)
		}
		pvcReconcileError := r.reconcilePVC(broker, brokerClusterIndex, false, 0)
		if pvcReconcileError != nil {
			reqLogger.Error(pvcReconcileError, "Failed to reconcile PVC " + broker.Namespace + "/" + pvcName)
		}

		// Check if the statefulSet already exists, if not create a new one
		foundMasterBroker := &appsv1.StatefulSet{}
		dep := r.statefulSetForMasterBroker(broker, brokerClusterIndex, pvcName)
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, foundMasterBroker)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Master Broker StatefulSet.", "StatefulSet.Namespace", dep.Namespace, "StatefulSet.Name", dep.Name)
			err = r.client.Create(context.TODO(), dep)

			if err != nil {
				reqLogger.Error(err, "Failed to create new StatefulSet of "+cons.BrokerClusterPrefix+strconv.Itoa(brokerClusterIndex), "StatefulSet.Namespace", dep.Namespace, "StatefulSet.Name", dep.Name)
			}
		} else if err != nil {
			reqLogger.Error(err, "Failed to get broker master StatefulSet.")
		}

		for slaveIndex := 1; slaveIndex <= slavePerGroup; slaveIndex++ {
			reqLogger.Info("Check Slave Broker of cluster-" + strconv.Itoa(brokerClusterIndex) + " " + strconv.Itoa(slaveIndex) + "/" + strconv.Itoa(slavePerGroup))

			// prepare pv and pvc for slave broker
			pvName := genPVName(broker, brokerClusterIndex, true, slaveIndex)
			pvcName := genPVCName(broker, brokerClusterIndex, true, slaveIndex)

			pvReconcileError := r.reconcilePV(broker, brokerClusterIndex, true, slaveIndex)
			if pvReconcileError != nil {
				reqLogger.Error(pvReconcileError, "Failed to reconcile PV " + broker.Namespace + "/" + pvName)
			}

			pvcReconcileError := r.reconcilePVC(broker, brokerClusterIndex, true, slaveIndex)
			if pvcReconcileError != nil {
				reqLogger.Error(pvcReconcileError, "Failed to reconcile PVC " + broker.Namespace + "/" + pvcName)
			}

			foundSlaveBroker := &appsv1.StatefulSet{}
			slaveDep := r.statefulSetForSlaveBroker(broker, brokerClusterIndex, slaveIndex, pvcName)
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: slaveDep.Name, Namespace: slaveDep.Namespace}, foundSlaveBroker)
			if err != nil && errors.IsNotFound(err) {
				reqLogger.Info("Creating a new Slave Broker StatefulSet.", "StatefulSet.Namespace", slaveDep.Namespace, "StatefulSet.Name", slaveDep.Name)
				err = r.client.Create(context.TODO(), slaveDep)
				if err != nil {
					reqLogger.Error(err, "Failed to create new StatefulSet of broker-"+strconv.Itoa(brokerClusterIndex)+"-slave-"+strconv.Itoa(slaveIndex), "StatefulSet.Namespace", slaveDep.Namespace, "StatefulSet.Name", slaveDep.Name)
				}
			} else if err != nil {
				reqLogger.Error(err, "Failed to get broker slave StatefulSet.")
			}
		}

		// The following code will restart all brokers to update NAMESRV_ADDR env
		/*if share.IsMetaServersStrUpdated {
			reqLogger.Info("Update Master Broker NAMESRV_ADDR of "+cons.BrokerClusterPrefix+strconv.Itoa(brokerClusterIndex))
			found.Spec.Template.Spec.Containers[0].Env[0].Value = share.MetaServersStr
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				reqLogger.Error(err, "Failed to update NAMESRV_ADDR of master broker in "+cons.BrokerClusterPrefix+strconv.Itoa(brokerClusterIndex), "StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name)
			}
			for slaveIndex := 1; slaveIndex <= slavePerGroup; slaveIndex++ {
				reqLogger.Info("Update Slave Broker NAMESRV_ADDR of cluster-" + strconv.Itoa(brokerClusterIndex) + " " + strconv.Itoa(slaveIndex) + "/" + strconv.Itoa(slavePerGroup))
				slaveDep := r.statefulSetForSlaveBroker(broker, brokerClusterIndex, slaveIndex)
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: slaveDep.Name, Namespace: slaveDep.Namespace}, found)
				found.Spec.Template.Spec.Containers[0].Env[0].Value = share.MetaServersStr
				err = r.client.Update(context.TODO(), found)
				if err != nil {
					reqLogger.Error(err, "Failed to update NAMESRV_ADDR of broker-"+strconv.Itoa(brokerClusterIndex)+"-slave-"+strconv.Itoa(slaveIndex), "StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name)
				}
			}
		}*/
	}
	share.IsMetaServersStrUpdated = false

	// Ensure the statefulSet size is the same as the spec
	//size := broker.Spec.Size
	//if *found.Spec.Replicas != size {
	//	found.Spec.Replicas = &size
	//	err = r.client.Update(context.TODO(), found)
	//	if err != nil {
	//		reqLogger.Error(err, "Failed to update StatefulSet.", "StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name)
	//		return reconcile.Result{}, err
	//	}
	//	// Spec updated - return and requeue
	//	return reconcile.Result{Requeue: true}, nil
	//}

	// Update the Broker status with the pod names
	// List the pods for this broker's statefulSet

	//podList := &corev1.PodList{}
	//labelSelector := labels.SelectorFromSet(labelsForBroker(broker.Name))
	//listOps := &client.ListOptions{
	//	Namespace:     broker.Namespace,
	//	LabelSelector: labelSelector,
	//}
	//err = r.client.List(context.TODO(), listOps, podList)
	//if err != nil {
	//	reqLogger.Error(err, "Failed to list pods.", "Broker.Namespace", broker.Namespace, "Broker.Name", broker.Name)
	//	return reconcile.Result{}, err
	//}
	//podNames := getPodNames(podList.Items)
	//
	//// Update status.Nodes if needed
	//if !reflect.DeepEqual(podNames, broker.Status.Nodes) {
	//	broker.Status.Nodes = podNames
	//	err := r.client.Status().Update(context.TODO(), broker)
	//	if err != nil {
	//		reqLogger.Error(err, "Failed to update Broker status.")
	//		return reconcile.Result{}, err
	//	}
	//}

	return reconcile.Result{true, time.Duration(3) * time.Second}, nil
}

func genName(cr *cachev1alpha1.Broker, middleName string, clusterIndex int, isSlave bool, slaveIndex int) string {
	name := ""
	if isSlave {
		name = cr.Spec.PersistentVolumeClaim.Name + "-" + middleName + "-cluster-" + strconv.Itoa(clusterIndex) + "-slave-" + strconv.Itoa(slaveIndex)
	} else {
		name = cr.Spec.PersistentVolumeClaim.Name + "-" + middleName + "-cluster-" + strconv.Itoa(clusterIndex) + "-master"
	}
	return name
}

func genPVName(cr *cachev1alpha1.Broker, clusterIndex int, isSlave bool, slaveIndex int) string {
	return genName(cr, "pv", clusterIndex, isSlave, slaveIndex)
}

func genPVCName(cr *cachev1alpha1.Broker, clusterIndex int, isSlave bool, slaveIndex int) string {
	return genName(cr, "pvc", clusterIndex, isSlave, slaveIndex)
}

// newPV creates the PV used by the PVCs.
func newPV(cr *cachev1alpha1.Broker, clusterIndex int, isSlave bool, slaveIndex int) *corev1.PersistentVolume {
	ls := labelsForBroker(cr.Name)
	name := genPVName(cr, clusterIndex, isSlave, slaveIndex)
	return &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolume",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cr.Namespace,
			Labels:    ls,
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: map[corev1.ResourceName]resource.Quantity{corev1.ResourceStorage: cr.Spec.PersistentVolumeClaim.Spec.Resources.Requests[corev1.ResourceStorage]},
			AccessModes: cr.Spec.PersistentVolumeClaim.Spec.AccessModes,
			PersistentVolumeReclaimPolicy: cr.Spec.PersistentVolumeReclaimPolicy,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server: cr.Spec.NfcServer,
					Path: cr.Spec.NfcPath,
				},
			},
		},
	}
}

// newPVCs creates the PVCs used by the application.
func newPVC(cr *cachev1alpha1.Broker, clusterIndex int, isSlave bool, slaveIndex int) *corev1.PersistentVolumeClaim {
	ls := labelsForBroker(cr.Name)
	name := genPVCName(cr, clusterIndex, isSlave, slaveIndex)
	return &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cr.Namespace,
			Labels:    ls,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: cr.Spec.PersistentVolumeClaim.Spec.AccessModes,
			Resources: cr.Spec.PersistentVolumeClaim.Spec.Resources,
		},
	}
}

func (r *ReconcileBroker) reconcilePV(cr *cachev1alpha1.Broker, clusterIndex int, isSlave bool, slaveIndex int) error {
	reqLogger := log.WithValues("Broker.Namespace", cr.Namespace, "Broker.Name", cr.Name)
	// Check if this Persistent Volume already exists
	name := genPVName(cr, clusterIndex, isSlave, slaveIndex)
	foundPV := &corev1.PersistentVolume{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: cr.Namespace}, foundPV)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Persistent Volume " + cr.Namespace + "/" + name)
		pv := newPV(cr, clusterIndex, isSlave, slaveIndex)

		// Set instance as the owner and controller
		if err = controllerutil.SetControllerReference(cr, pv, r.scheme); err != nil {
			reqLogger.Error(err, "Set PV " + pv.Name + " as the owner and controller failed")
			return err
		}

		err = r.client.Create(context.TODO(), pv)
		if err != nil {
			reqLogger.Error(err, "Create PV " + pv.Name +  " failed")
			return err
		}
		reqLogger.Info("Successfully created a new Persistent Volume: " + pv.Namespace + "/" + pv.Name)
		// Persitent Volume created successfully

		// TODO currently not support update PV

		return nil
	} else if err != nil {
		reqLogger.Error(err, "Get PV failed")
		return err
	}

	log.Info("Skip reconcile: Persistent Volume " + foundPV.Namespace + "/" + foundPV.Name + " because already exists")
	return nil
}

// reconcilePVC ensures the Persistent Volume Claim is created.
func (r *ReconcileBroker) reconcilePVC(cr *cachev1alpha1.Broker, clusterIndex int, isSlave bool, slaveIndex int) error {
	reqLogger := log.WithValues("Broker.Namespace", cr.Namespace, "Broker.Name", cr.Name)

	// Check if this Persitent Volume Claim already exists
	name := genPVCName(cr, clusterIndex, isSlave, slaveIndex)
	foundPVC := &corev1.PersistentVolumeClaim{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: cr.Namespace}, foundPVC)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Persistent Volume Claim " + cr.Namespace + "/" + cr.Spec.PersistentVolumeClaim.Name)
		pvc := newPVC(cr, clusterIndex, isSlave, slaveIndex)

		// Set instance as the owner and controller
		if err = controllerutil.SetControllerReference(cr, pvc, r.scheme); err != nil {
			reqLogger.Error(err, "Set PVC " + pvc.Name + " as the owner and controller failed")
			return err
		}

		err = r.client.Create(context.TODO(), pvc)
		if err != nil {
			reqLogger.Error(err, "Create PVC " + pvc.Name +  " failed")
			return err
		}
		reqLogger.Info("Successfully created a new Persistent Volume Claim: " + pvc.Namespace + "/" + pvc.Name)
		// Persitent Volume Claim created successfully

		// currently not support update PVC
		//cr.Status.PersistentVolumeClaimName = pvc.Name
		//err = r.client.Update(context.TODO(), cr)
		//if err != nil {
		//	reqLogger.Error(err, "Update PVC " + pvc.Name +  " failed")
		//	return err
		//}

		return nil
	} else if err != nil {
		reqLogger.Error(err, "Get PVC failed")
		return err
	}

	log.Info("Skip reconcile: Persistent Volume Claim " + foundPVC.Namespace + "/" + foundPVC.Name + " because already exists")
	return nil
}


// statefulSetForBroker returns a master broker StatefulSet object
func (r *ReconcileBroker) statefulSetForMasterBroker(m *cachev1alpha1.Broker, brokerClusterIndex int, pvcName string) *appsv1.StatefulSet {
	ls := labelsForBroker(m.Name)
	var a int32 = 1
	var c = &a
	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-" + strconv.Itoa(brokerClusterIndex) + "-master",
			Namespace: m.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: c,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           m.Spec.BrokerImage,
						Name:            cons.MasterBrokerContainerNamePrefix + strconv.Itoa(brokerClusterIndex),
						ImagePullPolicy: m.Spec.ImagePullPolicy,
						Env: []corev1.EnvVar{{
							Name:  cons.EnvNamesrvAddr,
							Value: m.Spec.NameServers,
						}, {
							Name:  cons.EnvReplicationMode,
							Value: m.Spec.ReplicationMode,
						}, {
							Name:  cons.EnvBrokerId,
							Value: "0",
						}, {
							Name:  cons.EnvBrokerClusterName,
							Value: cons.BrokerClusterPrefix + strconv.Itoa(brokerClusterIndex),
						}},
						//Command: []string{"cmd", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 10909,
							Name:          "10909port",
						}, {
							ContainerPort: 10911,
							Name:          "10911port",
						}, {
							ContainerPort: 10912,
							Name:          "10912port",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: cons.LogMountPath,
							Name: m.Name + "-" + strconv.Itoa(brokerClusterIndex) + "-master-logs",
							SubPath: m.Name + "-" + strconv.Itoa(brokerClusterIndex) + "-master-logs",
						},{
							MountPath: cons.StoreMountPath,
							Name: m.Name + "-" + strconv.Itoa(brokerClusterIndex) + "-master-store",
							SubPath: m.Name + "-" + strconv.Itoa(brokerClusterIndex) + "-master-store",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: m.Name + "-" + strconv.Itoa(brokerClusterIndex) + "-master-store-volume",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: pvcName,
							},
						},
					}},
				},
			},
		},
	}
	// Set Broker instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)

	return dep

}

// statefulSetForBroker returns a slave broker StatefulSet object
func (r *ReconcileBroker) statefulSetForSlaveBroker(m *cachev1alpha1.Broker, brokerClusterIndex int, slaveIndex int, pvcName string) *appsv1.StatefulSet {
	ls := labelsForBroker(m.Name)
	var a int32 = 1
	var c = &a
	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-" + strconv.Itoa(brokerClusterIndex) + "-slave-" + strconv.Itoa(slaveIndex),
			Namespace: m.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: c,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           m.Spec.BrokerImage,
						Name:            cons.SlaveBrokerContainerNamePrefix + strconv.Itoa(brokerClusterIndex),
						ImagePullPolicy: m.Spec.ImagePullPolicy,
						Env: []corev1.EnvVar{{
							Name:  cons.EnvNamesrvAddr,
							Value: m.Spec.NameServers,
						}, {
							Name:  cons.EnvReplicationMode,
							Value: m.Spec.ReplicationMode,
						}, {
							Name:  cons.EnvBrokerId,
							Value: strconv.Itoa(slaveIndex),
						}, {
							Name:  cons.EnvBrokerClusterName,
							Value: cons.BrokerClusterPrefix + strconv.Itoa(brokerClusterIndex),
						}},
						//Command: []string{"cmd", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 10909,
							Name:          "10909port",
						}, {
							ContainerPort: 10911,
							Name:          "10911port",
						}, {
							ContainerPort: 10912,
							Name:          "10912port",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: cons.LogMountPath,
							Name: m.Name + "-" + strconv.Itoa(brokerClusterIndex) + "-slave-" + strconv.Itoa(slaveIndex) + "-logs",
							SubPath: m.Name + "-" + strconv.Itoa(brokerClusterIndex) + "-slave-" + strconv.Itoa(slaveIndex) + "-logs",
						},{
							MountPath: cons.StoreMountPath,
							Name: m.Name + "-" + strconv.Itoa(brokerClusterIndex) + "-slave-" + strconv.Itoa(slaveIndex) + "-store",
							SubPath: m.Name + "-" + strconv.Itoa(brokerClusterIndex) + "-slave-" + strconv.Itoa(slaveIndex) + "-store",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: m.Name + "-" + strconv.Itoa(brokerClusterIndex) + "-slave-" + strconv.Itoa(slaveIndex) + "-store-volume",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: pvcName,
							},
						},
					}},
				},
			},
		},
	}
	// Set Broker instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)

	return dep

}

// labelsForBroker returns the labels for selecting the resources
// belonging to the given broker CR name.
func labelsForBroker(name string) map[string]string {
	return map[string]string{"app": "broker", "broker_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
