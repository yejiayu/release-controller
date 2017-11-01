package release

import (
	"time"

	informerrelease "github.com/caicloud/clientset/informers/release/v1alpha1"
	releasev1alpha1 "github.com/caicloud/clientset/kubernetes/typed/release/v1alpha1"
	listerrelease "github.com/caicloud/clientset/listers/release/v1alpha1"
	"github.com/caicloud/release-controller/pkg/kube"
	"github.com/caicloud/release-controller/pkg/release"
	"github.com/caicloud/release-controller/pkg/render"
	"github.com/caicloud/release-controller/pkg/storage"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// ReleaseController watches all resource related release and release history.
type ReleaseController struct {
	queue            workqueue.RateLimitingInterface
	manager          release.ReleaseManager
	releaseLister    listerrelease.ReleaseLister
	releaseHasSynced cache.InformerSynced
}

// NewReleaseController creates a release controller.
func NewReleaseController(
	clients kube.ClientPool,
	codec kube.Codec,
	releaseClient releasev1alpha1.ReleaseV1alpha1Interface,
	releaseInformer informerrelease.ReleaseInformer,
) (*ReleaseController, error) {
	client, err := kube.NewClient(clients, codec)
	if err != nil {
		return nil, err
	}
	handler := release.NewReleaseHandler(render.NewRender(), client)
	backend := storage.NewReleaseBackend(releaseClient)
	rc := &ReleaseController{
		queue:            workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		manager:          release.NewReleaseManager(backend, handler),
		releaseLister:    releaseInformer.Lister(),
		releaseHasSynced: releaseInformer.Informer().HasSynced,
	}
	releaseInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: rc.enqueueRelease,
		UpdateFunc: func(oldObj, newObj interface{}) {
			rc.enqueueRelease(newObj)
		},
		DeleteFunc: rc.enqueueRelease,
	})
	return rc, nil
}

// keyForObj returns the key of obj.
func (rc *ReleaseController) keyForObj(obj interface{}) (string, error) {
	return cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
}

// enqueueRelease enqueues the key of obj.
func (rc *ReleaseController) enqueueRelease(obj interface{}) {
	key, err := rc.keyForObj(obj)
	if err != nil {
		glog.Errorf("Can't get obj key: %v", err)
		return
	}

	glog.V(4).Infof("Enqueue: %s", key)
	// key must be a string
	rc.queue.Add(key)
}

// Run starts controller and checks releases
func (rc *ReleaseController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	glog.Info("Running ReleaseController")

	if !cache.WaitForCacheSync(stopCh, rc.releaseHasSynced) {
		glog.Errorf("Can't sync cache")
		return
	}
	glog.Info("Sync ReleaseController cache successfully")

	go wait.Until(rc.worker, time.Second, stopCh)

	<-stopCh
	glog.Info("Shutting down ReleaseController")
}

// worker checks improper resources. If controller unexpectedly terminated,
// some resources may not delete completely. worker should detect those
// resources and let them in a correct posture.
func (rc *ReleaseController) worker() {
	if err := rc.manager.Run(); err != nil {
		glog.Errorf("Can't run manager: %v", err)
	}
	glog.V(3).Infof("Processing ReleaseController releases")
	for rc.processNextWorkItem() {
	}
}

// processNextWorkItem processes next release
func (rc *ReleaseController) processNextWorkItem() bool {
	key, quit := rc.queue.Get()
	if quit {
		glog.Error("Unexpected quit of release queue")
		return false
	}
	defer rc.queue.Done(key)
	glog.V(4).Infof("Handle release by key: %s", key)
	namespace, name, err := cache.SplitMetaNamespaceKey(key.(string))
	if err != nil {
		glog.Errorf("Can't recognize key of release: %s", key)
		return false
	}
	release, err := rc.releaseLister.Releases(namespace).Get(name)
	if err != nil && !errors.IsNotFound(err) {
		glog.Errorf("Can't get release: %s", key)
		return false
	}
	if err != nil {
		// Deleted
		err = rc.manager.Delete(namespace, name)
	} else {
		// Added or Updated
		err = rc.manager.Trigger(release)
	}
	if err != nil {
		// Re-enqueue
		rc.queue.AddRateLimited(key)
		glog.Errorf("Can't handle release: %+v", release)
		return false
	}
	glog.V(4).Infof("Handled release: %s", key)
	return true
}