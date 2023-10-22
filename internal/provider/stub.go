package provider

import neon "github.com/kislerdm/neon-sdk-go"

type sdkClientStub struct {
	req interface{}
	err error
}

func (s *sdkClientStub) GetCurrentUserInfo() (neon.CurrentUserInfoResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) ListApiKeys() ([]neon.ApiKeysListResponseItem, error) {
	panic("not implemented")
}

func (s *sdkClientStub) CreateApiKey(cfg neon.ApiKeyCreateRequest) (neon.ApiKeyCreateResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) RevokeApiKey(keyID int64) (neon.ApiKeyRevokeResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) GetProjectOperation(projectID string, operationID string) (neon.OperationResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) ListProjects(cursor *string, limit *int) (neon.ListProjectsRespObj, error) {
	panic("not implemented")
}

func (s *sdkClientStub) CreateProject(cfg neon.ProjectCreateRequest) (neon.CreatedProject, error) {
	s.req = cfg
	return neon.CreatedProject{}, s.err
}

func (s *sdkClientStub) GetProject(projectID string) (neon.ProjectResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) UpdateProject(projectID string, cfg neon.ProjectUpdateRequest) (
	neon.UpdateProjectRespObj, error,
) {
	panic("not implemented")
}

func (s *sdkClientStub) DeleteProject(projectID string) (neon.ProjectResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) ListProjectOperations(projectID string, cursor *string, limit *int) (
	neon.ListOperations, error,
) {
	panic("not implemented")
}

func (s *sdkClientStub) ListProjectBranches(projectID string) (neon.BranchesResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) CreateProjectBranch(projectID string, cfg *neon.BranchCreateRequest) (
	neon.CreatedBranch, error,
) {
	panic("not implemented")
}

func (s *sdkClientStub) GetProjectBranch(projectID string, branchID string) (neon.BranchResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) UpdateProjectBranch(
	projectID string, branchID string, cfg neon.BranchUpdateRequest,
) (neon.BranchOperations, error) {
	panic("not implemented")
}

func (s *sdkClientStub) DeleteProjectBranch(projectID string, branchID string) (neon.BranchOperations, error) {
	panic("not implemented")
}

func (s *sdkClientStub) SetPrimaryProjectBranch(projectID string, branchID string) (neon.BranchOperations, error) {
	panic("not implemented")
}

func (s *sdkClientStub) ListProjectBranchEndpoints(projectID string, branchID string) (neon.EndpointsResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) ListProjectBranchDatabases(projectID string, branchID string) (neon.DatabasesResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) CreateProjectBranchDatabase(
	projectID string, branchID string, cfg neon.DatabaseCreateRequest,
) (neon.DatabaseOperations, error) {
	panic("not implemented")
}

func (s *sdkClientStub) GetProjectBranchDatabase(
	projectID string, branchID string, databaseName string,
) (neon.DatabaseResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) UpdateProjectBranchDatabase(
	projectID string, branchID string, databaseName string, cfg neon.DatabaseUpdateRequest,
) (neon.DatabaseOperations, error) {
	panic("not implemented")
}

func (s *sdkClientStub) DeleteProjectBranchDatabase(
	projectID string, branchID string, databaseName string,
) (neon.DatabaseOperations, error) {
	panic("not implemented")
}

func (s *sdkClientStub) ListProjectBranchRoles(projectID string, branchID string) (neon.RolesResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) CreateProjectBranchRole(
	projectID string, branchID string, cfg neon.RoleCreateRequest,
) (neon.RoleOperations, error) {
	panic("not implemented")
}

func (s *sdkClientStub) GetProjectBranchRole(projectID string, branchID string, roleName string) (
	neon.RoleResponse, error,
) {
	panic("not implemented")
}

func (s *sdkClientStub) DeleteProjectBranchRole(projectID string, branchID string, roleName string) (
	neon.RoleOperations, error,
) {
	panic("not implemented")
}

func (s *sdkClientStub) GetProjectBranchRolePassword(
	projectID string, branchID string, roleName string,
) (neon.RolePasswordResponse, error) {
	if s.err != nil {
		return neon.RolePasswordResponse{}, s.err
	}
	return neon.RolePasswordResponse{}, nil
}

func (s *sdkClientStub) ResetProjectBranchRolePassword(
	projectID string, branchID string, roleName string,
) (neon.RoleOperations, error) {
	panic("not implemented")
}

func (s *sdkClientStub) ListProjectEndpoints(projectID string) (neon.EndpointsResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) CreateProjectEndpoint(projectID string, cfg neon.EndpointCreateRequest) (
	neon.EndpointOperations, error,
) {
	panic("not implemented")
}

func (s *sdkClientStub) GetProjectEndpoint(projectID string, endpointID string) (neon.EndpointResponse, error) {
	panic("not implemented")
}

func (s *sdkClientStub) UpdateProjectEndpoint(
	projectID string, endpointID string, cfg neon.EndpointUpdateRequest,
) (neon.EndpointOperations, error) {
	panic("not implemented")
}

func (s *sdkClientStub) DeleteProjectEndpoint(projectID string, endpointID string) (neon.EndpointOperations, error) {
	panic("not implemented")
}

func (s *sdkClientStub) StartProjectEndpoint(projectID string, endpointID string) (neon.EndpointOperations, error) {
	panic("not implemented")
}

func (s *sdkClientStub) SuspendProjectEndpoint(projectID string, endpointID string) (neon.EndpointOperations, error) {
	panic("not implemented")
}

func (s *sdkClientStub) ListProjectsConsumption(cursor *string, limit *int) (
	neon.ListProjectsConsumptionRespObj, error,
) {
	panic("not implemented")
}
