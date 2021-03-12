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

package testutils

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getK8sObject(ctx context.Context, k8sClient client.Client, name string, obj client.Object) (exists bool, err error) {
	err = k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: DefaultK8sNamespace}, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			obj = nil
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func deleteK8sObject(ctx context.Context, k8sClient client.Client, name string, obj client.Object) error {
	exists, err := getK8sObject(ctx, k8sClient, name, obj)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	clientObject := obj.(client.Object)
	err = k8sClient.Delete(ctx, clientObject)
	if err != nil {
		return err
	}

	hasTimedOut := retryWithSleep(func() bool {
		exists, err = getK8sObject(ctx, k8sClient, name, obj)
		if !exists && err == nil {
			return true
		}

		return false
	})

	if hasTimedOut {
		errArgs := []interface{}{
			reflect.TypeOf(clientObject).String(),
			clientObject.GetNamespace(),
			clientObject.GetName(),
			clientObject.GetFinalizers(),
		}

		return fmt.Errorf("delete timeout for [%s] %s/%s exceeded, finalizers are %s", errArgs...)
	}

	return nil
}

func getResourceContentSHA(resource interface{}) string {
	serialized, _ := json.Marshal(reflect.ValueOf(resource).Elem().FieldByName("Spec").Interface())

	h := sha256.New()

	h.Write(serialized)
	result := fmt.Sprintf("%x", h.Sum(nil))

	return result
}

func retryWithSleep(action func() bool) bool {
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)

		shouldStop := action()
		if shouldStop {
			return false
		}
	}

	return true
}
