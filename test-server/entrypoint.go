// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"log"
	"net/http"

	v0Roles "github.com/canonical/identity-platform-api/v0/roles"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/types/known/emptypb"
)

type v0RoleService struct {
	v0Roles.UnimplementedRolesServiceServer
}

func (r *v0RoleService) ListRoles(ctx context.Context, empty *emptypb.Empty) (*v0Roles.ListRolesResp, error) {
	log.Println("ListRoles")
	return &v0Roles.ListRolesResp{}, nil
}

func (r *v0RoleService) GetRole(ctx context.Context, req *v0Roles.GetRoleReq) (*v0Roles.ListRolesResp, error) {
	log.Println("GetRole")
	return &v0Roles.ListRolesResp{}, nil
}

func (r *v0RoleService) CreateRole(ctx context.Context, req *v0Roles.CreateRoleReq) (*v0Roles.ListRolesResp, error) {
	log.Println("CreateRole")
	return &v0Roles.ListRolesResp{}, nil
}

func (r *v0RoleService) UpdateRole(ctx context.Context, req *v0Roles.UpdateRoleReq) (*v0Roles.UpdateRoleResp, error) {
	log.Println("UpdateRole")
	return &v0Roles.UpdateRoleResp{}, nil
}

func (r *v0RoleService) RemoveRole(ctx context.Context, req *v0Roles.RemoveRoleReq) (*v0Roles.RemoveRoleResp, error) {
	log.Println("RemoveRole")
	return &v0Roles.RemoveRoleResp{}, nil
}

func (r *v0RoleService) ListRoleEntitlements(ctx context.Context, req *v0Roles.ListRoleEntitlementsReq) (*v0Roles.ListRoleEntitlementsResp, error) {
	log.Println("ListRoleEntitlements")
	return &v0Roles.ListRoleEntitlementsResp{}, nil
}

func (r *v0RoleService) UpdateRoleEntitlements(ctx context.Context, req *v0Roles.UpdateRoleEntitlementsReq) (*emptypb.Empty, error) {
	log.Println("UpdateRoleEntitlements")
	return &emptypb.Empty{}, nil
}

func (r *v0RoleService) RemoveRoleEntitlement(ctx context.Context, req *v0Roles.RemoveRoleEntitlementReq) (*emptypb.Empty, error) {
	log.Println("RemoveRoleEntitlement")
	return &emptypb.Empty{}, nil
}

func (r *v0RoleService) GetRoleGroups(ctx context.Context, req *v0Roles.GetRoleGroupsReq) (*v0Roles.GetRoleGroupsResp, error) {
	log.Println("GetRoleGroups")
	return &v0Roles.GetRoleGroupsResp{}, nil
}

func HTTPHeaderMatcher(key string) (string, bool) {
	// pass all headers
	return key, true
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(HTTPHeaderMatcher))

	err := v0Roles.RegisterRolesServiceHandlerServer(ctx, mux, new(v0RoleService))
	if err != nil {
		return err
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	const listenAddress = ":8081"

	signal := make(chan error)
	go func() {
		if err := http.ListenAndServe(listenAddress, mux); err != nil {
			signal <- err
			log.Println("ListenAndServe:", err)
		}
	}()

	log.Printf("Test server running on %s\n", listenAddress)

	return <-signal
}

func main() {
	log.Println(run())
}
