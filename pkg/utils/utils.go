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

package utils

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ExtractCRDFromReq extracts a CRD from the reconcile request.
// This method knows how to handle errors in the context of the reconciliation loop.
func ExtractCRDFromReq(ctx context.Context, req ctrl.Request, reader client.Reader, logger logr.Logger, output client.Object) (*ctrl.Result, error) {
	err := reader.Get(ctx, req.NamespacedName, output)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected.
			// Return and don't requeue.
			logger.Info("resource not found, object must be deleted")

			return &ctrl.Result{}, nil
		}

		logger.Error(err, "failed to get resource")

		return &ctrl.Result{}, err
	}

	return nil, nil
}
