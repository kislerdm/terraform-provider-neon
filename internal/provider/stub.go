package provider

import neon "github.com/kislerdm/neon-sdk-go"

type sdkClientStub struct {
	req interface{}
	err error
}

func (s *sdkClientStub) UpdateProject(_ string, _ neon.ProjectUpdateRequest) (neon.UpdateProjectRespObj, error) {
	panic("implement me")
}

func (s *sdkClientStub) GetProject(_ string) (neon.ProjectResponse, error) {
	panic("implement me")
}

func (s *sdkClientStub) ListProjectBranches(_ string) (neon.BranchesResponse, error) {
	panic("implement me")
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

func (s *sdkClientStub) GetProjectBranchRolePassword(_ string, _ string, _ string) (neon.RolePasswordResponse, error) {
	if s.err != nil {
		return neon.RolePasswordResponse{}, s.err
	}
	return neon.RolePasswordResponse{}, nil
}
