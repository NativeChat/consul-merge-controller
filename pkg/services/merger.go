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
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type merger struct {
	reader                  client.Reader
	writer                  client.Writer
	log                     logr.Logger
	patchExpectedDefinition func(obj client.Object) client.Object
	mergeIntoPropertyName   string
	mergeItemPropertyName   string
	mergeDestinationType    reflect.Type
}

func (m *merger) Merge(ctx context.Context, destinationResourceName, namespace string, items []client.Object) (*ctrl.Result, error) {
	expected := m.getExpectedDefinition(destinationResourceName, namespace, items)

	actual := reflect.New(m.mergeDestinationType).Interface().(client.Object)
	destinationResourceKind := expected.GetObjectKind().GroupVersionKind().GroupVersion().String()
	err := m.reader.Get(ctx, types.NamespacedName{Namespace: namespace, Name: destinationResourceName}, actual)
	if err != nil {
		if !errors.IsNotFound(err) {
			m.log.Error(err, fmt.Sprintf("failed to get %s", destinationResourceKind))

			return &ctrl.Result{Requeue: true}, nil
		}

		m.log.Info(fmt.Sprintf("creating expected resource %s...", destinationResourceKind))

		err = m.writer.Create(ctx, expected)
		if err != nil {
			m.log.Error(err, fmt.Sprintf("failed to create %s", destinationResourceKind))

			return &ctrl.Result{Requeue: true}, nil
		}

		m.log.Info(fmt.Sprintf("%s created", destinationResourceKind))
		return nil, nil
	}

	expectedSpec := m.getSpec(expected)
	actualSpec := m.getSpec(actual)

	if reflect.DeepEqual(expectedSpec.Interface(), actualSpec.Interface()) {
		m.log.Info(fmt.Sprintf("%s is up to date", destinationResourceKind))

		return nil, nil
	}

	if m.getMergeDestinationProp(expected).Len() == 0 {
		m.log.Info(fmt.Sprintf("no %s left for %s, it will be deleted", m.mergeIntoPropertyName, destinationResourceKind))

		err = m.writer.Delete(ctx, actual)
		if err != nil {
			m.log.Error(err, fmt.Sprintf("failed to delete %s", destinationResourceKind))

			return &ctrl.Result{}, err
		}

		m.log.Info(fmt.Sprintf("successfully deleted %s", destinationResourceKind))

		return nil, nil
	}

	m.log.Info(fmt.Sprintf("updating %s...", destinationResourceKind))

	actualSpec.Set(expectedSpec)
	actual.SetOwnerReferences(expected.GetOwnerReferences())
	err = m.writer.Update(ctx, actual)
	if err != nil {
		m.log.Error(err, fmt.Sprintf("failed to update %s", destinationResourceKind))

		return &ctrl.Result{}, err
	}

	m.log.Info(fmt.Sprintf("%s updated", destinationResourceKind))

	return nil, nil
}

func (m *merger) getSpec(obj client.Object) reflect.Value {
	spec := reflect.ValueOf(obj).Elem().FieldByName("Spec")

	return spec
}

func (m *merger) getMergeDestinationProp(obj client.Object) reflect.Value {
	destination := m.getSpec(obj).FieldByName(m.mergeIntoPropertyName)

	return destination
}

func (m *merger) getExpectedDefinition(destinationResourceName, namespace string, items []client.Object) client.Object {
	expectedReflectValue := reflect.New(m.mergeDestinationType)
	expected := expectedReflectValue.Interface().(client.Object)

	expected.SetName(destinationResourceName)
	expected.SetNamespace(namespace)

	mergeDestinationProp := m.getMergeDestinationProp(expected)
	for _, item := range items {
		mergeDestinationProp.Set(reflect.Append(mergeDestinationProp, m.getSpec(item).FieldByName(m.mergeItemPropertyName)))

		ownerReference := metav1.OwnerReference{
			APIVersion: item.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			Kind:       item.GetObjectKind().GroupVersionKind().Kind,
			Name:       item.GetName(),
			UID:        item.GetUID(),
		}

		expected.SetOwnerReferences(append(expected.GetOwnerReferences(), ownerReference))
	}

	if m.patchExpectedDefinition != nil {
		expected = m.patchExpectedDefinition(expected)
	}

	return expected
}

// NewMerger creates new merger instance.
func NewMerger(
	reader client.Reader,
	writer client.Writer,
	log logr.Logger,
	patchExpectedDefinition func(obj client.Object) client.Object,
	mergeIntoPropertyName string,
	mergeItemPropertyName string,
	mergeDestinationType reflect.Type,
) Merger {
	m := new(merger)
	m.reader = reader
	m.writer = writer
	m.log = log
	m.mergeIntoPropertyName = mergeIntoPropertyName
	m.mergeItemPropertyName = mergeItemPropertyName
	m.mergeDestinationType = mergeDestinationType
	m.patchExpectedDefinition = patchExpectedDefinition

	return m
}
