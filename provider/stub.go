package provider

import (
	"sync"
	"time"

	"github.com/google/uuid"
	neon "github.com/kislerdm/neon-sdk-go"
)

type sdkClientStub struct {
	stubProjectPermission
	stubProjectRolePassword
	stubVPCEndpoint
	mockOpsReader

	req interface{}
	err error
}

type stubVPCEndpoint struct {
	VPCEndpointDetails neon.VPCEndpointDetails
	err                error
}

func (s *stubVPCEndpoint) AssignOrganizationVPCEndpoint(_, _, _ string, _ neon.VPCEndpointAssignment) error {
	return s.err
}

func (s *stubVPCEndpoint) GetOrganizationVPCEndpointDetails(_, _, _ string) (neon.VPCEndpointDetails, error) {
	if s.err != nil {
		return neon.VPCEndpointDetails{}, s.err
	}
	return s.VPCEndpointDetails, nil
}

func (s *stubVPCEndpoint) DeleteOrganizationVPCEndpoint(_, _, _ string) error {
	return s.err
}

func (s *sdkClientStub) UpdateProject(_ string, cfg neon.ProjectUpdateRequest) (neon.UpdateProjectRespObj, error) {
	s.req = cfg
	return neon.UpdateProjectRespObj{}, s.err
}

func (s *sdkClientStub) GetProject(_ string) (neon.ProjectResponse, error) {
	return neon.ProjectResponse{}, nil
}

func (s *sdkClientStub) ListProjectBranches(_ string, _ *string, _ *string, _ *string, _ *string, _ *int) (neon.ListProjectBranchesRespObj, error) {
	return neon.ListProjectBranchesRespObj{}, nil
}

func (s *sdkClientStub) ListProjectBranchEndpoints(_ string, _ string) (neon.EndpointsResponse, error) {
	panic("implement me")
}

func (s *sdkClientStub) DeleteProject(_ string) (neon.ProjectResponse, error) {
	panic("implement me")
}

func (s *sdkClientStub) ListProjectBranchDatabases(_ string, _ string) (neon.DatabasesResponse, error) {
	panic("implement me")
}

func (s *sdkClientStub) CreateProject(cfg neon.ProjectCreateRequest) (neon.CreatedProject, error) {
	s.req = cfg
	return neon.CreatedProject{}, s.err
}

type stubProjectRolePassword struct {
	Password string
	err      error
}

func (s *stubProjectRolePassword) GetProjectBranchRolePassword(_ string, _ string, _ string) (neon.RolePasswordResponse, error) {
	if s.err != nil {
		return neon.RolePasswordResponse{}, s.err
	}
	return neon.RolePasswordResponse{Password: s.Password}, nil
}

type stubProjectPermission struct {
	ProjectPermissions neon.ProjectPermissions
	err                error
}

func (s *stubProjectPermission) GrantPermissionToProject(_ string, cfg neon.GrantPermissionToProjectRequest) (neon.ProjectPermission, error) {
	if s.err != nil {
		return neon.ProjectPermission{}, s.err
	}

	resp := neon.ProjectPermission{
		GrantedAt:      time.Now().UTC(),
		GrantedToEmail: cfg.Email,
		ID:             uuid.NewString(),
	}

	s.ProjectPermissions.ProjectPermissions = append(s.ProjectPermissions.ProjectPermissions, resp)
	return resp, nil
}

func (s *stubProjectPermission) RevokePermissionFromProject(_ string, permissionID string) (neon.ProjectPermission, error) {
	if s.err != nil {
		return neon.ProjectPermission{}, s.err
	}

	now := time.Now().UTC()
	return neon.ProjectPermission{
		GrantedAt:      now.Add(-1 * time.Second),
		GrantedToEmail: "foo@bar.baz",
		ID:             permissionID,
		RevokedAt:      &now,
	}, nil
}

func (s *stubProjectPermission) ListProjectPermissions(_ string) (neon.ProjectPermissions, error) {
	if s.err != nil {
		return neon.ProjectPermissions{}, s.err
	}
	return s.ProjectPermissions, nil
}

type mockOpsReader struct {
	rec         map[string][]time.Time
	maxRequests map[string]int
	mu          *sync.Mutex
}

func (m mockOpsReader) GetProjectOperation(_ string, operationID string) (o neon.OperationResponse, err error) {
	o = neon.OperationResponse{Operation: neon.Operation{
		ID:     operationID,
		Status: neon.OperationStatusFinished,
	}}
	m.mu.Lock()
	m.rec[operationID] = append(m.rec[operationID], time.Now())
	if m.maxRequests[operationID] > 0 {
		o.Operation.Status = neon.OperationStatusRunning
		m.maxRequests[operationID]--
	}
	m.mu.Unlock()
	return o, nil
}
