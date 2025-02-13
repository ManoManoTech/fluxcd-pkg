/*
Copyright 2022 The Flux authors

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

package login

import (
	"context"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"

	"github.com/fluxcd/pkg/oci"
	"github.com/fluxcd/pkg/oci/auth/aws"
	"github.com/fluxcd/pkg/oci/auth/azure"
	"github.com/fluxcd/pkg/oci/auth/gcp"
)

// ImageRegistryProvider analyzes the provided image and returns the identified
// container image registry provider.
func ImageRegistryProvider(image string, ref name.Reference) oci.Provider {
	_, _, ok := aws.ParseImage(image)
	if ok {
		return oci.ProviderAWS
	}
	if gcp.ValidHost(ref.Context().RegistryStr()) {
		return oci.ProviderGCP
	}
	if azure.ValidHost(ref.Context().RegistryStr()) {
		return oci.ProviderAzure
	}
	return oci.ProviderGeneric
}

// ProviderOptions contains options for registry provider login.
type ProviderOptions struct {
	// AwsAutoLogin enables automatic attempt to get credentials for images in
	// ECR.
	AwsAutoLogin bool
	// GcpAutoLogin enables automatic attempt to get credentials for images in
	// GCP.
	GcpAutoLogin bool
	// AzureAutoLogin enables automatic attempt to get credentials for images in
	// ACR.
	AzureAutoLogin bool
}

// Manager is a login manager for various registry providers.
type Manager struct {
	ecr *aws.Client
	gcr *gcp.Client
	acr *azure.Client
}

// NewManager initializes a Manager with default registry clients
// configurations.
func NewManager() *Manager {
	return &Manager{
		ecr: aws.NewClient(),
		gcr: gcp.NewClient(),
		acr: azure.NewClient(),
	}
}

// WithECRClient allows overriding the default ECR client.
func (m *Manager) WithECRClient(c *aws.Client) *Manager {
	m.ecr = c
	return m
}

// WithGCRClient allows overriding the default GCR client.
func (m *Manager) WithGCRClient(c *gcp.Client) *Manager {
	m.gcr = c
	return m
}

// WithACRClient allows overriding the default ACR client.
func (m *Manager) WithACRClient(c *azure.Client) *Manager {
	m.acr = c
	return m
}

// Login performs authentication against a registry and returns the
// authentication material. For generic registry provider, it is no-op.
func (m *Manager) Login(ctx context.Context, image string, ref name.Reference, opts ProviderOptions) (authn.Authenticator, error) {
	switch ImageRegistryProvider(image, ref) {
	case oci.ProviderAWS:
		return m.ecr.Login(ctx, opts.AwsAutoLogin, image)
	case oci.ProviderGCP:
		return m.gcr.Login(ctx, opts.GcpAutoLogin, image, ref)
	case oci.ProviderAzure:
		return m.acr.Login(ctx, opts.AzureAutoLogin, image, ref)
	}
	return nil, nil
}
