package applicationsets

import (
	"testing"

	"github.com/argoproj-labs/applicationset/api/v1alpha1"
	. "github.com/argoproj-labs/applicationset/test/e2e/fixture/applicationsets"
	"github.com/argoproj-labs/applicationset/test/e2e/fixture/applicationsets/utils"
	argov1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSimpleListGenerator(t *testing.T) {

	expectedApp := argov1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "argoproj.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "my-cluster-guestbook",
			Namespace:  utils.ArgoCDNamespace,
			Finalizers: []string{"resources-finalizer.argocd.argoproj.io"},
		},
		Spec: argov1alpha1.ApplicationSpec{
			Project: "default",
			Source: argov1alpha1.ApplicationSource{
				RepoURL:        "https://github.com/argoproj/argocd-example-apps.git",
				TargetRevision: "HEAD",
				Path:           "guestbook",
			},
			Destination: argov1alpha1.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: "guestbook",
			},
		},
	}
	var expectedAppNewNamespace *argov1alpha1.Application
	var expectedAppNewMetadata *argov1alpha1.Application

	Given(t).
		// Create a ListGenerator-based ApplicationSet
		When().Create(v1alpha1.ApplicationSet{ObjectMeta: metav1.ObjectMeta{
		Name: "simple-list-generator",
	},
		Spec: v1alpha1.ApplicationSetSpec{
			Template: v1alpha1.ApplicationSetTemplate{
				ApplicationSetTemplateMeta: v1alpha1.ApplicationSetTemplateMeta{Name: "{{cluster}}-guestbook"},
				Spec: argov1alpha1.ApplicationSpec{
					Project: "default",
					Source: argov1alpha1.ApplicationSource{
						RepoURL:        "https://github.com/argoproj/argocd-example-apps.git",
						TargetRevision: "HEAD",
						Path:           "guestbook",
					},
					Destination: argov1alpha1.ApplicationDestination{
						Server:    "{{url}}",
						Namespace: "guestbook",
					},
				},
			},
			Generators: []v1alpha1.ApplicationSetGenerator{
				{
					List: &v1alpha1.ListGenerator{
						Elements: []v1alpha1.ListGeneratorElement{
							{Cluster: "my-cluster", Url: "https://kubernetes.default.svc"},
						},
					},
				},
			},
		},
	}).Then().Expect(ApplicationsExist([]argov1alpha1.Application{expectedApp})).

		// Update the ApplicationSet template namespace, and verify it updates the Applications
		When().
		And(func() {
			expectedAppNewNamespace = expectedApp.DeepCopy()
			expectedAppNewNamespace.Spec.Destination.Namespace = "guestbook2"
		}).
		Update(func(appset *v1alpha1.ApplicationSet) {
			appset.Spec.Template.Spec.Destination.Namespace = "guestbook2"
		}).Then().Expect(ApplicationsExist([]argov1alpha1.Application{*expectedAppNewNamespace})).

		// Update the metadata fields in the appset template, and make sure it propagates to the apps
		When().
		And(func() {
			expectedAppNewMetadata = expectedAppNewNamespace.DeepCopy()
			expectedAppNewMetadata.ObjectMeta.Annotations = map[string]string{"annotation-key": "annotation-value"}
			expectedAppNewMetadata.ObjectMeta.Labels = map[string]string{"label-key": "label-value"}
		}).
		Update(func(appset *v1alpha1.ApplicationSet) {
			appset.Spec.Template.Annotations = map[string]string{"annotation-key": "annotation-value"}
			appset.Spec.Template.Labels = map[string]string{"label-key": "label-value"}
		}).Then().Expect(ApplicationsExist([]argov1alpha1.Application{*expectedAppNewMetadata})).

		// Delete the ApplicationSet, and verify it deletes the Applications
		When().
		Delete().Then().Expect(ApplicationsDoNotExist([]argov1alpha1.Application{*expectedAppNewMetadata}))

}

func TestSimpleGitDirectoryGenerator(t *testing.T) {
	generateExpectedApp := func(name string) argov1alpha1.Application {
		return argov1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "argoproj.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:       name,
				Namespace:  utils.ArgoCDNamespace,
				Finalizers: []string{"resources-finalizer.argocd.argoproj.io"},
			},
			Spec: argov1alpha1.ApplicationSpec{
				Project: "default",
				Source: argov1alpha1.ApplicationSource{
					RepoURL:        "https://github.com/argoproj/argocd-example-apps.git",
					TargetRevision: "HEAD",
					Path:           name,
				},
				Destination: argov1alpha1.ApplicationDestination{
					Server:    "https://kubernetes.default.svc",
					Namespace: name,
				},
			},
		}
	}

	expectedApps := []argov1alpha1.Application{
		generateExpectedApp("kustomize-guestbook"),
		generateExpectedApp("helm-guestbook"),
		generateExpectedApp("ksonnet-guestbook"),
	}

	var expectedAppsNewNamespace []argov1alpha1.Application
	var expectedAppsNewMetadata []argov1alpha1.Application

	Given(t).
		When().
		// Create a GitGenerator-based ApplicationSet
		Create(v1alpha1.ApplicationSet{ObjectMeta: metav1.ObjectMeta{
			Name: "simple-git-generator",
		},
			Spec: v1alpha1.ApplicationSetSpec{
				Template: v1alpha1.ApplicationSetTemplate{
					ApplicationSetTemplateMeta: v1alpha1.ApplicationSetTemplateMeta{Name: "{{path.basename}}"},
					Spec: argov1alpha1.ApplicationSpec{
						Project: "default",
						Source: argov1alpha1.ApplicationSource{
							RepoURL:        "https://github.com/argoproj/argocd-example-apps.git",
							TargetRevision: "HEAD",
							Path:           "{{path}}",
						},
						Destination: argov1alpha1.ApplicationDestination{
							Server:    "https://kubernetes.default.svc",
							Namespace: "{{path.basename}}",
						},
					},
				},
				Generators: []v1alpha1.ApplicationSetGenerator{
					{
						Git: &v1alpha1.GitGenerator{
							RepoURL: "https://github.com/argoproj/argocd-example-apps.git",
							Directories: []v1alpha1.GitDirectoryGeneratorItem{
								{
									Path: "*guestbook*",
								},
							},
						},
					},
				},
			},
		}).Then().Expect(ApplicationsExist(expectedApps)).

		// Update the ApplicationSet template namespace, and verify it updates the Applications
		When().
		And(func() {
			for _, expectedApp := range expectedApps {
				newExpectedApp := expectedApp.DeepCopy()
				newExpectedApp.Spec.Destination.Namespace = "guestbook2"
				expectedAppsNewNamespace = append(expectedAppsNewNamespace, *newExpectedApp)
			}
		}).
		Update(func(appset *v1alpha1.ApplicationSet) {
			appset.Spec.Template.Spec.Destination.Namespace = "guestbook2"
		}).Then().Expect(ApplicationsExist(expectedAppsNewNamespace)).

		// Update the metadata fields in the appset template, and make sure it propagates to the apps
		When().
		And(func() {
			for _, expectedApp := range expectedAppsNewNamespace {
				expectedAppNewMetadata := expectedApp.DeepCopy()
				expectedAppNewMetadata.ObjectMeta.Annotations = map[string]string{"annotation-key": "annotation-value"}
				expectedAppNewMetadata.ObjectMeta.Labels = map[string]string{"label-key": "label-value"}
				expectedAppsNewMetadata = append(expectedAppsNewMetadata, *expectedAppNewMetadata)
			}
		}).
		Update(func(appset *v1alpha1.ApplicationSet) {
			appset.Spec.Template.Annotations = map[string]string{"annotation-key": "annotation-value"}
			appset.Spec.Template.Labels = map[string]string{"label-key": "label-value"}
		}).Then().Expect(ApplicationsExist(expectedAppsNewMetadata)).

		// Delete the ApplicationSet, and verify it deletes the Applications
		When().
		Delete().Then().Expect(ApplicationsDoNotExist(expectedAppsNewNamespace))
}

func TestSimpleGitFilesGenerator(t *testing.T) {

	generateExpectedApp := func(name string) argov1alpha1.Application {
		return argov1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "argoproj.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:       name,
				Namespace:  utils.ArgoCDNamespace,
				Finalizers: []string{"resources-finalizer.argocd.argoproj.io"},
			},
			Spec: argov1alpha1.ApplicationSpec{
				Project: "default",
				Source: argov1alpha1.ApplicationSource{
					RepoURL:        "https://github.com/argoproj/argocd-example-apps.git",
					TargetRevision: "HEAD",
					Path:           "guestbook",
				},
				Destination: argov1alpha1.ApplicationDestination{
					Server:    "https://kubernetes.default.svc",
					Namespace: "guestbook",
				},
			},
		}
	}

	expectedApps := []argov1alpha1.Application{
		generateExpectedApp("engineering-dev-guestbook"),
		generateExpectedApp("engineering-prod-guestbook"),
	}

	var expectedAppsNewNamespace []argov1alpha1.Application
	var expectedAppsNewMetadata []argov1alpha1.Application

	Given(t).
		When().
		// Create a GitGenerator-based ApplicationSet
		Create(v1alpha1.ApplicationSet{ObjectMeta: metav1.ObjectMeta{
			Name: "simple-git-generator",
		},
			Spec: v1alpha1.ApplicationSetSpec{
				Template: v1alpha1.ApplicationSetTemplate{
					ApplicationSetTemplateMeta: v1alpha1.ApplicationSetTemplateMeta{Name: "{{cluster.name}}-guestbook"},
					Spec: argov1alpha1.ApplicationSpec{
						Project: "default",
						Source: argov1alpha1.ApplicationSource{
							RepoURL:        "https://github.com/argoproj/argocd-example-apps.git",
							TargetRevision: "HEAD",
							Path:           "guestbook",
						},
						Destination: argov1alpha1.ApplicationDestination{
							Server:    "https://kubernetes.default.svc",
							Namespace: "guestbook",
						},
					},
				},
				Generators: []v1alpha1.ApplicationSetGenerator{
					{
						Git: &v1alpha1.GitGenerator{
							RepoURL: "https://github.com/argoproj-labs/applicationset.git",
							Files: []v1alpha1.GitFileGeneratorItem{
								{
									Path: "examples/git-generator-files-discovery/cluster-config/**/config.json",
								},
							},
						},
					},
				},
			},
		}).Then().Expect(ApplicationsExist(expectedApps)).

		// Update the ApplicationSet template namespace, and verify it updates the Applications
		When().
		And(func() {
			for _, expectedApp := range expectedApps {
				newExpectedApp := expectedApp.DeepCopy()
				newExpectedApp.Spec.Destination.Namespace = "guestbook2"
				expectedAppsNewNamespace = append(expectedAppsNewNamespace, *newExpectedApp)
			}
		}).
		Update(func(appset *v1alpha1.ApplicationSet) {
			appset.Spec.Template.Spec.Destination.Namespace = "guestbook2"
		}).Then().Expect(ApplicationsExist(expectedAppsNewNamespace)).

		// Update the metadata fields in the appset template, and make sure it propagates to the apps
		When().
		And(func() {
			for _, expectedApp := range expectedAppsNewNamespace {
				expectedAppNewMetadata := expectedApp.DeepCopy()
				expectedAppNewMetadata.ObjectMeta.Annotations = map[string]string{"annotation-key": "annotation-value"}
				expectedAppNewMetadata.ObjectMeta.Labels = map[string]string{"label-key": "label-value"}
				expectedAppsNewMetadata = append(expectedAppsNewMetadata, *expectedAppNewMetadata)
			}
		}).
		Update(func(appset *v1alpha1.ApplicationSet) {
			appset.Spec.Template.Annotations = map[string]string{"annotation-key": "annotation-value"}
			appset.Spec.Template.Labels = map[string]string{"label-key": "label-value"}
		}).Then().Expect(ApplicationsExist(expectedAppsNewMetadata)).

		// Delete the ApplicationSet, and verify it deletes the Applications
		When().
		Delete().Then().Expect(ApplicationsDoNotExist(expectedAppsNewNamespace))
}
