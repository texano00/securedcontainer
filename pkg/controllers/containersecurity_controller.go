package controllers

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	securityv1alpha1 "github.com/texano00/securedcontainer/pkg/apis/security/v1alpha1"
	"github.com/texano00/securedcontainer/pkg/utils"
)

// ContainerSecurityReconciler reconciles a ContainerSecurity object
type ContainerSecurityReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=security.securedcontainer.io,resources=containersecurities,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.securedcontainer.io,resources=containersecurities/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.securedcontainer.io,resources=containersecurities/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=core,resources=pods;secrets,verbs=get;list;watch

func (r *ContainerSecurityReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Get the ContainerSecurity resource
	var containerSecurity securityv1alpha1.ContainerSecurity
	if err := r.Get(ctx, req.NamespacedName, &containerSecurity); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Reconciling ContainerSecurity",
		"namespace", req.Namespace,
		"name", req.Name,
		"scanInterval", containerSecurity.Spec.ScanInterval)

	// List deployments matching the selector
	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(containerSecurity.Spec.Selector.MatchLabels),
		Namespace:     req.Namespace,
	}); err != nil {
		log.Error(err, "Failed to list deployments")
		return ctrl.Result{}, err
	}

	// Process each deployment
	for _, deployment := range deployments.Items {
		for i, container := range deployment.Spec.Template.Spec.Containers {
			// Scan the container image
			scanResult, err := utils.ScanImage(ctx, container.Image)
			if err != nil {
				log.Error(err, "Failed to scan image", "image", container.Image)
				continue
			}

			// If vulnerabilities found and AutoPatch is enabled
			if len(scanResult.Vulnerabilities) > 0 && containerSecurity.Spec.AutoPatch {
				// Generate new image name
				suffix := containerSecurity.Spec.TagSuffix
				if suffix == "" {
					suffix = "-sc"
				}
				newImage := container.Image + suffix

				// Create patched image
				if err := utils.PatchImage(ctx, container.Image, newImage); err != nil {
					log.Error(err, "Failed to patch image", "image", container.Image)
					continue
				}

				// Update container image in deployment
				deployment.Spec.Template.Spec.Containers[i].Image = newImage
			}
		}

		// Update the deployment
		if err := r.Update(ctx, &deployment); err != nil {
			log.Error(err, "Failed to update deployment", "deployment", deployment.Name)
			continue
		}
	}

	// Update status
	now := metav1.Now()
	containerSecurity.Status.LastScanTime = &now
	if containerSecurity.Spec.AutoPatch {
		containerSecurity.Status.LastUpdateTime = &now
	}

	if err := r.Status().Update(ctx, &containerSecurity); err != nil {
		log.Error(err, "Failed to update ContainerSecurity status")
		return ctrl.Result{}, err
	}

	// Requeue based on scan interval
	return ctrl.Result{
		RequeueAfter: time.Duration(containerSecurity.Spec.ScanInterval) * time.Hour,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ContainerSecurityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.ContainerSecurity{}).
		Complete(r)
}
