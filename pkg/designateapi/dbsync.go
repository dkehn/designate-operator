package designateapi

import (
	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	designatev1beta1 "github.com/openstack-k8s-operators/designate-operator/api/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DbSyncJob func
func DbSyncJob(
	cr *designatev1beta1.DesignateAPI,
	labels map[string]string,
) *batchv1.Job {

	runAsUser := int64(0)
	initVolumeMounts := GetInitVolumeMounts()
	volumeMounts := GetAPIVolumeMounts()
	volumes := GetAPIVolumes(cr.Name)

	envVars := map[string]env.Setter{}
	envVars["KOLLA_CONFIG_FILE"] = env.SetValue(KollaConfigDbSync)
	envVars["KOLLA_CONFIG_STRATEGY"] = env.SetValue("COPY_ALWAYS")

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-db-sync",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy:      "OnFailure",
					ServiceAccountName: ServiceAccountName,
					Containers: []corev1.Container{
						{
							Name:  cr.Name + "-db-sync",
							Image: cr.Spec.ContainerImage,
							SecurityContext: &corev1.SecurityContext{
								RunAsUser: &runAsUser,
							},
							Env:          env.MergeEnvs([]corev1.EnvVar{}, envVars),
							VolumeMounts: volumeMounts,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}
	initContainerDetails := InitContainer{
		ContainerImage:       cr.Spec.ContainerImage,
		DatabaseHost:         cr.Status.DatabaseHostname,
		Database:             Database,
		DesignateSecret:      cr.Spec.Secret,
		DBPasswordSelector:   cr.Spec.PasswordSelectors.Database,
		UserPasswordSelector: cr.Spec.PasswordSelectors.Service,
		NovaPasswordSelector: cr.Spec.PasswordSelectors.NovaService,
		VolumeMounts:         initVolumeMounts,
	}
	job.Spec.Template.Spec.InitContainers = GetInitContainer(initContainerDetails)
	return job
}
