/*
Copyright 2021 Progress Software Corporation.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"

	e "github.com/NativeChat/consul-merge-controller/pkg/errors"
	"github.com/NativeChat/consul-merge-controller/pkg/utils"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type crdService struct {
	reader           client.Reader
	writer           client.Writer
	log              logr.Logger
	finalizer        string
	resourceType     reflect.Type
	resourceListType reflect.Type
}

func (c *crdService) GetResourceFromRequest(ctx context.Context, req ctrl.Request) (client.Object, *ctrl.Result, error) {
	resource := reflect.New(c.resourceType).Interface().(client.Object)
	res, err := utils.ExtractCRDFromReq(ctx, req, c.reader, c.log, resource)
	if err != nil || res != nil {
		return nil, res, err
	}

	return resource, nil, nil
}

func (c *crdService) GetAllResourcesForService(ctx context.Context, label, serviceName, namespace string) ([]client.Object, error) {
	resourceListReflectValue := reflect.New(c.resourceListType)
	listItemsReflectValue := resourceListReflectValue.Elem().FieldByName("Items")

	resourceObjectList := resourceListReflectValue.Interface().(client.ObjectList)

	requirement, err := labels.NewRequirement(label, selection.Equals, []string{serviceName})
	if err != nil {
		return nil, e.NewReconcileError(err, false)
	}

	err = c.reader.List(ctx, resourceObjectList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.Everything().Add(*requirement),
	})

	if err != nil {
		return nil, err
	}

	notMarkedForDeletion := &[]client.Object{}
	notMarkedForDeletionReflectValue := reflect.ValueOf(notMarkedForDeletion).Elem()
	for i := 0; i < listItemsReflectValue.Len(); i++ {
		item := listItemsReflectValue.Index(i)
		if !c.IsDeleted(item.Addr().Interface().(client.Object)) {
			notMarkedForDeletionReflectValue.Set(reflect.Append(notMarkedForDeletionReflectValue, item.Addr()))
		}
	}

	return *notMarkedForDeletion, nil
}

func (c *crdService) UpdateFinalizer(ctx context.Context, obj client.Object) error {
	var err error = nil
	containsFinalizer := controllerutil.ContainsFinalizer(obj, c.finalizer)

	if c.IsDeleted(obj) {
		if containsFinalizer {
			c.log.Info("removing finalizer")
			controllerutil.RemoveFinalizer(obj, c.finalizer)

			err = c.writer.Update(ctx, obj)
			if err == nil {
				c.log.Info("finalizer removed")
			}
		}
	} else if !containsFinalizer {
		c.log.Info("adding finalizer")
		controllerutil.AddFinalizer(obj, c.finalizer)

		err = c.writer.Update(ctx, obj)
		if err == nil {
			c.log.Info("finalizer added")
		}
	}

	return err
}

func (c *crdService) IsDeleted(obj client.Object) bool {
	isDeleted := obj.GetDeletionTimestamp() != nil

	return isDeleted
}

func (c *crdService) IsNew(obj client.Object) bool {
	isNew := c.getCurrentContentSHA(obj) == ""

	return isNew
}

func (c *crdService) IsChanged(obj client.Object) bool {
	isChanged := c.getCurrentContentSHA(obj) != c.GetContentSHA(obj)

	return isChanged
}

func (c *crdService) GetContentSHA(obj client.Object) string {
	serialized, _ := json.Marshal(c.getSpec(obj).Interface())

	h := sha256.New()

	h.Write(serialized)
	result := fmt.Sprintf("%x", h.Sum(nil))

	return result
}

func (c *crdService) SetUpdatedAt(obj client.Object, updatedAt string) {
	c.getStatus(obj).FieldByName("UpdatedAt").SetString(updatedAt)
}

func (c *crdService) SetContentSHA(obj client.Object, contentSHA string) {
	c.getStatus(obj).FieldByName("ContentSHA").SetString(contentSHA)
}

func (c *crdService) getCurrentContentSHA(obj client.Object) string {
	contentSHA := c.getStatus(obj).FieldByName("ContentSHA").String()

	return contentSHA
}

func (c *crdService) getStatus(obj client.Object) reflect.Value {
	spec := reflect.ValueOf(obj).Elem().FieldByName("Status")

	return spec
}

func (c *crdService) getSpec(obj client.Object) reflect.Value {
	spec := reflect.ValueOf(obj).Elem().FieldByName("Spec")

	return spec
}

// NewCRDService returns new CRD service.
func NewCRDService(
	reader client.Reader,
	writer client.Writer,
	log logr.Logger,
	finalizer string,
	resourceType reflect.Type,
	resourceListType reflect.Type,
) CRDService {
	svc := new(crdService)
	svc.reader = reader
	svc.writer = writer
	svc.log = log
	svc.finalizer = finalizer
	svc.resourceType = resourceType
	svc.resourceListType = resourceListType

	return svc
}
