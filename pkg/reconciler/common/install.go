/*
Copyright 2020 The Knative Authors

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

package common

import (
	"fmt"

	mf "github.com/manifestival/manifestival"
	"knative.dev/operator/pkg/apis/operator/v1alpha1"
)

var (
	role        mf.Predicate = mf.Any(mf.ByKind("ClusterRole"), mf.ByKind("Role"))
	rolebinding mf.Predicate = mf.Any(mf.ByKind("ClusterRoleBinding"), mf.ByKind("RoleBinding"))
)

// Install applies the manifest resources for the given version and updates the given
// status accordingly.
func Install(manifest *mf.Manifest, version string, status v1alpha1.KComponentStatus) error {
	// The Operator needs a higher level of permissions if it 'bind's non-existent roles.
	// To avoid this, we strictly order the manifest application as (Cluster)Roles, then
	// (Cluster)RoleBindings, then the rest of the manifest.
	if err := manifest.Filter(role).Apply(); err != nil {
		status.MarkInstallFailed(err.Error())
		return fmt.Errorf("failed to apply (cluster)roles: %w", err)
	}
	if err := manifest.Filter(rolebinding).Apply(); err != nil {
		status.MarkInstallFailed(err.Error())
		return fmt.Errorf("failed to apply (cluster)rolebindings: %w", err)
	}
	if err := manifest.Filter(mf.None(role, rolebinding)).Apply(); err != nil {
		status.MarkInstallFailed(err.Error())
		return fmt.Errorf("failed to apply non rbac manifest: %w", err)
	}
	status.MarkInstallSucceeded()
	status.SetVersion(version)
	return nil
}
