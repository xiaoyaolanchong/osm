package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/onsi/gomega/gstruct"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	fake "k8s.io/client-go/kubernetes/fake"

	"github.com/openservicemesh/osm/pkg/constants"
)

var (
	testNamespace     = "namespace"
	testMeshName      = "mesh"
	incorrectMeshName = "incorrectMesh"
)

var _ = Describe("Running the namespace add command", func() {

	Describe("with pre-existing namespace", func() {
		var (
			out           *bytes.Buffer
			fakeClientSet kubernetes.Interface
			err           error
		)

		Context("given one namespace as an arg without sidecar injection enabled", func() {

			BeforeEach(func() {
				out = new(bytes.Buffer)
				fakeClientSet = fake.NewSimpleClientset()

				nsSpec := createNamespaceSpec(testNamespace, "", false)
				fakeClientSet.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metav1.CreateOptions{})

				namespaceAddCmd := &namespaceAddCmd{
					out:        out,
					meshName:   testMeshName,
					namespaces: []string{testNamespace},
					clientSet:  fakeClientSet,
				}

				err = namespaceAddCmd.run()
			})

			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should give a message confirming the successful install", func() {
				Expect(out.String()).To(Equal(fmt.Sprintf("Namespace [%s] successfully added to mesh [%s]\n", testNamespace, testMeshName)))
			})

			It("should correctly add a label to the namespace", func() {
				ns, err := fakeClientSet.CoreV1().Namespaces().Get(context.TODO(), testNamespace, metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(ns.Labels[constants.OSMKubeResourceMonitorAnnotation]).To(Equal(testMeshName))
			})

			It("should correctly not add an inject annotation to the namespace", func() {
				ns, err := fakeClientSet.CoreV1().Namespaces().Get(context.TODO(), testNamespace, metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(ns.Annotations).ShouldNot(HaveKey(constants.SidecarInjectionAnnotation))
			})
		})

		Context("given one namespace as an arg with sidecar injection enabled", func() {

			BeforeEach(func() {
				out = new(bytes.Buffer)
				fakeClientSet = fake.NewSimpleClientset()

				nsSpec := createNamespaceSpec(testNamespace, "", false)
				_, err = fakeClientSet.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metav1.CreateOptions{})
				Expect(err).ToNot(HaveOccurred())

				namespaceAddCmd := &namespaceAddCmd{
					out:                    out,
					meshName:               testMeshName,
					namespaces:             []string{testNamespace},
					enableSidecarInjection: true,
					clientSet:              fakeClientSet,
				}

				err = namespaceAddCmd.run()
			})

			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should give a message confirming the successful install", func() {
				Expect(out.String()).To(Equal(fmt.Sprintf("Namespace [%s] successfully added to mesh [%s]\n", testNamespace, testMeshName)))
			})

			It("should correctly add a label to the namespace", func() {
				ns, err := fakeClientSet.CoreV1().Namespaces().Get(context.TODO(), testNamespace, metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(ns.Labels[constants.OSMKubeResourceMonitorAnnotation]).To(Equal(testMeshName))
			})

			It("should correctly add an inject annotation to the namespace", func() {
				ns, err := fakeClientSet.CoreV1().Namespaces().Get(context.TODO(), testNamespace, metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(ns.Annotations[constants.SidecarInjectionAnnotation]).To(Equal("enabled"))
			})
		})

		Context("given two namespaces as args", func() {

			var (
				testNamespace2 string
			)

			BeforeEach(func() {
				out = new(bytes.Buffer)
				fakeClientSet = fake.NewSimpleClientset()
				testNamespace2 = "namespace2"

				nsSpec := createNamespaceSpec(testNamespace, "", false)
				fakeClientSet.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metav1.CreateOptions{})

				nsSpec2 := createNamespaceSpec(testNamespace2, "", false)
				fakeClientSet.CoreV1().Namespaces().Create(context.TODO(), nsSpec2, metav1.CreateOptions{})

				namespaceAddCmd := &namespaceAddCmd{
					out:        out,
					meshName:   testMeshName,
					namespaces: []string{testNamespace, testNamespace2},
					clientSet:  fakeClientSet,
				}

				err = namespaceAddCmd.run()
			})

			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should give a message confirming the successful install", func() {
				Expect(out.String()).To(Equal(fmt.Sprintf("Namespace [%s] successfully added to mesh [%s]\nNamespace [%s] successfully added to mesh [%s]\n", testNamespace, testMeshName, testNamespace2, testMeshName)))
			})

			It("should correctly add a label to all the namespaces", func() {
				ns1, err := fakeClientSet.CoreV1().Namespaces().Get(context.TODO(), testNamespace, metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(ns1.Labels[constants.OSMKubeResourceMonitorAnnotation]).To(Equal(testMeshName))

				ns2, err := fakeClientSet.CoreV1().Namespaces().Get(context.TODO(), testNamespace2, metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(ns2.Labels[constants.OSMKubeResourceMonitorAnnotation]).To(Equal(testMeshName))
			})
		})

		Context("given one namespace with osm-controller installed in it as an arg", func() {
			BeforeEach(func() {
				out = new(bytes.Buffer)
				fakeClientSet = fake.NewSimpleClientset()
				// mimic osm controller deployment in testNamespace
				deploymentSpec := createDeploymentSpec(testNamespace, defaultMeshName)
				fakeClientSet.AppsV1().Deployments(testNamespace).Create(context.TODO(), deploymentSpec, metav1.CreateOptions{})

				nsSpec := createNamespaceSpec(testNamespace, "", false)
				fakeClientSet.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metav1.CreateOptions{})

				namespaceAddCmd := &namespaceAddCmd{
					out:        out,
					meshName:   testMeshName,
					namespaces: []string{testNamespace},
					clientSet:  fakeClientSet,
				}

				err = namespaceAddCmd.run()
			})
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should give a warning message", func() {
				Expect(out.String()).To(Equal(fmt.Sprintf("Namespace [%s] already has [%s] installed and cannot be added to mesh [%s]\n", testNamespace, constants.OSMControllerName, testMeshName)))
			})
		})
	})

	Describe("with non-existent namespace", func() {
		var (
			out           *bytes.Buffer
			fakeClientSet kubernetes.Interface
			err           error
		)

		BeforeEach(func() {
			out = new(bytes.Buffer)
			fakeClientSet = fake.NewSimpleClientset()

			namespaceAddCmd := &namespaceAddCmd{
				out:        out,
				meshName:   testMeshName,
				namespaces: []string{testNamespace},
				clientSet:  fakeClientSet,
			}

			err = namespaceAddCmd.run()
		})

		It("should error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Could not add namespace [%s] to mesh [%s]: namespaces \"%s\" not found", testNamespace, testMeshName, testNamespace)))
		})
	})
})

var _ = Describe("Running the namespace remove command", func() {

	Describe("with pre-existing namespace and correct label", func() {
		var (
			out           *bytes.Buffer
			fakeClientSet kubernetes.Interface
			err           error
		)

		BeforeEach(func() {
			out = new(bytes.Buffer)
			fakeClientSet = fake.NewSimpleClientset()

			nsSpec := createNamespaceSpec(testNamespace, testMeshName, false)
			fakeClientSet.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metav1.CreateOptions{})

			namespaceRemoveCmd := &namespaceRemoveCmd{
				out:       out,
				meshName:  testMeshName,
				namespace: testNamespace,
				clientSet: fakeClientSet,
			}

			err = namespaceRemoveCmd.run()
		})

		It("should not error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should give a message confirming the successful install", func() {
			Expect(out.String()).To(Equal(fmt.Sprintf("Namespace [%s] successfully removed from mesh [%s]\n", testNamespace, testMeshName)))
		})

		It("should correctly remove the label on the namespace", func() {
			ns, err := fakeClientSet.CoreV1().Namespaces().Get(context.TODO(), testNamespace, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(ns.Labels).ShouldNot(HaveKey(constants.OSMKubeResourceMonitorAnnotation))
		})
	})

	Describe("with pre-existing namespace, correct label and annotation", func() {
		var (
			out           *bytes.Buffer
			fakeClientSet kubernetes.Interface
			err           error
		)

		BeforeEach(func() {
			out = new(bytes.Buffer)
			fakeClientSet = fake.NewSimpleClientset()

			nsSpec := createNamespaceSpec(testNamespace, testMeshName, true)
			_, err = fakeClientSet.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())

			namespaceRemoveCmd := &namespaceRemoveCmd{
				out:       out,
				meshName:  testMeshName,
				namespace: testNamespace,
				clientSet: fakeClientSet,
			}

			err = namespaceRemoveCmd.run()
		})

		It("should not error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should give a message confirming the successful install", func() {
			Expect(out.String()).To(Equal(fmt.Sprintf("Namespace [%s] successfully removed from mesh [%s]\n", testNamespace, testMeshName)))
		})

		It("should correctly remove the label on the namespace", func() {
			ns, err := fakeClientSet.CoreV1().Namespaces().Get(context.TODO(), testNamespace, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(ns.Labels).ShouldNot(HaveKey(constants.OSMKubeResourceMonitorAnnotation))
		})

		It("should correctly remove the inject annotation on the namespace", func() {
			ns, err := fakeClientSet.CoreV1().Namespaces().Get(context.TODO(), testNamespace, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(ns.Annotations).ShouldNot(HaveKey(constants.SidecarInjectionAnnotation))
		})
	})

	Describe("with pre-existing namespace and incorrect label", func() {
		var (
			out           *bytes.Buffer
			fakeClientSet kubernetes.Interface
			err           error
		)

		BeforeEach(func() {
			out = new(bytes.Buffer)
			fakeClientSet = fake.NewSimpleClientset()

			nsSpec := createNamespaceSpec(testNamespace, testMeshName, false)
			fakeClientSet.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metav1.CreateOptions{})

			namespaceRemoveCmd := &namespaceRemoveCmd{
				out:       out,
				meshName:  incorrectMeshName,
				namespace: testNamespace,
				clientSet: fakeClientSet,
			}

			err = namespaceRemoveCmd.run()
		})

		It("should error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Namespace belongs to mesh [%s], not mesh [%s]. Please specify the correct mesh", testMeshName, incorrectMeshName)))

		})
	})

	Describe("with pre-existing namespace and no label", func() {
		var (
			out           *bytes.Buffer
			fakeClientSet kubernetes.Interface
			err           error
		)

		BeforeEach(func() {
			out = new(bytes.Buffer)
			fakeClientSet = fake.NewSimpleClientset()

			nsSpec := createNamespaceSpec(testNamespace, "", false)
			fakeClientSet.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metav1.CreateOptions{})

			namespaceRemoveCmd := &namespaceRemoveCmd{
				out:       out,
				meshName:  testMeshName,
				namespace: testNamespace,
				clientSet: fakeClientSet,
			}

			err = namespaceRemoveCmd.run()
		})

		It("should not error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should give a message saying namespace was not part of mesh", func() {
			Expect(out.String()).To(Equal(fmt.Sprintf("Namespace [%s] already does not belong to any mesh\n", testNamespace)))
		})
	})

	Describe("with non-existent namespace", func() {
		var (
			out           *bytes.Buffer
			fakeClientSet kubernetes.Interface
			err           error
		)

		BeforeEach(func() {
			out = new(bytes.Buffer)
			fakeClientSet = fake.NewSimpleClientset()

			namespaceRemoveCmd := &namespaceRemoveCmd{
				out:       out,
				meshName:  testMeshName,
				namespace: testNamespace,
				clientSet: fakeClientSet,
			}

			err = namespaceRemoveCmd.run()
		})

		It("should error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Could not get namespace [%s]: namespaces \"%s\" not found", testNamespace, testNamespace)))
		})
	})
})

var _ = Describe("Running the namespace list command", func() {

	Describe("with multiple namespaces enlisted", func() {
		var (
			out           *bytes.Buffer
			fakeClientSet kubernetes.Interface
			err           error
			namespaces    *v1.NamespaceList
			listCmd       *namespaceListCmd
		)

		// helper function that adds a name space to the clientset
		addNamespace := func(name, mesh string) {
			ns := createNamespaceSpec(name, mesh, false)
			fakeClientSet.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
		}

		// helper function that takes element from slice and returns the name for gomega struct
		// https://onsi.github.io/gomega/#gstruct-testing-complex-data-types
		idSelector := func(element interface{}) string {
			return string(element.(v1.Namespace).ObjectMeta.Name)
		}

		BeforeEach(func() {
			out = new(bytes.Buffer)
			fakeClientSet = fake.NewSimpleClientset()

			listCmd = &namespaceListCmd{
				out:       out,
				clientSet: fakeClientSet,
			}
		})

		It("should only have namespaces enlisted", func() {
			addNamespace("enlisted1", "mesh1")
			addNamespace("enlisted2", "mesh2")
			addNamespace("not-enlisted", "")

			namespaces, err = listCmd.selectNamespaces()

			Expect(err).NotTo(HaveOccurred())
			Expect(namespaces.Items).To(gstruct.MatchAllElements(idSelector, gstruct.Elements{
				"enlisted1": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"ObjectMeta": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Labels": gstruct.MatchKeys(gstruct.IgnoreMissing, gstruct.Keys{
							constants.OSMKubeResourceMonitorAnnotation: Equal("mesh1"),
						}),
					}),
				}),
				"enlisted2": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"ObjectMeta": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Labels": gstruct.MatchKeys(gstruct.IgnoreMissing, gstruct.Keys{
							constants.OSMKubeResourceMonitorAnnotation: Equal("mesh2"),
						}),
					}),
				}),
			}))
		})

		It("should only have namespaces from mesh requested", func() {
			addNamespace("enlisted1", "mesh1")
			addNamespace("enlisted2", "mesh2")
			listCmd.meshName = "mesh2"

			namespaces, err = listCmd.selectNamespaces()

			Expect(err).NotTo(HaveOccurred())
			Expect(namespaces.Items).To(gstruct.MatchAllElements(idSelector, gstruct.Elements{
				"enlisted2": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"ObjectMeta": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Labels": gstruct.MatchKeys(gstruct.IgnoreMissing, gstruct.Keys{
							constants.OSMKubeResourceMonitorAnnotation: Equal("mesh2"),
						}),
					}),
				}),
			}))
		})

		It("should have empty list if mesh doesn't have any namespaces assigned", func() {
			addNamespace("enlisted1", "mesh1")
			addNamespace("enlisted2", "mesh2")

			listCmd.meshName = "someothermesh"

			namespaces, err = listCmd.selectNamespaces()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(namespaces.Items)).To(Equal(0))
		})

		It("should print no namespaces message if requested mesh doesn't have any namespaces assigned", func() {
			addNamespace("enlisted1", "mesh1")
			addNamespace("enlisted2", "mesh2")

			listCmd.meshName = "someothermesh"
			err = listCmd.run()
			Expect(err).NotTo(HaveOccurred())
			Expect(out.String()).To(Equal(fmt.Sprintf("No namespaces in mesh [%s]\n", "someothermesh")))
		})

		It("should print no namespaces message if there are no namespaces assigned", func() {
			addNamespace("not-enlisted", "")

			err = listCmd.run()
			Expect(err).NotTo(HaveOccurred())
			Expect(out.String()).To(Equal(fmt.Sprintf("No namespaces in any mesh\n")))
		})

		It("should print no namespaces message if there are no namespaces", func() {
			err = listCmd.run()
			Expect(err).NotTo(HaveOccurred())
			Expect(out.String()).To(Equal(fmt.Sprintf("No namespaces in any mesh\n")))
		})
	})
})

func createNamespaceSpec(namespace, meshName string, enableSideCarInjection bool) *v1.Namespace {
	labelMap := make(map[string]string)
	if meshName != "" {
		labelMap[constants.OSMKubeResourceMonitorAnnotation] = meshName
	}
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespace,
			Labels: labelMap,
		},
	}

	if enableSideCarInjection {
		ns.Annotations = map[string]string{constants.SidecarInjectionAnnotation: "enabled"}
	}

	return ns
}
