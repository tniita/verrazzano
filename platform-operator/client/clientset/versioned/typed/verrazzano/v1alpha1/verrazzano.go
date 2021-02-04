// Copyright (c) 2020, 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	scheme "github.com/verrazzano/verrazzano/platform-operator/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// VerrazzanosGetter has a method to return a VerrazzanoInterface.
// A group's client should implement this interface.
type VerrazzanosGetter interface {
	Verrazzanos(namespace string) VerrazzanoInterface
}

// VerrazzanoInterface has methods to work with Verrazzano resources.
type VerrazzanoInterface interface {
	Create(ctx context.Context, verrazzano *v1alpha1.Verrazzano, opts v1.CreateOptions) (*v1alpha1.Verrazzano, error)
	Update(ctx context.Context, verrazzano *v1alpha1.Verrazzano, opts v1.UpdateOptions) (*v1alpha1.Verrazzano, error)
	UpdateStatus(ctx context.Context, verrazzano *v1alpha1.Verrazzano, opts v1.UpdateOptions) (*v1alpha1.Verrazzano, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Verrazzano, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.VerrazzanoList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Verrazzano, err error)
	VerrazzanoExpansion
}

// verrazzanos implements VerrazzanoInterface
type verrazzanos struct {
	client rest.Interface
	ns     string
}

// newVerrazzanos returns a Verrazzanos
func newVerrazzanos(c *VerrazzanoV1alpha1Client, namespace string) *verrazzanos {
	return &verrazzanos{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the verrazzano, and returns the corresponding verrazzano object, and an error if there is any.
func (c *verrazzanos) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Verrazzano, err error) {
	result = &v1alpha1.Verrazzano{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("verrazzanos").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Verrazzanos that match those selectors.
func (c *verrazzanos) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.VerrazzanoList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.VerrazzanoList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("verrazzanos").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested verrazzanos.
func (c *verrazzanos) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("verrazzanos").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a verrazzano and creates it.  Returns the server's representation of the verrazzano, and an error, if there is any.
func (c *verrazzanos) Create(ctx context.Context, verrazzano *v1alpha1.Verrazzano, opts v1.CreateOptions) (result *v1alpha1.Verrazzano, err error) {
	result = &v1alpha1.Verrazzano{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("verrazzanos").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(verrazzano).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a verrazzano and updates it. Returns the server's representation of the verrazzano, and an error, if there is any.
func (c *verrazzanos) Update(ctx context.Context, verrazzano *v1alpha1.Verrazzano, opts v1.UpdateOptions) (result *v1alpha1.Verrazzano, err error) {
	result = &v1alpha1.Verrazzano{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("verrazzanos").
		Name(verrazzano.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(verrazzano).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *verrazzanos) UpdateStatus(ctx context.Context, verrazzano *v1alpha1.Verrazzano, opts v1.UpdateOptions) (result *v1alpha1.Verrazzano, err error) {
	result = &v1alpha1.Verrazzano{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("verrazzanos").
		Name(verrazzano.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(verrazzano).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the verrazzano and deletes it. Returns an error if one occurs.
func (c *verrazzanos) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("verrazzanos").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *verrazzanos) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("verrazzanos").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched verrazzano.
func (c *verrazzanos) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Verrazzano, err error) {
	result = &v1alpha1.Verrazzano{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("verrazzanos").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
