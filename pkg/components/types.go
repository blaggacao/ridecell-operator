/*
Copyright 2018 Ridecell, Inc..

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

package components

import (
	"context"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// // A ComponentController is the data for a type of controller.
// type ComponentController struct {
// 	client.Client
// 	Scheme     *runtime.Scheme
// 	Top        runtime.Object
// 	Templates  http.FileSystem
// 	Components []Component
// }

type componentReconciler struct {
	top        runtime.Object
	templates  http.FileSystem
	components []Component
	client     client.Client
	manager    manager.Manager
}

// A ComponentContext is the state for a single reconcile request to the controller.
type ComponentContext struct {
	client.Client
	reconciler *componentReconciler
	Context    context.Context
	Top        runtime.Object
	Scheme     *runtime.Scheme
}

// A component is a Promise Theory actor inside a controller.
type Component interface {
	WatchTypes() []runtime.Object
	IsReconcilable(*ComponentContext) bool
	Reconcile(*ComponentContext) (reconcile.Result, error)
}

type Status interface{}

type Statuser interface {
	GetStatus() Status
	SetStatus(Status)
	SetErrorStatus(string)
}
